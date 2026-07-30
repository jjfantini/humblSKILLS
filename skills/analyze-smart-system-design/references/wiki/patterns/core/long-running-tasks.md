---
title: "Long-Running Tasks"
context: patterns
category: core
concept: long-running-tasks
description: "Offload work too slow for a synchronous request to a queue + worker pool, return a job ID immediately, and report results via polling or push."
tags: async, queue, worker-pool, dead-letter, jobs
sources:
  - "references/raw/learn/system-design/patterns/long-running-tasks.md"
last_ingested: 2026-06-01
---

## Long-Running Tasks

Work too slow for a synchronous request: video transcoding, report generation, ML inference, bulk email.

**Incorrect (do it in the request):**

```text
POST /reports holds the HTTP connection open for 90s while it renders ->
timeouts, retries that re-trigger the work, exhausted server threads.
```

**Correct (accept, enqueue, return, report):**

```text
1. Accept the request, enqueue a job (queue + worker pool).
2. Return immediately with a job ID.
3. Client learns the result via polling, long-polling, or a push (SSE/WebSocket).
```

Make jobs **idempotent**, support **retries with backoff**, and route poison messages to a **dead-letter queue**. Scale workers independently based on **queue depth / consumer lag**, not request rate. This decouples slow work from the request path and keeps the API responsive.

## Sources

- `references/raw/learn/system-design/patterns/long-running-tasks.md` - queue + worker pattern, result delivery, idempotency, DLQ, lag-based scaling.
