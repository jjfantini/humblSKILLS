---
title: "Flink: Stateful Stream Processing"
context: tech
category: core
concept: flink
description: "Distributed dataflow engine for real-time aggregation with exactly-once state and windowing, typically sourced from Kafka; justify it over batch."
tags: flink, stream-processing, windowing, exactly-once, kafka
sources:
  - "references/raw/learn/system-design/deep-dives/flink.md"
last_ingested: 2026-06-01
---

## Flink: Stateful Stream Processing

Distributed stream-processing dataflow engine: stateful operators over a dataflow graph, exactly-once state, windowing.

**Reach for it when:** real-time aggregation/analytics over streams (ad-click counting, top-K, fraud, metrics) where per-event processing with managed state and scaling matters.

**Incorrect (stream when batch would do):**

```text
Use Flink to compute a daily report that has no latency requirement ->
you took on the hardest class of system (stateful streaming) for nothing.
```

**Correct (Kafka source, keyed windows):**

```text
Source from Kafka (partitions map to Flink parallelism).
Apply keyed windows / aggregations; Flink manages distributed state + rescaling.
If the problem reduces to batch, use Spark/batch instead.
```

Stream processing is genuinely hard, so justify it: name the real-time aggregation requirement that batch cannot meet.

## Sources

- `references/raw/learn/system-design/deep-dives/flink.md` - dataflow model, exactly-once state, Kafka sourcing, stream-vs-batch caution.
