# skillopt_lab — optimizing your own skills

This workspace holds pindex's custom SkillOpt environments. `exactmatch` is a complete, minimal
example; copy it to optimize a real skill.

## Is your skill optimizable by SkillOpt?

SkillOpt computes a "gradient" by **measuring** whether a skill edit improved outcomes on held-out
data. That only works if you can answer **two** questions programmatically:

1. **What is the task?** A dataset of inputs (`train` / `val` / `test`) the skill is applied to.
2. **What is the reward?** A function `score(model_output, gold) → number` where higher = better,
   computable **without a human in the loop**.

| Good fit (has a programmatic reward) | Poor fit (subjective; no auto-scorer) |
|---|---|
| QA with gold answers (exact match / F1) | "write a better commit message" |
| Code generation scored by tests passing | "be more helpful in PR reviews" |
| Spreadsheet/data ops verified by execution | "summarize this nicely" |
| Classification / extraction vs labels | open-ended writing quality |

If your skill's success is subjective, you must first invent a proxy reward (e.g. an LLM-judge
rubric, or a checkable sub-goal) — otherwise there is no gradient to descend.

## Add a new environment

```bash
cp -r skillopt_lab/envs/exactmatch skillopt_lab/envs/myskill
```

Then, in `skillopt_lab/envs/myskill/`:

1. **`evaluator.py`** — replace `evaluate()` with **your reward**. Return `em` (→ `hard`, the 0/1
   gate metric) and `f1` (→ `soft`, a graded signal in `[0,1]`).
2. **`loader.py`** — adjust `_normalize_item()` to your raw record shape. Only `id` is required.
3. **`rollout.py`** — change `_build_system`/`_build_user` and how the answer is produced.
   **Must** write `<out_dir>/predictions/<id>/conversation.json` per item (the reflect stage reads
   it; no file ⇒ no trajectory ⇒ no skill edit).
4. **`adapter.py`** — rename the class; the 5 abstract methods and the `reflect()` delegation stay
   the same. Update `get_task_types()`.
5. **`skills/initial.md`** — your seed skill (can be near-empty; the optimizer fills it in).
6. **Register it** in [`../train.py`](train.py): add `registry["myskill"] = MyskillAdapter` in
   `_custom_envs()`.
7. **Config**: copy `configs/exactmatch.yaml` → `configs/myskill.yaml`, set `env.name: myskill`,
   `env.data_path`, and `env.skill_init`.
8. **Data**: drop a `data/myskill/raw.jsonl` (one JSON object per line; `split_mode: ratio` will
   auto-split it into train/val/test).

Run it:

```bash
uv run python -m skillopt_lab.train --config skillopt_lab/configs/myskill.yaml
```

## The contract in one screen

- **`rollout(env_manager, skill_content, out_dir) -> list[dict]`** — `env_manager` is a `list[dict]`
  of items. For each: build a prompt with `skill_content` in the system message, call the target
  model, score it, write `predictions/<id>/conversation.json`, and return a result dict with at
  least `id` (str), `hard` (0/1), `soft` (float). Extras (`task_description`, `task_type`,
  `fail_reason`) help the analyst.
- **`reflect(results, skill_content, out_dir) -> list[dict|None]`** — delegate to
  `skillopt.gradient.reflect.run_minibatch_reflect` (already wired in `adapter.py`). Passing
  `error_system=None`/`success_system=None` makes it use SkillOpt's generic analyst prompts.
- **The gate is mandatory** (`evaluation.use_gate: true`); a candidate is kept only if it strictly
  beats the current skill on the `val` split.

Full details and the reasoning behind each choice: [`../docs/skillopt-integration.md`](../docs/skillopt-integration.md).
