---
title: "Design Bitly (URL Shortener)"
context: breakdown
category: core
concept: bitly
description: "Read-heavy URL shortener: generate unique short codes, redirect under 100ms, and scale 1B URLs / 100M DAU with a counter, cache, and CDN."
tags: bitly, url-shortener, scaling-reads, base62, cache
sources:
  - "references/raw/learn/system-design/problem-breakdowns/bitly.md"
last_ingested: 2026-06-01
---

## Design Bitly (URL Shortener)

**Functional:** shorten a long URL (optional custom alias + expiration); redirect a short code to the original URL.

**Non-functional:** unique short codes; redirect under 100ms; 99.99% available (availability > consistency); 1B URLs, 100M DAU. Read-to-write is heavily skewed (~1000:1).

**Core entities:** Original URL, Short URL, User.

**Key API:**

```text
POST /urls { long_url, custom_alias?, expiration? } -> { short_url }
GET  /{short_code} -> 302 redirect (410 Gone if expired)
```

**High-level design:**

> Assumed flow (diagram in source omitted; inferred): Client -> API Gateway -> Write Service (generate code, persist) / Read Service (lookup -> 302). DB + Redis cache; CDN in front.

Use **302** (not 301) so you keep control over expiry and analytics. Designate the short code as primary key for an O(log n) B-tree lookup.

**Key deep dives:**
- **Unique code generation:** prefer a **counter + base62** (Redis `INCR`, atomic, no collisions) over hashing (collision retries). 1B in base62 is 6 chars.
- **Fast redirects:** cache-aside Redis (memory ~1000x faster than SSD); CDN / edge for popular codes.
- **Scaling writes:** split Read/Write services; centralize the counter in Redis with **batching** (hand out 1000 ids at a time); disjoint counter ranges per region.

## Sources

- `references/raw/learn/system-design/problem-breakdowns/bitly.md` - requirements, code-generation options, caching, counter scaling.
