---
title: "PostgreSQL: The Relational Default"
context: tech
category: core
concept: postgresql
description: "Battle-tested ACID relational DB with joins, B-tree/FTS/geo indexes, and replication; handles millions of users on one tuned primary plus read replicas."
tags: postgresql, relational, acid, replicas, fts
sources:
  - "references/raw/learn/system-design/deep-dives/postgres.md"
last_ingested: 2026-06-01
---

## PostgreSQL: The Relational Default

Battle-tested relational DB: strong consistency, ACID transactions, rich querying, B-tree indexes, full-text and geospatial (PostGIS/GiST), replication.

**Reach for it when:** you need transactions, joins, flexible queries, or just a solid default. One well-tuned primary + replicas handles millions of users and terabytes.

**Incorrect (premature store zoo):**

```text
Add Elasticsearch for search, a separate geo DB, and a NoSQL store on
day one -> three systems to operate before any requirement demands them.
```

**Correct (stretch Postgres first):**

```text
Index for read latency: B-tree default, GIN for FTS, GiST for geo.
Scale reads with read replicas (accept lag).
Scale writes vertically, then partition/shard.
Use built-in full-text search before reaching for Elasticsearch.
```

Prefer one boring Postgres over a zoo of specialized stores whenever it meets the requirements; it is the strongest default answer for most product designs.

## Sources

- `references/raw/learn/system-design/deep-dives/postgres.md` - ACID/indexes, read-replica scaling, FTS-first, prefer-one-store guidance.
