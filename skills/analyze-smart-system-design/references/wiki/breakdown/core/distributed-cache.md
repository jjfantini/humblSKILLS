---
title: "Design a Distributed Cache"
context: breakdown
category: core
concept: distributed-cache
description: "An in-memory key-value cache with TTL and LRU eviction, scaled across nodes with consistent hashing to hold 1TB at 100k req/s under 10ms."
tags: distributed-cache, lru, ttl, consistent-hashing, sharding
sources:
  - "references/raw/learn/system-design/problem-breakdowns/distributed-cache.md"
last_ingested: 2026-06-01
---

## Design a Distributed Cache

**Functional:** set/get/delete key-value pairs; configurable expiration (TTL); LRU eviction.

**Non-functional:** highly available (eventual consistency OK); <10ms get/set; scale to 1TB and 100k req/s. Out of scope: durability, strong consistency.

**Core entities:** keys, values.

**Key API:**

```text
POST /:key { value } ;  GET /:key ;  DELETE /:key
```

**High-level design (single node first):**

> Assumed flow (diagram in source omitted; inferred): Client -> cache server holding an in-memory hash table; sharded across nodes via a consistent-hashing router.

A cache is a hash table (O(1) get/set/delete).

**Key deep dives:**
- **TTL:** store `(value, expiry)`; expire lazily on read **and** with a periodic **janitor** sweep so unaccessed expired keys do not leak memory.
- **LRU eviction:** **hash table + doubly linked list** - hash for O(1) lookup, list for O(1) recency reordering; evict from the tail at capacity.
- **Scale to 1TB / 100k req/s:** shard across nodes via **consistent hashing** (minimal movement on node add/remove; virtual nodes for balance); replicate hot shards for availability and hot keys.

## Sources

- `references/raw/learn/system-design/problem-breakdowns/distributed-cache.md` - hash-table core, TTL janitor, LRU hashmap+DLL, consistent-hashing scale.
