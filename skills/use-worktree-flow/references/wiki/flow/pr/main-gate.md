---
title: "Gate Develop Into Main Or Master"
context: flow
category: pr
concept: main-gate
description: "Honor auto-merge versus human review before production branch merges"
tags: pull-request, main, master, auto-merge, hitl
sources:
  - "references/raw/user-request.md"
last_ingested: 2026-06-12
---

## Main Gate

After `develop` contains the feature, open the production PR from `develop` to
`main` or `master`. This gate follows the user's upfront mode decision.

Vibe mode means the agent may merge when tests, linting, verifications, and
CI/CD are green. HITL mode means the agent gathers the evidence, reports it,
and waits for explicit approval before merging.

**Incorrect:**

```bash
# User asked for manual review, but the agent merges anyway.
gh pr merge --merge
```

**Correct:**

```bash
gh pr create --base main --head develop
gh pr checks --watch
# In HITL mode, stop here and ask for approval.
```

Prefer merge commits when release tooling such as release-please needs the
original conventional commits on the production branch. Avoid squash-merging
away the `feat:` or `fix:` commit that should drive the version bump.

## Sources

- `references/raw/user-request.md` - auto-merge and manual review requirements.
