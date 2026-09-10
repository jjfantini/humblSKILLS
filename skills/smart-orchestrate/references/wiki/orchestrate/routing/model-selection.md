---
title: "Route Subtasks by Scope, Difficulty, and Blast Radius"
context: orchestrate
category: routing
concept: model-selection
description: "Bias cheap and narrow, escalate on the four named signals instead of on vibes - and route on (tier, effort), not tier alone"
tags: routing, model-tier, effort, cost, escalation, blast-radius
sources:
  - "references/raw/orchestrate-SKILL.md"
  - "references/raw/anthropic-prompting-claude-fable-5-1-2026-09-10.md"
  - "references/raw/orchestrator-model-sweep-2026-09-10.md"
last_ingested: 2026-09-10
---

## Routing Guidelines

These are judgment calls, not hard rules. Bias cheap and narrow; escalate when unsure.

| Signal | Prefer | Effort |
|---|---|---|
| Clear brief, small blast radius, mechanical change | Fastest / cheapest worker | `low` |
| Multi-file but well-specified; some judgment | Mid-tier worker (e.g. Sonnet 5, Terra), or a frontier model dialled down | `medium` |
| Ambiguous design, high blast radius, tricky correctness | Stronger worker (e.g. Opus 5, Fable 5.1) or keep on parent | `high` |
| Cross-cutting architecture, phase planning, final integration check | Parent frontier only | `high`, `max` for the plan itself |

This table is deliberately generic so it survives model churn. For which
concrete `cursor-agent --model` IDs are actually reachable — and which are
measured-broken — see
`references/wiki/orchestrate/routing/cursor-cli-models.md`. Pick the tier here,
then pick a verified ID there.

## Routing Is Now Two Axes: Tier and Effort

On the current frontier models (Claude Fable 5.1, GPT-6 Astra) the reasoning
effort carries most of the cost/latency trade-off, and Anthropic states it
plainly: *"Effort is the primary control for trading off intelligence, latency,
and cost."* Two consequences for the table above:

1. **A cheap slot no longer implies a small model.** Fable 5.1 at `low` is, per
   Anthropic's own comparison, "often competitive with Claude Opus and Claude
   Sonnet models on cost per task while scoring higher." Routing *down a tier*
   and routing *down in effort* are different moves, and the second is often the
   better one because the model keeps its capability ceiling for the parts of the
   brief that need it.
2. **Effort names don't transfer across models.** A sweep run on one model tells
   you nothing about where the same label sits on another. Re-measure per model.

**A routing decision that names only the model is incomplete.** It leaves the
dominant cost variable at the harness default. Write both into the brief — see
`references/wiki/orchestrate/prompting/brief-prompting-by-vendor.md`.

## Two Non-Performance Constraints That Override the Table

The tier table asks "how strong a worker does this deserve." Two constraints can
veto its answer outright:

- **Data retention.** Fable 5.1 carries 30-day retention and is not offered
  under zero data retention without express authorization — Cursor tags every
  `claude-fable-5-1-*` ID `(NO ZDR)`. Proprietary code plus a ZDR requirement
  means the ID is unusable regardless of how well it scores.
- **Account entitlement, per CLI.** A model your organisation has not enabled
  fails in ~3s with `ActionRequiredError: Model Blocked`, no matter how correct
  the routing was. Availability is a property of the *backend account*, not the
  model: on 2026-09-10 Fable 5.1 was 0/9 through `cursor-agent` and 4/4 through
  Claude Code on the same machine within the same hour. Verify per CLI you
  intend to dispatch through.

## Escalate to the Parent (or a Stronger Worker) When

- The brief is underspecified or contradicts the codebase
- The change touches shared contracts, auth, data, or many callers
- Verification fails and the fix isn't obvious
- A worker wants to expand scope

## The Expensive Mis-Route

**Incorrect:**

```
Subtask: "Add the rate-limit check to the auth middleware."
Routed to: cheapest worker (clear-sounding one-liner)
```

It reads mechanical, but "auth middleware" is a shared contract with many callers
— row two of the escalation list. The worker guesses at the contract, the parent
re-verifies every caller by hand, and the round trip costs more than routing it
correctly the first time.

**Correct:**

Read the *blast radius*, not the sentence length. A one-line diff in a file that
forty callers depend on is a stronger-worker or parent task. A forty-line diff in
a leaf file nothing imports is a cheapest-worker task.

## Sources

- `references/raw/orchestrate-SKILL.md` — the tier table and escalation signals.
- `references/raw/anthropic-prompting-claude-fable-5-1-2026-09-10.md` — effort as
  the primary cost control, and the `low`-tier cost comparison.
- `references/raw/orchestrator-model-sweep-2026-09-10.md` — the entitlement-block
  measurements.
