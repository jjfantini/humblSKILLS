---
title: "Approved TweetClaw Actions"
context: tweetclaw
category: workflows
concept: approved-actions
description: "Route TweetClaw reads and write-like actions through safe OpenClaw approval boundaries"
tags: tweetclaw, openclaw, approval, twitter, x
sources: []
last_ingested: 2026-06-09
---

## Approved TweetClaw Actions

TweetClaw belongs at the boundary between agent planning and X/Twitter account
work. The agent decides what is needed, the operator approves write-like
actions, and TweetClaw performs the specific fetch or action through OpenClaw.

**Incorrect:**

```text
Post this thread now using TweetClaw.
```

This skips exact operator review and treats a write-like action as routine.

**Correct:**

```text
Draft the thread, show the exact final text, wait for approval, then use
TweetClaw only after the operator approves.
```

This keeps account-changing actions inside an explicit review boundary.

## Routing Rules

- Read-only jobs: search tweets, search replies, inspect users, export
  followers, download media, collect monitor context, or fetch giveaway inputs
  with the minimum query or identifier needed.
- Write-like jobs: post tweets, post replies, send DMs, upload media, start
  monitors, configure webhooks, or run giveaway draws only after exact operator
  approval.
- Setup jobs: if TweetClaw is missing, point the operator to the npm install and
  runtime inspection commands in `SKILL.md`.
- Safety jobs: never place credentials, cookies, account secrets, or raw session
  material into prompts, issues, logs, or this skill.

## Sources

- No raw sources yet. Public installation and package references are linked from
  `SKILL.md`.
