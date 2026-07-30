---
title: "Feature Engineering: The Five Signal Sources"
context: concepts
category: core
concept: feature-engineering
description: "Structure the feature discussion around five recurring signal sources instead of dumping features, then choose encodings and flag the pitfalls that disqualify candidates."
tags: features, signal-sources, encoding, leakage, drift
sources:
  - "references/raw/learn/ml-system-design/core-concepts/feature-engineering.md"
last_ingested: 2026-06-01
---

## Feature Engineering: The Five Signal Sources

The feature discussion is the biggest tarpit in the interview. Candidates dump
30 features and run out the clock, or freeze on domain detail. The fix is
structure: enumerate **signal sources, not low-level features**, sketch the
sources, populate each with 2-3 representative signals, and stop cleanly when
the interviewer moves on. Modern transformer/DLRM/multimodal models consume raw
text, images, and event sequences directly, so hand-crafted features are a
shrinking minority. But which data sources you put in front of the model is a
design decision the model cannot make for itself, which is why this is still
~30% of the interview.

**The five recurring sources** (same set across recs, harmful content, bots):

1. **Content / item** - what the thing is (post text, thumbnail, audio,
   metadata). Highest signal when the answer is "what does this look like"
   (dominates harmful content); often weak for recommendations.
2. **Actor / creator / user** - who is involved (profile, history, reputation,
   account age). Usually a slow upstream embedding + a few explicit attributes.
3. **Behavioral / engagement** - what happened (last 100 watches, view velocity,
   posting cadence, negative-reaction rate). Highest value for recs; where most
   drift, leakage, and train/serve skew pain lives.
4. **Network / graph** - relationships (follows, co-engagement, shared IPs). Fed
   as a GNN embedding. Heavy in adversarial and social domains.
5. **Context / request** - time of day, device, locale, current session. Cheap,
   real-time, predictive, and the most forgotten; naming it is a cheap signal.

**Encoding by shape:** raw text/images -> transformer/ViT encoders; sparse IDs
-> embedding tables (hashing trick for high cardinality); sequences -> mean
aggregation (light ranker) or sequence-as-tokens (heavy ranker); numeric scalars
-> log-scale power laws, bucket-then-embed, Bayesian smoothing for low-
denominator rates, exponential decay + multiple time windows for recency.

**Pitfalls that disqualify:** leakage (a feature unavailable at serve time);
cold start (sentinel/missing handling, not naive imputation); feedback loops
(name it, fix with exploration or inverse-propensity weighting); adversarial
robustness (content signals are easy to fool, temporal/network signals harder);
drift (monitor distributions, drop corrupted features).

**Most powerful sentence:** "There is more I would cover on behavioral signals
but I want to keep moving; pull me back if useful." Earmark depth, do not burn
time.

## Sources

- `references/raw/learn/ml-system-design/core-concepts/feature-engineering.md` -
  the five sources, encoding decision table, and the five pitfalls.
