---
title: "Anatomy of the Eight-Section Code CLAUDE.md"
context: claudeinit
category: template
concept: eight-sections
description: "What each of the eight canonical CLAUDE.md sections must contain and what makes it actionable versus filler, so the generated file actually steers the agent."
tags: template, sections, claude-md, anatomy, structure
sources:
  - "references/raw/user-brief.md"
  - "references/raw/example-claude-md.md"
last_ingested: 2026-06-02
---

## The eight sections

The code template (`assets/claude-code.md.tmpl`) has exactly eight top-level
`##` sections. Each must contain **concrete, project-specific directives** — a
sentence the agent can act on, not a platitude. The contract is enforced by
`scripts/validate-claudemd.sh`, which fails if any section header is missing or
any `{{PLACEHOLDER}}` / `TODO` survives.

| # | Section | Must answer | Filler to avoid |
|---|---------|-------------|-----------------|
| 1 | **Project Intent** | What it does (one line), who for, explicit non-goals | "A great app." |
| 2 | **Architecture & Stack** | Languages, frameworks, runtime, layout, entry points, services | "Modern tech." |
| 3 | **Engineering Preferences** | Plan-first rule, subagent use, dev philosophy, lessons loop | "Write good code." |
| 4 | **Code Quality** | Formatter/linter, naming, comments, simplicity bar, size limits | "Be clean." |
| 5 | **Testing** | Framework, run command, coverage rule, verify-before-done | "Test things." |
| 6 | **Performance** | Budgets (or explicit "none"), hot paths, do-not-optimize list | "Be fast." |
| 7 | **Bug Protocol** | Autonomous-fix rule, root-cause discipline, regression-test rule | "Fix bugs." |
| 8 | **Task Management & Core Principles** | Tracking method + 3-5 non-negotiables | "Stay organized." |

### What "actionable" looks like

**Incorrect (vague, un-actionable — the agent learns nothing):**

```markdown
## Testing
We care a lot about testing and quality. Please write good tests.
```

**Correct (concrete — the agent knows exactly what to do):**

```markdown
## Testing
- Runner: `pytest`. Run the suite with `make test` (or `uv run pytest -q`).
- Every bug fix ships with a regression test that fails before the fix.
- Coverage bar: 85% on `src/`; do not lower it to make a PR pass.
- Verify before done: run `make test` and exercise the changed path; never
  mark a task complete on a green typecheck alone.
```

### Section ordering rationale

Intent and Architecture come first because they are context every later
section depends on. Engineering Preferences and Code Quality define *how* to
work. Testing, Performance, and Bug Protocol define *quality gates*. Task
Management & Core Principles close with the operating rules. This top-to-bottom
flow also matches the interview's dependency order (see
`references/wiki/claudeinit/interview/question-bank.md`).

### Keep it skimmable

A CLAUDE.md the agent re-reads every session must be short. Prefer tight
bullets over prose. If a section grows past ~12 bullets, it is probably mixing
concerns — split the detail into a linked doc and keep the directive in the
file. The generated file should read in under two minutes.

## Sources

- `references/raw/user-brief.md` — the canonical eight-section list and the
  items each consolidates.
- `references/raw/example-claude-md.md` — the actionable phrasing for
  workflow, bug protocol, task management, and core principles.
