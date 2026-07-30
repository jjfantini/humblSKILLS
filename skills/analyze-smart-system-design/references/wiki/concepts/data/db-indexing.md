---
title: "Database Indexing for Read Latency"
context: concepts
category: data
concept: db-indexing
description: "Indexes turn O(n) scans into fast lookups; default B-tree, use specialized indexes for full-text and geo, and add them deliberately since they slow writes."
tags: indexing, b-tree, inverted-index, geospatial
sources:
  - "references/raw/learn/system-design/core-concepts/db-indexing.md"
last_ingested: 2026-06-01
---

## Database Indexing for Read Latency

Auxiliary structures that turn full-table O(n) scans into fast lookups. A core deep-dive lever whenever a query filters, sorts, or joins on a column at scale.

**Incorrect (index everything):**

```text
Add an index on every column "to be safe" -> writes slow down,
storage balloons, and the planner still scans for the queries that matter.
```

**Correct (index the query, deliberately):**

```text
B-tree on (status, created_at) to serve the actual filter+sort.
Covering index so the query is answered from the index alone.
```

Picking the index type:
- **B-tree** - default; equality and range queries.
- **Hash** - exact match only.
- **Inverted index** - full-text (Elasticsearch, Postgres FTS).
- **Geospatial** - GiST/PostGIS, Redis geo for proximity.
- **Composite** - order the columns to match the query pattern.

Indexes speed reads but slow writes and cost storage, so add them for specific queries, not by default. Cover the query with the index where you can.

## Sources

- `references/raw/learn/system-design/core-concepts/db-indexing.md` - index types, the read/write trade-off, covering indexes.
