---
title: "Scaffold and Flesh Out a New Smart Skill"
context: smart
category: create
concept: workflow
description: "Consistent skill structure from day one, zero rework later"
tags: create, scaffold, new-skill, smart, brain
sources: []
last_ingested: 2026-04-16
command: scripts/scaffold.sh
---

## Create a New Smart Skill

Starting a new skill with the Smart Skill (CCCCC + brain) pattern avoids the
cost of migrating later. The scaffold script generates the full directory
(wiki/, raw/, brain meta files, seeded templates) in one call.

## Workflow

### 1. Choose name and location

Pick a kebab-case name (1-64 chars, lowercase, no consecutive hyphens).
Decide personal (`~/.cursor/skills/`) vs project (`.cursor/skills/`).

### 2. Run the scaffold

```bash
bash scripts/scaffold.sh <skill-name> [--scripts] [--assets] [--location personal|project]
```

| Flag         | Effect                                      |
|--------------|---------------------------------------------|
| `--scripts`  | Creates `scripts/` directory                |
| `--assets`   | Creates `assets/` directory                 |
| `--location` | `personal` (default) or `project`           |

The scaffold generates:
- `SKILL.md` with Brain Protocol block pre-injected (name pre-filled)
- `references/_template.md`, `_brain.md`
- `references/_index.md` (seeded with sentinel markers)
- `references/patterns.md`, `decisions.md`, `log.md`
- `references/wiki/` (empty)
- `references/raw/.gitkeep`

Idempotent - skips files that already exist.

### 3. Edit SKILL.md

Fill in:
- `description` - include WHAT the skill does and WHEN to trigger it
- `compatibility` - only if the skill requires specific env (bash, git, docker, specific runtime, network). Delete the scaffold's TODO line if not needed. See `references/wiki/smart/spec/skill-frontmatter.md`.
- `metadata.author`, `metadata.version`
- When-to-Use bullets
- How-to-Use section pointing to your wiki concepts

Keep SKILL.md under 120 lines. It is a router + brain protocol, not a manual.

### 4. Write wiki concepts

For each concept:

1. Create `references/wiki/<context>/<category>/<concept>.md`
2. Follow the structure in `references/_template.md`
3. Required frontmatter: `title`, `context`, `category`, `concept`,
   `description`, `tags`, `sources`, `last_ingested`
4. Body: explanation, incorrect example, correct example

The `context`/`category`/`concept` triple MUST match the filesystem path.
One concept per file. The taxonomy is derived from where files live - no
registry file to maintain.

### 5. Add commands (optional)

If a concept has a deterministic, executable action:

1. Create `scripts/<command>.sh` or `scripts/<command>.py`
2. Add `command: scripts/<command>.sh` to the concept's frontmatter
3. Scripts must be self-contained, idempotent, non-interactive

### 6. Regenerate the index

```bash
bash scripts/lint.sh
```

`lint.sh` walks the filesystem, validates every wiki file, and rewrites
`references/_index.md` between its sentinel markers.

### 7. Validate

See `references/wiki/smart/create/validation-checklist.md` before shipping.

## Sources

- (none) - authored from the Smart Skill architecture.

## Command

```bash
bash scripts/scaffold.sh my-new-skill --scripts
```
