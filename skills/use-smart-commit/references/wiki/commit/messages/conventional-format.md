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

- **Type** is one of the canonical conventional types (see the table in `SKILL.md`). The choice drives release-please's version bump on merge to `main`.
- **Scope** is optional but encouraged. It names the area of the repo being changed. Real scopes from this repo's history: `skills`, `cli`, `eval`, `docs`, `ci`, `registry`. Compound scopes are fine when they add signal: `refactor(eval/grader): ...`.
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

### Verbatim examples from this repo

Use these as canonical references when picking a type + scope:

```
docs(eval): publish indie-launch full 4-arm × 3-run ablation (#99)
fix(docs): co-locate eval HTML with scenario pages so iframes resolve (#97)
refactor(eval/grader): default grader model to claude-sonnet-4-6
chore(registry): regenerate registry.json
feat(eval): add flat_skill_wiki arm + cumulative retention outcome
```

Note the `(#NN)` PR-number suffix on commits that landed through PRs — that's a release-please / merge convention in this repo, not something you author. When committing locally before the PR exists, omit it.

### Breaking changes

Two ways to mark a breaking change:

1. Append `!` after the type/scope: `feat(api)!: remove deprecated /v1 endpoint`
2. Add a `BREAKING CHANGE:` footer in the body:

   ```
   feat(api): switch default tier to scoped tokens

   BREAKING CHANGE: existing global tokens are rejected after this commit.
   Consumers must migrate to scoped tokens per docs/migrating-tokens.md.
   ```

Either form triggers a major version bump in release-please.

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
