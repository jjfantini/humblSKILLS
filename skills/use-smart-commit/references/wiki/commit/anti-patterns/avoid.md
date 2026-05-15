---
title: "Commit Anti-Patterns to Avoid"
context: commit
category: anti-patterns
concept: avoid
description: "Concrete don'ts when authoring commits in this repo - blind bulk staging, the skip-CI token rule, body-restates-subject, unrequested AI attribution, and hook bypass."
tags: anti-patterns, dont-do, skip-ci, blind-bulk, attribution
sources: []
last_ingested: 2026-05-15
---

## Anti-patterns

These are the failure modes the SKILL exists to prevent. Each one has burned someone before.

### 1. The blind-bulk commit

**Incorrect:**

```bash
git add -A
git commit -m "update files"
```

`git add -A` sweeps in every change including unrelated work, generated artefacts, debug prints, and accidental edits. The subject `update files` says nothing. The commit is unrevertable in any useful sense — rolling it back would undo five unrelated things.

**Correct:**

Stage the specific files for one intent, write a real subject, repeat for the next bucket. See `commit/workflow/atomic-grouping.md`.

### 2. The skip-CI token

**Never write any of these literal strings in a commit message** — subject or body, anywhere:

- `[skip ci]`
- `[ci skip]`
- `[no ci]`
- `[skip actions]`
- `[actions skip]`

**Why:** GitHub Actions scans the **entire commit message** for these tokens — not just a trailer, not just the subject line. A single occurrence anywhere in any commit on a push suppresses every workflow for that push. This has bitten this repo before: a commit that explained how the skip-CI feature worked silently killed CI on its own push.

**If you must discuss the mechanism in prose**, write it as `skip-ci` (with a hyphen, no brackets), or wrap as `skip ci` in backticks **without** the square brackets. Example:

```
docs(ci): document the skip-ci marker behaviour

Explains how the skip-ci marker suppresses workflows when present
in a commit subject or body. See .github/docs/ci-controls.md.
```

The hyphenated form will never match GitHub's parser.

### 3. The body that restates the subject

**Incorrect:**

```
fix(parser): handle empty input

This commit handles empty input in the parser.
```

Drop the body if it adds nothing. The body should explain **why** the change is being made and the **impact** (what it resolves or unblocks). If you can't add new information, leave it off.

### 4. Unrequested AI attribution

Don't append `Co-Authored-By: Claude <noreply@anthropic.com>` or `Generated with Claude Code` footers unless the user explicitly asks for them. The repo's history doesn't carry these by default; adding them changes the commit signature and pollutes `git log --author` queries.

### 5. Bypassing hooks

**Incorrect:**

```bash
git commit --no-verify -m "fix(...): ..."
```

If a pre-commit hook fails, fix the underlying issue and re-commit. `--no-verify` skips lint, format, secret-scan, and any other gate the maintainer wired up. The right escape hatch is to fix the failure, not to suppress the gate.

### 6. Amending pushed commits without confirmation

`git commit --amend` rewrites history. If the original commit was already pushed, amending forces a force-push to align the remote, which can clobber teammates' work and breaks PR review threads. Default to a new commit; only amend a pushed commit when the user explicitly asks for it.

### 7. Generic or absent scopes

`misc`, `stuff`, `update`, `chore: changes` — these are scopes and subjects that hide intent. Pick what's already conventional in the current repo by checking `git log --oneline -50` for existing scopes and reusing them. If you can't think of a scope, omit it (`feat: ...` is valid without one) — but don't invent a meaningless one.
