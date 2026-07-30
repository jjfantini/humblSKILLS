---
title: "Design Facebook News Feed"
context: breakdown
category: core
concept: fb-news-feed
description: "Create posts, follow users, and view a reverse-chron feed at 2B scale; solve fan-out with a hybrid precomputed feed + read merge for celebrity accounts."
tags: news-feed, fan-out, precompute, hot-key, scaling-reads
sources:
  - "references/raw/learn/system-design/problem-breakdowns/fb-news-feed.md"
last_ingested: 2026-06-01
---

## Design Facebook News Feed

**Functional:** create posts; follow users; view a reverse-chronological feed; page through it.

**Non-functional:** availability > consistency (tolerate ~1 min staleness); post/feed <500ms; 2B users; unlimited follows/followers.

**Core entities:** User, Follow (uni-directional), Post.

**Key API:**

```text
POST /posts { content }
PUT  /users/{id}/follow
GET  /feed?pageSize=&cursor={timestamp} -> { items, nextCursor }
```

**High-level design:**

> Assumed flow (diagram in source omitted; inferred): Client -> API Gateway -> Post Service / Follow Service / Feed Service -> DynamoDB (Posts, Follows with GSI, PrecomputedFeed).

Naive feed = fan-out-on-read: get follows, query each user's posts (GSI on creatorID+createdAt), merge, sort. Flag that this will not scale.

**Key deep dives:**
- **Many follows:** **fan-out-on-write** to a PrecomputedFeed table (~200 post IDs/user, ~4TB for 2B users).
- **Many followers (celebrity):** async workers off a queue; **hybrid** - skip precompute for high-follow accounts and merge their recent posts at read time.
- **Hot posts:** **redundant (replicated) post cache** so a viral post's reads spread across all N instances instead of one shard.

## Sources

- `references/raw/learn/system-design/problem-breakdowns/fb-news-feed.md` - fan-out on read vs write, hybrid feeds, hot-key caching.
