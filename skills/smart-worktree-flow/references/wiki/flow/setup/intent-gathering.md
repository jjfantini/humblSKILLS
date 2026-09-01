---
title: "Gather Workflow Intent Before Acting"
context: flow
category: setup
concept: intent-gathering
description: "Ask the routing questions once so the agent can execute without guessing later"
tags: intent, vibe, hitl, defaults
sources:
  - "references/raw/user-request.md"
last_ingested: 2026-09-01
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
4. Should the generated release PRs auto-merge on green checks? There are two:
   develop cuts `vX.Y.Z-pre.N` (GitHub pre-release + `humblskills-pre` formula);
   main cuts the real `vX.Y.Z` and updates the stable `humblskills` formula.
   `brew upgrade humblskills` is a post-check after the main stable only.
5. Should stale worktrees and branches be cleaned up at the end?

If the user defers, use: Vibe mode, worktree isolation, auto-merge on green,
both release PRs auto-merge on green, brew post-check after main, cleanup
enabled.

**Incorrect:**

```markdown
I'll just start coding and figure out the merge path after CI.
```

**Correct:**

```markdown
Defaulting because you deferred: Vibe mode, worktree isolation, auto-merge
develop→main and both release PRs on green checks, brew post-check after
main only, cleanup enabled.
```

Before finalizing the worktree choice, also inspect local reality with
`git status`, `git worktree list`, branch tracking, and active terminals or
agent sessions. If anything suggests concurrent work, use a worktree.

## Sources

- `references/raw/user-request.md` - required upfront question set and defaults.
