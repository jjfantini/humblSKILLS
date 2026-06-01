---
title: "Whiteboard: Uber (Simplicity Bar)"
context: examples
category: whiteboards
concept: uber
description: "The simplicity bar for ride-sharing: two services over a main DB plus a location store, where a lock satisfies contention and a geo-index satisfies proximity."
tags: uber, ride-sharing, geo-index, contention, whiteboard
sources:
  - "references/raw/whiteboards/uber.png"
last_ingested: 2026-06-01
---

## Whiteboard: Uber (Simplicity Bar)

A reference design holding the simplicity bar for location-based matching.

**Incorrect (over-built):**

```text
Event-sourced everything, a streaming pipeline for fares, and global sharding
before the contention and proximity requirements force the lock and geo-index.
```

**Correct (the simple shape):**

```text
Rider/Driver clients -> API Gateway -> Ride Service + Driver Service.
A DB lock guards trip assignment (consistency, no double-booking).
A Location DB optimized for high write throughput; a geospatial index
(geohash/quadtree) finds nearby drivers. Notification Service pushes requests.
```

- **Functional:** fare estimate; request a ride; driver accepts/declines and navigates.
- **Non-functional:** 100M DAU, 10M rides/day; proximity matching; strong consistency for matching; resilient under bursts.
- **Why simple:** two services + main DB + a dedicated location store. The contention requirement justifies the lock; proximity justifies the geo-index.

Lesson: each added component traces to one named requirement.

## Sources

- `references/raw/whiteboards/uber.png` - the reference whiteboard for the Uber ride-sharing design.
