---
title: "Handling Large Blobs"
context: patterns
category: core
concept: large-blobs
description: "Store large files in object storage with only metadata in the DB; move bytes client-to-S3 via presigned URLs, chunk big files, and serve via CDN."
tags: blobs, object-storage, presigned-url, chunking, cdn
sources:
  - "references/raw/learn/system-design/patterns/large-blobs.md"
last_ingested: 2026-06-01
---

## Handling Large Blobs

Large files (images, video, backups) that do not belong in your primary DB. Uploads, media, documents, anything MB-GB sized.

**Incorrect (bytes through the app and into the DB):**

```text
Client uploads a 2GB video to the app server, which streams it into a
BLOB column. App memory blows up; the DB bloats; throughput collapses.
```

**Correct (object storage + presigned URLs):**

```text
- Store blobs in object storage (S3); keep only metadata + the object
  key/URL in the database.
- Upload/download directly client <-> S3 via presigned URLs (bytes skip
  your servers).
- Chunk large files for resumable uploads.
- Serve via CDN.
```

For very large files discuss **multipart upload** and integrity hashing per chunk. The database tracks only metadata and chunk state, so the app tier stays thin and stateless.

## Sources

- `references/raw/learn/system-design/patterns/large-blobs.md` - object storage, presigned URLs, chunking, CDN delivery.
