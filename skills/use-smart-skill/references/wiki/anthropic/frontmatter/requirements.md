---
title: "Frontmatter Field Requirements (Anthropic Spec)"
context: anthropic
category: frontmatter
concept: requirements
description: "What the top-level YAML frontmatter MUST include, what's optional, character limits, and which fields humblSKILLS nests under metadata."
tags: frontmatter, yaml, compliance, anthropic
sources:
  - "references/raw/anthropic-skill-building-guide.pdf"
last_ingested: 2026-04-19
---

## Anthropic-compliant Frontmatter

The YAML frontmatter is the first level of progressive disclosure - it is always loaded in Claude's system prompt. Keep the top level clean so Claude can route correctly without loading the body.

### Required at top level

| Field         | Constraint                                             |
|---------------|--------------------------------------------------------|
| `name`        | kebab-case, no spaces/capitals/underscores, matches folder name |
| `description` | <=1024 chars, no `<` or `>`, includes WHAT + WHEN + trigger phrases |

### Optional at top level

| Field            | Purpose                                               |
|------------------|-------------------------------------------------------|
| `license`        | For open-source skills (e.g. `MIT`, `Apache-2.0`)     |
| `compatibility`  | 1-500 chars; declares env requirements (bash, Python 3.14+, network, specific MCP) |
| `allowed-tools`  | Restricts tool access, e.g. `"Bash(python:*) WebFetch"` |
| `metadata`       | Bag for custom key-value pairs (see below)            |

### humblSKILLS fields live under `metadata:`

Anthropic's "All optional fields" example treats `version`, `author`, `tags`, `mcp-server`, etc. as custom metadata. humblSKILLS places every extension key there too, so top-level stays Anthropic-compliant.

**Incorrect (humblSKILLS fields leaking to top level):**

```yaml
---
name: my-skill
description: ...
version: 1.0.0        # humblSKILLS extension, not Anthropic-owned
tags: [productivity]  # humblSKILLS extension
platforms: [claude-code]
preserve:
  - references/log.md
---
```

**Correct (humblSKILLS fields nested under metadata):**

```yaml
---
name: my-skill
description: ...
license: MIT
compatibility: Requires Python 3.14+ and the github MCP server.
allowed-tools: "Bash(python:*) Read Write"
metadata:
  author: jjfantini
  version: 1.0.0
  tags: [productivity, automation]
  platforms: [claude-code, cursor]
  preserve:
    - references/log.md
---
```

### Character budget

- Full frontmatter block ideally under ~1.5 KB (it always loads).
- `description` hard cap: 1024 characters.
- `compatibility`: 1-500 characters.
- Prefer one description paragraph with embedded triggers over multi-paragraph prose.

### Reserved names

- Skill names starting with `claude` or `anthropic` are reserved and will be rejected on upload.

## Sources

- `references/raw/anthropic-skill-building-guide.pdf` - Chapter 2 "Planning and design" / YAML frontmatter section + Reference B "YAML frontmatter"
