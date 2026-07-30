---
title: "Pair Worktree And Branch Names"
context: flow
category: setup
concept: worktree-naming
description: "Use matched conventional names so isolated feature work stays obvious locally and remotely"
tags: git, worktree, branch, naming
sources:
  - "references/raw/user-request.md"
last_ingested: 2026-06-12
command: scripts/create-worktree.sh
---

## Worktree Naming

The worktree directory and branch should describe the same implementation. Use
a conventional type plus a direct slug.

Pattern:

```text
worktree: <type>-<slug>
branch:   <type>/<slug>
```

Good types: `feat`, `fix`, `docs`, `test`, `refactor`, `perf`, `build`, `ci`,
`chore`.

**Incorrect:**

```bash
git worktree add ../new-stuff -b updates origin/develop
```

This hides intent, makes cleanup risky, and does not show whether the work is a
feature, fix, or chore.

**Correct:**

```bash
git worktree add ../feat-add-data -b feat/add-data origin/develop
```

Default to worktrees when the current repo is dirty, multiple agents are active,
or the user wants reliable isolation. An in-place branch is acceptable only when
the user explicitly chooses it or the repo is clean and no parallel work exists.

## Command

```bash
bash scripts/create-worktree.sh feat add-data origin/develop
```

## Sources

- `references/raw/user-request.md` - naming examples and worktree default.
