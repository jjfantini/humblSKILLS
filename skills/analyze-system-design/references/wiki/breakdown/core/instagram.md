---
title: "Design Instagram (Photo Sharing)"
context: breakdown
category: core
concept: instagram
description: "Create photo/video posts, follow users, and view a chronological feed at 500M DAU; solve feed latency with hybrid fan-out and serve media via S3 + CDN."
tags: instagram, feed, fan-out, media, cdn
sources:
  - "references/raw/learn/system-design/problem-breakdowns/instagram.md"
last_ingested: 2026-06-01
---

## Design Instagram (Photo Sharing)

**Functional:** create posts (photo/video + caption); follow users; view a chronological feed of followed users.

**Non-functional:** availability > consistency (eventual, up to 2 min); feed <500ms; instant media rendering; 500M DAU, 100M posts/day.

**Core entities:** User, Post, Media (S3), Follow.

**Key API:**

```text
POST /posts { media, caption } -> postId   (media via presigned URL)
POST /follows { followedId }
GET  /feed?cursor=&limit= -> Post[]
```

**High-level design:**

> Assumed flow (diagram in source omitted; inferred): Client -> API Gateway -> Post/Follow services -> DynamoDB (Posts partitioned by userId + createdAt sort, Follows with GSI); media bytes to S3 via presigned URLs, served via CDN.

Naive feed = fan-out-on-read (query follows, query each user's posts, merge) - flag it as not scaling.

**Key deep dives:**
- **Feed latency:** **hybrid fan-out** - precompute feeds (fan-out-on-write) for normal users; for celebrity accounts fall back to read-time merge (same pattern as fb-news-feed).
- **Media delivery:** presigned-URL upload to S3, **CDN** for instant global rendering; handle large video like YouTube/Dropbox (chunking).

## Sources

- `references/raw/learn/system-design/problem-breakdowns/instagram.md` - feed fan-out trade-offs, media via S3/CDN, DynamoDB indexing.
