---
title: "Gate Feature Branches Into Develop"
context: flow
category: pr
concept: develop-gate
description: "Merge feature work into develop only after local verification and CI are green"
tags: pull-request, develop, ci, verification
sources:
  - "references/raw/user-request.md"
last_ingested: 2026-06-12
---

## Develop Gate

The feature branch goes into `develop` first. This keeps integration testing
separate from the production branch.

Minimum gate:

1. Run the relevant local tests, linting, and verification commands.
2. Push the feature branch.
3. Open a PR with base `develop`.
4. Watch CI/CD until all required checks pass.
5. Merge only when local evidence and remote checks agree.

**Incorrect:**

```bash
gh pr merge --merge
# CI is still running or local verification was skipped.
```

**Correct:**

```bash
gh pr create --base develop --head feat/add-data
gh pr checks --watch
gh pr merge --merge
```

If checks fail, investigate the root cause in the failing logs and fix the
feature branch. Do not bypass checks or merge a red PR.

## Sources

- `references/raw/user-request.md` - develop PR and green CI/CD requirement.
