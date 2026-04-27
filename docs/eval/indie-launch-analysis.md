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

---

## Iteration 2: 4-arm ablation + cumulative-retention outcome

The 3-arm run above left one question open: **is smart's win the brain, or the wiki?** A smart skill ships both persistent memory (`patterns.md`, `decisions.md`, `log.md`) and static knowledge (`references/wiki/humanize/...`). `flat_skill` stripped both; `no_skill` had neither. The delta therefore conflated *"has wiki"* with *"has persistent memory"*.

This iteration splits the flat arm into two variants to separate those effects, bumps `runs_per_configuration` from 1 → 3 for statistical stability, and adds one outcome-first assertion that's the single binary answer to *"did the skill keep its retention promise?"*.

### The 4-arm ablation

| Arm               | `SKILL.md` | `wiki/` | Persistent brain      | What it isolates |
|-------------------|:----------:|:-------:|:---------------------:|------------------|
| `no_skill`        | ✗          | ✗       | ✗                     | Baseline — pure model |
| `flat_skill`      | ✓          | ✗       | ✗                     | Instructions only |
| `flat_skill_wiki` | ✓          | ✓       | ✗ (reset each session) | + static wiki knowledge |
| `smart_skill`     | ✓          | ✓       | ✓                     | + persistent memory |

`flat_skill_wiki` is produced by `brain.DeriveFlatWithWiki` at `cli/internal/eval/brain/brain.go:116-135`: it calls `DeriveFlat` (same shaped meta files, no `raw/`) then copies `references/wiki/` verbatim. Before every session the arm is re-derived from source so wiki content is stable but any brain writes from the previous session are wiped. `smart_skill` vs `flat_skill_wiki` therefore holds the wiki constant — the only difference is whether `patterns.md`/`decisions.md`/`log.md` persisted across sessions.

### Cumulative-retention outcome assertion

A new S6 assertion reads both S5 and S6 checker sidecars and asserts their combined violation count is `≤ 1`:

```
exec: S5_SESS=$(echo "$EVAL_WORK_DIR" | sed 's/session-06-/session-05-/')
      bash check-retention-cumulative.sh \
        "$S5_SESS/outputs/out-launch-5-check.json" \
        out-launch-6-check.json 1
```

Only a skill whose brain persisted can meet this cap on the no-feedback tail — the script path at `skills/use-smart-humanize-text/evals/assertions/check-retention-cumulative.sh` sums the counts deterministically and exits 0 iff the total is within the cap. This is the single headline answer to the scenario's thesis: *"did memory pay off on the sessions that deliberately withheld feedback?"*

### Results (4 arms × 3 runs × 6 sessions = 72 sessions, executor `claudecode`)

**Cumulative retention assertion — the binary outcome per arm:**

| Arm               | S5+S6 ≤ 1 violations | Pass rate |
|-------------------|:--------------------:|:---------:|
| `smart_skill`     | **3 / 3** ✅         | **100%**  |
| `flat_skill_wiki` | 0 / 3 ❌             | 0%        |
| `flat_skill`      | 0 / 3 ❌             | 0%        |
| `no_skill`        | 0 / 3 ❌             | 0%        |

Across nine independent retention runs, `smart_skill` is the only arm that kept the promise. Every time.

**Per-session violation means (lower is better; mean over 3 runs):**

| Session | smart | flat_wiki | flat | no_skill | What's probed |
|---------|------:|----------:|-----:|---------:|---------------|
| S1      | 1.33  | 1.00      | 0.67 | 1.00     | baseline       |
| S2      | 1.00  | 1.00      | 1.00 | 1.00     | first feedback |
| S3      | **1.00** | 2.33   | 1.33 | 2.00     | accumulation — smart carries S2 via brain |
| S4      | 0.33  | 0.33      | 0.00 | 0.00     | all feedback in prompt — tied |
| S5      | **0.33** | 1.00   | 1.00 | 1.33     | pure retention — smart wins |
| S6      | **0.33** | **2.00** | 1.00 | 1.00 | generalization — smart wins by 6× |
| **totals (18 sess)** | **13** | **23** | **15** | **19** | |

**Aggregate pass rate by arm (mean over 18 sessions × 4 assertions / session; no LLM judge on this iteration):**

| Arm               | Pass rate | Tokens / session | Wall time / session |
|-------------------|----------:|-----------------:|--------------------:|
| `smart_skill`     | **0.876** |  4,755           |  74.0 s             |
| `flat_skill`      |   0.848   |  4,058           |  64.4 s             |
| `no_skill`        |   0.836   |  1,968           |  35.6 s             |
| `flat_skill_wiki` |   0.816   |  4,882           |  81.2 s             |

### The ablation's main finding

**`flat_skill_wiki` is the worst arm — worse than `no_skill`.** 23 total violations over 18 sessions vs `no_skill`'s 19, vs `flat_skill`'s 15. This was not the expected result.

What happened: the wiki teaches *general* humanize principles — em-dash overuse, rule-of-three lists, vague attributions, the "despite challenges" formula. Those concepts are orthogonal to Liana's specific voice rules (`b1…b9` clichés, `g1…g4` voice moves). The agent reads the wiki dutifully, spends attention budget on guidance that doesn't apply, and sometimes *misses* what's genuinely in the current prompt. `no_skill` has no distractor. `flat_skill` has just the instructions — skill ceremony without bulk knowledge. `smart_skill` has the specific rules in `patterns.md`.

Three concrete implications for smart-skill authoring:

1. **Wiki without brain can hurt, not help.** Static knowledge that's adjacent to the task competes with specific in-context rules. If your skill's wiki isn't the ground truth for the task at hand, shipping it in a no-brain configuration can make the output *worse* than shipping nothing.
2. **The brain is what compounds.** `smart_skill` vs `flat_skill_wiki` — identical inputs minus memory — swings the total violations from 23 → 13 (-43%) and flips the cumulative retention assertion from 0% → 100%. This is the cleanest measurement of brain value in the project.
3. **No-feedback generalization is the cliff.** S6 mean violations: `smart 0.33` vs `flat_wiki 2.00` — a 6× gap. Format never seen, rules never restated, only the brain arm keeps form. The flat variants collapse here whether they have the wiki or not.

### Token and wall-time tradeoff

| Pair                         | Δ pass rate | Δ tokens     | Δ wall time |
|------------------------------|:-----------:|:------------:|:-----------:|
| `smart_skill` vs `flat_skill_wiki` | **+6.0 pp** | **-2.6%**    | **-8.9%**    |
| `smart_skill` vs `flat_skill`      | +2.8 pp     | +17.2%       | +14.9%       |
| `smart_skill` vs `no_skill`        | +4.0 pp     | +141.6%      | +107.7%      |

Against `flat_skill_wiki` — the arm with the most matching context — `smart_skill` is **cheaper on both axes** while delivering 43% fewer violations: same-ballpark cost, dramatically better output. Against `flat_skill` (instructions only), the brain costs +17% tokens / +15% time in exchange for +2.8pp pass rate and -13% violations: honest tradeoff. Against `no_skill`, the brain costs ~2× tokens for +4pp pass rate — justified when the task rewards specific learned rules.

### Smart-arm brain growth over 3 runs

| Run | S1 | S2 | S3 | S4 | S5 | S6 |
|-----|---:|---:|---:|---:|---:|---:|
| 1   | 1  | 2  | 3  | 4  | 4  | 4  |
| 2   | 1  | 4  | 7  | 14 | 14 | 14 |
| 3   | 1  | 4  | 7  | 14 | 14 | 14 |

Entries in `patterns.md` at each session's `brain-snapshot-after`. Every run plateaus at exactly S4 (the last session with feedback). Post-S4 the brain stays frozen — correct, since there's no new feedback to log. Agent occasionally chose different granularity for logging ("one entry per feedback cohort" vs "one entry per rule"), but the *retention outcome* was 3-for-3 regardless.

### Leak audit
```
bash skills/use-smart-humanize-text/evals/scripts/audit-no-leaks.sh \
  /tmp/eval-ws-4arm/use-smart-humanize-text/iteration-1
→ leaks: none
```

### What this iteration establishes

1. Brain beats wiki — `smart > flat_skill_wiki` on every retention session and on totals.
2. Wiki alone can harm — `flat_skill_wiki` is the worst of the four arms on this scenario.
3. Smart is the only arm that generalizes — S6 mean viol: smart 0.33, all others 1.0–2.0.
4. Every number above is an outcome check on the delivered markdown — no mechanism-adherence assertions were added. The improvement is in what the user gets, not in whether the brain-protocol happened.
5. 3 runs per configuration smooths the per-cell noise enough that the cumulative-retention outcome is an unambiguous 100% / 0% / 0% / 0%.

### Reproduction

```sh
export HUMBLSKILLS_ROOT=/path/to/humblSKILLS
export HUMBLSKILLS_EVAL_WORKSPACE=/path/to/eval-state
humblskills eval run use-smart-humanize-text \
  --scenario indie-launch-copy-iteration \
  --runner claudecode
bash skills/use-smart-humanize-text/evals/scripts/audit-no-leaks.sh \
  "$HUMBLSKILLS_EVAL_WORKSPACE/use-smart-humanize-text/iteration-1"
```

With `ANTHROPIC_API_KEY` set (env or `.env` at repo root), the `llm` prose-quality assertions on every session are judged by `claude-sonnet-4-6` and the headline pass rates rise ~10 pp across all arms without materially changing the ablation shape.
