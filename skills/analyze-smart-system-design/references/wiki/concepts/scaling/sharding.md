---
title: "Sharding: Horizontal Partitioning"
context: concepts
category: scaling
concept: sharding
description: "Split data across machines only when one tuned instance truly cannot cope; pick a high-cardinality shard key that matches query patterns."
tags: sharding, partitioning, shard-key, horizontal-scale
sources:
  - "references/raw/learn/system-design/core-concepts/sharding.md"
last_ingested: 2026-06-01
---

## Sharding: Horizontal Partitioning

Splitting data across multiple machines to scale beyond one node. (Partitioning is within one instance; sharding is across machines. Do not get hung up on the wording.)

**Incorrect (sharding too early):**

```text
"100K users, so let's shard across 10 nodes." A single primary +
read replicas handles millions of users and terabytes. Premature.
```

**Correct (shard only when forced, key chosen well):**

```text
Do the math: one well-tuned instance cannot hold the data or throughput.
Shard key = user_id (high cardinality, even, matches access pattern).
Bad keys: created_at (all new writes hit one shard), country/boolean (low cardinality).
```

Strategies:
- **Range-based** - simple, scan-friendly, risks hot shards.
- **Hash-based** - even distribution, but kills range queries.
- **Directory / lookup** - flexible, costs an extra hop.

Note that managed stores (DynamoDB, Cassandra) shard for you under the hood, so often the right answer is "pick a store that handles it."

## Sources

- `references/raw/learn/system-design/core-concepts/sharding.md` - when to shard, shard-key selection, the three strategies.
