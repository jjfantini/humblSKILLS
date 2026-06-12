---
title: "Handle Generated Release PRs"
context: flow
category: release
concept: release-pr
description: "Decide whether the agent or user owns the release PR before the main merge"
tags: release, release-please, version, brew
sources:
  - "references/raw/user-request.md"
last_ingested: 2026-06-12
---

## Release PR

Some repos generate a release PR after `main` or `master` receives a
conventional commit. The release PR usually updates changelog and version files;
merging it cuts the tag and starts artifact publication.

Ask up front whether the user wants the agent to merge this PR on green checks.
If they choose manual release review, stop after the release PR is ready and
report its URL plus check state.

**Incorrect:**

```bash
# Main was merged, a release PR appeared, and the agent silently ignores it.
```

**Correct:**

```bash
gh pr list --search "release-please" --state open
gh pr checks --watch <release-pr-number>
gh pr merge <release-pr-number> --merge
```

After the release PR merges, verify the tag, release workflow, package
artifacts, and Homebrew tap update before claiming the release is available.

## Sources

- `references/raw/user-request.md` - release PR merge-or-manual decision.
