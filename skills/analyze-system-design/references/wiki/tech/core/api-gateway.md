---
title: "API Gateway: Microservices Entry Point"
context: tech
category: core
concept: api-gateway
description: "Single entry point for routing, auth, rate limiting, and SSL termination in a microservices architecture; skip it for a simple client-server app."
tags: api-gateway, routing, auth, microservices
sources:
  - "references/raw/learn/system-design/deep-dives/api-gateway.md"
last_ingested: 2026-06-01
---

## API Gateway: Microservices Entry Point

The entry point for client requests: routing, authentication, rate limiting, SSL termination, sometimes protocol translation.

**Reach for it when:** you have a **microservices** architecture. Skip it for a simple client-server app.

**Incorrect (over-investing interview time):**

```text
Spend 8 minutes detailing gateway plugin chains and routing tables for a
two-service design -> time lost on the least interesting component.
```

**Correct (centralize cross-cutting concerns, move on):**

```text
Put the gateway between clients and services so auth, rate limiting, and
routing live in one place. Note it, then return to the interesting parts.
```

Over-explaining the gateway is a bigger risk than under-explaining it; it is rarely where the signal is.

## Sources

- `references/raw/learn/system-design/deep-dives/api-gateway.md` - what it centralizes, when to use it, the do-not-over-invest warning.
