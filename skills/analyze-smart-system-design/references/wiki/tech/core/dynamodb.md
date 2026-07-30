---
title: "DynamoDB: Managed Key-Value/Document NoSQL"
context: tech
category: core
concept: dynamodb
description: "Fully managed AWS NoSQL with single-digit-ms latency and transactions; choose a partition key and optional sort key, add GSIs for alternate access."
tags: dynamodb, nosql, managed, partition-key, gsi
sources:
  - "references/raw/learn/system-design/deep-dives/dynamodb.md"
last_ingested: 2026-06-01
---

## DynamoDB: Managed Key-Value/Document NoSQL

Fully managed AWS key-value/document NoSQL. Auto-scales, supports transactions, single-digit-ms latency. A great default NoSQL in interviews; it does almost everything.

**Reach for it when:** you want a managed, scalable store and are not barred from cloud vendor lock-in. Ask the interviewer if it is allowed.

**Incorrect (no access-pattern design):**

```text
Pick a low-cardinality partition key (e.g. country) -> hot partitions,
throttling, and no clean way to run the queries you actually need.
```

**Correct (key design from access patterns):**

```text
Partition key: even distribution + common query (e.g. user_id).
Optional sort key: range queries / sorting within a partition.
GSIs: alternate access patterns.
Strong or eventual consistency selectable per read.
```

Because it is managed and handles sharding, replication, and transactions for you, it is often the lowest-effort scalable choice when cloud is allowed.

## Sources

- `references/raw/learn/system-design/deep-dives/dynamodb.md` - partition/sort keys, GSIs, per-read consistency, managed scaling.
