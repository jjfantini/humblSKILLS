---
title: "Design an Ad Click Aggregator"
context: breakdown
category: core
concept: ad-click-aggregator
description: "Track ad clicks via server-side redirect and serve near-real-time metrics; stream into Flink, store in an OLAP DB, and reconcile with a batch layer."
tags: ad-click, streaming, flink, olap, scaling-writes
sources:
  - "references/raw/learn/system-design/problem-breakdowns/ad-click-aggregator.md"
last_ingested: 2026-06-01
---

## Design an Ad Click Aggregator

**Functional:** click an ad and redirect to advertiser; advertisers query click metrics over time at >=1 min granularity.

**Non-functional:** 10k clicks/s peak; sub-second analytics queries; fault-tolerant and lossless; as real-time as possible; idempotent (no double-counting).

**Core entities / interface:** input = ad click events; output = aggregated metrics.

**Data flow:** click -> track + store -> redirect -> advertiser queries aggregates.

**High-level design:**

> Assumed flow (diagram in source omitted; inferred): Client -> Click Processor (302 redirect) -> Kafka/Kinesis -> Flink aggregation -> OLAP DB; advertisers query the OLAP DB.

Use a **server-side 302 redirect** (not client-side) so every click is tracked.

**Key deep dives:**
- **Real-time analytics:** stream into **Kafka -> Flink** (event-time windows, watermarks) -> **OLAP DB** (columnar, high-cardinality), not batch Spark alone.
- **Scale to 10k/s:** shard stream/processor by AdId; mitigate **hot shards** (viral ad) by appending a random suffix to the key, stripped on write.
- **No data loss:** stream retention + replay; reconsider checkpointing given tiny windows.
- **Correctness:** dump raw events to S3 and run a periodic **batch reconciliation** (lambda architecture).
- **Idempotency:** dedup clicks with an impression/click ID.

## Sources

- `references/raw/learn/system-design/problem-breakdowns/ad-click-aggregator.md` - stream vs batch, OLAP, hot shards, reconciliation.
