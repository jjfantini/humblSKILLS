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

The body is where the commit earns its keep. The subject says **what**; the body says **why** and **impact**.

Body rules:

- One blank line between subject and body.
- Wrap lines at ~72 characters.
- Cover two things: (1) **why** the change is being made — the motivation, the user-visible problem, the constraint that forced this — and (2) **impact** — what it resolves, unblocks, or prevents.
- Reference issues, PRs, or upstream tickets when relevant (`Resolves #214`, `Fixes ENG-1029`).
- If you can't add new information beyond the subject, omit the body.

### Worked example

**Incorrect (body just restates the subject):**

```
fix(parser): handle empty input

This commit handles empty input in the parser.
```

The body adds no information the subject doesn't already convey. Drop the body.

**Correct (body explains *why* and *impact*):**

```
fix(parser): handle empty input without panicking

The parser called .unwrap() on the first token, which panicked on
empty stdin. Now it returns ParseError::Empty so callers can recover
gracefully. Resolves #214 and unblocks the CLI's --from-stdin flow.
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

Always use a HEREDOC so the body's newlines survive shell quoting verbatim:

```bash
git commit -m "$(cat <<'EOF'
fix(parser): handle empty input without panicking

The parser called .unwrap() on the first token, which panicked on
empty stdin. Now it returns ParseError::Empty so callers can recover.
Resolves #214 and unblocks the CLI's --from-stdin flow.

Authored by humblSKILLS; "use-smart-commit"
EOF
)"
```

The single-quoted `'EOF'` is important — it prevents the shell from interpolating `$variables` or backticks inside the body. The `Authored by humblSKILLS; "use-smart-commit"` line is the skill's authorship footer; it is default-on and the user can disable it per-conversation or persistently (see the *Authorship footer* section in `SKILL.md`).
