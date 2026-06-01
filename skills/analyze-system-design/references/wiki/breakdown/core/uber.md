---
title: "Design Uber (Ride Sharing)"
context: breakdown
category: core
concept: uber
description: "Fare estimate, ride request, and proximity driver matching; absorb 2M location writes/s in Redis geo, and prevent double-assignment with a distributed lock + TTL."
tags: uber, ride-sharing, geospatial, redis, matching
sources:
  - "references/raw/learn/system-design/problem-breakdowns/uber.md"
last_ingested: 2026-06-01
---

## Design Uber (Ride Sharing)

**Functional:** input pickup + destination, get fare estimate; request a ride at the estimate; match with a nearby available driver; driver accepts/declines and navigates.

**Non-functional:** low-latency matching (<1 min to match or fail); strong consistency (no driver assigned two rides at once); high throughput including bursts (100k requests from one location).

**Core entities:** Rider, Driver, Fare, Ride, Location.

**Key API:**

```text
POST  /fare { pickup, destination } -> Fare
POST  /rides { fareId } -> Ride
POST  /drivers/location { lat, long }
PATCH /rides/{rideId} { accept | deny }
```

**High-level design:**

> Assumed flow (diagram in source omitted; inferred): Rider/Driver clients -> API Gateway -> Ride Service (fare via third-party maps), Location Service, Ride Matching Service; APNS/FCM notifications.

**Key deep dives:**
- **Location writes + proximity (~2M writes/s):** **Redis geospatial** (`GEOADD`/`GEOSEARCH`, geohash) for real-time updates and nearby queries; overwrites keep latest position. (PostGIS/quadtree is the batch alternative.)
- **Reduce update load:** **adaptive client-side update intervals** based on speed/direction.
- **No double-assignment:** **distributed lock with TTL** (like Ticketmaster) - offer to one driver for ~10s, then move to the next; DB consistency as safety net.

## Sources

- `references/raw/learn/system-design/problem-breakdowns/uber.md` - geospatial store, adaptive updates, matching consistency.
