# evals/ for use-smart-humanize-text

One scenario: `adaptive-brand-voice-discovery`. Six sessions. Three arms (smart_skill, flat_skill, no_skill). Designed to prove Smart Skills learn idiosyncratic, non-derivable rules over repeated use — with hard numbers.

## Running

```sh
# Short form (canonical demo, opens report in browser):
humblskills eval brand-voice

# Equivalent long form:
humblskills eval run use-smart-humanize-text --scenario adaptive-brand-voice-discovery --open

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

## Files

### `assertions/`

- `check-brand-voice.sh` — deterministic ArcFactor rule detector. Emits JSON `{violations, count, rules_violated}`. Invoked by harness assertions (not staged to the agent), so its comments cannot leak rules to the model.

### `files/`

- `arcfactor-brief-1.md` … `arcfactor-brief-6.md` — content briefs with casual/incorrect formats that must be translated to ArcFactor house style. Each brief includes company context, required data, and an output-path instruction.
