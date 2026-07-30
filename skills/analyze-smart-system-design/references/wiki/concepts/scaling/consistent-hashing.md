---
title: "Consistent Hashing"
context: concepts
category: scaling
concept: consistent-hashing
description: "A hash ring that minimizes data movement when nodes join or leave; name-drop it for managed stores, go deep only for infrastructure-from-scratch problems."
tags: consistent-hashing, hash-ring, virtual-nodes, rebalancing
sources:
  - "references/raw/learn/system-design/core-concepts/consistent-hashing.md"
last_ingested: 2026-06-01
---

## Consistent Hashing

A hash ring that minimizes data movement when nodes join or leave. Virtual nodes spread each physical node across many ring points for even load.

**Incorrect (modulo hashing):**

```text
node = hash(key) % N. Add or remove one node and N changes ->
almost every key remaps and the whole cache/DB reshuffles.
```

**Correct (ring placement):**

```text
Place nodes and keys on a ring; a key belongs to the next node clockwise.
Add a node -> only the keys between it and its predecessor move.
Virtual nodes even out the load; replicate to the next N clockwise nodes.
```

When to use it: infrastructure-from-scratch interviews (design a distributed cache, DB, or message broker). Otherwise just say "DynamoDB/Cassandra/CDNs use consistent hashing under the hood" and move on.

Go deep only when asked: ring placement, virtual nodes for balance, replication for fault tolerance, hot-spot mitigation (read replicas, key-space salting). Note Redis Cluster uses fixed 16,384 hash slots instead, a real trade-off worth raising.

## Sources

- `references/raw/learn/system-design/core-concepts/consistent-hashing.md` - ring vs modulo, virtual nodes, replication, Redis Cluster slots.
