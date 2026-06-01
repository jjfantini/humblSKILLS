---
title: "Cassandra: Write-Optimized Wide-Column NoSQL"
context: tech
category: core
concept: cassandra
description: "Distributed, highly available, write-optimized wide-column store; design tables per query, pick a partition key for even distribution and access pattern."
tags: cassandra, wide-column, nosql, write-optimized, lsm
sources:
  - "references/raw/learn/system-design/deep-dives/cassandra.md"
last_ingested: 2026-06-01
---

## Cassandra: Write-Optimized Wide-Column NoSQL

Distributed wide-column NoSQL. Partitioned via consistent hashing, eventually consistent (tunable), last-write-wins on conflicts, horizontally scalable, extremely available and write-optimized (LSM).

**Reach for it when:** massive write throughput, huge data footprint, high availability, known query patterns (time-series, messaging, feeds).

**Incorrect (relational habits):**

```text
Normalize into many tables and JOIN across partitions at read time ->
cross-partition scans, hot partitions, terrible latency.
```

**Correct (model per query, denormalize):**

```text
One table per query pattern; denormalize freely.
Partition key for even distribution + your access pattern.
Clustering keys for sort order within a partition.
Set replication factor + tunable consistency (e.g. QUORUM).
```

Avoid cross-partition queries and unbounded partitions. If you need ad-hoc joins and transactions, this is the wrong store.

## Sources

- `references/raw/learn/system-design/deep-dives/cassandra.md` - LSM write path, per-query modeling, partition/clustering keys, tunable consistency.
