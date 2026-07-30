---
title: "Sync And Prune After Release"
context: flow
category: cleanup
concept: sync-and-prune
description: "Leave local branches, remote branches, and worktrees clean after the release"
tags: cleanup, sync, branch, worktree
sources:
  - "references/raw/user-request.md"
last_ingested: 2026-06-12
command: scripts/cleanup.sh
---

## Sync And Prune

The flow is not done when the release merges. The local repo must reflect
upstream, and stale branches or worktrees should be removed by default.

Default cleanup:

1. Fetch origin.
2. Fast-forward local `main` or `master`.
3. Fast-forward local `develop` to upstream or to the released production
   branch, depending on the repo's branching policy.
4. Remove the feature worktree.
5. Delete the local feature branch.
6. Delete the remote feature branch.
7. Run `git worktree prune`.

**Incorrect:**

```bash
# Release shipped, but local branches and remote feature branch remain stale.
git status
```

**Correct:**

```bash
bash scripts/cleanup.sh feat/add-data ../feat-add-data
```

In HITL mode, ask before cleanup if the user opted out. Otherwise cleanup is
the default because commit history remains in `develop`, `main`, and tags.

## Sources

- `references/raw/user-request.md` - cleanup requirement and default behavior.
