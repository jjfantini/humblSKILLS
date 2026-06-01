---
title: "Generalization: Over/Underfit, Drift, Regularization, Leakage"
context: concepts
category: core
concept: generalization
description: "How a model performs on unseen data: spot over/underfitting via loss curves, balance capacity with data, handle drift, and regularize so the model survives production."
tags: generalization, overfitting, drift, regularization, leakage
sources:
  - "references/raw/learn/ml-system-design/core-concepts/generalization.md"
last_ingested: 2026-06-01
---

## Generalization: Over/Underfit, Drift, Regularization, Leakage

Generalization is performing well on new, unseen data, and it is the primary
goal of almost all industrial ML. It is not a binary switch: it is a gray area,
and you can overfit some data while underfitting other data at once. Be specific
in interviews, do not just say "avoid overfitting".

**Overfitting (high variance).** Memorizes noise and quirks; great on train,
bad on new data. Like a student who memorizes practice exams verbatim.

**Underfitting (high bias).** Too simple to capture patterns; bad on both train
and test. Usually a wrong-model-for-the-job problem.

**Detect** by holding out a validation set and plotting training vs validation
loss over epochs. Good fit: both decrease, small gap. Overfit: train keeps
dropping while validation stalls or rises. Choosing what to hold out is an art:
random splits leak market-wide trends in stock prediction, so slice by time.
The other tell is production underperforming the notebook.

**Capacity vs data.** High-capacity models (deep nets, billions of params) need
more data or they overfit immediately. Training a huge model end-to-end on tiny
data is a classic red flag interviewers are sensitive to. Mitigate with a
smaller model, **transfer learning** (freeze a pre-trained base, train a small
head), data augmentation, or self-/semi-supervised learning on unlabeled pools.

**Data drift** is the world changing after deployment (distinct from
overfitting). Types: covariate shift (input distribution moves), prior/label
shift (target rate moves, e.g. fraud 1% -> 3%), concept drift (the
feature-label relationship itself changes, the nastiest). Detect by monitoring
prediction distributions, feature distributions, and performance on a fixed
holdout. Remediate by retraining regularly (free to suggest), online or online-
embedding learning for fast adaptation, ensembles, or human-in-the-loop.

**Regularization** constrains the model so it learns robust patterns: L2 /
weight decay (the default, mention first), L1 (sparse, feature selection),
dropout and layer normalization for deep nets, and early stopping (free,
always). Leakage (a feature encoding the answer or the future) is harder to spot
and a rich probing area: ask whether every feature is knowable at request time.

## Sources

- `references/raw/learn/ml-system-design/core-concepts/generalization.md` -
  over/underfit diagnosis, capacity, drift types and handling, regularization.
