---
title: "Dealing With Contention"
context: patterns
category: core
concept: dealing-with-contention
description: "Keep data correct when many actors mutate the same thing; prefer the database's own guarantees (optimistic vs pessimistic) and make operations idempotent."
tags: contention, locking, concurrency, idempotency
sources:
  - "references/raw/learn/system-design/patterns/dealing-with-contention.md"
last_ingested: 2026-06-01
---

## Dealing With Contention

Correctness under concurrent mutation of the same thing: double-booking, oversold inventory, double-spend. Bookings (Ticketmaster), inventory, ride matching (Uber), payments, coupons.

**Incorrect (read-modify-write race):**

```text
read seats_left=1 -> two requests both see 1 -> both write 0 -> two buyers,
one seat. The lost-update classic.
```

**Correct (use the DB's guarantees, by contention level):**

```text
Low contention  -> optimistic concurrency: version / compare-and-set,
                   retry on conflict.
High contention -> pessimistic locking: SELECT ... FOR UPDATE, row locks.
```

When the contended resource is not a single DB row, use a **distributed lock** (Redis `INCR`+TTL, or Redlock). When correctness must trump performance (financial, hours-long locks), use a **ZooKeeper** lock.

Make operations **idempotent** with idempotency keys so retries are safe. Favor atomic operations and transactions; never do read-modify-write without a guard.

## Sources

- `references/raw/learn/system-design/patterns/dealing-with-contention.md` - optimistic vs pessimistic, distributed locks, idempotency.
