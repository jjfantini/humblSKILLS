---
title: "Whiteboard: URL Shortener (Simplicity Bar)"
context: examples
category: whiteboards
concept: url-shortener
description: "The simplicity bar for a URL shortener: two thin services over one DB plus a cache, with the only clever piece existing to hit latency and uniqueness."
tags: url-shortener, bitly, simplicity, whiteboard
sources:
  - "references/raw/whiteboards/url_shortener.png"
last_ingested: 2026-06-01
---

## Whiteboard: URL Shortener (Simplicity Bar)

A reference design that sets the simplicity bar: one client, an API gateway, one or two services, one DB, plus exactly one extra component per hard requirement.

**Incorrect (busier than the bar):**

```text
Microservice mesh + Kafka + sharded DB + multi-region active-active before
any requirement forces it. Over-built for a redirect service.
```

**Correct (the simple shape):**

```text
Client -> API Gateway -> Write Service (generate short code via a global
counter, save to DB) and Read Service (look up, return 302).
DB + Redis cache keyed short_code -> original_url; CDN in front.
```

- **Functional:** long URL -> short URL; optional custom alias + expiration; redirect to original.
- **Non-functional:** highly available + eventually consistent; redirect under 100ms; URLs permanent by default.
- **Why simple:** two thin services over one DB; the only clever piece (counter + cache) exists purely to hit the latency and uniqueness requirements.

Lesson: if your step-5 sketch is busier than this, cut back.

## Sources

- `references/raw/whiteboards/url_shortener.png` - the reference whiteboard for the URL shortener simplicity bar.
