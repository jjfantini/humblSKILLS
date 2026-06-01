---
title: "Design Facebook Live Comments"
context: breakdown
category: core
concept: fb-live-comments
description: "Broadcast comments on a live video to millions of viewers under 200ms using SSE, cursor pagination for history, and partitioned pub/sub with viewer co-location."
tags: live-comments, sse, pubsub, cursor-pagination, realtime
sources:
  - "references/raw/learn/system-design/problem-breakdowns/fb-live-comments.md"
last_ingested: 2026-06-01
---

## Design Facebook Live Comments

**Functional:** post comments on a live video; see new comments in near-real-time while watching; see comments made before joining.

**Non-functional:** millions of concurrent videos, thousands of comments/s per video; availability > consistency (eventual OK); <200ms end-to-end broadcast.

**Core entities:** User, Live Video, Comment.

**Key API:**

```text
POST /comments/{liveVideoId} { message }
GET  /comments/{liveVideoId}?cursor={lastId}&pageSize=&sort=desc
```

**High-level design:**

> Assumed flow (diagram in source omitted; inferred): Commenter -> Comment Management Service -> DynamoDB; viewers <-SSE- Realtime Messaging Servers fed by pub/sub.

Use **cursor pagination** (not offset - unstable under fast inserts) for history.

**Key deep dives:**
- **Realtime broadcast:** **SSE** (one-way) over WebSockets - read/write ratio is hugely read-skewed, so a bidirectional socket per viewer is wasteful.
- **Scale to millions of viewers:** separate Realtime Messaging Servers; viewers of one video span servers -> **partitioned pub/sub** (`hash(liveVideoId) % N` channels) with **L7 consistent-hashing routing** to co-locate viewers, or a **dispatcher service** that routes comments to the right servers. Redis pub/sub fits (fire-and-forget OK since comments are persisted).

## Sources

- `references/raw/learn/system-design/problem-breakdowns/fb-live-comments.md` - SSE choice, cursor pagination, partitioned pub/sub, co-location.
