---
title: "Problem Breakdown: Harmful Content Detection"
context: breakdown
category: core
concept: harmful-content
description: "Facebook post moderation as multi-modal classification: minimize views of harmful content subject to a precision guardrail, with content signals dominating."
tags: harmful-content, classification, multi-modal, moderation, breakdown
sources:
  - "references/raw/learn/ml-system-design/problem-breakdowns/harmful-content.md"
last_ingested: 2026-06-01
---

## Problem Breakdown: Harmful Content Detection

Moderate Facebook posts (text + images) at 1B posts/day, harmful content < 1%.

**Framing.** Business objective is NOT accuracy. The strong objective is
**minimize views of harmful content, subject to a precision guardrail** (e.g.
95% precision before automated removal). Weighting by views matters: a harmful
post with many views matters far more than one with none, and the abstraction
answers "must we classify immediately or can we wait?" objectively (waiting
accrues views). ML objective: **binary classification**, harmful or not, with an
explicit precision/recall trade-off. Deep insight: harmful content gets easier
to detect as more people react to it.

**Data / signals.** Supervised: a small labeled set (50k, balanced). Semi-
supervised: user reports (~10M, noisy proxy labels). Self-supervised: predict
comments from post body for richer representations. **Content signals dominate**
(post text + image), supported by actor/creator signals (user embedding + real-
time tallies like reports, account age) and behavioral signals (negative-
reaction rate, share-per-view, Bayesian-smoothed for low view counts). Handle
class imbalance with balanced sampling + loss weighting.

**Model.** Baseline: logistic regression on numeric + embedding features. Then
late fusion -> **multi-modal transformer** (ViT for images, text encoder, cross-
attention to fuse, plus behavioral/user embeddings). Loss: view-weighted BCE
(log term, not raw views, since views are power-law) + a multi-task report-
prediction head to exploit semi-supervised data.

**Evaluation.** Offline: PR-AUC and recall@precision95, impression-weighted.
Online: run candidate vs control side-by-side with importance sampling on scores
(label fewer obvious cases); use proxies (reports, appeals). Watch positive
suppression from re-triggering on removed content.

**Key deep dives.** User embedding training (simple ID embedding -> transductive
GCN -> inductive GraphSAGE for cold start); two-stage cascade (distilled light
filter -> heavy model) for 1B/day; calibration; re-trigger on behavioral updates.

> Assumed flow (HLD diagram in source omitted; inferred): post creation (or
> behavioral update) -> lightweight filter -> if score high, heavy multi-modal
> classifier -> calibration layer enforcing the precision guardrail -> action
> router (auto-remove / demote / queue for human moderator).

## Sources

- `references/raw/learn/ml-system-design/problem-breakdowns/harmful-content.md` -
  full framing, data buckets, multi-modal model, evaluation, and deep dives.
