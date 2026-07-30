---
title: "Scaling Reads"
context: patterns
category: core
concept: scaling-reads
description: "Serve high read volume cheaply by layering indexes, cache, read replicas, CDN, and denormalization in order of increasing complexity."
tags: scaling, reads, cache, replicas, cdn, denormalization
sources:
  - "references/raw/learn/system-design/patterns/scaling-reads.md"
last_ingested: 2026-06-01
---

## Scaling Reads

The common case: read-heavy systems (feeds, catalogs, profiles). Layer techniques in order of simplicity, stopping when the requirement is met.

**Incorrect (jump to the heaviest tool):**

```text
Reads are slow -> immediately shard and run fan-out-on-write before
even adding an index or a cache. Massive complexity, premature.
```

**Correct (climb the ladder):**

```text
1. Add an index for the query.
2. Cache hot reads (cache-aside Redis).
3. Read replicas to spread load (accept replica lag -> eventual consistency).
4. CDN for static / cacheable content at the edge.
5. Denormalize / precompute (materialized views, fan-out-on-write) to
   avoid expensive joins at read time.
```

Each step trades freshness or write cost for read speed. Name the requirement that justifies each addition; do not add a CDN or fan-out without a read-latency or read-scale requirement demanding it.

## Sources

- `references/raw/learn/system-design/patterns/scaling-reads.md` - the index/cache/replica/CDN/denormalize ladder.
