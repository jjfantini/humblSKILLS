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

- **Type** is one of the canonical conventional types (see the table in `SKILL.md`). The type carries semver intent: `feat` is minor, `fix`/`perf` are patch, breaking changes are major. Many release tools (release-please, semantic-release, changesets, conventional-changelog) parse the type to drive automated version bumps; the type is still meaningful when no such tool is in use.
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

### Breaking changes

Two ways to mark a breaking change:

1. Append `!` after the type/scope: `feat(api)!: remove deprecated /v1 endpoint`
2. Add a `BREAKING CHANGE:` footer in the body:

   ```
   feat(api): switch default tier to scoped tokens

   BREAKING CHANGE: existing global tokens are rejected after this commit.
   Consumers must migrate to scoped tokens per docs/migrating-tokens.md.
   ```

Either form signals a major version bump under semver, and is what automated release tooling (release-please, semantic-release, changesets, etc.) keys on when present.

### Committing with a multi-line body

Always use a HEREDOC so the body's newlines survive shell quoting verbatim:

```bash
git commit -m "$(cat <<'EOF'
fix(parser): handle empty input without panicking

The parser called .unwrap() on the first token, which panicked on
empty stdin. Now it returns ParseError::Empty so callers can recover.
Resolves #214 and unblocks the CLI's --from-stdin flow.
EOF
)"
```

The single-quoted `'EOF'` is important — it prevents the shell from interpolating `$variables` or backticks inside the body.
