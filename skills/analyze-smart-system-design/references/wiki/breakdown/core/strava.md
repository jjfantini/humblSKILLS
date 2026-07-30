---
title: "Design Strava (Activity Tracking)"
context: breakdown
category: core
concept: strava
description: "Record runs/rides with GPS, view your and friends' activities; do tracking on-device for offline support and sync only on completion."
tags: strava, gps, offline, client-side, activities
sources:
  - "references/raw/learn/system-design/problem-breakdowns/strava.md"
last_ingested: 2026-06-01
---

## Design Strava (Activity Tracking)

**Functional:** start/pause/stop/save runs and rides; view live activity data (route, distance, time); view your own and friends' completed activities.

**Non-functional:** availability >> consistency; works offline in remote areas; accurate live local stats; scale to 10M concurrent activities.

**Core entities:** User, Activity, Route (GPS coordinates), Friend.

**Key API:**

```text
POST  /activities { type }
PATCH /activities/{id} { state: STARTED|PAUSED|COMPLETE }
POST  /activities/{id}/routes { location }
GET   /activities?mode=USER|FRIENDS&page=
```

**High-level design:**

> Assumed flow (diagram in source omitted; inferred): Mobile client -> Activity Service -> DB (activities, routes); friends list resolved for the feed query.

Track elapsed time via a **log of status+timestamp pairs** so pauses are handled correctly. Distance from consecutive GPS points via the Haversine formula.

**Key deep dives:**
- **Offline + scale:** the key insight - since live sharing is not required, **record entirely on-device** and sync only when the activity completes / connectivity returns. This removes the 10M-concurrent-write load on the server entirely.
- **Friends feed:** bi-directional friends table (composite PK), query completed activities filtered by friend IDs.

## Sources

- `references/raw/learn/system-design/problem-breakdowns/strava.md` - status-log timing, on-device tracking, friends feed.
