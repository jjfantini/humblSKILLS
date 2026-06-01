---
title: "Design Google Docs (Collaborative Editing)"
context: breakdown
category: core
concept: google-docs
description: "Concurrent real-time document editing with presence; transmit edits (not snapshots) over WebSockets and converge with OT or CRDTs."
tags: google-docs, collaboration, websocket, ot, crdt
sources:
  - "references/raw/learn/system-design/problem-breakdowns/google-docs.md"
last_ingested: 2026-06-01
---

## Design Google Docs (Collaborative Editing)

**Functional:** create documents; multiple users edit the same doc concurrently; see each other's changes in real-time; see cursors/presence.

**Non-functional:** eventual consistency (all converge); updates <100ms; millions of concurrent users across billions of docs; <=100 concurrent editors per doc; durable across restarts.

**Core entities:** Editor, Document, Edit, Cursor.

**Key API:**

```text
POST /docs { title } -> docId
WS   /docs/{docId}  SEND insert/delete/updateCursor ; RECV update
```

**High-level design:**

> Assumed flow (diagram in source omitted; inferred): Client <-WebSocket-> Document Service (per-doc edit ordering) -> persisted edits/snapshots; metadata in Postgres.

Transmit **edits, not snapshots** (sending the whole doc loses concurrent edits and wastes bandwidth).

**Key deep dives:**
- **Consistency:** raw positional edits conflict (a delete at index 6 means different things after an insert). Resolve with **Operational Transformation** (central server transforms ops; low memory, fits the 100-editor cap) or **CRDTs** (commutative, subdividable position IDs + tombstones; no central server, offline-friendly).
- **Scale:** route each document's editors to one server/shard; persist edit log + periodic snapshots for durability and late joiners.
- **Presence/cursors:** ephemeral state broadcast over the same socket.

## Sources

- `references/raw/learn/system-design/problem-breakdowns/google-docs.md` - edits vs snapshots, OT vs CRDT, durability.
