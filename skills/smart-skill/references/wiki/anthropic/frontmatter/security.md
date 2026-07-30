---
title: "Frontmatter Security Restrictions (Anthropic Spec)"
context: anthropic
category: frontmatter
concept: security
description: "Forbidden characters and reserved name prefixes in YAML frontmatter, plus why they exist (prompt injection + namespace protection)."
tags: frontmatter, security, prompt-injection, anthropic
sources:
  - "references/raw/anthropic-skill-building-guide.pdf"
last_ingested: 2026-04-19
---

## Forbidden in Frontmatter

### No XML angle brackets

`<` and `>` are banned anywhere in the frontmatter block.

**Why:** the frontmatter loads into Claude's system prompt. XML-style tags are how the host injects structural directives (`<user_query>`, `<system_reminder>`, etc.). Any `<...>` in the frontmatter could be interpreted as a new directive, a closing tag for the host's wrappers, or a vector for prompt injection from untrusted skill authors.

**Incorrect:**

```yaml
---
name: my-skill
description: "Handles <code>TypeScript</code> and other web tools."
---
```

**Correct:**

```yaml
---
name: my-skill
description: "Handles TypeScript code and other web tools."
---
```

### Reserved name prefixes

Skills whose `name:` starts with `claude` or `anthropic` are rejected on upload. These namespaces are reserved for first-party skills.

```yaml
# Wrong - reserved prefix
name: claude-pr-reviewer
name: anthropic-deploy-helper

# Correct - unique namespace
name: acme-pr-reviewer
name: acme-deploy-helper
```

### What IS allowed

- All standard YAML types (strings, numbers, booleans, lists, objects)
- Custom `metadata:` fields (humblSKILLS uses this for version, tags, platforms, preserve, requires)
- Descriptions up to 1024 characters
- Any valid YAML escape sequence that does NOT produce `<` or `>`

### Out of YAML, also banned

- Code execution in YAML is blocked by Anthropic's safe YAML parser (no `!!python/object`, no custom tags that instantiate code).

## Sources

- `references/raw/anthropic-skill-building-guide.pdf` - "Security restrictions" (Chapter 2) and Reference B "Security notes"
