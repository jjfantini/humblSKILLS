---
title: "Detect Project Type and Pick the Template"
context: claudeinit
category: template
concept: code-vs-general
description: "How to decide whether a project is code or non-code and which bundled template to fill, so non-code projects get a lighter file instead of an awkward eight-section code template."
tags: detection, project-type, template-selection, non-code, general
sources:
  - "references/raw/user-brief.md"
last_ingested: 2026-06-02
---

## Choosing the template

The skill ships two templates in `assets/`:

- `claude-code.md.tmpl` — the full **eight-section** code template.
- `claude-general.md.tmpl` — a lighter **four-section** template for non-code
  projects (research, writing, ops, content, design).

Pick by detecting project type first; confirm with the user when ambiguous.

### Detection signals (code project)

Treat the project as **code** if any of these are present in the target
directory or are described by the user:

- A dependency manifest: `package.json`, `pyproject.toml`, `requirements.txt`,
  `go.mod`, `Cargo.toml`, `pom.xml`, `build.gradle`, `Gemfile`, `composer.json`.
- A source layout: `src/`, `lib/`, `cmd/`, `app/`, test directories.
- Build/CI config: `Makefile`, `.github/workflows/`, `Dockerfile`, a lockfile.
- The user describes building software (an app, service, library, CLI).

If none of these signals are present and the user describes a non-software
effort (a research project, a book, a marketing site's content plan, an
operations runbook), use the **general** template.

### When ambiguous, ask one question

Do not guess silently on a coin-flip. Ask the gating question with a
recommended default (see the question bank), e.g.:

```markdown
I don't see a dependency manifest here, so I'd treat this as a **non-code**
project and use the lighter template. Is that right, or is this a codebase
that just hasn't been scaffolded yet?
```

### The general (non-code) template

`claude-general.md.tmpl` keeps four sections so the file stays useful without
forcing code concepts onto a non-code effort:

1. **Project Intent** — one-liner, who for, non-goals (same as code).
2. **Working Preferences** — how the agent should operate: plan-first vs
   just-do-it, when to use subagents, tone, the self-improvement loop.
3. **Quality Bar** — what "good" looks like for this work's *deliverables*
   (a citation standard for research, a voice guide for writing, a checklist
   for ops) and how to verify before calling it done.
4. **Core Principles** — the 3-5 non-negotiables.

The same interview methodology applies — relentless, one at a time, recommend
an answer — just with fewer sections to resolve.

### Don't force code sections onto non-code work

**Incorrect:** asking a non-code project for its "test runner" and "performance
budget", then leaving those sections as empty placeholders.

**Correct:** switch to the general template, drop the code-only sections, and
translate the intent (e.g. "Testing" becomes "how do we verify the deliverable
is correct?").

## Sources

- `references/raw/user-brief.md` — the requirement that the project need not be
  code, and that a code project gets the standard eight-section template while
  non-code work gets a lighter file.
