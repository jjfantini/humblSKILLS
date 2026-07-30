---
title: "Design Facebook Post Search"
context: breakdown
category: core
concept: fb-post-search
description: "Keyword search over trillions of posts sorted by recency or likes, without a search engine; build an inverted index in Redis with separate sorted indexes."
tags: post-search, inverted-index, redis, sorting, scaling-reads
sources:
  - "references/raw/whiteboards/fb_post_search.png"
  - "references/raw/learn/system-design/problem-breakdowns/fb-post-search.md"
last_ingested: 2026-06-01
---

## Design Facebook Post Search

Constraint: the interviewer disallows Elasticsearch / Postgres FTS - build the index yourself.

**Functional:** create and like posts; search posts by keyword; sort results by recency or like count.

**Non-functional:** median search <500ms; high volume (write-heavy: ~10k posts/s, ~100k likes/s vs ~10k searches/s); new posts searchable <1 min; all posts discoverable; highly available. ~3.6PB total.

**Core entities:** User, Post, Like.

**Key API:**

```text
POST /posts, POST /likes
GET  /search?keyword=&sort=recency|likes
```

**High-level design (from whiteboard - Full Design):**

```text
Write: Post/Like Service -> Event Writer -> Kafka -> Like Batcher -> Ingestion Service
       -> Index (Redis): Creation Keyword->[PostIds], Likes Keyword->[PostIds]
       -> Cold Indexes (Blob Storage)
Read:  Client -> CDN -> API Gateway -> Search Service -> Search Cache (Redis) + Index lookup
       -> Query Likes -> Like Service (fresh ranking signal)
```

**Key deep dives (visible on whiteboard):**
- **No search engine:** build **inverted index** yourself (keyword -> post IDs); tokenize on ingest, not SQL `LIKE`.
- **Dual indexes:** separate **Creation** and **Likes** mappings per keyword for recency vs popularity sort.
- **Kafka ingestion:** Event Writer + scaled Ingestion Service decouple ~10k posts/s and ~100k likes/s from index writes.
- **Like Batcher:** aggregate like events before index update.
- **Hot/cold tier:** Redis for hot index; **Cold Indexes** in blob storage for older data (~3.6PB).
- **Read caching:** CDN + Search Cache (Redis); no personalization makes caching effective.
- **Fresh likes at read time:** Search Service queries Like Service during search for up-to-date ranking.
- **Phrase queries (stretch):** intersect keyword sets or index bigrams/shingles.

See also `wiki/examples/whiteboards/fb-post-search.md`.

## Sources

- `references/raw/whiteboards/fb_post_search.png` - full design with ingest/search paths, dual indexes, cold storage.
- `references/raw/learn/system-design/problem-breakdowns/fb-post-search.md` - inverted index, dual sort indexes, bigrams prose.
