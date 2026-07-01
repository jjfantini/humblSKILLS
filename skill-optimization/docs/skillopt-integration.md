# SkillOpt integration — contract reference

How SkillOpt (internal name **ReflACT**) is wired into pindex, and the exact contracts a custom
environment must satisfy. Everything here was verified against the pinned submodule source
(`external/SkillOpt`, commit `3f194d5`).

## 1. The optimization loop

SkillOpt optimizes one Markdown **skill document** with a deep-learning-style loop. No model weights
change; the only trainable state is the skill text. Per training step (`skillopt/engine/trainer.py`):

```
① ROLLOUT    target model runs the batch under the current skill  → results (hard/soft)
② REFLECT    optimizer model reads failure/success trajectories   → candidate edits ("patches")
③ AGGREGATE  hierarchically merge patches (failure-first)
④ SELECT     optimizer ranks/clips edits to the edit budget L      ("learning rate")
⑤ UPDATE     apply edits as literal string ops to the skill        → candidate skill
⑥ EVALUATE   roll out candidate on the val split; GATE accepts iff strictly better
```

Plus epoch-level **slow update** (longitudinal momentum) and **meta-skill** (cross-epoch memory),
both disabled in the demo config. Final artifact: `best_skill.md`.

Key framings (not metaphors that map to weights):
- `optimizer.learning_rate` is an **integer max-edits-per-step budget**, not a float LR.
- There is **no `evaluate()` method** on the env — scoring lives inside your rollout code.
- The **validation gate is mandatory**: `evaluation.use_gate: false` raises in `flatten_config`.

## 2. The env contract (`skillopt/envs/base.py`)

A custom env subclasses `EnvAdapter` and must implement **five abstract methods**:

```python
def build_train_env(self, batch_size: int, seed: int, **kwargs)            # -> "env manager"
def build_eval_env(self, env_num: int, split: str, seed: int, **kwargs)    # -> "env manager"
def rollout(self, env_manager, skill_content: str, out_dir: str, **kwargs) -> list[dict]
def reflect(self, results: list[dict], skill_content: str, out_dir: str, **kwargs) -> list[dict | None]
def get_task_types(self) -> list[str]
```

The official template, `searchqa`, and `docvqa` all also override `build_env_from_batch` to
`return list(batch.payload or [])`, making the "env manager" a plain `list[dict]` of items. We follow
that (see `skillopt_lab/envs/exactmatch/adapter.py`). Optional hooks you may override: `setup(cfg)`,
`get_dataloader()`, `requires_ray()`, `get_error_minibatch_prompt()`, `get_success_minibatch_prompt()`.

### rollout result contract
Each result dict needs **`id` (str), `hard` (0|1), `soft` (float ∈ [0,1])**. `hard` is the gate /
accuracy metric; `soft` is a graded signal. Extras (`task_description`, `task_type`, `fail_reason`,
`n_turns`) are preserved for the analyst.

### ⚠️ The non-obvious requirement: write `conversation.json`
`reflect → run_minibatch_reflect → fmt_minibatch_trajectories` reads
`<prediction_dir>/<id>/conversation.json` for **every** result and **silently skips** any item
without one. No file ⇒ no trajectory ⇒ no patch ⇒ **the skill never changes** (the loop still "runs"
and looks fine). So `rollout` **must** write that file per item. We include a trailing
`{"role":"system","content": "[EVALUATION RESULT] ..."}` message so the analyst can see *why* the
answer scored as it did. (Mirrors `skillopt/envs/searchqa/rollout.py`.)

### reflect
Delegate to `skillopt.gradient.reflect.run_minibatch_reflect(...)`. Passing `error_system=None` /
`success_system=None` makes it load SkillOpt's **generic** analyst prompts
(`skillopt/prompts/analyst_error.md` / `analyst_success.md`) via `load_prompt`, which falls back to
generic even when the env name can't be resolved — so an env living **outside** `skillopt.envs.*`
(like ours) works without shipping custom prompts.

## 3. Data format (`skillopt/datasets/base.py`)

`SplitDataLoader` gives you both split modes for free; you typically only implement
`load_split_items` (or `load_raw_items` for ratio mode):

- `split_mode: ratio` → reads `data_path` (one `.json`/`.jsonl`), shuffles deterministically by
  `split_seed`, and materializes `train/` `val/` `test/` dirs of `items.json` under
  `<out_root>/_generated_splits/`.
- `split_mode: split_dir` → reads an existing `train/`/`val/`/`test/` tree.

Each item is a `dict`; the **only hard requirement is `id` (str)**. Everything else is whatever your
rollout/evaluator reads.

## 4. Config (`skillopt/config.py`)

Structured YAML with `model` / `train` / `gradient` / `optimizer` / `evaluation` / `env` sections,
flattened to a flat dict for the trainer. Inheritance via **`_base_` (a string path, resolved
RELATIVE TO THE CONFIG FILE'S DIRECTORY)** — ours points at the submodule's base so there's a single
source of truth. Env paths like `data_path` / `skill_init` are resolved relative to the **CWD** (run
from the repo root). `evaluation.use_gate: false` is rejected.

## 5. Registration (`scripts/train.py`)

There is **no decorator registry**. The live registry is the module-level dict
`scripts.train._ENV_REGISTRY`, populated *additively* by `_register_builtins()` (lazy
`try/except ImportError` imports). `get_adapter(cfg)` looks up `cfg["env"]` there.

We register **without editing upstream** (`skillopt_lab/train.py`): wrap `_register_builtins` so it
also injects our envs, then delegate to `scripts.train.main()`. This keeps the submodule pristine.

## 6. Backends (`skillopt/model/`)

Optimizer/target are configured independently (`model.optimizer_backend` / `model.target_backend`).
`skillopt.model.chat_target` / `chat_optimizer` dispatch on the configured backend:

| backend | how it calls the model | credentials |
|---|---|---|
| `openai_chat` | Azure-OpenAI client; with `azure_openai_auth_mode=openai_compatible` it talks to plain OpenAI | `AZURE_OPENAI_*` (or OpenAI key reused) |
| `claude_chat` *(default)* | **subprocess to the local `claude` CLI** (`claude -p --output-format json ...`) — not the Anthropic SDK | the `claude` CLI's own auth |
| `claude_code_exec` | agentic `claude` CLI exec (target only) | `claude` CLI |
| `qwen_chat` / `minimax_chat` | OpenAI-compatible HTTP | per-backend base_url + key |

**Default Claude model:** `claude-sonnet-4-6`. Because `claude_chat` shells out to `claude`, run it
**outside** a Claude Code session — nested `claude -p` is unreliable (it inherits the parent
session's env/MCP and is blocked by the in-session safety classifier).

## 7. Edit / patch schema (`skillopt/optimizer/skill.py`)

The analyst returns `{"patch": {"edits": [ ... ]}, "source_type": "failure"|"success"}`. Each edit is
`{"op": ..., "content": ..., "target": ...}`. Valid ops:

| op | effect |
|---|---|
| `append` | add `content` at the end of the skill (no target) |
| `insert_after` | insert `content` after the line containing `target` (falls back to append) |
| `replace` | replace first occurrence of `target` with `content` |
| `delete` | remove first occurrence of `target` |

Unknown ops are skipped. (Our offline demo's stub initially used `op:"add"` → skipped → no-op →
every candidate rejected; switching to `op:"append"` made the gate accept.)

## 8. Why this integration shape

- **Submodule + editable install.** `configs/` is not part of the installed package, and the `.md`
  prompt files aren't declared as package data — a non-editable `pip install git+…` would break the
  reflect stage at runtime. An editable install from the vendored submodule keeps every file on disk
  at its expected path. The submodule is pinned and stays unmodified, so upstream updates are a
  `git submodule update`.
- **Custom env in our own package + launcher injection.** Keeps our code in our repo and the
  framework pristine; the registry is a plain dict, so injection needs no upstream patch.

## 9. Gotchas checklist

- [ ] `rollout` writes `predictions/<id>/conversation.json` for every item (else: silent no-op loop).
- [ ] Result dicts carry `id` / `hard` / `soft`.
- [ ] `_base_` path is relative to the config file; data/skill paths are relative to the CWD.
- [ ] Don't set `evaluation.use_gate: false` (it raises).
- [ ] Edit ops are `append` / `insert_after` / `replace` / `delete` only.
- [ ] Run `claude_chat` outside a Claude Code session.
- [ ] Reinstall editable + `git submodule update --init` after cloning/merging (venv is gitignored).
