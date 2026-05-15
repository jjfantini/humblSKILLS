---
name: use-smart-commit
description: >
  Inspect the working tree, group pending changes into atomic related buckets,
  and author one conventional commit per bucket. Every message covers what
  changed, why, and the impact (what it resolves or unblocks). Use when the
  user says "commit", "commit my changes", "make a commit", "stage and
  commit", or after finishing a unit of work and wanting to land it. Do NOT
  use for squashing existing history, rebasing, force-pushing, or writing PR
  descriptions.
license: MIT
metadata:
  author: jjfantini
  version: "1.0.0"
  tags: [git, commits, conventional-commits, workflow, humblskill]
  platforms: [claude-code, cursor]
  preserve:
    - references/decisions.md
    - references/log.md
    - references/patterns.md
---

# Smart Commit

Turn a messy `git status` into a clean, reviewable history. Read the diff, bucket changes by intent, write one conventional commit per bucket with a body that explains the **why** and the **impact**.

## Brain Protocol (read BEFORE creating anything)

1. `references/_index.md`       - what this skill knows (map)
2. `references/patterns.md`     - what worked, with numbers
3. `references/decisions.md`    - past reasoning, don't repeat mistakes
4. `references/log.md`          - last 5 session entries
5. Relevant `references/wiki/commit/<category>/` concepts per task

After completing work, UPDATE the brain:
- Performance data (if reported) -> `patterns.md`
- Non-obvious decisions -> `decisions.md`
- Session summary (always) -> append to `log.md`
- Run `scripts/lint.sh` to regenerate `_index.md` and verify structure

_Full spec: `references/_brain.md`._

## Workflow

1. **Inspect.** Run `git status` and `git diff` (and `git diff --staged` if anything is already staged). Read every hunk before deciding anything.
2. **Bucket by intent.** Group changes by *what they accomplish*, not by directory or file extension. A bug fix and its regression test belong together. A documentation tweak unrelated to the bug is its own commit. A drive-by typo fix in an unrelated file is its own commit.
3. **Draft each message.** For every bucket, decide `type(scope)`, write a subject line in the imperative mood (≤ 72 chars, no trailing period), then a blank line, then a 1-3 line body explaining **why** the change is being made and the **impact** (what it resolves, unblocks, or prevents). See `references/wiki/commit/messages/conventional-format.md` for the full anatomy.
4. **Ambiguity check.** If a single file mixes intents, if you can't decide which bucket a hunk belongs to, or if a judgment call is required (is this refactor part of the feature or its own commit?), **stop and ask the user before committing**. When grouping is obvious, proceed autonomously.
5. **Stage and commit.** For each bucket: `git add <specific files>` (never `git add -A` or `git add .` blindly), then commit with a HEREDOC so the body's newlines survive verbatim. Example:
   ```bash
   git commit -m "$(cat <<'EOF'
   fix(parser): handle empty input without panicking

   Previously the parser called .unwrap() on the first token,
   which panicked on empty stdin. Now it returns ParseError::Empty.
   Resolves #214 and unblocks the CLI's `--from-stdin` flow.
   EOF
   )"
   ```
6. **Repeat** until `git status` is clean.
7. **Confirm.** Show the user `git log --oneline -n <count>` so they can see what landed.

## Conventional commit types

The canonical conventional-commits set. Many automated tools (release-please, semantic-release, changesets, conventional-changelog, etc.) parse these to author changelogs and decide version bumps — but the types are valid even when no such tool is in use.

| Type       | Use when                                                  | Semver impact under conventional-commits |
|------------|-----------------------------------------------------------|------------------------------------------|
| `feat`     | New user-visible feature                                  | minor                                    |
| `fix`      | Bug fix                                                   | patch                                    |
| `perf`     | Performance improvement (no behaviour change)             | patch                                    |
| `refactor` | Internal restructure (no behaviour change)                | none                                     |
| `docs`     | Documentation only                                        | none                                     |
| `test`     | Tests only                                                | none                                     |
| `build`    | Build system, dependencies, packaging                     | none                                     |
| `ci`       | CI/CD pipeline only                                       | none                                     |
| `chore`    | Maintenance that doesn't fit above                        | none                                     |
| `style`    | Formatting / whitespace only                              | none                                     |

Breaking change: append `!` after the type/scope (`feat(api)!: ...`) **or** add a `BREAKING CHANGE:` footer in the body. Either form signals a major version bump under semver.

Scope is optional but encouraged. Pick whatever names the area of the repo being changed — common examples across repos: `(api)`, `(auth)`, `(ui)`, `(docs)`, `(ci)`. Check `git log --oneline -50` in the current repo to see which scopes are already in use and reuse them for consistency.

## Examples

### Example 1: bug fix bundled with its test

`git status` shows two changed files: `src/parser.rs` (the fix) and `tests/parser_test.rs` (a regression test for the bug). These share one intent — they ship as **one** commit.

```
fix(parser): handle empty input without panicking

The parser called .unwrap() on the first token, which panicked on
empty stdin. Now it returns ParseError::Empty so callers can recover.
Resolves #214 and unblocks the CLI's --from-stdin flow.
```

### Example 2: feature plus unrelated doc tweak

`git status` shows three changed files: `src/auth/oauth.go` and `src/auth/oauth_test.go` (a new feature) and `README.md` (fixing a typo unrelated to auth). These are **two** intents, so **two** commits.

Commit 1 (stage `src/auth/oauth.go` and `src/auth/oauth_test.go`):
```
feat(auth): add Google OAuth provider

Adds the missing OAuth flow for Google so users can sign in with
their Workspace accounts. Resolves the manual-account-creation step
that has been blocking enterprise onboarding.
```

Commit 2 (stage `README.md`):
```
docs: fix typo in installation step

The README pointed at the wrong package name in the brew install
example. Caught during onboarding review.
```

## DO NOT

- **Don't dump everything into one commit** with a generic subject like `update files`, `wip`, or `misc changes`. Read the diff and split.
- **Don't run `git add -A` or `git add .`** without first reading the diff. You will sweep in unrelated changes and lose atomicity.
- **Don't write a body that just restates the subject.** The body explains *why* and *what it resolves*. If you can't add new information, leave the body off.
- **Don't write the skip-CI token in any commit message** — not in the subject, not in the body, not as an explanation. Specifically, the literal strings `[skip ci]`, `[ci skip]`, `[no ci]`, `[skip actions]`, and `[actions skip]` are all parsed by GitHub Actions across the entire commit message and will suppress every workflow for that push. If you must discuss the mechanism in prose, refer to it as `skip-ci` (with a hyphen, no brackets) or wrap as `skip ci` in backticks without brackets.
- **Don't add AI attribution lines** like `Co-Authored-By: Claude <noreply@anthropic.com>` or `Generated with Claude Code` unless the user explicitly asks for them.
- **Don't bypass hooks** with `--no-verify`. If a pre-commit hook fails, fix the underlying issue and re-commit. The hook is there for a reason.
- **Don't amend a commit that has been pushed** without explicit user confirmation. Create a new commit instead.
- **Don't pick a generic scope** (`misc`, `stuff`, `update`) when a real one fits. Check `git log --oneline -50` to see which scopes the repo already uses and reuse them.

## How to Use

**Live enumeration of categories and concepts:**
Read `references/_index.md` (auto-regenerated by `scripts/lint.sh`).

**Brain protocol, naming conventions, writing principles, linking contract, ingest workflow, lint checks, `patterns.md` entry shape:**
Read `references/_brain.md`.

**Wiki concept file shape:**
Read `references/_template.md`.

### Primary workflows

**Group pending changes into atomic buckets:**
Read `references/wiki/commit/workflow/atomic-grouping.md`.

**Write a conventional commit message that covers why and impact:**
Read `references/wiki/commit/messages/conventional-format.md`.

**Avoid the common anti-patterns (blind bulk, skip-CI token, body-restates-subject):**
Read `references/wiki/commit/anti-patterns/avoid.md`.

## Success Signals

- Every commit on a clean branch maps to exactly one intent (revertable without breaking unrelated work).
- Every commit subject starts with a conventional type and stays under 72 characters.
- Every non-trivial commit has a body that explains *why* and *what it resolves*.
- Zero commits contain the skip-CI token in any form.
- `git log --oneline -n 10` is readable as a changelog.
