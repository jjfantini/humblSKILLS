---
title: "Elasticsearch: Search Over an Inverted Index"
context: tech
category: core
concept: elasticsearch
description: "Distributed search/analytics engine; treat it as a secondary read store fed from your source of truth, and try Postgres FTS first."
tags: elasticsearch, search, inverted-index, full-text
sources:
  - "references/raw/learn/system-design/deep-dives/elasticsearch.md"
last_ingested: 2026-06-01
---

## Elasticsearch: Search Over an Inverted Index

Distributed search and analytics engine over an inverted index. Documents grouped in indices (like tables); clean REST API.

**Reach for it when:** full-text search, fuzzy/typeahead, faceted or aggregation queries, log and observability search.

**Incorrect (system of record):**

```text
Write user accounts and orders only to Elasticsearch and read them back ->
no transactions, eventual consistency, lost writes on rebalance.
```

**Correct (secondary read store):**

```text
Source of truth = Postgres/Dynamo. Feed Elasticsearch via CDC or dual-write.
Define mappings, index documents, query the index for search only.
```

Start with **Postgres full-text search** for simple needs and add Elasticsearch only when search requirements outgrow it; that keeps the architecture simpler. Never make it your system of record.

## Sources

- `references/raw/learn/system-design/deep-dives/elasticsearch.md` - inverted index, secondary-store pattern, Postgres-FTS-first rule.
