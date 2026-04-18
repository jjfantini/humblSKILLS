---
title: "Migrate a Flat Skill to the Smart Skill Pattern"
context: smart
category: migrate
concept: workflow
description: "Scalable structure, self-learning memory, cheaper loads"
tags: migrate, refactor, restructure, smart, brain
sources: []
last_ingested: 2026-04-16
---

## Migrate a Flat Skill

A flat or monolithic SKILL.md that exceeds ~300 lines becomes expensive
to load and hard to maintain. Migrating to the Smart Skill pattern splits
it into a thin router + nested wiki concepts + self-learning memory.

## When to Migrate

- SKILL.md is over ~300 lines
- The skill covers 3+ distinct topics
- You keep adding content and the file keeps growing
- Multiple concerns are tangled in one document
- You want the skill to compound across sessions (patterns, decisions, log)

## Workflow

### 1. Identify contexts and categories

Read the existing skill and identify 3-8 distinct groupings (contexts).
For each, pick a short prefix (3-8 chars, lowercase, no hyphens) and
enumerate its categories.

The taxonomy is implicit in the filesystem - you don't register it
anywhere. You encode it by placing files at
`references/wiki/<context>/<category>/<concept>.md`.

### 2. Create the wiki + raw + brain skeleton

```bash
mkdir -p references/wiki references/raw
touch references/raw/.gitkeep
```

Copy the canonical templates from use-smart-skill:
- `references/_template.md`
- `references/_brain.md`

Create the four brain meta files with canonical headers:
- `references/_index.md` (with `<!-- GENERATED:START --> ... <!-- GENERATED:END -->` markers)
- `references/patterns.md`
- `references/decisions.md`
- `references/log.md`

(For greenfield skills `scripts/scaffold.sh <skill-name>` does this in one
call. Migration does it by hand.)

### 3. Extract wiki concepts

See `from-flat.md` for the deterministic conversion from flat references
to nested wiki concepts.

### 4. Move raw sources (if any)

If the original skill includes any source material (transcripts, PDFs,
screenshots, dumps), move them into `references/raw/` with their natural
filenames. Do NOT rename to match the taxonomy - raw is human territory.

For each wiki concept that draws from a raw source, add that raw file's
relative path to the concept's `sources:` frontmatter array.

### 5. Add commands (optional)

If a concept has a deterministic, executable action:

1. Create `scripts/<command>.sh` or `scripts/<command>.py`
2. Add `command: scripts/<command>.sh` to the concept's frontmatter
3. Scripts must be self-contained, idempotent, non-interactive

### 6. Rewrite the thin SKILL.md

Replace the monolithic content with a router. The new SKILL.md contains:

1. YAML frontmatter - `name`, `description` (trigger-rich), `license`, `compatibility` (if env-specific), `metadata`. See `references/wiki/smart/spec/skill-frontmatter.md` for the full field schema.
2. Brain Protocol block (mandatory, appears FIRST in body)
3. CCCCC Architecture table
4. When-to-Use bullets
5. How-to-Use section routing to specific wiki concepts
6. Pointer to `references/_index.md` for the live enumeration

Do NOT embed full concept content in SKILL.md. Point to paths.

Bump `metadata.version` major.

### 7. Populate the brain

- Add a bootstrap entry to `references/log.md` noting the migration.
- If the migration involved non-obvious choices (splitting vs merging
  concepts, context naming, etc.), add a `decisions.md` entry.

### 8. Lint + validate

```bash
bash scripts/lint.sh
```

`lint.sh` walks the filesystem, validates every wiki file, and rewrites
`references/_index.md`. See `references/wiki/smart/create/validation-checklist.md`
for the full pre-ship checklist.

## Sources

- (none)
