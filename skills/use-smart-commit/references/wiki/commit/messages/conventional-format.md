---
title: "Conventional Commit Format: Type, Scope, Subject, Body"
context: commit
category: messages
concept: conventional-format
description: "The anatomy of a conventional commit message in this repo - type(scope): subject + blank line + body explaining why and impact."
tags: conventional-commits, message-format, subject, body, release-please
sources: []
last_ingested: 2026-05-15
---

## Conventional commit format

Every commit message has the shape:

```
type(scope): subject

body (optional, but expected for non-trivial commits)
```

### The subject line

- **Type** is one of the canonical conventional types (see the table in `SKILL.md`). The type carries [semver](https://semver.org) intent:
  - `feat` -> **minor** (new functionality, backward-compatible)
  - `fix` / `perf` -> **patch** (backward-compatible bug fix or perf tweak)
  - any breaking change -> **major** (incompatible API change), regardless of type
  - `refactor` / `docs` / `test` / `build` / `ci` / `chore` / `style` -> no bump

  Many release tools (release-please, semantic-release, changesets, conventional-changelog) parse the type to drive automated version bumps; the mapping is still correct when no such tool is in use.
- **Scope** is optional but encouraged. It names the area of the repo being changed. Pick what's already conventional in the current repo — run `git log --oneline -50` to see existing scopes and reuse them. Compound scopes are fine when they add signal: `refactor(eval/grader): ...`.
- **Subject** is the one-line summary.
  - Imperative mood (`add`, not `added`; `fix`, not `fixed` or `fixes`).
  - Lowercase first character after the colon.
  - No trailing period.
  - Hard ceiling of **72 characters** for the entire subject line including the type/scope prefix. GitHub truncates beyond that in many views.

### The body

The body is where the commit earns its keep. The subject says **what**; the body answers three questions: **what changed**, **why**, and **what the impact is**.

For non-trivial commits, the body uses three labeled sections, each label on its own line, content on the line below:

```
Changed:
<one or two sentences naming what concretely changed in the code>

Why:
<one or two sentences on the motivation — problem, constraint, user need>

Impact:
<one or two sentences on what this resolves, unblocks, or prevents>
```

Body rules:

- One blank line between subject and body.
- One blank line between each labeled section.
- **Labels on their own line**, content on the line below. Don't write `Changed: ParseError::Empty ...` on one line.
- Use **plain labels** (`Changed:`), not markdown headers (`## Changed:`). Git's default `commentChar` is `#`, which silently strips `#`-prefixed lines on editor-based amends (`git commit --amend` without `-m`). Plain labels survive every git operation. They are also greppable: `git log --grep "^Why:"`.
- Wrap content at ~72 characters.
- Reference issues, PRs, or upstream tickets in the Impact section when relevant (`Resolves #214`, `Fixes ENG-1029`).

### When the labeled structure is optional

Free-form one-paragraph bodies are fine for **trivial commits only**:

- Single-character / single-word typo fix
- Formatting-only changes, lockfile re-sorts
- Dependency version bump with no behaviour change
- Auto-generated file regeneration (registry, lockfiles, generated types)
- Commits where Changed/Why/Impact would just paraphrase the subject

If you're not sure, use the structure.

### Worked example

**Incorrect (body just restates the subject):**

```
fix(parser): handle empty input

This commit handles empty input in the parser.
```

The body adds no information the subject doesn't already convey. Either drop the body, or expand it into the three labeled sections with real content.

**Correct (non-trivial, structured body):**

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
```

**Correct (trivial, free-form body):**

```
docs: fix typo in installation step

The README pointed at the wrong package name in the brew install
example. Caught during onboarding review.
```

### Reference examples

Real-world conventional commit messages that illustrate the shape:

```
docs(eval): publish indie-launch full 4-arm × 3-run ablation (#99)
fix(docs): co-locate eval HTML with scenario pages so iframes resolve (#97)
refactor(eval/grader): default grader model to claude-sonnet-4-6
chore(registry): regenerate registry.json
feat(eval): add flat_skill_wiki arm + cumulative retention outcome
```

The `(#NN)` PR-number suffix on some of these is a GitHub squash-merge convention (GitHub auto-appends the PR number when you squash-and-merge a PR using its title) — it's not something you author when committing locally. Omit it from your subject line; it'll be added by the merge if your repo uses that pattern.

### Breaking changes (MAJOR bumps)

Per [semver](https://semver.org), **any change that breaks backward compatibility forces a MAJOR bump** — irrespective of whether the change looks like a feature, a fix, or a refactor. Removing a public function, renaming a CLI flag, changing a wire format, dropping support for an old runtime — all MAJOR.

Two equivalent ways to mark a breaking change:

1. Append `!` after the type/scope: `feat(api)!: remove deprecated /v1 endpoint`
2. Add a `BREAKING CHANGE:` footer in the body:

   ```
   feat(api): switch default tier to scoped tokens

   BREAKING CHANGE: existing global tokens are rejected after this commit.
   Consumers must migrate to scoped tokens per docs/migrating-tokens.md.
   ```

Either form signals MAJOR to humans reading the log and to automated release tooling (release-please, semantic-release, changesets, etc.) that keys on it.

### Pre-release and build metadata

Semver allows extensions on top of `MAJOR.MINOR.PATCH`:

- **Pre-release identifier** (`1.4.0-rc.1`, `2.0.0-alpha`, `1.0.0-beta.3`) — precedes the released version and has lower precedence than the final release.
- **Build metadata** (`1.4.0+sha.abc123`, `1.0.0+20251015`) — appended after `+`, ignored for precedence comparison.

This skill does not author release commits — but if your repo cuts a pre-release, the version your commits ultimately ship under may include one of these tags. The conventional-commit type you write does not change; the pre-release label is added by whoever cuts the release.

### Committing with a multi-line body

Always use a HEREDOC so the body's newlines and labeled sections survive shell quoting verbatim:

```bash
git commit -m "$(cat <<'EOF'
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
EOF
)"
```

The single-quoted `'EOF'` is important — it prevents the shell from interpolating `$variables` or backticks inside the body. The `Authored by humblSKILLS; "use-smart-commit"` line is the skill's authorship footer; it is default-on and the user can disable it per-conversation or persistently (see the *Authorship footer* section in `SKILL.md`).
