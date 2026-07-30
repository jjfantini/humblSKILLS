---
title: "Design a Distributed Job Scheduler"
context: breakdown
category: core
concept: job-scheduler
description: "Schedule immediate, future, and recurring jobs at 10k/s within 2s of target; separate job definitions from executions and use a DB-query + queue scheduler."
tags: job-scheduler, cron, queue, at-least-once, time-bucket
sources:
  - "references/raw/whiteboards/job_scheduler.png"
  - "references/raw/learn/system-design/problem-breakdowns/job-scheduler.md"
last_ingested: 2026-06-01
---

## Design a Distributed Job Scheduler

**Functional:** schedule jobs to run immediately, at a future date, or on a recurring (CRON) schedule; monitor job status.

**Non-functional:** availability > consistency; execute within 2s of scheduled time; 10k jobs/s; at-least-once execution.

**Core entities:** Task (reusable unit of work), Job (a scheduled instance), Schedule (CRON or DateTime), User.

**Key API:**

```text
POST /jobs { task_id, schedule, parameters }
GET  /jobs?user_id=&status=&start=&end= -> Job[]
```

**High-level design (from whiteboard):**

```text
Client -> API Gateway -> Scheduler Service (create) / Query Service (status)
                      -> Job Store (DynamoDB): Jobs + Executions tables
Watcher (polls due jobs) -> Message Queue (SQS) -> Worker -> Execute Job -> Update Executions
```

Key insight: **separate job definition from execution instances** (Jobs vs Executions tables in DynamoDB).

**Key deep dives (visible on whiteboard):**
- **Watcher + SQS:** decouple "find due jobs" from "run on time"; avoids polling DB every second at 10k/s.
- **SQS delivery delay:** hold message until exact execution time (2s precision).
- **Retry with exponential backoff** on the queue; partition by execution time for scale.
- **At-least-once:** idempotent tasks; status + attempt tracked on Executions table.

See also `wiki/examples/whiteboards/job-scheduler.md`.

## Sources

- `references/raw/whiteboards/job_scheduler.png` - Jobs/Executions split, Watcher, SQS timing.
- `references/raw/learn/system-design/problem-breakdowns/job-scheduler.md` - time buckets, two-layer scheduler, retries prose.
