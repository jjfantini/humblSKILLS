---
title: "Whiteboard: Web Crawler (Scale and Efficiency)"
context: examples
category: whiteboards
concept: web-crawler
description: "Reference pipelined crawler: Frontier SQS with DLQ, Redis rate limiter/DNS cache, raw HTML in S3, separate Parsing Queue and worker with hash dedup."
tags: web-crawler, sqs, s3, politeness, pipeline, whiteboard
sources:
  - "references/raw/whiteboards/web_crawler.png"
  - "references/raw/learn/system-design/problem-breakdowns/web-crawler.md"
last_ingested: 2026-06-01
---

## Whiteboard: Web Crawler (Scale and Efficiency)

Labeled "Scale and Efficiency" on the board. Core loop is frontier -> fetch -> store -> parse -> re-enqueue. Deep dives are pipelining, politeness, and dedup.

**Crawl loop:**

```text
Frontier Queue (SQS + DLQ) -> Crawler (fetch & store webpage)
  -> DNS + Webpage
  -> Save raw HTML -> S3 HTML Data
  -> Update URL status -> URL Metadata DB
  -> Parsing Queue (SQS)
Parsing Worker (check hash first)
  -> fetch S3 URL from Metadata -> download HTML -> save parsed text to S3
  -> extract URLs -> back to Frontier Queue
Retry on failure w/ backoff -> Frontier Queue
```

**Metadata schemas (from diagram):**

- **URL:** id, url, s3Link, lastCrawlTime, hash (GSI), depth
- **Domain:** domain, lastCrawlTime, robots

**Deep-dive components:**

- **Rate Limiter + DNS Cache (Redis):** "Just use same Redis cluster" - politeness per domain + cached DNS resolution.
- **Separate Parsing Queue:** crawl and parse scale independently; queue carries URL id, not HTML bytes.
- **Hash check first:** content dedup before re-parsing identical pages.
- **DLQ + backoff:** failed crawls retry with exponential backoff; poison messages to DLQ.

Lesson: a single-process crawler satisfies functional requirements. SQS stages, S3 blob storage, Redis politeness, and hash dedup are step-6 answers to fault tolerance, politeness, and 10B-page efficiency.

## Sources

- `references/raw/whiteboards/web_crawler.png` - pipelined fetch/parse, S3 storage, Redis politeness, hash dedup.
- `references/raw/learn/system-design/problem-breakdowns/web-crawler.md` - robots.txt, visibility timeout backoff prose.
