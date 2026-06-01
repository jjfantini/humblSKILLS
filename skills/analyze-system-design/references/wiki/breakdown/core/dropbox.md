---
title: "Design Dropbox (File Sync)"
context: breakdown
category: core
concept: dropbox
description: "Cloud file storage with upload/download/share/sync; move bytes client-to-S3 via presigned URLs, chunk 50GB files for resume, and sync via WebSocket + polling."
tags: dropbox, file-sync, presigned-url, chunking, large-blobs
sources:
  - "references/raw/learn/system-design/problem-breakdowns/dropbox.md"
last_ingested: 2026-06-01
---

## Design Dropbox (File Sync)

**Functional:** upload, download, share with other users, auto-sync across devices.

**Non-functional:** availability > consistency (a few seconds of staleness is fine); files up to 50GB; secure and recoverable; low-latency upload/download/sync.

**Core entities:** File (raw bytes), FileMetadata (name, size, mime, owner, chunks), User.

**Key API:**

```text
POST /files/presigned-url { FileMetadata } -> PresignedUrl   # then PUT to S3
GET  /files/{id}/presigned-url -> PresignedUrl               # then GET from CDN
POST /files/{id}/share { User[] }
GET  /files/changes?since={ts} -> ChangeEvent[]
```

**High-level design:**

> Assumed flow (diagram in source omitted; inferred): Client <-> API Gateway -> File Service (metadata + presigned URLs); bytes go client <-> S3 directly; metadata DB (Dynamo/Postgres); CDN for downloads.

Store blobs in **S3**, only metadata + chunk state in the DB. Shares get their own normalized SharedFiles table (userId PK, fileId SK).

**Key deep dives:**
- **Large files:** client-side **chunking** (5-10MB) for progress + resumable uploads; track chunk status in metadata; trust-but-verify via S3 ETags / ListParts.
- **Fingerprinting:** content hash (SHA-256) identifies files for dedup/resume; **content-defined chunking** keeps delta sync efficient across edits.
- **Sync:** hybrid **WebSocket push + periodic polling** fallback; last-write-wins conflict resolution.
- **Speed/security:** CDN download, optional compression (compress before encrypt); short-lived signed URLs.

## Sources

- `references/raw/learn/system-design/problem-breakdowns/dropbox.md` - presigned uploads, chunking/resume, sync, security.
