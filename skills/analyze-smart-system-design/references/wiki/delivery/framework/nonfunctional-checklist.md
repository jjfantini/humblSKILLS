---
title: "Non-Functional Requirements Checklist"
context: delivery
category: framework
concept: nonfunctional-checklist
description: "The 8-point checklist (CAP, scalability, latency, durability, security, fault tolerance, environment, compliance) for finding the top 3-5 system qualities."
tags: non-functional, cap, scalability, durability, checklist
sources:
  - "references/raw/learn/system-design/in-a-hurry/delivery.md"
last_ingested: 2026-06-01
---

## Non-Functional Requirements Checklist

Coming up with non-functional requirements is hard, especially in an unfamiliar domain. Walk this checklist and pick the top 3-5 most relevant, each quantified and in context.

**Incorrect (vague, untargeted):**

```text
"The system should be scalable, available, fast, and secure."
```

True of every system; signals nothing.

**Correct (targeted to this system):**

```text
- CAP: availability over consistency (a stale feed is fine)
- Scalability: 100M+ DAU, read-heavy, bursty around events
- Latency: feed renders under 200ms p99
- Durability: some tweet loss tolerable (not a bank)
```

The checklist:

1. **CAP** - consistency vs availability under partition. Ask "does every read need the most recent write?" Yes -> consistency; no -> availability. Discuss this first; it shapes everything. Often per-feature.
2. **Scalability** - bursty traffic? read- vs write-heavy? holiday/event spikes?
3. **Latency** - which operation must be fast, and how fast (with a number)?
4. **Durability** - how bad is data loss? Social feed vs bank.
5. **Security** - data protection, access control, compliance.
6. **Fault tolerance** - redundancy, failover, recovery.
7. **Environment constraints** - mobile battery, limited memory or bandwidth (video on 3G).
8. **Compliance** - GDPR, PCI, and other legal or regulatory limits.

## Sources

- `references/raw/learn/system-design/in-a-hurry/delivery.md` - the 8-point checklist and the quantify-in-context rule.
