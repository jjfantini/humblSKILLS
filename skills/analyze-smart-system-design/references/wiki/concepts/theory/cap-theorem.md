---
title: "CAP Theorem: Consistency vs Availability"
context: concepts
category: theory
concept: cap-theorem
description: "Under a network partition you must choose consistency or availability; the single most important early decision, often made per-feature."
tags: cap, consistency, availability, partition
sources:
  - "references/raw/learn/system-design/core-concepts/cap-theorem.md"
last_ingested: 2026-06-01
---

## CAP Theorem: Consistency vs Availability

Under a network partition (a given in distributed systems) you must choose consistency or availability. This is the first non-functional decision and it shapes every later choice.

**Incorrect (treating it globally):**

```text
"The whole system is eventually consistent." But booking a ticket and
browsing events have opposite needs - one global choice is wrong.
```

**Correct (per-feature, driven by one question):**

```text
Ask: "does every read need the most recent write?"
Ticketmaster: booking -> consistency; browsing -> availability.
Tinder: matches -> consistency; profile views -> availability.
```

Map the answer to mechanisms:
- **Consistency** - RDBMS, Spanner, DynamoDB strong mode, distributed transactions, single-node.
- **Availability** - read replicas + async replication, CDC, Cassandra, Redis clusters, eventual consistency.

Know the spectrum between the extremes: strong, causal, read-your-own-writes, eventual. Stating the right point on that spectrum per feature is a senior signal.

## Sources

- `references/raw/learn/system-design/core-concepts/cap-theorem.md` - the choose-one rule, the per-feature framing, the consistency spectrum.
