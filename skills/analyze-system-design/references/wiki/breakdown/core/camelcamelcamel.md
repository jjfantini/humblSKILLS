---
title: "Design CamelCamelCamel (Price Tracker)"
context: breakdown
category: core
concept: camelcamelcamel
description: "Track Amazon price history and send price-drop alerts for 500M products under rate limits; crowdsource collection via the Chrome extension and tier crawling by interest."
tags: price-tracker, crawling, time-series, notifications, crowdsource
sources:
  - "references/raw/learn/system-design/problem-breakdowns/camelcamelcamel.md"
last_ingested: 2026-06-01
---

## Design CamelCamelCamel (Price Tracker)

**Functional:** view price history for Amazon products (web + Chrome extension); subscribe to price-drop notifications with a threshold.

**Non-functional:** availability > consistency; 500M products; price-history queries <500ms; notifications within 1 hour of a price change. Amazon rate-limits ~1 req/s/IP.

**Core entities:** Product, User, Subscription, Price (time-series).

**Key API:**

```text
GET  /products/{id}/price?period=&granularity= -> PriceHistory[]
POST /subscriptions { product_id, price_threshold, notification_type }
```

**High-level design:**

> Assumed flow (diagram in source omitted; inferred): Web Crawler / extension -> Price DB (time-series, append-only) + Primary DB (Users, Products, Subscriptions); Price History Service serves charts; notification job emails on drops.

Split the append-only **time-series Price DB** from the CRUD Primary DB.

**Key deep dives:**
- **Track 500M products under rate limits:** a naive crawl takes 15+ years. **Tier crawling by user interest** (Pareto) and, best, **crowdsource via the Chrome extension** (1M users report prices for pages they view) - turning the constraint into the data source.
- **Malicious reports:** validate crowdsourced prices (cross-check, anomaly detection) before triggering alerts.
- **Notifications:** start with a 2h cron; move toward event-driven on price change to hit the 1h SLA.

## Sources

- `references/raw/learn/system-design/problem-breakdowns/camelcamelcamel.md` - crowdsourced collection, tiered crawling, data validation, notifications.
