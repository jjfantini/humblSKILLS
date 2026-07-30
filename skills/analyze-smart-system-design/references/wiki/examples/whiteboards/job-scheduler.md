---
title: "Whiteboard: Distributed Job Scheduler"
context: examples
category: whiteboards
concept: job-scheduler
description: "Reference design separating Jobs from Executions in DynamoDB, a Watcher pushing to SQS with delivery delay, and workers that update execution status."
tags: job-scheduler, dynamodb, sqs, cron, whiteboard
sources:
  - "references/raw/whiteboards/job_scheduler.png"
  - "references/raw/learn/system-design/problem-breakdowns/job-scheduler.md"
last_ingested: 2026-06-01
---

## Whiteboard: Distributed Job Scheduler

Shows the two-service + two-table core with queue-based timing as the deep dive for 2-second precision.

**Management path:**

```text
Client -> API Gateway -> Scheduler Service (create) / Query Service (status)
                      -> Job Store (DynamoDB)
```

**Job Store schemas:**

- **Jobs:** id, userId, taskId, params, schedule
- **Executions:** time, jobId, userId, status, attempt

**Execution path (deep dive):**

```text
Watcher (polls Jobs/Executions for jobs due soon)
       -> Message Queue (SQS)
       -> Worker (consume jobs) -> Execute Job (parallel instances)
       -> Update Status -> Executions table
```

**SQS implementation notes on the board:**

1. **Delivery delay** - hold message until exact execution time
2. **Retry with exponential backoff** - transient failure handling
3. **Partition by execution time** - scale queue processing

**Why this shape:** Scheduler + Query services satisfy CRUD. Watcher + SQS decouple "find due jobs" from "run on time" without hammering the DB every second at 10k jobs/s.

## Sources

- `references/raw/whiteboards/job_scheduler.png` - Jobs vs Executions split, Watcher, SQS timing tricks.
- `references/raw/learn/system-design/problem-breakdowns/job-scheduler.md` - time buckets, at-least-once, two-layer scheduler prose.
