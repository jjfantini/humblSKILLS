# User brief for smart-claude-init (raw request)

Captured verbatim-in-spirit from the request that created this skill.

## Goal

A skill, `smart-claude-init`, that grills and questions the individual to find
out the preferences they want their CLAUDE.md to have in a **new project**.
Relentlessly ask questions about what the project is used for, to create the
cleanest, best, iterated CLAUDE.md that will guide future use.

The project does not need to be code-specific. But if it **is** code, there
must be a standard template to drop answers into. That template describes
things like:

- project intent (one-liner)
- stack
- user preferences
- coding guidelines
- development philosophy
- bug protocol
- task management
- core principles

## The 8 core sections (consolidated, confirmed with the user)

Dial in 8 canonical sections that drive the best codebase. Consolidate the
overlapping items above plus these additional requested areas — architecture,
engineering preferences, code quality preferences, testing preferences,
performance — into exactly 8 non-overlapping top-level sections:

1. **Project Intent** — one-liner, what/why, target users, explicit non-goals
2. **Architecture & Stack** — languages, frameworks, runtime, repo layout,
   entry points, data flow, external services
3. **Engineering Preferences** — plan-first workflow, subagent strategy,
   development philosophy, self-improvement loop
4. **Code Quality** — style, naming, comments, simplicity/elegance,
   file-size limits
5. **Testing** — framework, how to run, coverage bar, verify-before-done
6. **Performance** — budgets, hot paths, optimize vs deliberately-not
7. **Bug Protocol** — autonomous fix, root-cause discipline, regression tests,
   no temporary hacks
8. **Task Management & Core Principles** — todo/lessons tracking +
   non-negotiable principles

## Behaviour decisions (confirmed with the user)

- Ship **both** a code template (8 sections) and a lighter general (non-code)
  template; the skill auto-detects project type.
- Interview style: **relentless, one question at a time**, recommend an answer
  for each, explore the codebase instead of asking when the answer is
  discoverable, continue until every section is resolved. Escape hatch: the
  user can say "use your best guesses".

## Inspiration

The interview behaviour is modelled on the `grill-me` skill — see
`references/raw/grill-me-skill.md`. The desired output shape is modelled on the
example in `references/raw/example-claude-md.md`.
