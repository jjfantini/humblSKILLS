---
title: "Design a Web Crawler"
context: breakdown
category: core
concept: web-crawler
description: "Crawl 10B pages in under 5 days to extract text; a pipelined fetch/parse design over SQS with backoff, politeness via robots.txt, and content dedup."
tags: web-crawler, pipeline, sqs, politeness, dedup
sources:
  - "references/raw/whiteboards/web_crawler.png"
  - "references/raw/learn/system-design/problem-breakdowns/web-crawler.md"
last_ingested: 2026-06-01
---

## Design a Web Crawler

**Functional:** crawl from seed URLs; extract and store text data for later processing (e.g. LLM training).

**Non-functional:** fault tolerant (resume without losing progress); polite (respect robots.txt, ~1 req/s per domain); efficient (10B pages in <5 days); scalable.

**System interface:** input = seed URLs; output = text data in S3.

**Data flow:** frontier URL -> DNS -> fetch HTML -> extract text + links -> store -> repeat.

**High-level design (from whiteboard - Scale and Efficiency):**

```text
Frontier Queue (SQS + DLQ) -> Crawler -> DNS/Webpage -> raw HTML in S3 -> URL Metadata DB
                           -> Parsing Queue -> Parsing Worker (hash check) -> text in S3
                           -> extracted URLs back to Frontier ; retry w/ backoff on failure
Rate Limiter + DNS Cache (Redis, same cluster) <- Crawler
```

**Key deep dives (visible on whiteboard):**
- **Pipeline stages:** separate Frontier/Parsing queues so fetch and parse scale independently; queue carries URL id, not HTML.
- **S3 blob storage:** raw HTML and parsed text in S3; Metadata DB tracks url, s3Link, hash (GSI), depth, robots per Domain.
- **Politeness:** Redis rate limiter per domain; robots.txt in Domain schema.
- **Hash dedup:** Parsing Worker checks hash first before re-parsing identical content.
- **Fault tolerance:** DLQ after repeated failures; backoff re-enqueues to Frontier.

See also `wiki/examples/whiteboards/web-crawler.md`.

## Sources

- `references/raw/whiteboards/web_crawler.png` - pipelined fetch/parse, S3, Redis politeness, hash dedup.
- `references/raw/learn/system-design/problem-breakdowns/web-crawler.md` - robots.txt, visibility timeout backoff prose.
