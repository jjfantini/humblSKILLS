# evals/ for use-smart-humanize-text

Two scenarios, both 6 sessions. The `indie-launch-copy-iteration` scenario runs a **4-arm ablation** — `smart_skill`, `flat_skill_wiki`, `flat_skill`, `no_skill` — at 3 runs per arm, to isolate which layer of the skill actually produces the compounding win. `adaptive-brand-voice-discovery` still runs the original 3-arm setup. Both cover different axes of learning:

| Scenario | Tests | What "bad" looks like | What "good" looks like |
|----------|-------|-----------------------|------------------------|
| `adaptive-brand-voice-discovery` | **Rejection of rigid surface rules** (naming, spelling, formatting, terminology) | A rule is violated (e.g. `Arc Factor` instead of `ArcFactor`) | Every rule absent |
| `indie-launch-copy-iteration`    | **Rejection of clichés AND reinforcement of positive voice moves** | A banned cliché appears **or** a required voice move is missing | Zero banned clichés, all four voice moves present |

The two scenarios together prove the brain learns **both** "don't do X" (banned patterns fade over time) and "always do Y" (good patterns get reinforced over time). Dashboards render them side by side as part of a single iteration.

## Running

```sh
# Short form (canonical demo, opens report in browser):
humblskills eval brand-voice

# Equivalent long form:
humblskills eval run use-smart-humanize-text --scenario adaptive-brand-voice-discovery --open

# Run only the new scenario:
humblskills eval run use-smart-humanize-text --scenario indie-launch-copy-iteration --open

# Run both scenarios (default when --scenario is omitted):
humblskills eval run use-smart-humanize-text --open

# Head-to-head subset:
humblskills eval run use-smart-humanize-text --config smart_skill,no_skill
```

## Scenario: `adaptive-brand-voice-discovery`

### What makes it hard

Ten style rules for a fictional fintech company ("ArcFactor Capital"). The rules are invented — no frontier model can infer them from pretraining. Each content brief uses **dirty** formats (`$35M`, "April 18, 2026", "clients", "Arc-Factor", "Pulse Index", "Bridgewater", "AI-driven"...) that must be translated into ArcFactor house style (`CAD 35,000,000`, `2026-04-18`, "customers", "ArcFactor", "PulseIndex", "a major incumbent", "machine learning"...).

Rules are disclosed progressively via prompt feedback:

| Session | Content type          | New rules disclosed in prompt | Cumulative rules the arm knows |
|---------|-----------------------|-------------------------------|--------------------------------|
| 1       | Quarterly memo        | none                          | 0 (baseline)                   |
| 2       | Customer newsletter   | rules 1, 2, 5                 | 3                              |
| 3       | Market commentary     | rules 3, 6                    | smart: 5 / flat+no: 2          |
| 4       | Signal-decay blog     | rules 4, 7, 8, 9, 10          | smart: 10 / flat+no: 5         |
| 5       | Press release         | **none — pure retention**     | smart: 10 / flat+no: 0         |
| 6       | Annual letter excerpt | **none — generalization**     | smart: 10 / flat+no: 0         |

Session 5 and 6 are the crucial tests: no feedback in the prompt, so only a skill with a persistent brain can produce compliant output.

### The three arms

- **smart_skill**: brain persists across sessions. The agent logs each round's prompt feedback to `patterns.md` and applies rules from brain on every subsequent session.
- **flat_skill**: has `SKILL.md` but the brain is **re-derived before every session** (framework guarantee). Applies only the rules in the current session's prompt.
- **no_skill**: no skill at all. Relies purely on in-prompt guidance plus model priors.

### What gets measured

The harness emits per-session for every arm:
- `pass_rate` — share of assertions passed
- `tokens` — prompt + completion
- `duration_ms`, `cost_usd`
- `violations` — summed count field from `out-brand-N-check.json` (new field, added 2026-04-20)
- `patterns_entries`, `wiki_concepts`, `brain_bytes`, `reads_from_brain` (smart arm only)

The report renders these as per-arm time series (pass_rate, patterns as lines; tokens, violations as bars), per-arm summary tables, and absolute + **percent-change** deltas (`smart_vs_none`, `smart_vs_flat`).

### Graduated violation ceilings

Each session has a count-ceiling assertion (inline `awk` parsing `out-brand-N-check.json`):

| Session | Ceiling |
|---------|---------|
| 1       | ≤ 8 rules violated (baseline) |
| 2       | ≤ 6 (3 rules disclosed)       |
| 3       | ≤ 4 (smart: 5 in brain + 2 in prompt; flat: 2 in prompt) |
| 4       | ≤ 2 (smart: 10 in brain + 5 disclosed) |
| 5       | ≤ 1 (retention — smart has 10 in brain; flat/no start fresh) |
| 6       | = 0 (generalization) |

### Brain-retention assertions

Sessions 3, 4, 5 include explicit brain-retention checks: rules disclosed in earlier sessions' feedback **must not** reappear in the checker output. The only channel carrying those rules forward is the brain, so passing these proves the agent read and applied `patterns.md`.

## Scenario: `indie-launch-copy-iteration`

### What makes it hard

Liana Koval is a solo indie maker. Her voice is **specific**: she hates SaaS clichés and insists every blurb has a named audience, a concrete number, a first-person sentence, and a named limitation. Those rules are not in any model's priors — they can only be learned by reading Liana's feedback and carrying it forward.

Each session ships a new micro-product (Thinkmoss → Tabpile → Spritemash → Queuedeck → Warpshelf → Plipspace). The feedback is progressive: 9 clichés to **avoid** and 4 moves to **include**, disclosed across sessions 2–4, never restated after.

| Session | Product        | New rules in prompt (bad / good)                | Cumulative rules the arm knows |
|---------|----------------|-------------------------------------------------|--------------------------------|
| 1       | Thinkmoss      | none                                             | 0 (baseline)                   |
| 2       | Tabpile        | b1 (powerful), b2 (seamless) / g1 (audience)    | 3                              |
| 3       | Spritemash     | b3 (leverage), b4 (unleash) / g2 (number)       | smart: 6 / flat+no: 3          |
| 4       | Queuedeck      | b5–b9 (intuitive, effortless, revolutionary, game-changer, cutting-edge) / g3 (first-person), g4 (limitation) | smart: 13 / flat+no: 7 |
| 5       | Warpshelf      | **none — pure retention**                        | smart: 13 / flat+no: 0         |
| 6       | Plipspace (thread) | **none — generalization to 3-post thread**  | smart: 13 / flat+no: 0         |

Sessions 5 and 6 are the crucial tests: zero feedback in the prompt, so only a skill with a persistent brain produces compliant output.

### The 4-arm ablation

| Arm               | `SKILL.md` | `wiki/` | Persistent brain | What it isolates |
|-------------------|:----------:|:-------:|:----------------:|------------------|
| `no_skill`        | ✗          | ✗       | ✗                | Baseline — pure model |
| `flat_skill`      | ✓          | ✗       | ✗                | Instructions only (no static knowledge, no memory) |
| `flat_skill_wiki` | ✓          | ✓       | ✗ (reset each session) | + static wiki knowledge |
| `smart_skill`     | ✓          | ✓       | ✓                | + persistent memory |

`flat_skill_wiki` vs `flat_skill` measures whether the wiki's static content helps on its own; `smart_skill` vs `flat_skill_wiki` measures the pure compounding value of the persistent brain (identical inputs minus memory). The scenario's premise — Liana's voice rules are non-derivable from general humanize knowledge — predicts that `flat_skill_wiki` will **not** beat `flat_skill` on Liana-specific outcomes; the only arm that should compound is `smart_skill`. If this prediction holds, you're looking at a clean measurement of brain value, not just skill-vs-no-skill.

### Cumulative retention outcome

Session 6's assertion list ends with an **outcome-based cumulative check**: `violations(S5) + violations(S6) ≤ 1`. This is the single headline that answers *"did the skill keep its retention promise on the no-feedback tail?"* across the retention + generalization sessions. Only `smart_skill` should hit this cap; every other arm should fail it because its brain was reset (flat variants) or never existed (no_skill). The check runs `evals/assertions/check-retention-cumulative.sh`, which reads both sidecar JSONs directly.

### Why it pairs with the ArcFactor scenario

Where `adaptive-brand-voice-discovery` tests whether the brain can learn **surface-level rules that must not be violated**, `indie-launch-copy-iteration` tests whether it can learn **positive voice moves that must be present**. The checker's `count` field sums both kinds of deviations (banned cliché present == violation; required move missing == violation), so the same ceiling / retention assertions work for both axes.

### Graduated violation ceilings

| Session | Ceiling (of 13 rule units) | Total assertions |
|---------|----------------------------|------------------|
| 1       | ≤ 11 (baseline)            | 4                |
| 2       | ≤ 9 (3 units disclosed)    | 7                |
| 3       | ≤ 6 (smart: 6 in brain; flat: 3 in prompt) | 10 |
| 4       | ≤ 2 (smart: 13 in brain; flat: 7 in prompt) | 10 |
| 5       | ≤ 1 (retention — smart has 13 in brain) | 17 |
| 6       | ≤ 1 (generalization to a new format)    | 17 |

Sessions 5 and 6 carry the same 13 per-rule brain-retention assertions (one `! grep 'bN cliche' check.json` or `! grep 'gN voice' check.json` for each of the 13 rule units). The ceiling assertion is the coarse "smart arm met the bar" signal; the 13 per-rule asserts show the **profile** of which specific rules survived in each arm and make the smart-vs-flat gap visible even when both arms' ceilings move together.

### Leak audit

`scripts/audit-no-leaks.sh <iteration-dir>` proves the three-arm comparison is honest:

- For `no_skill` and `flat_skill`, the session-5 and session-6 prompts and transcripts contain **none** of the rule-disclosure fragments from sessions 2–4.
- For `flat_skill`, every `brain-snapshot-before/references/patterns.md` is free of scenario-specific entries (proving the harness resets the flat brain per session).
- For `smart_skill`, `patterns.md` grows monotonically session over session.

Run it after any `eval run` that included this scenario; it exits 0 with `leaks: none` when the comparison is valid.

## Files

### `assertions/`

- `check-brand-voice.sh` — deterministic ArcFactor rule detector for `adaptive-brand-voice-discovery`. Emits JSON `{violations, count, rules_violated}`.
- `check-launch-voice.sh` — deterministic Liana-voice detector for `indie-launch-copy-iteration`. Same JSON shape (`{violations, count, rules_checked, rules_violated}`) so the ceiling / retention awk snippet is identical across scenarios. Violations cover both banned clichés (`b1`…`b9`) and missing positive moves (`g1`…`g4`).
- `check-retention-cumulative.sh` — reads two `*-check.json` sidecars (S5 + S6) and asserts their combined violation count is at or below a cap. Used by the S6 cumulative-retention outcome assertion.

Both checkers are invoked by harness assertions, not staged to the agent, so their comments cannot leak rules to the model.

### `files/`

- `arcfactor-brief-1.md` … `arcfactor-brief-6.md` — content briefs for the ArcFactor scenario.
- `indie-brief-1.md` … `indie-brief-6.md` — content briefs for the Liana-voice scenario. Each brief names one of Liana's products, lists raw analyst-style bullets (with dirty formats and clichés seeded), and instructs the agent to write to `out-launch-N.md`.

### `scripts/`

- `audit-no-leaks.sh` — post-eval auditor for the `indie-launch-copy-iteration` three-arm comparison (see "Leak audit" above).
