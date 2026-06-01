---
title: "Whiteboard: Metrics Monitoring and Alerting"
context: examples
category: whiteboards
concept: metrics-monitoring
description: "Final design for 5M metrics/s: agent ingestion through Kafka to a time-series DB with rollups, query cache, dual alert paths (batch + Flink realtime), and multi-channel notifications."
tags: monitoring, time-series, kafka, flink, alerting, whiteboard
sources:
  - "references/raw/whiteboards/metrics_monitoring.png"
  - "references/raw/learn/system-design/problem-breakdowns/metrics-monitoring.md"
last_ingested: 2026-06-01
---

## Whiteboard: Metrics Monitoring and Alerting

Labeled "Final Design" in the reference whiteboard. Shows a complete metrics platform with separate ingest, query, and alert paths - more components than step 5 needs, but each maps to a deep dive.

**Ingestion path:**

```text
Servers (agents) -> Ingestion Service <-> Card Tracker (Redis)
                 -> Kafka -> Ingestion Consumer -> Time-Series DB (with Rollups loop)
```

**Query path:**

```text
Dashboard <-> Query Service <-> Time-Series DB
                         <-> Redis (precomputed / partial query results)
```

**Alerting paths (two deep dives):**

```text
Standard:  Ingestion Service <-> Alert Service <-> Policy DB
Realtime:  Kafka -> Realtime Alerts (Flink) <-> Policy DB
Both fire -> Noti Service -> Noti DB -> Slack / Pager / SMS
```

**Key design choices visible on the board:**

- **Card Tracker (Redis):** dedup or stream-state at ingest (same Redis cluster reused elsewhere).
- **Rollups:** TSDB self-loop aggregates raw points into coarser windows for retention and faster queries.
- **Dual alerting:** batch/threshold path via Alert Service plus low-latency Flink stream evaluation on Kafka.
- **Query cache:** Redis holds precomputed or partial results so dashboards stay in seconds, not minutes.

Lesson: start with agents -> queue -> TSDB -> query service. Add Flink realtime alerts and query cache only when sub-minute alert latency or heavy dashboard load forces them.

## Sources

- `references/raw/whiteboards/metrics_monitoring.png` - final design with ingest, query, dual alert, and notification paths.
- `references/raw/learn/system-design/problem-breakdowns/metrics-monitoring.md` - cardinality, agent buffering, rollups prose.
