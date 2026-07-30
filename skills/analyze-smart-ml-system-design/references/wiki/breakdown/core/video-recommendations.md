---
title: "Problem Breakdown: Video Recommendations"
context: breakdown
category: core
concept: video-recommendations
description: "YouTube 'up next' as a multi-stage ranking problem: candidate generation -> light ranker -> heavy ranker -> re-rank, optimizing quality-adjusted watch time."
tags: recommendations, ranking, two-tower, multi-stage, breakdown
sources:
  - "references/raw/learn/ml-system-design/problem-breakdowns/video-recommendations.md"
last_ingested: 2026-06-01
---

## Problem Breakdown: Video Recommendations

YouTube "up next" recommendations: 1B videos, 1B DAU, 5 slots, 250ms budget.

**Framing.** Business objective: maximize **quality-adjusted watch time** (or,
stronger, long-term satisfaction balancing users, creators, and platform), NOT
raw CTR (rewards clickbait). ML objective: a **ranking problem** given a user
and context (current session + the video being watched), responsive as the user
reveals intent. Ranking is a function of predicted future behavior (click,
watch, share) combined in a value model.

**High-level design.** Standard multi-stage architecture: **candidate
generation** (hundreds of generators in parallel, some universal like "top 10k",
some personalized like subscriptions; highly cacheable) -> **light ranker** (10k
-> 100 candidates, recall-oriented, fast, often GBDT on CPU) -> **heavy ranker**
(100 candidates, precision-oriented, full features) -> **re-ranking** (value
model, diversity, new-creator promotion).

**Data / signals.** Explicit feedback (likes, subscriptions, "not interested") -
high quality but rare. Implicit feedback (watch time absolute/relative, returns,
attrition as a strong negative) - the bulk. Contextual data (current video,
prior searches, time, device). **Behavioral signals dominate**; content alone
says little about engagement. Candidate generators correspond closely to
informative features.

**Model.** Baselines: random blend, simple collaborative filtering. Candidate
generation: **two-tower** user/item embeddings with triplet loss, hard-negative
mining, served from a vector DB via ANN. Heavy ranker: simple MLP -> DLRM ->
**transformer sequence ranker** (multi-task heads: watch time, CTR, like, share,
completion, return). Loss: watch-time-weighted engagement + auxiliary heads +
position-bias correction.

**Evaluation.** Offline: NDCG, MAP, precision/recall per head + diversity;
evaluate candidate-generator recall with unbiased inputs. Online: A/B on session
watch time, return rate, long-term trends; beware novelty effects (run long
enough).

**Key deep dives.** Feedback loops (popularity bias, filter bubbles; fix with
counterfactual logging, inverse-propensity reweighting, diversity constraints);
cold start (onboarding clusters for new users, content features + controlled
exploration for new videos); explore/exploit (Thompson sampling, contextual
bandits, per-slot risk budget).

> Assumed flow (HLD diagram in source omitted; inferred): a watch request fans
> out to parallel candidate generators (cached embeddings + vector-DB ANN) ->
> light ranker trims to ~100 -> heavy transformer ranker scores -> re-ranker
> applies the value model and diversity -> 5 slots served within 250ms.

## Sources

- `references/raw/learn/ml-system-design/problem-breakdowns/video-recommendations.md` -
  multi-stage design, two-tower retrieval, transformer ranker, deep dives.
