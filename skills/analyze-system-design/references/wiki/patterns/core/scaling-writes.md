---
title: "Scaling Writes"
context: patterns
category: core
concept: scaling-writes
description: "Absorb high write throughput by scaling vertically first, then sharding by a high-cardinality key, buffering with a queue, and batching writes."
tags: scaling, writes, sharding, queue, batching
sources:
  - "references/raw/learn/system-design/patterns/scaling-writes.md"
last_ingested: 2026-06-01
---

## Scaling Writes

Write-heavy systems: metrics, logs, chat, click streams, IoT.

**Incorrect (synchronous fan-in to one row):**

```text
Every click writes synchronously to a single counter row -> lock
contention and a write bottleneck under burst traffic.
```

**Correct (buffer, parallelize, batch):**

```text
1. Vertical scale first (bigger box, batching).
2. Shard by a high-cardinality key to parallelize writes.
3. Put a queue (Kafka) in front to buffer bursts and decouple producers
   from consumers.
4. Batch writes; use write-back caching where eventual durability is OK.
```

Use a write-optimized store (Cassandra, LSM-tree) when appropriate, and partition hot keys so one key does not pin one shard. The queue both smooths spikes and lets you scale consumers independently off lag.

## Sources

- `references/raw/learn/system-design/patterns/scaling-writes.md` - vertical-first, sharding, queue buffering, batching, write-optimized stores.
