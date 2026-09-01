---
title: "Handle Generated Release PRs"
context: flow
category: release
concept: release-pr
description: "Develop cuts a pre-release; main opens a stable PR from the last stable, not the -pre tag"
tags: release, release-please, version, brew, prerelease
sources:
  - "references/raw/user-request.md"
last_ingested: 2026-09-01
---

## Two release PRs

humblSKILLS release-please runs on both integration branches. Each opens its
own release PR (changelog + that branch's manifest). Merging the PR is what
creates the tag and starts GoReleaser. Merging `develop` into `main` is **not**
the tag: it is the commit that lets main's release-please open the stable PR.

| Base | Manifest | Tag | GitHub Release | Homebrew tap |
|---|---|---|---|---|
| `develop` | `.release-please-manifest.develop.json` | `vX.Y.Z-pre.N` | pre-release | `humblskills-pre` only |
| `main` | `.release-please-manifest.json` | `vX.Y.Z` | latest / stable | `humblskills` (stable formula) |

The manifests are split because release-please matches GitHub releases to the
version in the manifest, then considers only commits **after that tag**. If
main records `2.52.0-pre`, it treats `v2.52.0-pre` as already released and
skips with "No user facing commits found" — even though the feat commits that
produced the pre are on `main`. The develop→main merge subject is not a
conventional commit, so it does not count either. Toggling `prerelease:
false` does not bypass that empty-changelog gate.

So: **do not write a `-pre` version into `.release-please-manifest.json`.**
That file is last **stable** only. After a pre is cut, promote by merging
`develop` → `main` (merge commit). Then wait for the **stable** release PR
and merge it. A tag does not appear from the promote merge alone.

Ask up front whether the user wants the agent to merge these PRs on green
checks. If they choose manual release review, stop after each PR is ready and
report its URL plus check state.

Same-major bumps auto-merge when repo automation is enabled. A major bump
(`2.x` → `3.0.0` or `3.0.0-pre.1`) is left for a human.

**Incorrect:**

```bash
# develop was merged, a pre-release PR appeared, and the agent silently ignores it.
# Or: develop was merged to main and brew is claimed updated because "a release
# always appears." The stable release PR is what cuts vX.Y.Z and brew.
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

After `develop` is merged to `main`, wait for the **main** release PR. After
*that* PR merges, verify the stable tag and artifacts, then run
`brew upgrade humblskills` as a **post-check**. Confirm
[homebrew-humbl](https://github.com/jjfantini/homebrew-humbl) `Formula/humblskills.rb`
matches that version before claiming the release is available.

Users switch channels with the same profile field Homebrew and `upgrade` read:
`humblskills profile set channel beta`, `profile get channel`, or the existing
Profile TUI (`humblskills` → Profile → **install channel**). Unset means
stable.

## Sources

- `references/raw/user-request.md` - release PR merge-or-manual decision.
