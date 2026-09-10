---
title: "The Nine-Step Orchestration Session Loop"
context: orchestrate
category: loop
concept: session-loop
description: "One ordered loop keeps gates closed in sequence instead of overlapping into mush"
tags: loop, workflow, gates, dispatch, verification
sources:
  - "references/raw/orchestrate-SKILL.md"
  - "references/raw/anthropic-prompting-claude-fable-5-1-2026-09-10.md"
last_ingested: 2026-09-10
---

## The Loop

1. **Clarify goal** — feature, fix, goal, or loop; success criteria
2. **Plan or delegate planning** — phases, dependencies, risks, verification points
3. **Isolate** — `smart-worktree-flow` worktree + branch (unless the user opts out)
4. **Gate subtasks** — each task carries: scope, brief, model **and effort**,
   done-when, return contract
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

## A Gate Is a Commit Boundary, Not a Blocking Boundary

Step 8 does not mean the parent sits idle. Anthropic's guidance is explicit that
on coding tasks, "letting the lead continue while subagents run lowers average
time to completion at similar quality, token usage, and cost" — so a dispatch
tool that returns immediately, with results arriving in a later message, is the
better harness shape.

Both hold at once, because they constrain different things:

| | May overlap | Must not overlap |
|---|---|---|
| Within one gate | Parent dispatches worker B, reads code, prepares the next brief while worker A runs | — |
| Across gates | — | Gate 2 may not open until gate 1 is verified and committed |

Non-blocking dispatch buys wall clock *inside* a gate. Committing between gates
buys a verified baseline *across* them. Conflating the two produces the failure
below — three phases in flight with nothing committed — and calls it
parallelism.

## Step 9's Second Branch Is Not Optional

When reality diverges from the plan, replan **before** more dispatch. Continuing to
dispatch against a stale plan converts one wrong assumption into N worker diffs
that all encode it.

## Sources

- `references/raw/orchestrate-SKILL.md` — the nine steps.
- `references/raw/anthropic-prompting-claude-fable-5-1-2026-09-10.md` — "Let the
  lead agent keep working while subagents run", the source of the
  commit-boundary / blocking-boundary distinction.
