---
title: "Design Top-K (Trending Videos)"
context: breakdown
category: core
concept: top-k
description: "Precise top-K most-viewed videos over tumbling windows from a massive view stream; aggregate with Flink, shard by video ID, and serve precomputed results from cache."
tags: top-k, streaming, flink, kafka, precompute
sources:
  - "references/raw/whiteboards/youtube_top_k.png"
  - "references/raw/learn/system-design/problem-breakdowns/top-k.md"
last_ingested: 2026-06-01
---

## Design Top-K (Trending Videos)

**Functional:** query top K videos for all-time and tumbling windows of 1 hour / 1 day / 1 month (max 1k results). Out of scope: arbitrary time periods.

**Non-functional:** <=1 min view-to-tabulation delay; precise (no approximation, initially); responses in tens of ms; massive view volume (~700k views/s) and ~3.6B videos.

**Core entities:** Video, View, Time Window.

**Key API:**

```text
GET /views/top-k?window={WINDOW}&k={K} -> [{ videoId, views }]
```

**High-level design (from whiteboard):**

```text
Write: ViewEvent (Kafka) -> Flink Window Aggregation <-> Views DB (Postgres, sharded by VideoId)
Read:  Top-K Cron -> Cache (Redis) ; Client GET /views/top-k -> Top-K Service -> Cache
```

Summary tables: `VideoViewsLast{Hour,Day,Month,AllTime}` with **index on Views**, sharded by VideoId.

**Key deep dives (visible on whiteboard):**
- **Flink tumbling windows:** aggregate view stream; bulk writes cut shard load from ~700k tps.
- **Shard by VideoId:** Kafka partition -> consumer -> DB shard; merge per-shard top-K safely.
- **Precompute on cron:** 1-min grace lets cron refresh Redis; reads in tens of ms, not global sort per request.
- **Full aggregates per window:** maintain complete counts per tumbling window (Hour/Day/Month/AllTime).
- **Approximation (stretch):** count-min sketch / heap if precision can be relaxed.

See also `wiki/examples/whiteboards/youtube-top-k.md`.

## Sources

- `references/raw/whiteboards/youtube_top_k.png` - Flink, sharded Postgres, cron precompute, Redis serve.
- `references/raw/learn/system-design/problem-breakdowns/top-k.md` - precompute, sharding, approximation stretch prose.
