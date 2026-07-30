---
title: "Kafka: Durable Event-Streaming Log"
context: tech
category: core
concept: kafka
description: "Distributed durable log of topics split into ordered partitions; buffer write bursts and decouple services, designing idempotent at-least-once consumers."
tags: kafka, queue, event-log, partitions, consumer-group
sources:
  - "references/raw/learn/system-design/deep-dives/kafka.md"
last_ingested: 2026-06-01
---

## Kafka: Durable Event-Streaming Log

Distributed, durable event-streaming log. Topics (logical) split into partitions (physical, ordered, append-only) across brokers. Up to ~1M msgs/sec/broker, 1-5ms latency, weeks-to-months retention.

**Reach for it when:** message queue, buffering write bursts, decoupling services, event sourcing, stream-processing source, smoothing spikes.

**Incorrect (assume exactly-once, ignore ordering):**

```text
Consumer assumes each message arrives once and processes payments
non-idempotently -> a redelivery double-charges the customer.
```

**Correct (idempotent consumers, partition by key):**

```text
Producers write to topics; partitions give parallelism + per-partition order.
Consumer groups assign each partition to one consumer (default at-least-once).
Partition by a key (e.g. user_id) to preserve per-key ordering.
Design consumers to be idempotent; tune replication + acks for durability.
```

Sub-5ms latency means Kafka can sit in a synchronous request path as a buffer, not only in async pipelines.

## Sources

- `references/raw/learn/system-design/deep-dives/kafka.md` - topics/partitions, consumer groups, at-least-once, partition-by-key ordering.
