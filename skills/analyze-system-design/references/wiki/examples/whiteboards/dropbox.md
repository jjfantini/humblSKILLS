---
title: "Whiteboard: Dropbox (Simplicity Bar)"
context: examples
category: whiteboards
concept: dropbox
description: "The simplicity bar for file sync: blobs go client-to-S3 via presigned URLs while the DB tracks only metadata and chunk state; chunking satisfies 50GB + resume."
tags: dropbox, file-sync, presigned-url, chunking, simplicity, whiteboard
sources:
  - "references/raw/whiteboards/dropbox.png"
last_ingested: 2026-06-01
---

## Whiteboard: Dropbox (Simplicity Bar)

A reference design holding the simplicity bar for file sync.

**Incorrect (over-built):**

```text
Stream file bytes through the app tier into a sharded DB and build a custom
replication protocol before the requirements force chunking or presigned URLs.
```

**Correct (the simple shape):**

```text
Client (local DB + folder) -> API Gateway -> File Service (presigned URLs,
store metadata in DB) and Update Service (sync). Bytes go client <-> S3
directly via presigned URLs; DB holds file_metadata + chunks; CDN for download.
```

- **Functional:** upload, download, auto-sync across devices.
- **Non-functional:** 100M DAU; files up to 50GB with chunking + auto-resume; sync in under a minute; availability-first (eventual consistency).
- **Why simple:** blobs live in S3 and skip the app servers entirely; the DB only tracks metadata and chunk state. Chunking exists to satisfy the 50GB + resumable requirement.

Lesson: push bytes around your servers, not through them.

## Sources

- `references/raw/whiteboards/dropbox.png` - the reference whiteboard for the Dropbox file-sync design.
