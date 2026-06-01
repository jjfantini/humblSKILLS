---
title: "Evaluation: The 5-Layer Framework"
context: concepts
category: core
concept: evaluation
description: "A general evaluation structure for any ML system design problem: work top-down from business objective to product metrics, ML metrics, methodology, then challenges."
tags: evaluation, metrics, offline-online, precision-recall, ndcg
sources:
  - "references/raw/learn/ml-system-design/core-concepts/evaluation.md"
last_ingested: 2026-06-01
---

## Evaluation: The 5-Layer Framework

Every ML system design problem needs an evaluation, and showcasing your ability
to evaluate (and so improve) a system is core to the interview. Production ML is
hard to evaluate: subjectivity, long feedback loops, many components. Use a
5-layer structure, working top-down so the metric stays tethered to something
real, not a vanity number.

1. **Business objective** - the real, valuable goal from problem framing. Anchor
   everything here. Ask "what action follows the prediction, and how does it
   impact the business?"
2. **Product metrics** - user-facing signals of success (CTR, conversion,
   retention, time-to-resolution, operational review cost, appeal rate).
3. **ML metrics** - technical metrics aligned to the product goal, measurable
   without new inputs. Classification: precision, recall, PR-AUC, F1, ROC-AUC.
   Ranking/IR: NDCG, MAP, MRR, recall@k, coverage.
4. **Evaluation methodology** - offline (historical data, fast iteration, a
   proxy for online) and online (shadow mode, then A/B tests measuring real
   impact). The offline metric only matters if it correlates with the online
   outcome; validate that correlation repeatedly.
5. **Address challenges** - class imbalance, labeling cost, fairness, feedback
   loops, and how you would mitigate them.

**Key metric notes.** For strong class imbalance (99% negative), **PR-AUC beats
ROC-AUC**: ROC-AUC can look great while the classifier is useless. Clicks are a
biased proxy (presentation/position bias); debias with inverse-propensity
weighting or interleaving (10-20x less traffic than A/B for the same power).
Short-term metrics (CTR) often cannibalize long-term ones (retention); measure
both and treat divergence as a trap.

**Per-domain pitfalls.** Recommenders: long evaluation horizon, feedback loops,
plan for replayability (log all candidates, not just the ranked list). Search:
query ambiguity, long-tail sparse judgments, freshness. Generative: subjective
quality, hallucination, safety/policy compliance, costly human review.

Always tie ML metrics back to the business objective: a model that wins on ML
metrics but does not move business outcomes is not valuable.

## Sources

- `references/raw/learn/ml-system-design/core-concepts/evaluation.md` - the
  5-layer framework, per-domain metrics, and metric definitions.
