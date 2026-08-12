---
title: "Isolate in a Worktree Before Any Worker Runs"
context: orchestrate
category: isolation
concept: worktree-first
description: "Prevents cross-CLI agents from colliding in one dirty checkout"
tags: worktree, isolation, parallel, smart-worktree-flow, git
sources:
  - "references/raw/orchestrate-SKILL.md"
last_ingested: 2026-08-12
---

## Isolation Is the Default, Not an Optimization

**Incorrect (farming across CLIs into one checkout):**

```
Parent (Claude)  ── dispatch ──> Codex worker  ┐
                 ── dispatch ──> Cursor worker ├─> ~/proj  (main checkout)
                 ── dispatch ──> Sonnet worker ┘
```

Three agents write to one working tree at once. Diffs interleave, `git status` is
meaningless, and the user's own uncommitted work is in the blast radius. Nothing
errors — the tree just quietly stops representing anyone's intent.

**Correct (paired worktree per run):**

1. Parent follows `smart-worktree-flow` to create a paired worktree + branch
2. Every worker is pointed at that worktree path — or a separate worktree when the
   subtasks are truly independent
3. Parallel workers get **non-overlapping file scopes** inside the same worktree,
   or separate worktrees when scopes would collide
4. Parent alone merges, opens PRs, and cleans up via `smart-worktree-flow`

## When Skipping Isolation Is Legitimate

Only two cases:

- The user explicitly wants in-place work on the current checkout
- The change is a trivial single-file edit with no parallel workers

Anything else — multiple workers, multiple CLIs, a dirty tree, or a change that
spans files — gets a worktree.

## Scope Collision Is the Failure to Design Against

Two parallel workers on the same file is the one collision a worktree cannot
prevent, because they share it. Decide file ownership at dispatch time, in the
brief's `Scope` and `Out of scope` lines, not after the diffs come back.

## Sources

- `references/raw/orchestrate-SKILL.md`
