---
title: "Design YouTube (Video Sharing)"
context: breakdown
category: core
concept: youtube
description: "Upload and stream large videos; store segments in multiple formats via a transcoding DAG and serve adaptive-bitrate streams from a CDN with resumable uploads."
tags: youtube, video, transcoding, adaptive-bitrate, cdn
sources:
  - "references/raw/whiteboards/youtube.png"
  - "references/raw/learn/system-design/problem-breakdowns/youtube.md"
last_ingested: 2026-06-01
---

## Design YouTube (Video Sharing)

**Functional:** upload videos; watch (stream) videos.

**Non-functional:** availability > consistency; support 10s-of-GB videos; low-latency streaming even on low bandwidth; ~1M uploads/day, 100M watches/day; resumable uploads.

**Core entities:** User, Video, VideoMetadata.

**Key API:**

```text
POST /presigned_url { VideoMetadata } -> presigned S3 URL  (then multipart upload)
GET  /videos/{id} -> VideoMetadata (URL to manifest)
```

**High-level design (from whiteboard):**

```text
Upload:  Client -> API Gateway -> Video Service -> presigned S3 URL + Metadata DB
         Client uploads direct to S3; Upload Monitor (Lambda) tracks chunks
Process: S3 trigger -> Video Processing (split -> transcode/audio/transcript -> manifest) -> S3
Watch:   Client -> Video Service -> Metadata Cache/DB -> CDN serves manifest + segments
```

Store videos as **segments in multiple formats**, not the raw file.

**Key deep dives (visible on whiteboard):**
- **Presigned multipart upload:** client uploads direct to S3; merchant/server never holds video bytes.
- **Upload Monitor (Lambda):** S3 events update chunk status in Metadata DB (resumable uploads).
- **Processing DAG:** splitter -> parallel transcode/audio/transcript -> build manifest -> mark done.
- **Metadata cache (Redis):** hot video metadata off DB.
- **CDN + adaptive bitrate:** client picks segment quality; CDN caches S3 segments/manifests.

See also `wiki/examples/whiteboards/youtube.md`.

## Sources

- `references/raw/whiteboards/youtube.png` - upload, processing pipeline, CDN playback.
- `references/raw/learn/system-design/problem-breakdowns/youtube.md` - segment storage, adaptive bitrate, resumable upload prose.
