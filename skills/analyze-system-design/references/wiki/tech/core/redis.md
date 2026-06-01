---
title: "Redis: In-Memory Data-Structure Store"
context: tech
category: core
concept: redis
description: "Fast single-threaded data-structure store for caching, distributed locks, rate limiting, leaderboards, geo, and pub/sub; not durable by default."
tags: redis, cache, lock, leaderboard, rate-limit
sources:
  - "references/raw/learn/system-design/deep-dives/redis.md"
last_ingested: 2026-06-01
---

## Redis: In-Memory Data-Structure Store

In-memory, single-threaded store of strings, hashes, lists, sets, sorted sets, streams, geo, and bloom. 100k+ ops/sec, microsecond reads, versatile. Not durable by default.

**Reach for it when:** caching, distributed locks, rate limiting, leaderboards, proximity search, pub/sub, work queues.

**Incorrect (durable event log on plain pub/sub):**

```text
Use Redis Pub/Sub as the source of truth for events -> it is at-most-once;
a disconnected subscriber loses messages permanently.
```

**Correct (match the structure to the job):**

```text
Cache: cache-aside with TTL.
Distributed lock: INCR + EXPIRE (or Redlock + fencing tokens).
Leaderboard: sorted sets (ZADD / ZRANGE).
Rate limit: fixed-window INCR + EXPIRE (sliding window via sorted set + Lua).
Proximity: GEOADD / GEOSEARCH.
Durable fan-out: Redis Streams or Kafka, not Pub/Sub.
```

Scaling is how you structure keys; watch the **hot-key** problem and mitigate with client-side caching, key replication, or read replicas.

## Sources

- `references/raw/learn/system-design/deep-dives/redis.md` - data structures, use cases, the durability caveat, hot-key mitigation.
