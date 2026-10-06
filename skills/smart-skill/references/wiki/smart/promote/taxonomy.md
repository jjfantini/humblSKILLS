---
title: "Fill Category, Role, and Tags So the CLI Can File the Skill"
context: smart
category: promote
concept: taxonomy
description: "metadata.category, metadata.role, and metadata.tags decide where a skill appears in humblskills search and the skill browser; promote.sh writes and validates them"
tags: promote, taxonomy, category, role, tags, search, metadata
sources:
  - "references/raw/user-request-promote.md"
last_ingested: 2026-10-02
---

## Taxonomy Is How Skills Get Found

The CLI groups the browser as **Registry › Category › Role › skill**, filters
with `humblskills search --category=<c> --role=<r>`, and fuzzy-matches tags.
A skill with a placeholder category doesn't build; one with vague tags is
never found. The category and role sets are closed and validated by
`build-registry`.

| Field | Values | Rule |
|---|---|---|
| `category` | `development`, `design`, `writing`, `meta` | required, exactly one |
| `role` | `fde`, `ds`, `sdr` | who it serves; required for happySKILLS, optional for humblSKILLS |
| `tags` | kebab-case keywords | required; 3+ recommended |

- **development** - git/workflow tooling, integrations, APIs, data and infra
- **design** - frontend, UI/UX, creative and media generation
- **writing** - content, copy, editing
- **meta** - skill authoring, project onboarding
- **fde** forward-deployed engineer, **ds** data scientist, **sdr** sales
  development representative

**Incorrect (scaffold placeholders and free-form values):**

```yaml
metadata:
  category: TODO  # required, one of: development, design, writing, meta
  role: TODO
  tags: [TODO, Code Review]
```

**Correct (written by `--category development --role fde --tags code-review,git`):**

```yaml
metadata:
  category: development
  role: fde
  tags: [code-review, git]
```

## Choosing

- Category: what the skill *does*, not who asked for it.
- Role: the job it serves. A skill any engineer would use is unscoped on
  humblSKILLS (`--role none` removes the field); happySKILLS always takes a role.
- Tags: the words a user would type into search - tools, file types, task
  verbs. `--tags` adds to the existing list and normalizes `Code Review` to
  `code-review`. Propose all three, then confirm them with the user.

Adding a category or role is a CLI change (`cli/internal/frontmatter/validate.go`),
not a per-skill decision.

## Sources

- `references/raw/user-request-promote.md` - "ensure the skill is tagged and
  categorized so role/responsibility is filled properly and the CLI can
  organize skills".
