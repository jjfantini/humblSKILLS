---
title: "Time-Series Databases"
context: tech
category: advanced
concept: time-series-databases
description: "Write-optimized stores for timestamped metrics using LSM and time-partitioned chunks; watch cardinality and stretch Postgres/Dynamo first."
tags: tsdb, metrics, lsm, downsampling, cardinality
sources:
  - "references/raw/learn/system-design/deep-dives/time-series-databases.md"
last_ingested: 2026-06-01
---

## Time-Series Databases

Write-optimized stores for timestamped metrics (InfluxDB, TimescaleDB, Prometheus): append-only + LSM, columnar compression, time-partitioned chunks.

**Reach for it when:** metrics/monitoring, IoT, high-volume append-only timestamped data (100k servers x 5 metrics / 10s = billions of points/day that crush a single Postgres).

**Incorrect (high-cardinality tags):**

```text
Tag every point with user_id and request_id -> cardinality explosion
destroys all the TSDB benefits.
```

**Correct (low-cardinality tags, downsample):**

```text
Append-only + LSM: random writes become sequential SSTable flushes -> huge throughput.
Time-based partitioning: retention = drop old partitions (no DELETE scans).
Only low-cardinality values (host, region) are tags; high-cardinality go in fields.
Downsample / roll up old data (5-min averages) for cheap historical reads.
```

Do not reach for a TSDB just because data has timestamps; stretch Postgres/DynamoDB first. Cross-series sort/aggregate (Top-K) can actually be worse on a TSDB. Knowing a store's data assumptions before proposing it is a staff+ signal.

## Sources

- `references/raw/learn/system-design/deep-dives/time-series-databases.md` - LSM write path, partitioning, downsampling, cardinality caveat.
