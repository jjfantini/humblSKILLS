---
title: "Numbers to Know (2026)"
context: concepts
category: theory
concept: numbers-to-know
description: "Order-of-magnitude capacity figures for cache, database, app server, and queue so inline math can decide one node vs sharding."
tags: numbers, capacity, throughput, latency
sources:
  - "references/raw/learn/system-design/core-concepts/numbers-to-know.md"
last_ingested: 2026-06-01
---

## Numbers to Know (2026)

Order-of-magnitude figures for sanity checks. Use them only when a number changes a decision (one node vs sharding, sync vs async), per the `estimation` rule.

**Incorrect (stale 2015 assumptions):**

```text
"A server can only hold a few thousand connections, RAM is tiny,
so we must shard immediately." Hardware moved on; this over-builds.
```

**Correct (current ladder):**

```text
- Cache (Redis): ~1ms, 100k+ ops/sec, memory-bound up to ~1TB.
  Scale at hit rate < 80% or memory > 80%.
- Database: up to ~50k TPS, sub-5ms cached reads, 64TiB+ storage.
  Scale at writes > 10k TPS or uncached reads > 5ms.
- App server: 100k+ concurrent connections, 8-64 cores, 64-512GB RAM.
  CPU is usually the first bottleneck, not memory.
- Message queue (Kafka): up to ~1M msgs/sec/broker, 1-5ms, 50TB/broker.
```

Key implications: a Kafka queue's sub-5ms latency means it can sit in a synchronous request path. Containers start in 30-60s, so aggressive autoscaling beats over-provisioning. A "single instance" is not a single point of failure if you run a primary + replicas; replication (availability) is separate from sharding (scale).

## Sources

- `references/raw/learn/system-design/core-concepts/numbers-to-know.md` - the capacity ladder and the scale-at thresholds.
