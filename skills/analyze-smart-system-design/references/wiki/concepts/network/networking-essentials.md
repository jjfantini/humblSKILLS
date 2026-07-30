---
title: "Networking Essentials for Interviews"
context: concepts
category: network
concept: networking-essentials
description: "The OSI layers that matter (L3/L4/L7) and L4 vs L7 load balancers, with the rule to pin WebSockets behind an L4 LB."
tags: networking, tcp, load-balancer, websocket
sources:
  - "references/raw/learn/system-design/core-concepts/networking-essentials.md"
last_ingested: 2026-06-01
---

## Networking Essentials

The layers worth knowing: **L3** IP (best-effort packets), **L4** TCP (connection-oriented, ordered, reliable) and UDP (connectionless, fire-and-forget), **L7** application protocols (HTTP, WebSocket, WebRTC, DNS).

Load balancers come in two flavors:
- **L4 LB** - transport level; pins a whole TCP session to one server.
- **L7 LB** - HTTP-aware; routing, TLS termination, header logic.

**Incorrect (persistent connections behind L7):**

```text
Clients open WebSockets through a plain L7 round-robin LB ->
the persistent TCP session gets confused across backends, connections drop.
```

**Correct (match LB to the protocol):**

```text
Default client traffic -> L7 LB (flexible HTTP routing).
WebSockets / long-lived TCP -> L4 LB (the session pins to one server).
```

Reach for **UDP** only when loss is tolerable and latency is king (video, games, some DNS). For everything request/response, HTTP over an L7 LB is the default.

## Sources

- `references/raw/learn/system-design/core-concepts/networking-essentials.md` - OSI layers, L4 vs L7 LB, the WebSocket pinning rule.
