---
title: "Bucket Pending Changes Into Atomic Commits"
context: commit
category: workflow
concept: atomic-grouping
description: "How to read git status/diff and split changes into one-intent buckets that are revertable without breaking unrelated work."
tags: atomic-commits, grouping, intent, git-add, staging
sources: []
last_ingested: 2026-05-15
---

## Atomic grouping

A commit is **atomic** when it represents one logical change that can be reverted on its own without breaking unrelated work. Atomicity is decided by *intent*, not by filesystem proximity.

### The procedure

1. Run `git status` to see every changed file.
2. Run `git diff` (and `git diff --staged` if anything is already staged) and read every hunk. Do not skim.
3. For each hunk, ask: **what is the one thing this change accomplishes?** Write that intent down in 3-6 words.
4. Group hunks that share an intent into one bucket. The same file can have hunks in different buckets — use `git add -p` to stage hunk-by-hunk when this happens.
5. Commit each bucket separately.

### Grouping heuristics

**Same bucket:**

- A bug fix and the test that exercises the regression
- A refactor and the simultaneous renames it forces (e.g., rename a function + update every caller)
- A new feature and the docs that describe it (only if the docs are *about* this feature — a typo fix is separate)
- A schema change and the migration that lands it

**Different buckets:**

- A feature + a typo fix in an unrelated README
- A bug fix + a drive-by formatting cleanup
- A dependency bump + a feature that happens to use the new dependency (separate the upgrade so it can be reverted independently)
- A new feature + tests for a *different* feature that you noticed were missing

### Counter-examples

**Incorrect (one commit, two intents):**

```bash
# git status:
#   modified: src/auth/oauth.go
#   modified: src/auth/oauth_test.go
#   modified: README.md         # unrelated typo
git add -A
git commit -m "feat: oauth and readme fix"
```

The OAuth feature and the README typo share no intent. Reverting the commit would also revert the typo fix.

**Correct (two commits, two intents):**

```bash
git add src/auth/oauth.go src/auth/oauth_test.go
git commit -m "feat(auth): add Google OAuth provider ..."

git add README.md
git commit -m "docs: fix typo in installation step"
```

### When a single file straddles two intents

Use `git add -p` to stage individual hunks, or check out the file fresh with `git checkout -p` to selectively keep changes. If splitting hunks feels arbitrary or requires you to interpret the user's intent, **stop and ask the user** before committing — this is the ambiguity escape hatch in the SKILL.md workflow.

### Why atomicity matters

- **Revertable.** `git revert <sha>` undoes exactly one intent.
- **Bisectable.** `git bisect` lands on the commit that introduced a regression, not on a sweep that mixed five things.
- **Reviewable.** Reviewers can reason about one change at a time. Mixed commits force them to context-switch mid-diff.
- **Changelog quality.** Automated changelog tools (release-please, semantic-release, changesets, conventional-changelog) read commit messages to author release notes. Atomic commits make the changelog readable; human-authored changelogs benefit just as much.
