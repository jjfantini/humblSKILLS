---
title: "SKILL.md Frontmatter Fields (Agent Skills Spec)"
context: smart
category: spec
concept: skill-frontmatter
description: "Canonical SKILL.md frontmatter schema - 2 required, 4 optional fields, with per-field constraints"
tags: frontmatter, spec, skill-md, compatibility, allowed-tools, metadata, license, agentskills
sources:
  - "references/raw/agentskills-spec.md"
last_ingested: 2026-04-17
---

## SKILL.md Frontmatter Schema

The agentskills.io spec defines 6 frontmatter fields. Only `name` and
`description` are required. The rest are optional - add them only when they
carry real signal. Wrong frontmatter means the skill won't load or triggers
unreliably.

### Field Overview

| Field           | Required | Max chars | Purpose                                        |
|-----------------|----------|-----------|------------------------------------------------|
| `name`          | Yes      | 64        | Kebab-case identifier, must match directory    |
| `description`   | Yes      | 1024      | What the skill does AND when to trigger        |
| `license`       | No       | -         | License name or bundled file reference         |
| `compatibility` | No       | 500       | Env requirements (runtime, tools, network)     |
| `metadata`      | No       | -         | Arbitrary key-value map (author, version, etc.)|
| `allowed-tools` | No       | -         | Pre-approved tools (experimental, skip)        |

### `name` - required

Lowercase `a-z`, digits, hyphens. No leading/trailing/consecutive hyphens.
MUST match the parent directory name.

**Incorrect:**

```yaml
name: PDF-Processing   # uppercase
name: -pdf             # leading hyphen
name: pdf--processing  # consecutive hyphens
```

**Correct:**

```yaml
name: pdf-processing
```

### `description` - required

Must describe BOTH what the skill does AND the trigger scenarios.
Keyword-rich so agents match intent to skill.

**Incorrect (no trigger signal):**

```yaml
description: Helps with PDFs.
```

**Correct (what + when):**

```yaml
description: Extracts text and tables from PDFs, fills forms, merges files. Use when the user mentions PDFs, forms, or document extraction.
```

### `compatibility` - include only when env-specific

Max 500 chars. Add when the skill needs a specific runtime, system packages,
network access, or a particular agent product. Omit for pure-prose skills.

**Incorrect (filler):**

```yaml
compatibility: Works on most systems.
```

**Correct (concrete requirements):**

```yaml
compatibility: Requires bash, standard POSIX utilities (awk, sed, find, grep), and a writable filesystem.
compatibility: Requires Python 3.14+ and uv.
compatibility: Designed for Claude Code (or similar products).
```

Rule of thumb: if the skill has `scripts/` invoking binaries, or depends on
a specific runtime/product, set it. Otherwise delete the field entirely.

### `metadata` - arbitrary key-value

Used for author, version, tags not covered elsewhere. Keep keys unique-ish
to avoid collisions.

```yaml
metadata:
  author: example-org
  version: "1.0"
```

### `license` - optional

Short: SPDX name or pointer to a bundled file.

```yaml
license: MIT
license: Proprietary. LICENSE.txt has complete terms
```

### `allowed-tools` - experimental, skip

Space-separated pre-approved tools. Support varies between agent
implementations. Skip unless the skill is a narrow command runner AND the
target agent honors it.

```yaml
allowed-tools: Bash(git:*) Bash(jq:*) Read
```

## Sources

- `references/raw/agentskills-spec.md` - full spec, all field definitions,
  constraints, and canonical examples.
