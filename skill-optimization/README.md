# skill-optimization — SkillOpt integration

Optimize humblSKILLS skill documents with **[microsoft/SkillOpt](https://github.com/microsoft/SkillOpt)**
("textual gradient descent"): a skill is treated as trainable state and improved through a
deep-learning-style loop — rollout → reflect → optimize → **validation gate** → `best_skill.md` —
without touching model weights.

> This is a **self-contained Python subproject** inside the (Go-based) humblSKILLS repo. It lives
> entirely under `skill-optimization/` and does not touch the CLI, registry, or release tooling.
> Run every command below **from this directory** (`cd skill-optimization`).

## ⚠️ What SkillOpt can and cannot optimize

SkillOpt improves a skill against a **programmatic reward over a dataset with known-good outcomes**
(QA vs. gold answers, code scored by tests, extraction vs. labels). It computes the "gradient" by
*measuring* whether the skill made outcomes better. It is **not** a generic "make my skills better"
button: a humblSKILL whose success is subjective (e.g. `use-smart-commit`, `use-smart-humanize-text`)
has no automatic scorer, so SkillOpt cannot optimize it until you define a task + reward (e.g. an
LLM-judge rubric or a checkable sub-goal). The humblSKILLS under `../skills/` are candidate targets
**only once each has a measurable reward**.

## What's here

| Path | What it is |
|------|------------|
| `external/SkillOpt/` | SkillOpt, vendored as a pinned git submodule (unmodified, editable-installed). |
| `skillopt_lab/envs/exactmatch/` | A complete minimal custom env — the **template to copy** for a real skill, and the smoke test. |
| `skillopt_lab/train.py` | Launcher that registers custom envs into SkillOpt's registry without editing upstream. |
| `skillopt_lab/offline_demo.py` | A **zero-network** end-to-end run of the real trainer (stubbed LLM). |
| `skillopt_lab/tests/` | Offline wiring tests (pytest). |
| `skillopt_lab/configs/`, `skillopt_lab/data/` | Training config + toy dataset. |
| `docs/skillopt-integration.md` | Deep reference: the env / backend / reward contracts. |

## Setup (requires `uv` + Python 3.12)

```bash
cd skill-optimization
git submodule update --init --recursive        # fetch external/SkillOpt
uv venv --python 3.12 && source .venv/bin/activate
uv pip install -e external/SkillOpt            # editable install (required — see docs)
uv pip install pytest
```

## Verify it works (no API keys, no tokens)

```bash
uv run pytest -q                               # 7 offline wiring tests
uv run python -m skillopt_lab.offline_demo     # full loop with a stubbed LLM
```

The offline demo runs the **real** `ReflACTTrainer`; the starting skill omits a formatting rule, so
answers fail exact-match (accuracy 0). The optimizer proposes the rule, the **validation gate
accepts it**, and held-out accuracy goes **0 → 1.0**. That acceptance is **rigged by construction**
(the oracle answerer and the matching edit are both written by the demo) — it proves the pipeline is
**wired correctly**, NOT that SkillOpt improves a real skill.

## Run a real optimization (spends tokens)

`claude_chat` (default) drives the **local `claude` CLI** as a subprocess — run it **outside** a
Claude Code session:

```bash
uv run python -m skillopt_lab.train --config skillopt_lab/configs/exactmatch.yaml
```

Or use OpenAI/Azure with `openai_chat` (no nested CLI):

```bash
export AZURE_OPENAI_AUTH_MODE=openai_compatible
export AZURE_OPENAI_ENDPOINT=https://api.openai.com/v1
export AZURE_OPENAI_API_KEY=sk-...
uv run python -m skillopt_lab.train --config skillopt_lab/configs/exactmatch.yaml \
  --cfg-options model.optimizer_backend=openai_chat model.target_backend=openai_chat \
                model.optimizer=gpt-4o model.target=gpt-4o
```

To optimize one of your real humblSKILLS, copy `skillopt_lab/envs/exactmatch` and supply a dataset +
a programmatic reward — see `skillopt_lab/README.md` and `docs/skillopt-integration.md`.

> **Status:** built and verified OFFLINE (7/7 pytest, ruff clean, gated loop emits best_skill.md). No
> live optimization run has been executed yet.
