---
title: "CLAUDE.md Anti-Patterns to Avoid"
context: claudeinit
category: quality
concept: anti-patterns
description: "The failure modes that make a CLAUDE.md get ignored - vague platitudes, bloat, contradictions, staleness, and duplicating the codebase - and how the interview prevents each."
tags: anti-patterns, quality, pitfalls, claude-md, maintenance
sources:
  - "references/raw/user-brief.md"
  - "references/raw/example-claude-md.md"
last_ingested: 2026-06-02
---

## Anti-patterns

A CLAUDE.md fails when the agent reads it and learns nothing it would not have
defaulted to anyway. The interview exists to avoid these five failure modes.

### 1. Vague platitudes

"Write clean code. Be thorough. Care about quality." These are true of every
project and steer nothing. **Fix:** the interview demands a concrete directive
per section — a command to run, a number to hit, a rule to enforce. See
`references/wiki/claudeinit/template/eight-sections.md`.

### 2. Bloat

A 600-line CLAUDE.md is read once and skimmed forever. Token cost is paid every
session. **Fix:** keep it skimmable (under ~2 minutes to read); push long
detail into linked docs and keep only the directive. If a section exceeds ~12
bullets it is mixing concerns.

### 3. Contradictions

"Always plan first" next to "just fix bugs autonomously without asking" reads
as a conflict unless the boundary is stated. **Fix:** when two directives could
collide, write the precedence explicitly (e.g. "plan first for features;
bug fixes from a clear report are autonomous"). The example file resolves this
by scoping autonomy to bug reports.

### 4. Staleness / duplicating the codebase

Listing every file and function turns the CLAUDE.md into a second, rotting copy
of the repo. It is wrong the moment the code changes. **Fix:** capture
*intent, preferences, and non-obvious rules* — the things the code cannot tell
you. Let the agent read the code for structure. (This is the line between
`smart-claude-init` and the built-in `/init`, which documents existing
structure.)

### 5. Generated-but-unfinished

Leftover `{{PLACEHOLDER}}` tokens or `TODO` markers from the template. **Fix:**
`scripts/validate-claudemd.sh` fails on any survivor; the workflow validates
before writing the file. Never ship a half-filled file "to finish later".

### Quick contrast

**Incorrect (vague + bloated + unfinished):**

```markdown
## Code Quality
We value high-quality, clean, maintainable, well-architected code across the
entire {{STACK}} stack. TODO: add naming rules.
```

**Correct (concrete + tight + finished):**

```markdown
## Code Quality
- Formatter: `ruff format`; lint with `ruff check` (fix before commit).
- Names: full words, no abbreviations except `id`, `url`, `db`.
- Prefer the simplest change that fully solves it; flag any hack for review.
```

## Sources

- `references/raw/user-brief.md` — the "cleanest, best, iterated" bar that
  motivates avoiding these failure modes, and the boundary against the built-in
  init.
- `references/raw/example-claude-md.md` — the example's scoped autonomy and
  simplicity-first principles inform the contradiction and platitude fixes.
