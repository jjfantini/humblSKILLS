---
title: "Data Modeling and Store Selection"
context: concepts
category: data
concept: data-modeling
description: "Turn core entities into schema and pick relational vs NoSQL from access patterns, not hype; model relationships and only write fields the design needs."
tags: data-modeling, schema, relational, nosql
sources:
  - "references/raw/learn/system-design/core-concepts/data-modeling.md"
last_ingested: 2026-06-01
---

## Data Modeling and Store Selection

Used in step 2 (names) and step 5 (fields and relationships). The store choice falls out of how you query, not what is trendy.

**Incorrect (picking by hype):**

```text
"We'll use MongoDB because it's web-scale" - then immediately need
multi-row transactions and joins it does not give you cleanly.
```

**Correct (pick from access patterns):**

```text
Need joins, transactions, flexible ad-hoc queries -> relational (Postgres).
Know your query patterns up front, need horizontal scale -> NoSQL (DynamoDB/Cassandra).
```

Model the relationships explicitly (1:1, 1:many, many:many). During step 5, write only the fields that matter to the design next to the database; the interviewer infers the obvious ones (a `User` has name, email, password hash). Defer the full schema until you know what state each request mutates.

When one boring relational DB meets the requirements, prefer it over a zoo of specialized stores.

## Sources

- `references/raw/learn/system-design/core-concepts/data-modeling.md` - relational vs NoSQL by access pattern, relationship modeling, minimal fields.
