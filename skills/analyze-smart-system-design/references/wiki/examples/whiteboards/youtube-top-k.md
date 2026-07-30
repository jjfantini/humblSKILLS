---
title: "Whiteboard: YouTube Top-K (Trending Views)"
context: examples
category: whiteboards
concept: youtube-top-k
description: "Reference design for precise top-K over tumbling windows: Kafka ViewEvent, Flink window aggregation, sharded Postgres, Top-K cron precompute, and Redis cache."
tags: top-k, flink, kafka, precompute, whiteboard
sources:
  - "references/raw/whiteboards/youtube_top_k.png"
  - "references/raw/learn/system-design/problem-breakdowns/top-k.md"
last_ingested: 2026-06-01
---

## Whiteboard: YouTube Top-K (Trending Views)

Caption on board: "Full Aggregates for Each Window." Shows write path (stream aggregation) and read path (precomputed cache) as separate deep dives.

**Write / aggregation path:**

```text
ViewEvent (Kafka) -> Flink Window Aggregation <-> Views DB (Postgres)
```

**Views DB schemas (from diagram):**

- **VideoViews:** VideoId, Views, Timestamp - sharded by VideoId
- **VideoViewsLast{Hour,Day,Month,AllTime}:** VideoId, Views - index on Views, sharded by VideoId

**Read path (deep dive - precompute):**

```text
Top-K Cron (reads aggregated tables) -> Cache (Redis)
Client GET /views/top-k -> API Gateway -> Top-K Service -> Cache
```

**Why each extra component:**

- **Flink:** tumbling-window aggregation collapses ~700k views/s into periodic bulk writes.
- **Shard by VideoId:** keeps hot videos localized; merge top-K per shard safely.
- **Index on Views:** fast sort/extract for cron job.
- **Cron + Redis:** sub-10ms reads; no on-the-fly global sort per request.

Lesson: the simple core is Kafka -> counters -> GET top-K. Flink, summary tables, cron, and cache are step-6 additions for write volume and read latency.

## Sources

- `references/raw/whiteboards/youtube_top_k.png` - Flink aggregation, sharded Postgres, cron precompute, Redis serve path.
- `references/raw/learn/system-design/problem-breakdowns/top-k.md` - precompute, sharding, approximation stretch prose.
