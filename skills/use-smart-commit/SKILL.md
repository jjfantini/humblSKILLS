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
3. **Draft each message.** For every bucket, decide `type(scope)`, write a subject line in the imperative mood (≤ 72 chars, no trailing period), then a blank line, then a structured body with three labeled sections (`Changed:`, `Why:`, `Impact:`) — see the **Body format** section below. For trivial commits (single-character typo fix, formatting-only, dependency bump with no behaviour change, auto-generated file regeneration), the labeled structure is optional and a free-form one-paragraph body is fine. See `references/wiki/commit/messages/conventional-format.md` for the full anatomy.
4. **Ambiguity check.** If a single file mixes intents, if you can't decide which bucket a hunk belongs to, or if a judgment call is required (is this refactor part of the feature or its own commit?), **stop and ask the user before committing**. When grouping is obvious, proceed autonomously.
5. **Stage and commit.** For each bucket: `git add <specific files>` (never `git add -A` or `git add .` blindly), then commit with a HEREDOC so the body's newlines survive verbatim. Append the authorship footer (see below) on a blank line after the body. Example:
   ```bash
   git commit -m "$(cat <<'EOF'
   fix(parser): handle empty input without panicking

   Changed:
   ParseError::Empty now returned from parse_first_token() instead
   of an unwrap() panic.

   Why:
   Empty stdin was reaching the CLI's --from-stdin flow and crashing
   the process with a non-actionable error.

   Impact:
   Resolves #214. Unblocks --from-stdin for piped workflows and
   removes the last unwrap() on the parser hot path.

   Authored by humblSKILLS; "use-smart-commit"
   EOF
   )"
   ```
6. **Repeat** until `git status` is clean.
7. **Confirm.** Show the user `git log --oneline -n <count>` so they can see what landed.

## Body format

Every non-trivial commit body uses three labeled sections, each on its own line, content on the line below, separated by blank lines:

```
Changed:
<one or two sentences naming what concretely changed in the code>

Why:
<one or two sentences on the motivation — the problem, constraint, or user need>

Impact:
<one or two sentences on what this resolves, unblocks, or prevents>
```

Rules:

- **Labels go on their own line**; content starts on the next line. Do not write `Changed: ParseError::Empty now ...` on one line.
- Wrap content at ~72 characters.
- Use **plain labels, not markdown headers** (`Changed:` not `## Changed:`). Git's default `commentChar` is `#`, which silently strips `#`-prefixed lines on editor-based amends. Plain labels survive every git operation and stay greppable: `git log --grep "^Why:"`.
- One blank line between sections.
- The authorship footer follows the last section after one blank line.

### When the structure is optional

Free-form one-paragraph bodies are acceptable for **trivial commits only**:

- Single-character / single-word typo fix
- Formatting-only changes (`style:` commits, lockfile re-sorts)
- Dependency version bump with no behaviour change
- Auto-generated file regeneration (registry, lockfiles, generated types)
- Commits where the subject line is genuinely self-explanatory and Changed/Why/Impact would just paraphrase it

If you're not sure whether a commit qualifies as trivial, use the structure.

## Authorship footer

Every commit authored by this skill ends with:

```
Authored by humblSKILLS; "use-smart-commit"
```

on its own line, separated from the body above by one blank line. This is **on by default** — it marks the skill's authorship in the commit's history so the user can later filter or audit which commits the skill wrote.

### How to turn it off

The user can disable the footer in two scopes:

- **For this conversation only.** If the user says any of "no footer", "skip the footer", "drop the humblSKILLS footer", "no authorship line", or similar, omit the footer for every remaining commit this session. Do not ask again in this session.
- **Persistently across sessions.** If the user says any of "always skip the footer", "never add the footer", "permanently disable the footer", or similar, omit the footer **and** save a feedback memory entry recording the preference (per the auto-memory protocol — write to a new file under the user's memory dir, then add a one-line pointer to `MEMORY.md`). Future sessions read that memory at conversation start and respect it.

Before authoring any commit, check this in order:
1. Memory file pointing at "disable use-smart-commit footer" → omit
2. User said "no footer" earlier in this conversation → omit
3. Otherwise → include the footer

### What the footer is NOT

This is **not** an AI attribution line. Do not also append `Co-Authored-By: Claude <noreply@anthropic.com>` or `Generated with Claude Code` — those remain opt-in and require an explicit user request (see the DO-NOT section).

## Conventional commit types and semver

The canonical conventional-commits set, with each type's [semver](https://semver.org) bump. Many automated tools (release-please, semantic-release, changesets, conventional-changelog, etc.) parse the type to decide version bumps — but the mapping is correct even when no such tool is in use.

> **Semver primer.** Given a version `MAJOR.MINOR.PATCH`, increment:
> - **MAJOR** when you make incompatible API changes
> - **MINOR** when you add functionality in a backward-compatible manner
> - **PATCH** when you make backward-compatible bug fixes
>
> Pre-release and build metadata extend the format: `1.4.0-rc.1`, `1.4.0+sha.abc123`.

| Type       | Use when                                              | Semver bump |
|------------|-------------------------------------------------------|-------------|
| `feat`     | New user-visible feature (backward-compatible)        | **minor**   |
| `fix`      | Bug fix (backward-compatible)                         | **patch**   |
| `perf`     | Performance improvement (no behaviour change)         | **patch**   |
| `refactor` | Internal restructure (no behaviour change)            | none        |
| `docs`     | Documentation only                                    | none        |
| `test`     | Tests only                                            | none        |
| `build`    | Build system, dependencies, packaging                 | none        |
| `ci`       | CI/CD pipeline only                                   | none        |
| `chore`    | Maintenance that doesn't fit above                    | none        |
| `style`    | Formatting / whitespace only                          | none        |

### Breaking changes (MAJOR)

**Any change that breaks backward compatibility forces a MAJOR bump, regardless of the type prefix.** Removing a public function, renaming a CLI flag, changing a wire format, dropping support for a Node version — these are breaking, even if the surrounding commit feels like a `fix` or a `refactor`.

Mark a breaking change with either form:

1. Append `!` after the type/scope: `feat(api)!: drop deprecated /v1 endpoint`
2. Add a `BREAKING CHANGE:` footer in the body (described in the body section of the wiki).

Either form signals MAJOR to release tooling and to humans reading the log.

### Scopes

Scope is optional but encouraged. Pick whatever names the area being changed — common examples across repos: `(api)`, `(auth)`, `(ui)`, `(docs)`, `(ci)`. Check `git log --oneline -50` in the current repo to see which scopes are already in use and reuse them for consistency.

## Examples

### Example 1: bug fix bundled with its test (non-trivial → structured body)

`git status` shows two changed files: `src/parser.rs` (the fix) and `tests/parser_test.rs` (a regression test for the bug). These share one intent — they ship as **one** commit.

```
fix(parser): handle empty input without panicking

Changed:
ParseError::Empty now returned from parse_first_token() instead
of an unwrap() panic, with a regression test covering empty stdin.

Why:
Empty stdin was reaching the CLI's --from-stdin flow and crashing
the process with a non-actionable error.

Impact:
Resolves #214. Unblocks --from-stdin for piped workflows and
removes the last unwrap() on the parser hot path.

Authored by humblSKILLS; "use-smart-commit"
```

### Example 2: feature plus unrelated doc tweak (two commits — one structured, one trivial)

`git status` shows three changed files: `src/auth/oauth.go` and `src/auth/oauth_test.go` (a new feature) and `README.md` (fixing a typo unrelated to auth). These are **two** intents, so **two** commits.

Commit 1 (stage `src/auth/oauth.go` and `src/auth/oauth_test.go`) — non-trivial, full structure:
```
feat(auth): add Google OAuth provider

Changed:
New GoogleOAuthProvider implementing the AuthProvider interface,
wired into the auth router and covered by oauth_test.go.

Why:
Workspace customers were stuck in manual account creation because
no SSO option existed. Sales has been blocked on this for two weeks.

Impact:
Unblocks enterprise onboarding. Closes the OAuth gap relative to
competing tools and removes the manual-account-creation handoff.

Authored by humblSKILLS; "use-smart-commit"
```

Commit 2 (stage `README.md`) — trivial typo fix, free-form body is fine:
```
docs: fix typo in installation step

The README pointed at the wrong package name in the brew install
example. Caught during onboarding review.

Authored by humblSKILLS; "use-smart-commit"
```

(Footer omitted in your examples? See the **Authorship footer** section above for how to disable it.)

## DO NOT

- **Don't dump everything into one commit** with a generic subject like `update files`, `wip`, or `misc changes`. Read the diff and split.
- **Don't run `git add -A` or `git add .`** without first reading the diff. You will sweep in unrelated changes and lose atomicity.
- **Don't write a body that just restates the subject.** The body explains *why* and *what it resolves*. If you can't add new information, leave the body off.
- **Don't write the skip-CI token in any commit message** — not in the subject, not in the body, not as an explanation. Specifically, the literal strings `[skip ci]`, `[ci skip]`, `[no ci]`, `[skip actions]`, and `[actions skip]` are all parsed by GitHub Actions across the entire commit message and will suppress every workflow for that push. If you must discuss the mechanism in prose, refer to it as `skip-ci` (with a hyphen, no brackets) or wrap as `skip ci` in backticks without brackets.
- **Don't add AI attribution lines** like `Co-Authored-By: Claude <noreply@anthropic.com>` or `Generated with Claude Code` unless the user explicitly asks for them. (Note: the `Authored by humblSKILLS; "use-smart-commit"` line is the skill's own authorship footer — see the **Authorship footer** section — and is separate from AI attribution.)
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
