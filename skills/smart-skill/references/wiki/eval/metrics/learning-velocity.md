---
title: "Learning Velocity - Does the Brain Actually Compound?"
context: eval
category: metrics
concept: learning-velocity
description: "The per-session slope of pass_rate. Zero on flat_skill; positive on smart_skill means the brain is compounding session over session."
tags: eval, metrics, longitudinal, humblskills
sources: []
last_ingested: 2026-04-19
---

## What it measures

For a scenario with K sessions run under the `smart_skill` arm:

```
learning_velocity = slope(pass_rate ~ session_number)
```

Computed as the least-squares regression slope. Units: pass_rate points
per session.

A flat_skill arm, by construction, has no brain state to carry forward -
its learning_velocity should sit near zero. A smart_skill arm that's
actually using its brain should have a positive slope.

## What good looks like

| Scenario length | Good velocity         | Interpretation                          |
|-----------------|-----------------------|-----------------------------------------|
| 3 sessions      | > 0.05                | Brain picked up 1-2 concepts worth using |
| 5 sessions      | > 0.08                | Compounding is visible                   |
| 10+ sessions    | > 0.10                | Clear trajectory                         |

Velocities below those numbers on a smart_skill arm indicate one of:

1. Scenarios don't actually require the brain (task is memory-free).
2. The skill isn't writing useful data to patterns.md / decisions.md.
3. The skill's Brain Protocol isn't being followed (check
   `reads_from_brain` in metrics.json).

## Incorrect interpretation

```
smart_skill  learning_velocity = 0.02   <-- "brain works!"
flat_skill   learning_velocity = 0.00
```

A velocity of 0.02 is within noise - you cannot claim compounding from
this. Add more sessions, or add assertions that explicitly test retention
across sessions.

## Correct interpretation

```
smart_skill  learning_velocity = 0.15   <-- pass_rate climbs a point every session and a half
flat_skill   learning_velocity = 0.01
retention_at_k[k=2] = 0.90              <-- session N+2 remembers lessons from session N
```

Now you have a claim: the brain compounds at 15 pass_rate points per
session, and retains lessons for at least two sessions.

## Sources

Implementation: [cli/internal/eval/metrics/metrics.go](../../../../../../cli/internal/eval/metrics/metrics.go) `computeDerived`.
