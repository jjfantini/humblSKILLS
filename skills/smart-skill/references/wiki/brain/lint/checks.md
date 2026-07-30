---
title: "What scripts/lint.sh Verifies and How to Fix Findings"
context: brain
category: lint
concept: checks
description: "Keeps brain self-consistent as the skill scales"
tags: lint, validation, checks, brain, health
sources: []
last_ingested: 2026-04-16
command: scripts/lint.sh
---

## Lint Checks

`scripts/lint.sh` is the brain's health check. Run every 2 weeks, before
releases, or after any wiki change. Exit code is 0 only when all hard
checks pass.

The taxonomy (valid `<context>`/`<category>` pairs) is **derived from the
filesystem** - there is no registry file. Any `<context>/<category>` path
that exists under `references/wiki/` is valid by definition.

## Checks Performed

### 1. Path / frontmatter triple match

Every wiki file at `references/wiki/<a>/<b>/<c>.md` MUST have frontmatter
`context: <a>`, `category: <b>`, `concept: <c>`. Catches misfiled files
and typos.

**Finding example:**
```
MISMATCH: references/wiki/brain/lint/checks.md
  path:   context=brain, category=lint, concept=checks
  front:  context=brain, category=lint, concept=lint-checks
```

**Fix:** rename the file OR change the `concept:` frontmatter so both agree.

### 2. Required frontmatter

Every wiki file MUST have: `title`, `context`, `category`, `concept`,
`description`, `tags`, `sources`, `last_ingested`.

**Finding example:**
```
MISSING FIELDS: references/wiki/brain/ingest/workflow.md
  missing: last_ingested
```

**Fix:** add the missing field.

### 3. Orphan raw files

Raw files not cited by any wiki concept's `sources:` array.

**Finding example:**
```
ORPHAN RAW: references/raw/old-dump.md
  not cited by any wiki concept
```

**Fix:** either (a) ingest into a wiki concept that cites it in `sources:`,
or (b) delete the raw file if it's no longer needed.

### 4. Broken sources

`sources:` paths in wiki frontmatter that don't resolve to a real file
under `references/raw/`.

**Finding example:**
```
BROKEN SOURCE: references/wiki/content/hooks/disaster.md
  sources[0]: references/raw/does-not-exist.md
```

**Fix:** correct the path OR remove the entry from `sources:`.

### 5. Orphan wiki concepts

Wiki concepts with empty `sources:` array. This is a warning, not an
error - sometimes a concept is pure synthesis with no raw origin. The
lint output flags them so you can audit.

### 6. Stale entries

Wiki files whose `last_ingested` is older than the configured threshold
(default: 180 days, override via `STALE_DAYS` env var). Surfaces
candidates for review; doesn't fail.

### 7. Contradictions (heuristic)

Two wiki files sharing the same `concept:` value (even under different
paths). Almost always a legitimate naming reuse (e.g. `workflow` across
multiple categories); flagged as WARN for human audit.

## Side Effect: _index.md Regeneration

`lint.sh` rewrites everything between the `<!-- GENERATED:START -->` and
`<!-- GENERATED:END -->` markers in `references/_index.md`. Walks the
filesystem, reads every wiki file's frontmatter, emits:

1. `## Summary` - context -> categories TOC (no concepts)
2. `## Wiki` - full `### context` / `#### category` / concept-bullet tree
3. `## Raw Sources` - raw files + their citers
4. `## Scripts` - script enumeration

Never hand-edit the generated region - edit the wiki frontmatter and
re-run lint.

## Exit Codes

| Code | Meaning                                              |
|------|------------------------------------------------------|
| 0    | All hard checks passed                               |
| 1    | One or more hard findings (mismatch, missing, broken)|
| 2    | Invocation error (wrong CWD, missing files)          |

## Sources

- (none)
