---
title: "The Parent Orchestrator: Seven Owned Responsibilities"
context: orchestrate
category: roles
concept: parent-orchestrator
description: "Keeps one owner holding context and policy, so frontier tokens buy judgment instead of typing"
tags: parent, frontier, planning, ownership, orchestration
sources:
  - "references/raw/orchestrate-SKILL.md"
last_ingested: 2026-08-12
---

## The Parent Owns Judgment, Not Typing

One agent holds the plan, the context across workers, and all commit/ship
authority. Everything else is delegable.

**Incorrect (parent as super-worker):**

```
Parent (Fable 5): reads 40 files, writes the middleware, writes the config,
                  writes the tests, commits.
Workers: none.
```

Frontier pricing for mechanical edits, and no gate anywhere — the plan and the
implementation are the same undifferentiated stream, so nothing is verifiable
against anything.

**Correct (parent as owner of seven things):**

1. **Plan** — understand the goal; break into phases and gated subtasks
2. **Isolate** — open a worktree via `smart-worktree-flow` before dispatch;
   workers never touch the user's main checkout
3. **Route** — assign each subtask a worker model by scope, difficulty, blast radius
4. **Dispatch** — farm clear, bounded work to worker agents/CLIs inside the worktree
5. **Glue** — hold context across workers; resolve conflicts; keep one coherent design
6. **Verify** — read worker handoffs, run checks, confirm the change fits the system
7. **Close out** — after each gate run `smart-commit`; when shipping run
   `smart-worktree-flow` for PR/merge/cleanup

## Model Tier for the Parent Slot

| CLI    | Parent model |
|--------|--------------|
| Claude | Opus 5 on max is the default — same or better performance than Fable 5 and much cheaper. Reach for Fable 5 (high–max) only when the user asks for it. |
| Codex  | GPT-5.6 Sol Max |
| Cursor | Auto Intelligence, or Grok 4.6 High |

**Ultracode is a spend decision, not a quality dial.** Use it only for a massive
refactor touching many services and e2e flows, and **always confirm with the user
first** — it can incur large cost. Planning with Opus 5 on ultracode is often the
better trade than running Fable 5 throughout.

## The Parent May Delegate Planning

The entry-point model does not have to produce the plan itself. It may delegate
planning or plan refinement to another capable agent, then retain ownership of the
resulting plan, execution gates, integration, and final verification. Ownership is
the invariant; authorship is not.

## Sources

- `references/raw/orchestrate-SKILL.md`
