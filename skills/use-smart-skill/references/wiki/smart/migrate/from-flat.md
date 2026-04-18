---
title: "Convert Flat references/*.md to Nested Wiki Concepts"
context: smart
category: migrate
concept: from-flat
description: "Deterministic path from flat skill to Smart Skill, no semantic loss"
tags: migrate, flat-to-nested, wiki, frontmatter, brain
sources: []
last_ingested: 2026-04-16
---

## Convert Flat References

Flat skills keep references as `references/<name>.md` files at a single
level. Smart Skills nest them at
`references/wiki/<context>/<category>/<concept>.md` with frontmatter
identity. This concept covers the deterministic conversion.

**Incorrect (keeping the flat layout):**

```
references/
  some-topic.md          # single level, mixes multiple concepts
  another-thing.md
```

The flat layout can't express more than one concept per topic and has no
structural hook for brain files or raw sources.

**Correct (nested wiki + brain):**

```
references/
  _template.md  _brain.md
  _index.md  patterns.md  decisions.md  log.md
  wiki/
    smart/
      create/
        workflow.md
        validation-checklist.md
      migrate/
        workflow.md
        from-flat.md
  raw/
    .gitkeep
```

## Conversion Steps

### 1. Create the new skeleton

```bash
cd path/to/skill
mkdir -p references/wiki references/raw
touch references/raw/.gitkeep

for f in index patterns decisions log; do
  [[ -f "references/$f.md" ]] || touch "references/$f.md"
done

# Copy _brain.md and _template.md from use-smart-skill/references/
```

Seed `references/_index.md` with the sentinel-bracketed placeholder so
`lint.sh` can regenerate it.

### 2. Convert each flat reference

For each `references/<name>.md` (that is not a brain meta file):

1. Read the file. Decide how many concepts it contains. A monolithic
   file may split into 2-5 concept files.

2. For each concept extracted, create
   `references/wiki/<context>/<category>/<concept>.md` with frontmatter:

   ```yaml
   ---
   title: "..."
   context: <context>
   category: <category>
   concept: <concept>          # MUST match filename stem
   description: "..."
   tags: ...
   sources: []                 # populate if you drop originating files into raw/
   last_ingested: <today>
   ---
   ```

3. If the flat file resists splitting, use `concept: overview` as the
   filename: `references/wiki/<context>/<category>/overview.md`.

4. Delete the old flat reference once all concepts are extracted and
   verified.

### 3. Move any source material to raw/

If the skill has any source files (transcripts, exports, notes,
screenshots), move them into `references/raw/` with their natural
filenames. Do NOT rename to match the taxonomy.

For each wiki concept that draws from a raw source, add its relative
path to the concept's `sources:` array and update `last_ingested`.

### 4. Update SKILL.md

- Inject the Brain Protocol block BEFORE When-to-Use.
- Rewrite the How-to-Use section to route to the new nested wiki paths.
- Point at `references/_index.md` for the live enumeration.
- Bump `metadata.version` major.

### 5. Populate the brain

- `log.md`: bootstrap entry documenting the migration.
- `decisions.md`: entry if any non-trivial splits/merges occurred.

### 6. Lint

```bash
bash scripts/lint.sh
```

`lint.sh` regenerates `_index.md`, validates every wiki file's
path/frontmatter match, and reports orphans. Fix findings, re-run until
clean. See `references/wiki/smart/create/validation-checklist.md`.

## Sources

- (none)
