---
title: "Design Gopuff (Local Delivery)"
context: breakdown
category: core
concept: gopuff
description: "Aggregate item availability across nearby distribution centers under 100ms and place strongly consistent orders without double-booking physical inventory."
tags: gopuff, delivery, inventory, geo, transactions
sources:
  - "references/raw/learn/system-design/problem-breakdowns/gopuff.md"
last_ingested: 2026-06-01
---

## Design Gopuff (Local Delivery)

**Functional:** query availability of items deliverable in 1 hour by location (union of nearby DCs); order multiple items at once.

**Non-functional:** availability reads fast (<100ms); ordering strongly consistent (no two customers buy the same physical unit); 10k DCs, 100k items; ~10M orders/day.

**Core entities:** Item (a type, e.g. Cheetos), Inventory (a physical unit at a DC), DistributionCenter, Order.

**Key API:**

```text
GET  /availability?lat=&long=&keyword=&cursor= -> Item[]
POST /orders { items[], lat, long }
```

**High-level design:**

> Assumed flow (diagram in source omitted; inferred): Client -> Availability Service -> Nearby Service (lat/long -> DC list) -> Inventory DB; Orders Service -> Postgres leader (atomic txn).

Postgres holds inventory + orders, partitioned by region (first 3 zip digits). Availability reads from replicas; orders write to the leader.

**Key deep dives:**
- **Drive-time, not crow-flies:** prune to candidate DCs within ~60mi, then call a travel-time service (traffic-aware) only on those.
- **No double-booking:** prefer a **single SERIALIZABLE Postgres transaction** over a distributed lock across two stores (avoids deadlocks and crash-after-order failure modes).
- **Scaling reads (~20k QPS):** Redis cache with a short TTL (1 min) invalidated on inventory writes; read replicas + region partitioning.

## Sources

- `references/raw/learn/system-design/problem-breakdowns/gopuff.md` - item vs inventory, drive-time, atomic transactions, read scaling.
