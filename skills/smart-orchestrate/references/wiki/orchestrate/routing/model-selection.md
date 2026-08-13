---
title: "Route Subtasks by Scope, Difficulty, and Blast Radius"
context: orchestrate
category: routing
concept: model-selection
description: "Bias cheap and narrow, escalate on the four named signals instead of on vibes"
tags: routing, model-tier, cost, escalation, blast-radius
sources:
  - "references/raw/orchestrate-SKILL.md"
last_ingested: 2026-08-12
---

## Routing Guidelines

These are judgment calls, not hard rules. Bias cheap and narrow; escalate when unsure.

| Signal | Prefer |
|---|---|
| Clear brief, small blast radius, mechanical change | Fastest / cheapest worker |
| Multi-file but well-specified; some judgment | Mid-tier worker (e.g. Sonnet 5, Terra) |
| Ambiguous design, high blast radius, tricky correctness | Stronger worker (e.g. Opus 5) or keep on parent |
| Cross-cutting architecture, phase planning, final integration check | Parent frontier only |

This table is deliberately generic so it survives model churn. For which
concrete `cursor-agent --model` IDs are actually reachable — and which are
measured-broken — see
`references/wiki/orchestrate/routing/cursor-cli-models.md`. Pick the tier here,
then pick a verified ID there.

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

- `references/raw/orchestrate-SKILL.md`
