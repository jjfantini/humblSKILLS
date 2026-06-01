---
title: "Problem Breakdown: Bot Detection"
context: breakdown
category: core
concept: bot-detection
description: "Adversarial binary classification on a social platform: minimize bot impact on legitimate users with graduated actions, behavioral and graph signals leading."
tags: bot-detection, adversarial, classification, graph, breakdown
sources:
  - "references/raw/learn/ml-system-design/problem-breakdowns/bot-detection.md"
last_ingested: 2026-06-01
---

## Problem Breakdown: Bot Detection

Separate bots from legitimate users on a 500M-DAU platform. Adversarial: bot
authors constantly adapt. Prevalence high before heuristics (~50%), < 1% after.
Ground-truth labels scarce (investigators handle low 100s/week).

**Framing.** Business objective: **minimize the impact of bot activity on
legitimate user experience, subject to false-positive guardrails** (e.g. < 1%
FPR), not "maximize bots caught" (perverse) or accuracy (meaningless at 50%
prevalence). Impact = spam impressions, rejected friend requests. ML objective:
**binary classification**, bot or not. Use **graduated actions**: ban when
confident, otherwise demote or limit visibility to cap false-positive damage.

**Data / signals.** Ground-truth investigator labels (scarce, for test sets and
calibration). User-generated signals: reports, appeal outcomes, abuse reports
(noisy, scalable). Network labels: IP clustering, behavioral similarity,
registration patterns -> propagate via the graph to catch whole networks.
Synthetic data (conditional GANs) for the rare class. **Behavioral and
network/graph signals dominate**: activity patterns (cadence, burst detection,
circadian rhythm, entropy over multiple windows), network topology
(follower/following ratios, clustering coefficients), account metadata. Content
signals are weakest (easiest to fool adversarially).

**Model.** Baseline: logistic regression on simple signals. Then a **graph-
sequence model**: a GraphSAGE branch (inductive, k=2 hops, relation-specific
edges) + a sequence branch (GRU over ~200 bucketed events, not a transformer
since we model behavior not language) fused via cross-attention -> small MLP ->
risk score. Self-supervised pretraining on each branch (masked nodes/edges,
masked events), then light supervised fine-tune on the precious labels. Loss:
capped weighted BCE.

**Evaluation.** Offline: precision@recall90, PR-AUC, impact-weighted; validation
stratified by time to reflect drift. Online: candidate vs control with
importance sampling near decision boundaries; proxies (reports, appeal success).

**Key deep dives.** Calibration (Platt scaling, given scarce labels); anomaly
detection for unknown bots (isolation forests + autoencoders ensemble); two-
stage cascade (80-90% compute reduction); holdouts to fight positive suppression
and obscure signal from adversaries.

> Assumed flow (HLD diagram in source omitted; inferred): account events trigger
> a lightweight filter; suspicious accounts pass to the heavy graph-sequence
> model and an unsupervised anomaly branch; scores merge and route to graduated
> actions (ban / demote / limit) with confident cases optionally sent to
> investigators.

## Sources

- `references/raw/learn/ml-system-design/problem-breakdowns/bot-detection.md` -
  adversarial framing, data sources, graph-sequence model, calibration, anomaly.
