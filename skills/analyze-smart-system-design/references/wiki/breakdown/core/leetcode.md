---
title: "Design LeetCode (Coding Judge)"
context: breakdown
category: core
concept: leetcode
description: "List/view problems, run untrusted user code safely in sandboxed containers within 5s, and serve a live competition leaderboard via a Redis sorted set."
tags: leetcode, code-execution, containers, sandbox, leaderboard
sources:
  - "references/raw/learn/system-design/problem-breakdowns/leetcode.md"
last_ingested: 2026-06-01
---

## Design LeetCode (Coding Judge)

**Functional:** view a list of problems; view a problem and code a solution; submit and get feedback; view a live competition leaderboard.

**Non-functional:** availability > consistency; secure, isolated code execution; results within 5s; competitions up to 100k users. Note: this is a small-scale system (hundreds of thousands of users, ~4k problems).

**Core entities:** Problem (statement, test cases, code stubs), Submission, Leaderboard.

**Key API:**

```text
GET  /problems?page=&limit=
GET  /problems/{id}?language=
POST /problems/{id}/submit { code, language }
GET  /leaderboard/{competitionId}
```

**High-level design:**

> Assumed flow (diagram in source omitted; inferred): Client -> API Server -> DynamoDB (problems w/ nested test cases) and language-specific execution containers.

**Key deep dives:**
- **Run code safely:** **Docker containers** per language (lighter than VMs, faster than serverless cold starts); read-only FS, CPU/memory limits, explicit timeout, no network (seccomp).
- **Leaderboard:** **Redis sorted set** (`ZADD`/`ZRANGE REV`) + 5s client polling, not WebSockets (overkill).
- **Scale execution:** **dynamic horizontal scaling** of containers + a **queue (SQS)** to buffer competition spikes and enable retries; async submit -> poll `GET /check/{id}`.
- **Test harness:** one serialized test-case set per problem; per-language harness deserializes input, runs, compares output.

## Sources

- `references/raw/learn/system-design/problem-breakdowns/leetcode.md` - sandboxed execution, leaderboard, scaling, test harness.
