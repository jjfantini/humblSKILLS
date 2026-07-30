---
title: "Step 5: Inference and Evaluation"
context: delivery
category: framework
concept: inference-eval
description: "The ~7-minute step: offline metrics as a proxy for online A/B tests tied to the business objective, plus practical inference constraints (scale, latency, cost)."
tags: evaluation, inference, offline-online, ab-test, latency
sources:
  - "references/raw/learn/ml-system-design/in-a-hurry/delivery.md"
last_ingested: 2026-06-01
---

## Step 5: Inference and Evaluation

About 7 minutes. Two halves: how you measure the model, and how you serve it.

**Evaluation.** Offline evaluation uses historical data (precision, recall,
NDCG, MAP on a held-out set) to estimate production performance. Online
evaluation measures real impact via A/B tests on business metrics (CTR,
conversion, AOV). Offline is a fast-iteration proxy for online; the offline
metric only matters if it correlates with the online outcome. Always tie
metrics back to the business objective: a model that wins on ML metrics but does
not move business outcomes is not valuable. (Full 5-layer structure in
`concepts/core/evaluation`.)

**Inference.** Problems get interesting when you operationalize them. Discuss
scale, latency, and cost: do we distill the model, cache, quantize, or prune?
If inference is offline and small-scale, there may be little to say. At massive
scale with tight latency, expect the interviewer to push hard here. Inference
depth is where lab-only candidates get exposed, which can be a deal-breaker for
applied roles.

**Incorrect (metric in a vacuum):**

```text
"I will report accuracy on the test set."
```

**Correct (tied to objective + serving plan):**

```text
"Offline: PR-AUC and recall@precision95, impression-weighted to match the
exposure objective. Online: shadow mode then A/B on harmful-content views and
false-removal rate. Serving: distilled first-stage filter + caching for viral
content."
```

- Green: clear offline + online metrics tied to business objectives; practical
  inference constraints considered; concrete optimizations where relevant.
- Red: metrics disconnected from business objectives; accuracy-only focus;
  gratuitous optimizations with no justification.

## Sources

- `references/raw/learn/ml-system-design/in-a-hurry/delivery.md` - evaluation
  design, inference considerations, and green/red flags.
