---
title: "Design Tinder (Dating / Swiping)"
context: breakdown
category: core
concept: tinder
description: "Serve a geo-filtered profile stack and record swipes with consistent match detection; use Redis atomic ops, geo-indexed feeds, and a bloom filter for seen profiles."
tags: tinder, swiping, geo, consistency, bloom-filter
sources:
  - "references/raw/learn/system-design/problem-breakdowns/tinder.md"
last_ingested: 2026-06-01
---

## Design Tinder (Dating / Swiping)

**Functional:** create profile with preferences + max distance; view a stack of nearby matches; swipe yes/no; get a match notification on mutual yes.

**Non-functional:** strong consistency for swiping (mutual yes -> immediate match); 20M DAU, ~100 swipes/user/day; stack loads <300ms; never re-show swiped profiles.

**Core entities:** User, Swipe (swiping_user -> target_user), Match.

**Key API:**

```text
POST /profile { age_min, age_max, distance, interestedIn }
GET  /feed?lat=&long=&distance= -> User[]
POST /swipe/{userId} { decision }
```

**High-level design:**

> Assumed flow (diagram in source omitted; inferred): Client -> API Gateway -> Profile Service (profile DB) and Swipe Service (Cassandra, write-optimized) -> APNS/FCM for match push.

Separate swipe service + write-optimized Cassandra (4B swipes/day) partitioned by swiping_user_id.

**Key deep dives:**
- **Consistent matches:** make reciprocal swipes hit the same shard (sorted `user_pair` key) and use **Redis atomic Lua** (or Cassandra single-partition LWT) to record + check in one op.
- **Low-latency stack:** **geo-indexed** store (Elasticsearch/OpenSearch) + precomputed cached feed with short TTL; refresh when near depletion; guard against stale profiles.
- **No repeats:** client-side recent-swipe cache + **bloom filter** for users with huge swipe histories (no false negatives).

## Sources

- `references/raw/learn/system-design/problem-breakdowns/tinder.md` - swipe consistency, feed generation, seen-profile filtering.
