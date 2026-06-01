---
title: "Whiteboard: Facebook Post Search (Full Design)"
context: examples
category: whiteboards
concept: fb-post-search
description: "Full design for keyword post search without Elasticsearch: Kafka ingestion, dual Redis inverted indexes (creation + likes), Like Batcher, cold blob storage, CDN + search cache."
tags: post-search, inverted-index, redis, kafka, whiteboard
sources:
  - "references/raw/whiteboards/fb_post_search.png"
  - "references/raw/learn/system-design/problem-breakdowns/fb-post-search.md"
last_ingested: 2026-06-01
---

## Whiteboard: Facebook Post Search (Full Design)

Labeled "Full Design" on the board. Separate write path (index ingestion) and read path (search). No Elasticsearch - build the inverted index yourself.

**Ingestion path (write):**

```text
Post Service (create Posts) + Like Service (create Likes)
  -> Load Balancer -> Event Writer -> Kafka
  -> Like Batcher (aggregates like events)
  -> Ingestion Service (scaled) -> Index (Redis)
       Creation: Keyword -> [PostIds]
       Likes:    Keyword -> [PostIds]
  -> Cold Indexes (Blob Storage) for older/unpopular data
```

**Search path (read):**

```text
Client -> CDN -> API Gateway -> Search Service
  -> Search Cache (Redis) for frequent queries
  -> Index (Redis) keyword lookup -> PostIds
  -> Query Likes -> Like Service (fresh like counts for ranking)
```

**Deep-dive components on the board:**

- **Kafka + Event Writer:** decouple high-volume post/like writes from index updates; ingestion scales horizontally.
- **Like Batcher:** collapse many like events before index update (write amplification control).
- **Dual inverted indexes in Redis:** separate **Creation** (recency sort) and **Likes** (popularity sort) mappings per keyword.
- **Cold Indexes (blob storage):** tier older index data off hot Redis (~3.6PB total constraint).
- **Search Cache + CDN:** no personalization makes result caching effective (<500ms median, <1 min freshness).
- **Read-path Like Service call:** side-fetch fresh like counts at query time so ranking stays current even if batch ingestion lags slightly.

Lesson: step-5 core is Post/Like services + Ingestion + inverted index + Search Service. Kafka, Like Batcher, cold tier, CDN, and dual caches are step-6 answers to write volume, storage cost, and read latency.

## Sources

- `references/raw/whiteboards/fb_post_search.png` - full design with ingest and search paths, dual indexes, cold storage.
- `references/raw/learn/system-design/problem-breakdowns/fb-post-search.md` - no-Elasticsearch constraint, bigrams, phrase queries prose.
