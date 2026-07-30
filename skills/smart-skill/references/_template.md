---
title: "Clear, Action-Oriented Title (5-10 words)"
context: smart                              # MUST match first path segment under wiki/
category: create                            # MUST match second path segment under wiki/
concept: workflow                           # MUST match filename stem (this file)
description: "What this concept covers and why it matters (quantify when possible, e.g. '3x faster agent context loading')"
tags: keyword1, keyword2, keyword3
sources:                                    # Raw files this concept distills (may be empty at creation)
  - "references/raw/example-dump.md"
  - "references/raw/Screenshot 2026-04-16 at 09.32.png"
last_ingested: 2026-04-16                   # ISO date; update when sources are re-read
command: scripts/example.sh                 # Optional - path to deterministic script
---

## [Concept Title]

[1-2 sentence explanation of what this concept covers and why it matters.
Focus on concrete impact - token cost, discoverability, maintainability,
behavioural change.]

**Incorrect (describe the problem):**

```markdown
<!-- or ```sql, ```python, etc. - match the skill's domain -->
<!-- Comment explaining what makes this wrong -->
[Bad example]
```

**Correct (describe the solution):**

```markdown
<!-- Comment explaining why this is better -->
[Good example]
```

[Optional: additional context, edge cases, trade-offs, or when to deviate.]

## Sources

[Optional prose annotation of the `sources:` array - explain which raw file
contributed what to this concept. Helps future ingest/lint passes reason
about provenance.]

- `references/raw/example-dump.md` - primary source for [claim X]
- `references/raw/Screenshot 2026-04-16 at 09.32.png` - evidence for [claim Y]

## Command

[Optional section - include only when `command` is set in frontmatter.]

Run the associated script:

```bash
bash scripts/example.sh [args]
```

Reference: [Relevant docs](https://example.com)

---

## Frontmatter Rules (delete this section in concrete concept files)

- `context` / `category` / `concept` triple MUST match the filesystem path
  `references/wiki/<context>/<category>/<concept>.md`. `scripts/lint.sh`
  enforces this.
- The taxonomy is derived from the filesystem by `lint.sh` - there is no
  separate registry. If a `<context>` or `<category>` path exists, it's valid.
- `concept` MUST equal the filename stem (kebab-case, 2-5 words).
- `description` is a one-line summary of what this concept covers and why it matters.
- `sources` is a YAML list of relative paths to files under `references/raw/`.
  Always quote paths (filenames may contain spaces or special characters).
  Empty list is allowed at creation time but `lint.sh` will flag it as an
  orphan.
- `last_ingested` is an ISO date (YYYY-MM-DD). Update whenever the concept
  is re-derived from its sources.
