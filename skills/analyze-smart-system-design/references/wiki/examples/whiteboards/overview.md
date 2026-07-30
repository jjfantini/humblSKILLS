---
title: "Whiteboard Examples Index"
context: examples
category: whiteboards
concept: overview
description: "Index of reference whiteboards: simplicity-bar sketches for step 5 and full final designs (with deep-dive components) for step 6 prep."
tags: whiteboard, examples, simplicity, overview
sources:
  - "references/raw/whiteboards/url_shortener.png"
  - "references/raw/whiteboards/uber.png"
  - "references/raw/whiteboards/whatsapp_chat.png"
  - "references/raw/whiteboards/dropbox.png"
  - "references/raw/whiteboards/chatgpt.png"
  - "references/raw/whiteboards/payment_system.png"
  - "references/raw/whiteboards/metrics_monitoring.png"
  - "references/raw/whiteboards/job_scheduler.png"
  - "references/raw/whiteboards/youtube.png"
  - "references/raw/whiteboards/youtube_top_k.png"
  - "references/raw/whiteboards/online_auction.png"
  - "references/raw/whiteboards/web_crawler.png"
  - "references/raw/whiteboards/fb_post_search.png"
last_ingested: 2026-06-01
---

## Whiteboard Examples Index

Two tiers of reference designs live under `wiki/examples/whiteboards/`.

### Simplicity bar (step 5 target)

Use these when sketching the **minimum working system**. Each extra box should trace to a functional requirement.

| Concept | Wiki | Raw image |
|---------|------|-----------|
| URL shortener | `url-shortener.md` | `raw/whiteboards/url_shortener.png` |
| Uber ride-sharing | `uber.md` | `raw/whiteboards/uber.png` |
| WhatsApp chat | `whatsapp.md` | `raw/whiteboards/whatsapp_chat.png` |
| Dropbox sync | `dropbox.md` | `raw/whiteboards/dropbox.png` |
| ChatGPT inference | `chatgpt.md` | `raw/whiteboards/chatgpt.png` |

### Full final designs (step 5 + deep dives)

Use these when the prompt forces async pipelines, realtime, or massive scale. Core path stays simple; Kafka, Flink, cron precompute, reconciliation, etc. are **step-6 additions**.

| Problem | Wiki | Raw image | Paired breakdown |
|---------|------|-----------|------------------|
| Payment system (Stripe-like) | `payment-system.md` | `payment_system.png` | `breakdown/core/payment-system.md` |
| Metrics monitoring | `metrics-monitoring.md` | `metrics_monitoring.png` | `breakdown/core/metrics-monitoring.md` |
| Job scheduler | `job-scheduler.md` | `job_scheduler.png` | `breakdown/core/job-scheduler.md` |
| YouTube upload/stream | `youtube.md` | `youtube.png` | `breakdown/core/youtube.md` |
| YouTube top-K | `youtube-top-k.md` | `youtube_top_k.png` | `breakdown/core/top-k.md` |
| Online auction | `online-auction.md` | `online_auction.png` | `breakdown/core/online-auction.md` |
| Web crawler | `web-crawler.md` | `web_crawler.png` | `breakdown/core/web-crawler.md` |
| Facebook post search | `fb-post-search.md` | `fb_post_search.png` | `breakdown/core/fb-post-search.md` |

**How to use:** deliver the simplicity-bar shape in steps 1-5, then name which deep-dive boxes from the full design you would add and why.

## Sources

All PNGs under `references/raw/whiteboards/`; Hello Interview prose under `references/raw/learn/system-design/problem-breakdowns/` where listed in each concept's `sources:`.
