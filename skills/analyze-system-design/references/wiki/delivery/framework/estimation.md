---
title: "Capacity Estimation Only When It Changes the Design"
context: delivery
category: framework
concept: estimation
description: "Skip up-front back-of-envelope math; do calculations only when a number actually changes a design decision (one node vs shard, sync vs async)."
tags: estimation, capacity, math, qps
sources:
  - "references/raw/learn/system-design/in-a-hurry/delivery.md"
  - "references/raw/learn/system-design/core-concepts/numbers-to-know.md"
last_ingested: 2026-06-01
---

## Capacity Estimation Only When It Changes the Design

Most guides tell you to compute storage/DAU/QPS up front. Skip it. Interviewers gain nothing from "ok, it's a lot."

**Incorrect (ritual math that changes nothing):**

```text
"100M DAU x 10 tweets x 300 bytes = ~300GB/day... ok, that's a lot.
Anyway, moving on to the design."
```

Wasted minutes, zero design impact.

**Correct (math that drives a decision):**

```text
Designing TopK trending topics: estimate the number of distinct topics.
If it fits one min-heap on a single node -> single instance.
If not -> shard the heap across N instances. The number decides the design.
```

State to the interviewer: "I'll skip estimates up front and do math inline when it changes a decision." Then do it only when a number resolves a real fork: does the working set fit one node? do we need to shard? can a queue sit in the synchronous path? Use `numbers-to-know` for the order-of-magnitude figures that answer these.

Learning to estimate fast still helps you reason about trade-offs; just deploy it surgically, not as an opening ritual.

## Sources

- `references/raw/learn/system-design/in-a-hurry/delivery.md` - the "do math only when it changes the design" rule and the TopK example.
- `references/raw/learn/system-design/core-concepts/numbers-to-know.md` - the figures used in inline estimates.
