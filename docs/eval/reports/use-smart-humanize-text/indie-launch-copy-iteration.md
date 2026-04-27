# Indie-launch copy iteration

**Skill:** `use-smart-humanize-text` · **Scenario:** `indie-launch-copy-iteration`
**Harness:** 2 arms (`smart_skill` / `no_skill`) × 6 sessions
**Executor:** `cursor-agent` · **Grader:** `claude-sonnet-4-6`
**Run:** 2026-04-27

## What we're testing

Indie-maker Liana Koval ships 13 voice rules — 9 banned clichés (`powerful`, `seamless`, `leverage`, …) and 4 required moves (named audience, concrete number, first-person, named limitation) — across six sessions of in-prompt feedback. Sessions 5 and 6 withhold all feedback so retention and generalization are the only path to passing.

## Why we're testing it

This is the floor-vs-skill comparison. `smart_skill` carries learned rules forward in `patterns.md` / `decisions.md` / `log.md`; `no_skill` runs the same prompts without any skill machinery. Any `smart > no_skill` delta on the no-feedback sessions (S5, S6) is the value the skill adds over the bare model.

## Headline numbers

| Arm           | Pass rate | Mean tokens / session | Mean wall time / session |
|---------------|----------:|----------------------:|-------------------------:|
| `smart_skill` | **0.882** |  73,523               |  55.6 s                  |
| `no_skill`    |   0.834   |  38,498               |  52.4 s                  |

## Verifiable results comparing the skills

- **smart vs no_skill:** +0.048 pass rate (**+5.7%**), +91.0% tokens, +6.1% wall time — the brain costs ~2× tokens for a meaningful pass-rate lift, concentrated entirely on the no-feedback tail.
- **Sessions 1–4 (feedback in-prompt):** smart and no_skill tie at 0.750, 0.857, 0.900, 0.900 — when the rules are restated each turn, both arms pass at the same rate.
- **Session 5 (pure retention, no in-prompt feedback):** smart **0.941** vs no_skill **0.765** — **+23% relative**. The brain remembers; the bare model doesn't.
- **Session 6 (generalization to a new 3-post thread format):** smart **0.944** vs no_skill **0.833** — **+13% relative**. Format the skill never saw, rules never restated; the brain still holds form.

[**→ Open the full interactive report**](indie-launch-copy-iteration-2026-04-27.html){ target="_blank" }

## Live preview

<iframe src="../indie-launch-copy-iteration-2026-04-27.html" width="100%" height="900" style="border: 1px solid #24262f; border-radius: 8px; background: #0c0d12;" loading="lazy"></iframe>

## Reproduction

```sh
humblskills eval run use-smart-humanize-text \
  --scenario indie-launch-copy-iteration \
  --config smart_skill,no_skill \
  --runs 1 \
  --runner cursor-agent
```
