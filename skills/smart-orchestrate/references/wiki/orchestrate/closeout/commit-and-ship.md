---
title: "Close-Out Is Parent-Only: Commit Per Gate, Ship Once"
context: orchestrate
category: closeout
concept: commit-and-ship
description: "One committer keeps history coherent and stops workers inventing ship policy"
tags: closeout, commit, ship, smart-commit, smart-worktree-flow
sources:
  - "references/raw/orchestrate-SKILL.md"
last_ingested: 2026-08-12
---

## Who Runs What, When

| When | Skill | Who |
|---|---|---|
| After each verified gate | `smart-commit` | Parent |
| Feature ready to ship | `smart-worktree-flow` (PR → develop → main, cleanup) | Parent |
| Plan still in flux | neither — replan first | Parent |

Workers never run these.

## Why the Parent Alone Commits

**Incorrect (workers commit their own gates):**

```
worker-a: "feat: rate limiter"
worker-b: "add limiter tests"          ← not conventional, invisible to release tooling
worker-c: "wip"                        ← ships to main as-is
```

Three models with three commit conventions produce history that release automation
reads wrong, changelogs that lose entries, and no single diff that maps to a phase.
The parent is the only agent holding the whole plan, so it is the only one that can
group changes into commits that mean something.

**Correct:**

Workers leave changes in the worktree and return the handoff contract. The parent
verifies the diff against the plan, then runs `smart-commit` — one commit boundary
per closed gate, authored with full context of what the phase was for.

## Shipping Is One Step, at the End

`smart-worktree-flow` owns PR, develop merge, main merge, release handling, and
cleanup. Do not interleave it with dispatch. A ship step run while gates are still
open merges a half-executed plan.

## Sources

- `references/raw/orchestrate-SKILL.md`
