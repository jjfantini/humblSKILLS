---
title: "Design a Distributed Rate Limiter"
context: breakdown
category: core
concept: distributed-rate-limiter
description: "Limit requests per client at the API gateway using a token bucket in Redis with atomic Lua, then shard Redis by client key to handle 1M req/s."
tags: rate-limiter, token-bucket, redis, sharding, contention
sources:
  - "references/raw/learn/system-design/problem-breakdowns/distributed-rate-limiter.md"
last_ingested: 2026-06-01
---

## Design a Distributed Rate Limiter

**Functional:** identify clients by user ID / IP / API key; limit requests per configurable rules; reject excess with HTTP 429 + helpful headers.

**Non-functional:** <10ms overhead per check; highly available (eventual consistency OK); 1M req/s across 100M DAU.

**Core entities:** Rule, Client, Request.

**Key interface:**

```text
isRequestAllowed(clientId, ruleId) -> { passes, remaining, resetTime }
429 + X-RateLimit-Limit / -Remaining / -Reset / Retry-After
```

**High-level design:**

> Assumed flow (diagram in source omitted; inferred): Client -> API Gateway (rate-limit check against Redis) -> app servers; gateway returns 429 on limit.

Place the limiter at the **API gateway** (edge). Centralized state in **Redis** so all gateways share counters.

**Key deep dives:**
- **Algorithm:** **token bucket** (handles bursts + steady rate, two values per client) over fixed-window (boundary bursts), sliding-log (memory-heavy), or sliding-counter.
- **Race conditions:** the read-calculate-update must be one atomic op -> **Redis Lua script** (MULTI/EXEC alone leaves the HMGET read outside the transaction).
- **Scale to 1M req/s:** one Redis does ~100k ops/s, so **shard via consistent hashing** on the client key (each client always hits one shard); fail-open vs fail-closed decision on Redis outage.

## Sources

- `references/raw/learn/system-design/problem-breakdowns/distributed-rate-limiter.md` - placement, algorithms, atomic Lua, Redis sharding.
