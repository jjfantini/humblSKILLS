---
title: "The Worker Agent: One Gated Subtask, One Contract Back"
context: orchestrate
category: roles
concept: worker-agent
description: "Bounded worker authority is what makes cheap models safe to dispatch to"
tags: worker, subagent, scope, cost, delegation
sources:
  - "references/raw/orchestrate-SKILL.md"
last_ingested: 2026-08-12
---

## A Worker Executes One Brief and Returns One Contract

Prefer the smallest, fastest model that can finish the brief cleanly:

- Grok 4.6, Composer 2.5
- GPT-5.6 Terra, Luna
- Claude Sonnet 5, Opus 5 (reserve Opus for the harder worker slots)

**Incorrect (unbounded worker):**

```
Parent: "Take the rate limiting work."
Worker: refactors the router, adds a config system, commits three times,
        opens a PR, and replies with a 4,000-token transcript.
```

The parent now owns a design it never approved, in commits it did not write, and
must read a transcript to find out what happened.

**Correct (bounded worker):**

Owns exactly three things:

1. Execute one gated subtask from the parent's brief, inside the assigned worktree
2. Stay inside the given scope and files
3. Return the handoff contract — summary only, not a full transcript

Workers do **not** re-plan the whole feature, expand scope, commit, open PRs, or
touch the main checkout without escalating.

## Why the Limits Are the Point

Cheap models are safe to dispatch to precisely because their authority is small.
Every constraint removed from a worker is a constraint the parent must re-verify
by hand, which cancels the cost saving that motivated delegation.

## Sources

- `references/raw/orchestrate-SKILL.md`
