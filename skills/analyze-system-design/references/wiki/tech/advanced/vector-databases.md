---
title: "Vector Databases: Similarity Search via ANN"
context: tech
category: advanced
concept: vector-databases
description: "Stores embeddings and finds similar items fast via approximate nearest neighbor (usually HNSW); start with pgvector/ES kNN, not a purpose-built DB."
tags: vector-db, embeddings, ann, hnsw, rag
sources:
  - "references/raw/learn/system-design/deep-dives/vector-databases.md"
last_ingested: 2026-06-01
---

## Vector Databases: Similarity Search via ANN

Vector DBs store **embeddings** and answer one question fast: "find the things most similar to this." That similarity primitive powers semantic search, recommendations, and RAG. They show up almost exclusively next to AI/ML problems; interviewers care how and where you use one, not the internals.

**Incorrect (exact match or premature platform):**

```text
Use a vector DB to "find document by ID" (use a regular DB), or reach for
Pinecone/Milvus on day one when pgvector handles millions of vectors.
```

**Correct (ANN index, start on a DB you run):**

```text
Embedding model: content in -> fixed-length vector out (128-1536 dims).
Index: HNSW by default (multi-layer graph, O(log n), 95%+ recall, memory-hungry).
  Alternatives: IVF (clusters), LSH, Annoy (static datasets).
Start: pgvector or Elasticsearch kNN; purpose-built only when scale/features demand.
```

Architecture: vector DB as a separate service - app embeds the query, gets similar IDs, fetches full records from the primary DB. Raise consistency (eventually consistent is usually fine), update strategy, and filtering (pre/post/hybrid). It is **not a system of record**; authoritative data lives elsewhere, and changing the embedding model invalidates all vectors (plan a re-embed).

## Sources

- `references/raw/learn/system-design/deep-dives/vector-databases.md` - embeddings, ANN index types, start-simple options, architecture patterns, gotchas.
