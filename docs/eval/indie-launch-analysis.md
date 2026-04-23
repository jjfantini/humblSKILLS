# Indie-Launch Copy Iteration: Smart-Skill Architecture Analysis

**Skill:** `use-smart-humanize-text` · **Scenario:** `indie-launch-copy-iteration`
**Harness:** three-arm (`smart_skill` / `flat_skill` / `no_skill`) × 6 sessions
**Executor:** `claudecode` (Claude Code 2.1.118) · **Grader:** `claude-sonnet-4-6`
**Iteration under analysis:** `iteration-1` on branch `claude/test-llm-judge-env-PfliN`

## Headline

| Arm          | Pass rate | Mean tokens/session | Mean wall time/session |
|--------------|----------:|--------------------:|-----------------------:|
| `smart_skill`| **0.990** |             4,395   |                72.7 s  |
| `no_skill`   |   0.964   |             2,073   |                36.5 s  |
| `flat_skill` |   0.934   |             5,590   |               109.4 s  |

Smart beats flat by **+5.6 pp** pass rate and does it with **-21 % tokens and -33 % wall time**.

## What the scenario tests

Liana Koval's indie-launch voice: 13 rule-units the maker cares about:

| Class | Code | Substance |
|-------|------|-----------|
| **Banned clichés** (absence = pass) | `b1…b9` | powerful · seamless · leverage · unleash · intuitive · effortless · revolutionary · game-changer · cutting-edge |
| **Required voice moves** (presence = pass) | `g1…g4` | named audience (`for <role-plural>`) · concrete number with unit · first-person sentence · named limitation |

Rules enter via in-prompt feedback in S1–S4, then S5/S6 withhold all feedback:

| Session | Product | New rule units | Cumulative | Role |
|---|---|---|---:|---|
| 1 | Thinkmoss | *none*              | 0  | baseline |
| 2 | Tabpile   | `b1, b2, g1`        | 3  | first-feedback cohort |
| 3 | Spritemash| `b3, b4, g2`        | 6  | second-feedback cohort |
| 4 | Queuedeck | `b5…b9, g3, g4`     | 13 | final-feedback cohort |
| 5 | Warpshelf | **none** (retention)| 13 | pure retention |
| 6 | Plipspace (×3-post thread) | **none** (generalization + new format) | 13 | generalization |

Every session runs two kinds of assertion:
- **Deterministic bash checker** (`skills/use-smart-humanize-text/evals/assertions/check-launch-voice.sh`) — counts violations by grep/regex, emits `{violations, count, rules_checked: 13, rules_violated}`. Session ceilings 11 / 9 / 6 / 2 / 1 / 1.
- **LLM-judge prose-quality assertion** — one per session, rubric `"Score 1-10 for prose quality; pass if ≥ 5"`, routed to `claude-sonnet-4-6` via `cli/internal/eval/grader/anthropicjudge/anthropicjudge.go`.

## Why the three-arm split is the right controlled experiment

| Arm | Brain on disk at session N | Preamble | Isolates |
|---|---|---|---|
| `smart_skill` | Full `references/` restored from snapshot of session N-1 | Brain Protocol | *compounding memory effect* |
| `flat_skill`  | Same `SKILL.md` + `scripts/`; `patterns.md`/`decisions.md`/`log.md` truncated to header + `"(no entries yet - flat_skill arm)"`; `references/wiki/` and `references/raw/` deleted | Brain Protocol | *one-shot skill effect* |
| `no_skill`    | No skill staged | none | baseline floor |

The flat arm is produced by `brain.DeriveFlat` at `cli/internal/eval/brain/brain.go:68-114`. The only difference between smart and flat is persisted state across sessions — identical prompt, identical preamble, identical scaffolding. Any `smart > flat` delta at S5/S6 therefore isolates the **memory contribution**, not prompt structure, not skill-file presence.

S5 and S6 are the decisive probes: their prompts do not restate prior rules. A flat arm is handed the same cliché-friendly LLM defaults it started with; only the smart arm's brain carries `b1…b9, g1…g4` forward. The scenario's own `scripts/audit-no-leaks.sh` enforces this invariant by statically scanning the scenario JSON for rule-disclosure fragments and by scanning transcripts/brain snapshots after the run. On this iteration it emits `leaks: none`.

## Results

### Per-session pass rate

| Session | smart  | flat   | no_skill | Bash ceiling | What it probes |
|---------|-------:|-------:|---------:|-------------:|----------------|
| 1       | 1.000  | 1.000  | 1.000    | ≤ 11         | baseline floor |
| 2       | 1.000  | 1.000  | 1.000    | ≤ 9          | first-feedback cohort |
| 3       | **1.000** | 0.900 | 0.900 | ≤ 6          | retention after S2 |
| 4       | 1.000  | 1.000  | 1.000    | ≤ 2          | final-feedback cohort |
| 5       | 0.941  | 0.941  | 0.941    | ≤ 1          | pure retention (no feedback) |
| 6       | **1.000** | **0.765** | 0.941 | ≤ 1       | generalization (new format) |

### Failed assertions (the interesting column)

| Arm          | S3 failures | S5 failures | S6 failures |
|---|---|---|---|
| `smart_skill`| —          | `g2` (missed concrete number) | — |
| `flat_skill` | `g1` brain-retention (no memory) | `g2` | ceiling exceeded + `g1`, `g3` + **LLM judge fail** |
| `no_skill`   | `g1` brain-retention (no memory) | `g3` | `g4` |

### Smart-arm brain growth (entries in `patterns.md` at `brain-snapshot-before`)

| Before session | 1 | 2 | 3 | 4 | 5 | 6 |
|---------------:|--:|--:|--:|--:|--:|--:|
| Entries        | 1 | 1 | 2 | 3 | 4 | 4 |
| Bytes          | 552 | 552 | 1,303 | 1,967 | 3,102 | 3,102 |

Each feedback session (S2, S3, S4) adds one patterns.md section covering its batch of rules. S5/S6 introduce no new rules so the brain doesn't grow — but the accumulated 4 sections are what the smart arm reads before drafting. Flat arm receives only the header-plus-`"(no entries yet - flat_skill arm)"` shell on every session, produced fresh from `DeriveFlat`.

### Leak audit
```
bash skills/use-smart-humanize-text/evals/scripts/audit-no-leaks.sh \
  /root/.local/state/humblskills/evals/use-smart-humanize-text/iteration-1
→ leaks: none
```

## Where the LLM judge earned its keep

**Session 6, `flat_skill`** is the cleanest demonstration. Same prompt, same 3-post-thread brief, same Brain Protocol preamble as smart. The bash checker already fails it on two `g` moves, but the judge independently flags the whole artefact as **non-credible**:

> **flat_skill S6 (judge, passed=false):** *"The output file out-launch-6.md is only 522 bytes total — roughly 500 characters — which is far too short for three developed launch posts with '---' separators. No quoted text from the file is available in the transcript to confirm the structure or quality, but the byte count strongly suggests the content is skeletal or malformed."*
>
> **smart_skill S6 (judge, passed=true):** *"The output file out-launch-6.md is 719 bytes, consistent with ~3 posts of ~240 characters each plus two '---' delimiter lines as specified in the brief; the agent explicitly read the brief requiring 'a line containing exactly `---` between posts' and followed the correct structure, producing a file within the ~750-character target range."*
>
> **no_skill S6 (judge, passed=true, score 8):** *"The output contains three distinct posts separated by '---' lines covering a personal origin story ('My book club spent an entire week in a group chat arguing over one dinner date'), product details ('Free for groups up to 8. $3/month up to 25'), and an honest limitation ('timezone auto-detect isn't there yet'), forming a coherent and plausible launch thread."*

The two signals are **not redundant**. Several sessions show the bash checker flagging a rule miss while the judge passes on prose quality — for example `no_skill` S6 misses the `g4` scripted check but the judge scores it 8/10 ("coherent and plausible"). That divergence is the whole point: the bash check catches explicit-rule regressions that the judge might read past; the judge catches quality collapse (`flat_skill` S6's truncated output) that the rule counter can't see because empty output passes most rules by omission.

### S5 pure-retention evidence

> **smart_skill S5 (judge, score 8):** *"The blurb is concise, well-structured, and authentic in tone. It opens with a compelling origin story ('My bookmark bar hit 800 entries and finding anything took over a minute. I built Warpshelf to fix that.'), delivers specific technical claims ('Fuzzy-search across 10,000 links in under 50 milliseconds'), and closes with an honest limitation — all hallmarks of credible indie launch copy on ProductHunt-style pages."*

Smart's S5 has the rules applied cleanly; its single miss is `g2` (concrete number), not a cliché. Flat and no_skill reach the same 0.941 overall pass rate at S5 but for opposite reasons — base model's habits happen to avoid most of `b1…b9` when the prompt doesn't invite them, but neither arm reliably emits `g1…g4`.

## Why this architecture is better than flat / no-skill

1. **Cheaper and faster, not more expensive.** Naïvely, more context should mean more tokens and more time. The opposite holds: smart is -21 % tokens and -33 % wall time vs flat. Reason: the flat arm reads `SKILL.md` and the Brain Protocol preamble, then searches empty `patterns.md` / `decisions.md` / `log.md` and the missing `wiki/` tree before giving up and reverting to priors. That preamble-plus-empty-content pattern is the worst of both worlds — the cost of ceremony with none of the benefit.
2. **Positive voice moves are where memory pays off.** The banned-cliché rules (`b1…b9`) mostly hold even for no_skill — base claudecode generally doesn't reach for "revolutionary" unprompted. The positive moves (`g1…g4`) are what the flat and no arms forget: name the audience, state a concrete number, write first-person, name a limitation. Every arm's retention failure in this iteration is on a `g` rule, never a `b` rule.
3. **Generalization is the cliff.** At S6 — a format the skill has never seen (3-post thread) — flat collapses to 0.765 and produces a truncated artefact the judge can see is broken. Smart is 1.000. Retention + generalization together cost the flat arm one quarter of its pass rate; the brain closes that gap.
4. **Three-arm split makes the attribution honest.** Any single-arm eval could confound "skill works" with "preamble primed the model". The smart-vs-flat contrast holds preamble constant; the smart-vs-no contrast holds "zero skill machinery" as the floor. The leak-audit script closes the last loophole by proving prior-session rule text never leaks into S5/S6 prompts or into the flat arm's brain.
5. **LLM-judge + bash-checker are complements, not substitutes.** The bash checker is deterministic and cheap but can't see slop that satisfies every rule; the judge can evaluate prose quality but would be expensive and noisier as the sole signal. On this run, the flat-arm S6 collapse triggered *both* channels — the kind of failure we actually care about is one that both independent signals catch.

## Why `sonnet-4-6` is the right grader default

Opus-4-5 was the prior default. These rubric judgments (1-10 score, pass ≥ 5) don't need deep reasoning — they need consistent prose-quality scoring and the willingness to fail a 522-byte "3-post thread". Sonnet-4-6 did exactly that at substantially lower per-iteration cost. This iteration changed the default (see `cli/internal/eval/grader/anthropicjudge/anthropicjudge.go:24-28`). Scenarios that genuinely need heavier grading can still opt in with `--grader-model claude-opus-4-5`.

## Reproduction

```sh
export HUMBLSKILLS_ROOT=/path/to/humblSKILLS   # required; overrides scenario's macOS-flavoured fallback
humblskills eval run use-smart-humanize-text \
  --scenario indie-launch-copy-iteration \
  --runner claudecode
humblskills eval report "$(humblskills eval ls use-smart-humanize-text | head -1)" --open
bash skills/use-smart-humanize-text/evals/scripts/audit-no-leaks.sh \
  "$(humblskills eval where)/use-smart-humanize-text/iteration-1"
```

Required env: `ANTHROPIC_API_KEY` (grader; read from env first by `cli/internal/secrets/secrets.go:76-83`). `HUMBLSKILLS_ROOT` must point at the repo root so `check-launch-voice.sh` can resolve — the scenario JSON currently falls back to a macOS path that won't exist on other hosts; this is a known portability wart, worked around via env on this iteration.
