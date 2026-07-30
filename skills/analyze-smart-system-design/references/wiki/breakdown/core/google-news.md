---
title: "Design Google News (Aggregator)"
context: breakdown
category: core
concept: google-news
description: "Aggregate articles from thousands of publishers into an infinite regional feed under 200ms; ingest via RSS polling, paginate with composite cursors, cache with CDC."
tags: google-news, aggregator, rss, cursor-pagination, cdc
sources:
  - "references/raw/learn/system-design/problem-breakdowns/google-news.md"
last_ingested: 2026-06-01
---

## Design Google News (Aggregator)

**Functional:** view an aggregated feed of articles from thousands of publishers; scroll infinitely; click through to the publisher's site.

**Non-functional:** availability > consistency; 100M DAU (spikes to 500M); feed loads <200ms.

**Core entities:** Article, Publisher, User (region).

**Key API:**

```text
GET /feed?cursor=&limit=&region= -> Article[]
```

**High-level design:**

> Assumed flow (diagram in source omitted; inferred): Data Collection Service polls publisher RSS -> stores articles + thumbnails (S3) in DB; Feed Service serves regional feeds behind API Gateway.

Separate the write-heavy collector from the read-heavy Feed Service. Download thumbnails into own object storage rather than hotlinking publishers.

**Key deep dives:**
- **Pagination:** replace offset (drift, duplicates) with **cursor pagination** - composite `(published_at, article_id)` cursor, or **monotonic/ULID article IDs** for a single-value cursor.
- **Low latency (<200ms):** cache regional feeds in **Redis sorted sets**; upgrade from TTL (thundering herd, staleness) to **CDC-driven precomputed feeds** (ZADD on publish, ZREMRANGEBYRANK to cap size).
- **Freshness:** poll publishers on a cadence; push new articles into caches within ~30 min.

## Sources

- `references/raw/learn/system-design/problem-breakdowns/google-news.md` - RSS ingestion, cursor pagination, CDC-cached feeds.
