---
title: "Whiteboard: YouTube (Video Upload and Streaming)"
context: examples
category: whiteboards
concept: youtube
description: "Reference design for large-video upload via presigned S3 URLs, Lambda upload monitor, transcoding DAG, metadata cache, and CDN-backed adaptive streaming."
tags: youtube, video, s3, cdn, transcoding, whiteboard
sources:
  - "references/raw/whiteboards/youtube.png"
  - "references/raw/learn/system-design/problem-breakdowns/youtube.md"
last_ingested: 2026-06-01
---

## Whiteboard: YouTube (Video Upload and Streaming)

Full upload + processing + playback design. Core is Video Service + Metadata DB + S3; deep dives are the processing pipeline, CDN, and upload monitor.

**Upload flow:**

```text
Client -> API Gateway (routing, auth, rate limit)
       -> Video Service -> getPresignedURL() from S3 -> client uploads direct to S3
       -> Video Metadata DB (videoId, uploaderId, name, description, chunks, S3 URLs)
S3 events -> Upload Monitor (Lambda) -> chunk progress in Metadata DB
```

**Processing pipeline (deep dive):**

```text
S3 (raw upload) -> Video Processing Service
  -> Video Splitter -> parallel Transcoding / Audio / Transcript workers
  -> Build + store manifest files -> segments + manifests back to S3
  -> mark done in Metadata DB
```

**Playback flow:**

```text
Client GET /video -> Video Service -> Video Metadata Cache (Redis) or Metadata DB
Client fetches manifest + segments -> CDN (cache miss pulls from S3) -> adaptive bitrate streaming
```

**Scale question on the board:** "How do we scale to a large number of videos uploaded / watched a day?" Answer visible in the design: decouple upload (presigned S3) from processing (S3-triggered pipeline) and reads (CDN + metadata cache).

## Sources

- `references/raw/whiteboards/youtube.png` - upload, processing DAG, CDN playback paths.
- `references/raw/learn/system-design/problem-breakdowns/youtube.md` - segment storage, resumable upload, adaptive bitrate prose.
