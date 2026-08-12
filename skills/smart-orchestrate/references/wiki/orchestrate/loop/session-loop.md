---
title: "The Nine-Step Orchestration Session Loop"
context: orchestrate
category: loop
concept: session-loop
description: "One ordered loop keeps gates closed in sequence instead of overlapping into mush"
tags: loop, workflow, gates, dispatch, verification
sources:
  - "references/raw/orchestrate-SKILL.md"
last_ingested: 2026-08-12
---

## The Loop

1. **Clarify goal** — feature, fix, goal, or loop; success criteria
2. **Plan or delegate planning** — phases, dependencies, risks, verification points
3. **Isolate** — `smart-worktree-flow` worktree + branch (unless the user opts out)
4. **Gate subtasks** — each task carries: scope, brief, model tier, done-when,
   return contract
5. **Dispatch workers** — one subtask per worker; no overlapping file ownership
   when avoidable
6. **Collect handoffs** — reject incomplete returns; re-dispatch or escalate
7. **Integrate + verify** — parent reviews diffs against the plan and the larger system
8. **Close the gate** — parent runs `smart-commit` for that phase; **only then**
   open the next gate
9. **Ship or replan** — when done, parent runs `smart-worktree-flow` for
   PR/merge/cleanup; if reality diverges, parent revises the plan before more dispatch

## Step 8 Is the Load-Bearing One

**Incorrect (gates left open):**

```
Dispatch phase 1 ─┐
Dispatch phase 2 ─┼─ all in flight, nothing committed
Dispatch phase 3 ─┘
→ phase 2 depended on phase 1's shape, which changed under it
```

With no commit between phases there is no verified baseline to build on, so a
late phase silently invalidates an early one and the parent cannot tell which
diff is at fault.

**Correct:**

Verify phase 1, `smart-commit` it, *then* dispatch phase 2. Each commit is the
checkpoint the next gate is briefed against.

## Step 9's Second Branch Is Not Optional

When reality diverges from the plan, replan **before** more dispatch. Continuing to
dispatch against a stale plan converts one wrong assumption into N worker diffs
that all encode it.

## Sources

- `references/raw/orchestrate-SKILL.md`
