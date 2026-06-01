---
title: "Real-Time Updates"
context: patterns
category: core
concept: realtime-updates
description: "Push server-side changes to clients with low latency across two hops: client connection (polling -> WebSockets) and server-side fan-out (pub/sub)."
tags: realtime, websocket, sse, polling, pubsub
sources:
  - "references/raw/learn/system-design/patterns/realtime-updates.md"
last_ingested: 2026-06-01
---

## Real-Time Updates

Two independent hops: (1) server -> client connection, (2) source -> server fan-out. Chat, live feeds, collaborative editing, notifications, dashboards, presence.

**Incorrect (reach for WebSockets reflexively):**

```text
"Realtime" -> immediately add WebSockets everywhere, including a
status page that updates every few minutes. Connection state for nothing.
```

**Correct (climb the ladder only as needed):**

```text
Hop 1, increasing complexity:
- Simple polling: request on an interval. Dead simple; great baseline/fallback.
- Long polling: server holds the request until data is ready. Good for
  "tell me when this async job finishes" (payment status).
- SSE: one-way server -> client stream over HTTP. Feeds, notifications.
- WebSockets: full-duplex persistent TCP. True bidirectional (chat). Pin with L4 LB.
- WebRTC: peer-to-peer media or direct peer data.
```

Hop 2 (server-side push): poll a store; route via **consistent hashing** (topic/user -> known server); or **pub/sub** (Redis, Kafka) to fan messages to the servers holding the relevant connections. Discuss reconnection, ordering, and the "one user with millions of followers" fan-out problem.

## Sources

- `references/raw/learn/system-design/patterns/realtime-updates.md` - the two hops and the connection-complexity ladder.
