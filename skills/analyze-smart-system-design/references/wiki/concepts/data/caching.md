---
title: "Caching Strategies and Pitfalls"
context: concepts
category: data
concept: caching
description: "A fast in-memory layer (usually Redis) in front of slower storage; default cache-aside with TTL, know the write strategies, and watch hot keys and stampedes."
tags: caching, cache-aside, ttl, eviction, hot-key
sources:
  - "references/raw/learn/system-design/core-concepts/caching.md"
last_ingested: 2026-06-01
---

## Caching Strategies and Pitfalls

A fast in-memory layer (usually Redis) in front of slower storage. Introduce it in deep dives to satisfy a latency or read-scale requirement, not in the first sketch.

**Incorrect (cache with no TTL or invalidation):**

```text
Read DB, store in cache forever. Data changes, cache never updates ->
users see stale results indefinitely.
```

**Correct (cache-aside + TTL):**

```text
1. App reads cache.
2. Miss -> read DB, populate cache with a TTL.
3. Hit -> serve from cache.
```

Write strategies to know: **write-through** (writes go through cache to DB synchronously; fresh reads, slower writes), **write-behind** (async flush; fast writes, risk of loss), **read-through** (cache fetches on miss; CDNs work this way).

Eviction: **LRU** default, LFU for steadily-popular keys, FIFO rarely.

Watch the **hot-key** problem (one key overwhelms a node) and **cache invalidation / stampede** (many misses hammer the DB at once; mitigate with request coalescing or staggered TTLs).

## Sources

- `references/raw/learn/system-design/core-concepts/caching.md` - cache-aside default, write strategies, eviction policies, hot-key and stampede.
