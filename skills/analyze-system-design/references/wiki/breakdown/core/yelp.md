---
title: "Design Yelp (Business Reviews)"
context: breakdown
category: core
concept: yelp
description: "Search and view businesses with reviews and leave 1-5 star ratings; precompute average ratings, add geospatial search, and enforce one review per user."
tags: yelp, reviews, geospatial, search, average-rating
sources:
  - "references/raw/learn/system-design/problem-breakdowns/yelp.md"
last_ingested: 2026-06-01
---

## Design Yelp (Business Reviews)

**Functional:** search businesses by name, location (lat/long), category; view a business and its reviews; leave a review (1-5 stars + optional text). Constraint: one review per user per business.

**Non-functional:** search <500ms; highly available (eventual consistency OK); 100M DAU, 10M businesses (data ~1TB - fits one DB).

**Core entities:** Business, User, Review.

**Key API:**

```text
GET  /businesses?query=&location=&category=&page=
GET  /businesses/{id}  and  /businesses/{id}/reviews?page=
POST /businesses/{id}/reviews { rating, text? }
```

**High-level design:**

> Assumed flow (diagram in source omitted; inferred): Client -> API Gateway -> Business Service (search/view) and Review Service -> shared DB.

Separate Review Service (writes are rare vs reads), but a shared DB is fine at this scale.

**Key deep dives:**
- **Average rating:** do not compute on the fly; **precompute** an `average_rating` column updated incrementally on each new review (running count + sum) so search results carry it.
- **Geospatial search:** add a geo index (PostGIS/quadtree or Elasticsearch geo) for location filtering, not a bounding-box scan.
- **One review per user:** unique constraint on `(user_id, business_id)`.
- **Scale reads:** cache hot businesses + CDN.

## Sources

- `references/raw/learn/system-design/problem-breakdowns/yelp.md` - average-rating precompute, geo search, read scaling.
