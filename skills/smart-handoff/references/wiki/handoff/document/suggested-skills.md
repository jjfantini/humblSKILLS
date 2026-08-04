---
title: "Build the Suggested Skills Section From What Is Installed"
context: handoff
category: document
concept: suggested-skills
description: "The receiving agent gets real, installable skill names instead of plausible-sounding ones that do not exist"
tags: document, skills, humblskills, install, suggestions
sources: []
last_ingested: 2026-08-03
command: scripts/preflight.sh
---

## The Problem

Naming skills from memory produces confident hallucinations — `smart-pr-review`,
`smart-debug`, `smart-refactor` all sound real and none exist. The receiving
agent then runs an install that 404s, or worse, silently skips the step.

## Enumerate, Then Filter

`scripts/preflight.sh` prints an `installed-skills` block (via
`humblskills list`, falling back to a scan of the skill directories). Every
name in the Suggested Skills table must appear in that block.

Then filter hard: suggest a skill **only if a specific Next Step calls for
it**. Three relevant skills beat twelve plausible ones.

## Emit Install Commands

The receiving harness may not have the skill. Every row carries its install
command so the next agent is one copy-paste from ready:

```markdown
## Suggested Skills

| Skill | Use it for | Install |
|-------|------------|---------|
| `smart-commit` | landing steps 2-4 as atomic conventional commits | `humblskills install smart-commit` |
| `smart-worktree-flow` | the feature → develop → main → release path in step 5 | `humblskills install smart-worktree-flow` |
```

## Tie Each Row to a Step

The "Use it for" column names the step number it serves. A skill that maps to
no step does not belong in the table.

### Incorrect

```markdown
## Suggested Skills
- smart-commit
- smart-worktree-flow
- smart-debug          <- does not exist
- better-typography    <- irrelevant to a backend task
```

### Correct

```markdown
| `smart-commit` | step 3: split the parser fix from the doc update | `humblskills install smart-commit` |
```

## When Nothing Fits

Write `No skills needed beyond the receiving agent's defaults.` An empty
honest section is better than a padded one.

## Sources

- (none) - authored from the smart-handoff design session.
