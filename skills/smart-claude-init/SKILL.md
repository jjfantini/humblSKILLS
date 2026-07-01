---
name: smart-claude-init
description: >
  Interview the user relentlessly, one question at a time with a recommended
  answer for each, to produce a clean iterated CLAUDE.md for a new project.
  Detects code vs non-code projects and fills a standard 8-section template
  (project intent, architecture, engineering preferences, code quality,
  testing, performance, bug protocol, task management and core principles).
  Use when the user says "init CLAUDE.md", "set up CLAUDE.md", "grill me about
  my project", "smart-claude-init", or wants a guided CLAUDE.md for a new repo.
  Do NOT use to mechanically document an existing codebase (use the built-in
  init skill), or to edit an already-good CLAUDE.md.
license: MIT
compatibility: Requires bash and POSIX utilities (grep, sed) plus python3 for scripts/lint.sh.
allowed-tools: "Bash(bash:*) Read Write Edit Glob Grep"
metadata:
  author: jjfantini
  version: "0.1.3"
  category: meta
  tags: [claude-md, onboarding, interview, project-setup, humblskill]
  platforms: [claude-code, cursor, codex]
  preserve:
    - references/raw/
    - references/wiki/
    - references/decisions.md
    - references/log.md
    - references/patterns.md
---

# Smart Claude Init

Turn a blank project into a sharp, opinionated `CLAUDE.md` by **grilling the
user one question at a time** — recommending an answer for each — then filling a
standard template and validating it before it lands. This is for *bootstrapping*
a new project's guide; for mechanically documenting an *existing* codebase, use
the built-in `init` skill instead.

## Brain Protocol (read BEFORE creating anything)

1. `references/_index.md`       - what this skill knows (map)
2. `references/patterns.md`     - what worked, with numbers
3. `references/decisions.md`    - past reasoning, don't repeat mistakes
4. `references/log.md`          - last 5 session entries
5. Relevant `references/wiki/claudeinit/<category>/` concepts per task

After completing work, UPDATE the brain:
- Performance data (if reported) -> `patterns.md`
- Non-obvious decisions -> `decisions.md`
- Session summary (always) -> append to `log.md`
- Run `scripts/lint.sh` to regenerate `_index.md` and verify structure

_Full spec: `references/_brain.md`._

## Workflow

Six steps. Steps 1-3 are judgment; step 5 is the deterministic gate. Full
detail in `references/wiki/claudeinit/generate/workflow.md`.

| Step | What | Driver |
|------|------|--------|
| 1 | Detect code vs non-code; pick the template | Agent + directory read |
| 2 | Explore the repo; pre-fill what the code answers | Agent |
| 3 | Interview relentlessly to resolve every section | Agent judgment |
| 4 | Substitute answers into the chosen template | Agent |
| 5 | Validate the draft (sections, no placeholders, no TODOs) | `scripts/validate-claudemd.sh` |
| 6 | Write `CLAUDE.md`, offer one refinement pass, log | Agent |

### Step-by-step

1. **Detect project type.** Read the target directory. A dependency manifest
   (`package.json`, `pyproject.toml`, `go.mod`, `Cargo.toml`, ...) or source
   layout means **code** -> `assets/claude-code.md.tmpl` (8 sections). No code
   signals and a non-software effort means **general** ->
   `assets/claude-general.md.tmpl` (4 sections). Ask one question only when
   ambiguous. See `references/wiki/claudeinit/template/code-vs-general.md`.
2. **Explore before asking.** If a repo exists, read manifests, configs, CI,
   test setup, and layout. Pre-fill every answer the code already gives so the
   interview only covers what the code cannot tell you.
3. **Interview relentlessly.** One question per turn, each with a recommended
   default, walking the decision tree until every section is resolved. Explore
   the codebase instead of asking when the answer is discoverable. Honour the
   "use your best guesses" escape hatch. See
   `references/wiki/claudeinit/interview/methodology.md` and the question pool
   in `.../interview/question-bank.md`.
4. **Substitute.** Replace every `{{TOKEN}}` in the chosen template with the
   resolved content. Never leave a token — even "no explicit budget" gets a
   real sentence.
5. **Validate before writing.** Run the gate against your draft:
   ```bash
   bash scripts/validate-claudemd.sh /path/to/draft-CLAUDE.md            # code
   bash scripts/validate-claudemd.sh --general /path/to/draft-CLAUDE.md  # general
   ```
   It exits non-zero and lists any missing section, surviving `{{placeholder}}`,
   or leftover `<!-- TODO -->`. Fix until it exits 0.
6. **Write, iterate, log.** Write the final file to `CLAUDE.md` at the project
   root (or the path the user gave), offer one refinement pass, then append a
   one-line entry to `references/log.md`.

## The eight code sections

`assets/claude-code.md.tmpl` carries exactly these top-level sections; each must
hold a concrete, project-specific directive (see
`references/wiki/claudeinit/template/eight-sections.md`):

1. **Project Intent** — one-liner, who for, explicit non-goals
2. **Architecture & Stack** — languages, frameworks, runtime, layout, services
3. **Engineering Preferences** — plan-first, subagents, philosophy, lessons loop
4. **Code Quality** — formatter/linter, naming, simplicity bar, size limits
5. **Testing** — framework, run command, coverage rule, verify-before-done
6. **Performance** — budgets, hot paths, do-not-optimize
7. **Bug Protocol** — autonomy, root-cause discipline, regression tests
8. **Task Management & Core Principles** — tracking + 3-5 non-negotiables

The general (non-code) template keeps four: Project Intent, Working Preferences,
Quality Bar, Core Principles.

## When to Use

- Bootstrapping a `CLAUDE.md` for a brand-new project or empty repo
- The user says "init CLAUDE.md", "set up CLAUDE.md", "grill me about my project"
- An existing repo with no `CLAUDE.md` where the user wants a guided, opinionated one
- NOT for mechanically dumping an existing codebase's structure — use `init`

## Examples

### Example 1: new code project, full grilling

User says: "Set up a CLAUDE.md for my new FastAPI service."

Actions:
1. Read the directory — `pyproject.toml` present -> code project, pick
   `assets/claude-code.md.tmpl`.
2. Read `references/wiki/claudeinit/interview/methodology.md` and `.../question-bank.md`.
3. Grill one question at a time, each with a recommended default (intent,
   non-goals, runtime, test runner, coverage bar, bug autonomy, principles),
   reading configs instead of asking where possible.
4. Substitute answers into the template; run
   `bash scripts/validate-claudemd.sh draft.md` until it exits 0.
5. Write `CLAUDE.md`, offer a refinement pass, append a `log.md` entry.

Result: a tight 8-section `CLAUDE.md` with no placeholders, every section
carrying a concrete directive.

### Example 2: non-code project, escape hatch

User says: "Grill me about my research project and make a CLAUDE.md — actually,
just use your best guesses after the first couple questions."

Actions:
1. No manifest, non-software effort -> `assets/claude-general.md.tmpl` (4 sections).
2. Ask the intent and audience questions, then honour the escape hatch.
3. Fill the remaining sections from sensible defaults, marking each inferred
   choice inline.
4. `bash scripts/validate-claudemd.sh --general draft.md` -> exit 0; write the file.

Result: a lighter 4-section `CLAUDE.md` for non-code work, with assumptions
flagged for later correction.

## How to Use

**Live enumeration of categories and concepts:**
Read `references/_index.md` (auto-regenerated by `scripts/lint.sh`).

**Brain protocol, naming, writing principles, linking, lint checks:**
Read `references/_brain.md`. **Wiki concept file shape:** `references/_template.md`.

### Scripts

- `scripts/validate-claudemd.sh` — the completeness gate (sections present, no
  placeholders, no leftover TODOs). Run at workflow step 5. `--general` for the
  4-section contract. `--help` for usage.
- `scripts/lint.sh` — brain health check (regenerates `_index.md`, reports
  orphans/stale/broken sources). Run after editing wiki concepts.

### Tests (opt-in)

`tests/run.sh` exercises the validator in an isolated `mktemp` dir — section
checks, placeholder/TODO detection, the `--general` contract, and invocation
errors. **Not wired into CI.** Run after editing the validator or templates.
See `tests/README.md`.

### Primary workflows

**Run a relentless, one-at-a-time interview:**
Read `references/wiki/claudeinit/interview/methodology.md`.

**Know which questions to ask per section:**
Read `references/wiki/claudeinit/interview/question-bank.md`.

**Understand the 8-section anatomy / pick code vs general:**
Read `references/wiki/claudeinit/template/eight-sections.md` and
`.../template/code-vs-general.md`.

**Run the full detect-interview-fill-validate-write pipeline:**
Read `references/wiki/claudeinit/generate/workflow.md`.

**Avoid the failure modes that get a CLAUDE.md ignored:**
Read `references/wiki/claudeinit/quality/anti-patterns.md`.

## Success Signals

- `scripts/validate-claudemd.sh` exits 0 on the draft before it is written.
- Every required section is present and carries a concrete, project-specific
  directive — no vague platitudes, no surviving `{{placeholders}}`.
- The generated `CLAUDE.md` reads in under two minutes (skimmable, not bloated).
- `scripts/lint.sh` exits 0; `log.md` grows by exactly one entry per session.
- `bash tests/run.sh` passes after any validator or template change.
