---
title: "Question Bank Mapped to the Eight Sections"
context: claudeinit
category: interview
concept: question-bank
description: "The canonical set of interview questions, grouped by the eight CLAUDE.md sections, each with a recommended default so the grilling stays concrete and ordered."
tags: questions, sections, defaults, checklist, interview
sources:
  - "references/raw/user-brief.md"
  - "references/raw/example-claude-md.md"
last_ingested: 2026-06-02
---

## Question bank

The agent owns the conversation, but this is the canonical pool to draw from so
no section ships empty. Ask one at a time (see
`references/wiki/claudeinit/interview/methodology.md`), recommend the default,
and skip any question the codebase already answers. Order top-to-bottom;
earlier answers gate later ones.

### 0. Gating question (asked first)

- Is this a **code** project or a **non-code** project (research, writing,
  ops, content)? *Default: infer from the directory — a manifest file means
  code.* This picks the template (see
  `references/wiki/claudeinit/template/code-vs-general.md`).

### 1. Project Intent

- In one sentence, what does this project do? *No default — must come from the user.*
- Who is it for (the primary user/consumer)?
- What are the explicit **non-goals** — things it will deliberately never do?

### 2. Architecture & Stack

- Primary language(s) and framework(s)? *Default: read the manifest.*
- Runtime / deployment target (server, CLI, library, browser, mobile)?
- Repo layout — monorepo or single package? Where are the entry points?
- External services / data stores it talks to?

### 3. Engineering Preferences

- Plan-first or just-do-it for non-trivial tasks? *Default: plan mode for
  anything 3+ steps or with architectural impact.*
- Should the agent use subagents for research/parallel work? *Default: yes, to
  keep the main context clean.*
- Self-improvement loop — capture lessons after corrections? *Default: yes,
  write a rule that prevents the same mistake.*
- Any stated development philosophy (ship-fast, correctness-first, etc.)?

### 4. Code Quality

- Style guide / formatter / linter in force? *Default: read configs
  (.prettierrc, ruff, eslint, gofmt).*
- Naming conventions and comment density expectations?
- Simplicity vs cleverness — how much elegance to demand? *Default: simplest
  change that fully solves it; flag hacky fixes.*
- File-size / function-size limits?

### 5. Testing

- Test framework and how to run the suite? *Default: read the test config.*
- Coverage bar or "tests for every bug/feature" rule? *Default: a regression
  test ships with every bug fix.*
- Is verify-before-done mandatory (run tests / the app before claiming done)?
  *Default: yes.*

### 6. Performance

- Are there performance budgets that matter (latency, bundle size, memory)?
  *Default: none stated — note "no explicit budget" rather than inventing one.*
- Known hot paths to protect?
- Anything to deliberately **not** optimize (premature-optimization traps)?

### 7. Bug Protocol

- Autonomous bug-fixing or always confirm first? *Default: just fix it from
  logs/errors/failing tests; no hand-holding.*
- Root-cause discipline — ban temporary patches? *Default: yes, fix the cause.*
- Does every fix need a regression test? *Default: yes.*

### 8. Task Management & Core Principles

- How should work be tracked (todo file, issue tracker, in-chat)? *Default: a
  checkable plan before implementation; mark items done as you go.*
- The 3-5 non-negotiable principles for this project? *Default: Simplicity
  First, No Laziness (root causes), Minimal Impact.*

## Sources

- `references/raw/user-brief.md` — the eight sections and the items each
  consolidates.
- `references/raw/example-claude-md.md` — the defaults for engineering
  preferences, bug protocol, task management, and core principles are drawn
  from the user's example file.
