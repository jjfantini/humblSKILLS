---
title: "Gather Workflow Intent Before Acting"
context: flow
category: setup
concept: intent-gathering
description: "Ask the routing questions once so the agent can execute without guessing later"
tags: intent, vibe, hitl, defaults
sources:
  - "references/raw/user-request.md"
last_ingested: 2026-06-12
---

## Intent Gathering

The failure mode is starting implementation before the merge, release, and
cleanup policy is known. Ask once at the beginning, then execute against those
answers.

Ask these questions before work:

1. Vibe mode or HITL mode?
2. Is parallel work happening in this repo, including other Codex, Claude, or
   Cursor agents?
3. Should the `develop` -> `main` or `master` PR auto-merge on green checks?
4. Should the generated release PR auto-merge on green checks?
5. Should stale worktrees and branches be cleaned up at the end?

If the user defers, use: Vibe mode, worktree isolation, auto-merge on green,
release PR auto-merge on green, cleanup enabled.

**Incorrect:**

```markdown
I'll just start coding and figure out the merge path after CI.
```

**Correct:**

```markdown
Defaulting because you deferred: Vibe mode, worktree isolation, auto-merge
main and release PR on green checks, cleanup enabled.
```

Before finalizing the worktree choice, also inspect local reality with
`git status`, `git worktree list`, branch tracking, and active terminals or
agent sessions. If anything suggests concurrent work, use a worktree.

## Sources

- `references/raw/user-request.md` - required upfront question set and defaults.
