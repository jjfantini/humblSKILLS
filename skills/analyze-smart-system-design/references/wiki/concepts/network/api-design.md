---
title: "API Design: The Client Contract"
context: concepts
category: network
concept: api-design
description: "Default REST with plural-noun resources and HTTP verbs; derive the user from the auth token, paginate with cursors, version the API."
tags: api, rest, websocket, grpc, pagination
sources:
  - "references/raw/learn/system-design/core-concepts/api-design.md"
last_ingested: 2026-06-01
---

## API Design: The Client Contract

Step 3, every interview. The API is the contract that satisfies the functional requirements, usually one endpoint per feature.

**Incorrect (trusting the body for identity):**

```text
POST /v1/tweets { "user_id": "123", "text": "hi" }
```

Never trust a user ID from the request body; it is forgeable.

**Correct (derive identity from the token):**

```text
POST /v1/tweets { "text": "hi" }      # current user from auth token
GET  /v1/feed?cursor=...              # cursor pagination
```

Defaults:
- **REST** with plural-noun resources and HTTP verbs (`POST /v1/tweets`, `GET /v1/feed`).
- **WebSocket / SSE** for realtime; **gRPC** for internal service-to-service when latency matters; **GraphQL** for diverse client data needs.
- Paginate list endpoints; prefer **cursor** over offset for large or changing sets.
- **Version** the API (`/v1/`).

Design the core REST API first even when you will add realtime; do not overthink the protocol choice.

## Sources

- `references/raw/learn/system-design/core-concepts/api-design.md` - REST default, protocol picks, auth-token rule, pagination and versioning.
