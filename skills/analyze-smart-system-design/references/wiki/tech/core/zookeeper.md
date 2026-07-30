---
title: "ZooKeeper: Distributed Coordination"
context: tech
category: core
concept: zookeeper
description: "Strongly consistent coordination service for leader election, service discovery, and correctness-first locks; prefer Redis locks for fast simple locking."
tags: zookeeper, coordination, leader-election, locks, ephemeral
sources:
  - "references/raw/learn/system-design/deep-dives/zookeeper.md"
last_ingested: 2026-06-01
---

## ZooKeeper: Distributed Coordination

Distributed coordination service: consistent hierarchical key-value (ZNodes), ephemeral nodes, watches, leader election, distributed locks. Strong consistency.

**Reach for it when:** cluster coordination, service discovery, leader election, configuration, or locks where **correctness > performance** (financial, hours-long locks).

**Incorrect (a ZNode per user):**

```text
Create millions of per-user ZNodes to track connections -> ZooKeeper is
not built for that volume; it buckles.
```

**Correct (ephemeral nodes per server):**

```text
Register each server as an ephemeral ZNode (auto-removed on disconnect ->
fast failure detection). Track servers, then use consistent hashing to map
users -> servers, rather than a node per user.
```

Choose your lock by need: **Redis locks** for fast, simple locking (Ticketmaster, Uber); **ZooKeeper locks** for financial-grade correctness or long-lived locks.

## Sources

- `references/raw/learn/system-design/deep-dives/zookeeper.md` - ZNodes, ephemeral nodes, server tracking, Redis-vs-ZooKeeper lock choice.
