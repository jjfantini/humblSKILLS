---
title: "The 6-Step ML Delivery Framework"
context: delivery
category: framework
concept: overview
description: "The full-interview scaffold for applied ML system design: six steps with time targets that take an ambiguous prompt to a production-minded solution."
tags: framework, delivery, interview-structure, timing
sources:
  - "references/raw/learn/ml-system-design/in-a-hurry/delivery.md"
  - "references/raw/diagrams/ml-delivery-framework.png"
last_ingested: 2026-06-01
---

## The 6-Step ML Delivery Framework

ML design interviews are less standardized than SWE system design, so a fixed
scaffold keeps you on the load-bearing parts. These steps are guideposts, not
hard rules: if the interviewer pulls you off course, follow their lead. Timings
assume a ~45 min interview.

| # | Step | Time | Output |
|---|------|------|--------|
| 1 | Problem framing | 5-7 min | Clarified problem + business objective + ML objective |
| 2 | High-level design | 2-3 min | Lifecycle block diagram (inputs -> features -> model -> action) |
| 3 | Data and features | ~10 min | Signal sources, training data buckets, representation |
| 4 | Modeling | ~10 min | Baseline -> model selection + trade-offs -> architecture |
| 5 | Inference and evaluation | ~7 min | Offline + online metrics tied to business objective; inference plan |
| 6 | Deep dives | remaining | Edge cases, scaling, monitoring and retraining |

The flow runs problem-framing -> high-level design -> data and features ->
modeling -> inference and evaluation -> deep dives. Each step has its own
concept under `delivery/framework/`.

> Assumed flow (diagram in source omitted; inferred): the framework PNG at
> `references/raw/diagrams/ml-delivery-framework.png` renders the six steps
> left-to-right as a single horizontal pipeline, with framing feeding design,
> design feeding data/features, and so on, then looping into deep dives.

The goal is not a perfect system in 45 minutes. It is demonstrating structured
thinking and reasonable trade-offs: taking an ambiguous business problem and
showing how ML drives real impact, not just a model.

**Green flags (whole interview):** clear business objective driving every
choice; structure the interviewer can follow; depth where the problem is hard.

**Red flags:** jumping to modeling before framing; feature dumps that burn the
clock; ML metrics disconnected from business outcomes; hand-waved architecture.

## Sources

- `references/raw/learn/ml-system-design/in-a-hurry/delivery.md` - step list,
  timings, and per-step green/red flags.
- `references/raw/diagrams/ml-delivery-framework.png` - the framework diagram
  cited across `delivery/framework/` concepts.
