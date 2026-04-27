# Indie-launch copy iteration

**Skill:** `use-smart-humanize-text` · **Scenario:** `indie-launch-copy-iteration`
**Harness:** 4 arms (`smart_skill` / `flat_skill_wiki` / `flat_skill` / `no_skill`) × 3 runs × 6 sessions = 72 sessions
**Executor:** `cursor-agent` · **Grader:** `claude-opus-4-5`
**Run:** 2026-04-27

## What we're testing

Indie-maker Liana Koval ships 13 voice rules — 9 banned clichés (`powerful`, `seamless`, `leverage`, …) and 4 required moves (named audience, concrete number, first-person, named limitation) — across six sessions of in-prompt feedback. Sessions 5 and 6 deliberately withhold all feedback so retention (S5) and generalization to a new 3-post-thread format (S6) are the only path to passing.

## Why we're testing it

The 4-arm ablation isolates two effects that conflate in simpler harnesses. `flat_skill_wiki` keeps `SKILL.md` + the static `wiki/` knowledge but resets the brain (`patterns.md` / `decisions.md` / `log.md`) every session — so `smart` vs `flat_skill_wiki` measures **persistent memory alone**. `flat_skill` strips the wiki too, so `flat_skill_wiki` vs `flat_skill` measures whether static knowledge helps when memory is absent. `no_skill` sets the floor. The decisive sessions are S5 and S6: any arm above the others there is winning on retention, not on rules-in-prompt.

## Headline numbers (mean over 3 runs × 6 sessions = 18 sessions per arm)

| Arm               | Pass rate | Mean tokens / session | Mean wall time / session |
|-------------------|----------:|----------------------:|-------------------------:|
| `smart_skill`     | **0.873** |  53,435               |  57.2 s                  |
| `no_skill`        |   0.851   |  55,387               |  42.6 s                  |
| `flat_skill_wiki` |   0.842   |  64,923               |  51.0 s                  |
| `flat_skill`      |   0.832   |  75,296               |  58.1 s                  |

## Verifiable results comparing the skills

The aggregate table buries the real signal because S1–S4 carry rules in-prompt and tie across arms. The story is the no-feedback tail (S5, S6 — mean over 3 runs):

| Session | smart | flat_wiki | flat   | no_skill | What it probes                            |
|---------|------:|----------:|-------:|---------:|-------------------------------------------|
| S5      | **0.922** | 0.843     | 0.902  | 0.843    | pure retention — no in-prompt feedback    |
| S6      | **0.944** | 0.833     | 0.852  | 0.889    | generalization to a new 3-post format     |

- **smart vs flat_skill_wiki (the brain-only delta):** +9.4% on S5, **+13.3% on S6**. Same `SKILL.md`, same `wiki/`; the only difference is whether `patterns.md` / `decisions.md` / `log.md` persisted. This is the cleanest measurement of memory value.
- **smart vs flat_skill:** +5% pass rate, **−29% tokens, −1.4% wall time**. Memory is *cheaper* than re-reading `SKILL.md` against an empty brain.
- **smart vs no_skill:** +2.6% pass rate aggregate; **+9.4% on S5, +6.2% on S6**. The full-skill apparatus pays off precisely on the sessions that test what was learned.
- **Smart is the only arm above 0.9 on both S5 and S6.** Every other arm fails at least one of the no-feedback probes.

[**→ Open the full interactive report**](indie-launch-copy-iteration-2026-04-27.html){ target="_blank" }

## Live preview

<iframe src="../indie-launch-copy-iteration-2026-04-27.html" width="100%" height="900" style="border: 1px solid #24262f; border-radius: 8px; background: #0c0d12;" loading="lazy"></iframe>

## Reproduction

```sh
humblskills eval run use-smart-humanize-text \
  --scenario indie-launch-copy-iteration \
  --runner cursor-agent
```
