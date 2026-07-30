---
title: "Clear, Action-Oriented Title"
context: flow
category: setup
concept: example-concept
description: "One sentence explaining this concept and why it matters"
tags: git, worktree, workflow
sources:
  - "references/raw/user-request.md"
last_ingested: 2026-06-12
command: scripts/example.sh
---

## Concept

Explain one workflow concept. Start with the failure mode, then give the
operational rule an agent can execute.

**Incorrect:**

```bash
# Vague branch name and no isolation from parallel work.
git checkout -b changes
```

**Correct:**

```bash
# Branch and worktree share a conventional type and clear slug.
git worktree add ../feat-add-data -b feat/add-data origin/develop
```

## Sources

- `references/raw/user-request.md` - workflow requirements and defaults.
