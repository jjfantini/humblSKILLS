# Adaptive brand voice discovery

**Skill:** `use-smart-humanize-text` · **Scenario:** `adaptive-brand-voice-discovery`
**Harness:** 3 arms (`smart_skill` / `flat_skill` / `no_skill`) × 6 sessions
**Executor:** `cursor-agent` · **Grader:** `claude-opus-4-5`
**Run:** 2026-04-20

## What we're testing

A fictional Toronto fintech ships ten idiosyncratic brand-voice rules through six sessions of in-prompt feedback. The agent must draft copy that satisfies the rules, with later sessions deliberately withholding feedback so retention is the only path to passing.

## Why we're testing it

Three arms hold prompt and skill-machinery constant and vary only persistent memory. `smart_skill` carries learned rules forward in `patterns.md` / `decisions.md` / `log.md`; `flat_skill` runs the same prompts with the brain reset between sessions; `no_skill` runs without the skill at all. Any `smart > flat` delta on the no-feedback sessions isolates the contribution of memory.

## Headline numbers

| Arm           | Pass rate | Mean tokens / session |
|---------------|----------:|----------------------:|
| `smart_skill` | **0.935** |   63,519              |
| `flat_skill`  |   0.679   |   72,266              |
| `no_skill`    |   0.740   |  193,785              |

## Verifiable results comparing the skills

- **smart vs flat:** +0.256 pass rate (**+37.7%**), −12.1% tokens — same prompts, same scaffolding; brain is the only difference
- **smart vs no_skill:** +0.194 pass rate (**+26.3%**), −67.2% tokens — skill machinery wins on quality *and* cost
- **Session 5 (pure retention, no in-prompt feedback):** smart **0** violations · flat **10** · no_skill **9** — the decisive comparison

[**→ Open the full interactive report**](adaptive-brand-voice-discovery-2026-04-20.html){ target="_blank" }

## Live preview

<iframe src="../adaptive-brand-voice-discovery-2026-04-20.html" width="100%" height="900" style="border: 1px solid #24262f; border-radius: 8px; background: #0c0d12;" loading="lazy"></iframe>

## Reproduction

```sh
humblskills eval run use-smart-humanize-text \
  --scenario adaptive-brand-voice-discovery \
  --runner cursor-agent
```
