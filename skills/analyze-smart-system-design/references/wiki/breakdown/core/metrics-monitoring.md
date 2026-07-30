---
title: "Design Metrics Monitoring and Alerting"
context: breakdown
category: core
concept: metrics-monitoring
description: "Ingest 5M metrics/s, query dashboards in seconds, and fire alerts in under a minute; agent buffering + Kafka into a time-series DB, taming cardinality explosion."
tags: monitoring, time-series, kafka, cardinality, alerting
sources:
  - "references/raw/whiteboards/metrics_monitoring.png"
  - "references/raw/learn/system-design/problem-breakdowns/metrics-monitoring.md"
last_ingested: 2026-06-01
---

## Design Metrics Monitoring and Alerting

**Functional:** ingest metrics from services; query/visualize on dashboards (filters, aggregations, time ranges); define alert rules over time windows; notify on alerts (email/Slack/PagerDuty).

**Non-functional:** 5M metrics/s from 500k servers (~1GB/s); dashboard queries in seconds; alerts fire <1 min; highly available (eventual OK for dashboards); handle late/out-of-order data.

**Core entities:** Label, Metric, Series (metric+label combo - the scaling unit), Alert Rule, Dashboard.

**Key API:**

```text
POST /metrics/ingest { metrics[] }   (batched, protobuf at scale)
GET  /metrics/query?query=<PromQL>&start=&end=&step=
POST /alerts/rules { query, for, notifications }
```

**High-level design (from whiteboard - Final Design):**

```text
Ingest:  Servers (agents) -> Ingestion Service <-> Card Tracker (Redis)
                              -> Kafka -> Ingestion Consumer -> Time-Series DB (Rollups)
Query:   Dashboard <-> Query Service <-> TSDB + Redis (precomputed/partial results)
Alert:   Alert Service <-> Policy DB (standard) ; Flink on Kafka (realtime)
         Both -> Noti Service -> Noti DB -> Slack / Pager / SMS
```

**Key deep dives (visible on whiteboard):**
- **Ingestion:** agents buffer/batch; Card Tracker (Redis) for dedup or stream state; Kafka for backpressure.
- **Storage:** TSDB with **Rollups** loop (raw -> 1-min -> 1-hr retention tiers); not Postgres.
- **Query cache:** Redis holds precomputed or partial query results for dashboard speed.
- **Dual alerting:** batch Alert Service on ingest path plus **Flink realtime** stream on Kafka; both fire Noti Service.
- **Cardinality:** keep labels low-cardinality; central scaling problem.

See also `wiki/examples/whiteboards/metrics-monitoring.md`.

## Sources

- `references/raw/whiteboards/metrics_monitoring.png` - final design with ingest, query cache, dual alert paths.
- `references/raw/learn/system-design/problem-breakdowns/metrics-monitoring.md` - agent buffering, cardinality, rollups prose.
