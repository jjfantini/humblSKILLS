---
title: "Handle Generated Release PRs"
context: flow
category: release
concept: release-pr
description: "Develop cuts a pre-release; main cuts the stable version and updates brew"
tags: release, release-please, version, brew, prerelease
sources:
  - "references/raw/user-request.md"
last_ingested: 2026-09-01
---

## Two release PRs

humblSKILLS release-please runs on both integration branches. Each opens its
own release PR (changelog + `.release-please-manifest.json`). Merging the PR
is what creates the tag and starts GoReleaser.

| Base | Tag | GitHub Release | Homebrew tap |
|---|---|---|---|
| `develop` | `vX.Y.Z-pre.N` | pre-release | `humblskills-pre` only |
| `main` | `vX.Y.Z` | latest / stable | `humblskills` (stable formula) |

Ask up front whether the user wants the agent to merge these PRs on green
checks. If they choose manual release review, stop after each PR is ready and
report its URL plus check state.

Same-major bumps auto-merge when repo automation is enabled. A major bump
(`2.x` → `3.0.0` or `3.0.0-pre.1`) is left for a human.

**Incorrect:**

```bash
# develop was merged, a pre-release PR appeared, and the agent silently ignores it.
# Or: main was merged and brew is claimed updated before the stable release PR landed.
```

**Correct:**

```bash
gh pr list --search "release-please" --state open
gh pr checks --watch <release-pr-number>
gh pr merge <release-pr-number> --merge
```

After the **develop** release PR merges, verify the `vX.Y.Z-pre.N` tag and that
the GitHub Release is marked pre-release. Optional tester check:
`brew upgrade humblskills-pre` (or `humblskills upgrade --channel beta`). Do
**not** run `brew upgrade humblskills` — that formula must stay on the last
stable.

After the **main** release PR merges, verify the stable tag and artifacts, then
run `brew upgrade humblskills` as a **post-check**. Confirm
[homebrew-humbl](https://github.com/jjfantini/homebrew-humbl) `Formula/humblskills.rb`
matches that version before claiming the release is available.

Users switch channels with the same profile field Homebrew and `upgrade` read:
`humblskills profile set channel beta`, `profile get channel`, or the existing
Profile TUI (`humblskills` → Profile → **install channel**). Unset means
stable.

## Sources

- `references/raw/user-request.md` - release PR merge-or-manual decision.
