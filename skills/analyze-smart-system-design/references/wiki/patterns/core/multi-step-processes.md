---
title: "Multi-Step Processes"
context: patterns
category: core
concept: multi-step-processes
description: "Reliably complete workflows spanning multiple services; model as a state machine, coordinate via orchestration or choreography, use sagas for rollback."
tags: workflow, state-machine, saga, orchestration
sources:
  - "references/raw/learn/system-design/patterns/multi-step-processes.md"
last_ingested: 2026-06-01
---

## Multi-Step Processes

Workflows with several stateful stages that can partially fail: orders, signups, provisioning.

**Incorrect (best-effort chain, no durable state):**

```text
charge card -> reserve inventory -> ship. Step 2 fails after step 1;
nothing tracks where it stopped, so the charge is never reversed.
```

**Correct (durable state machine + compensation):**

```text
Model the workflow as a state machine; persist state per step.
Coordinate via an orchestrator (workflow engine) or choreography (events).
On failure, run saga compensating actions to roll back completed steps.
```

Keep each step **idempotent and retryable** so a resumed or retried run is safe. Prefer the **saga pattern** with compensating actions over distributed two-phase commit, which does not scale. Durable per-step state lets you resume after a crash instead of restarting.

## Sources

- `references/raw/learn/system-design/patterns/multi-step-processes.md` - state machine, orchestration vs choreography, saga compensation.
