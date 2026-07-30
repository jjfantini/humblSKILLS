---
title: "Step 3: Data and Features"
context: delivery
category: framework
concept: data-features
description: "The ~10-minute data step: bucket training data as supervised/semi/unsupervised, enumerate signal sources (not a feature dump), and choose representations."
tags: data, features, training-data, representation, signal-sources
sources:
  - "references/raw/learn/ml-system-design/in-a-hurry/delivery.md"
last_ingested: 2026-06-01
---

## Step 3: Data and Features

About 10 minutes. Walk raw data -> features -> representation; jumping back and
forth is fine as new ideas surface.

**Training data.** Name your sources and ask what exists. Bucket as
**supervised / semi-supervised / unsupervised**. Most candidates only reach for
the supervised bucket; great solutions exploit the orders-of-magnitude larger
semi- and unsupervised pools. Design labels (direct labels or proxy signals
like clicks). Do not assume perfect data: collection and prep dominate real ML
work, and cold-start and bias concerns live here too.

**Features.** Enumerate **signal sources, not a feature dump**. There is a
nearly infinite list of possible features for most problems; rattling them off
shows no insight and burns time. Use domain knowledge and temporal aspects, and
prioritize by predictive power x implementation feasibility. (The five recurring
sources are in `concepts/core/feature-engineering`.)

**Representation.** Embeddings for high-cardinality categoricals and
users/items, one-hot for small categoricals, normalization for numerics,
pre-trained models for text/image, explicit missing-value handling. Distinguish
online/queried (fresh) from offline/batch features held in a feature store.

**Incorrect (laundry list):**

```text
"Features: age, country, device, last login, post count, friend count, photo
count, bio length, ... (continues for 3 minutes)"
```

**Correct (sources, then representative signals):**

```text
"Content signals (post text, image), actor signals (author history embedding +
account age), behavioral signals (negative-reaction rate, share velocity). Let
me populate content first and earmark behavioral for depth if you want it."
```

- Green: creative use of semi/unsupervised data; impactful features with a
  hypothesis; thoughtful representation choices.
- Red: a laundry-list feature dump; ambiguous representation.

## Sources

- `references/raw/learn/ml-system-design/in-a-hurry/delivery.md` - the
  data/features walkthrough, buckets, and green/red flags.
