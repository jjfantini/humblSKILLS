---
title: "Inline-Header Vertical Lists"
context: humanize
category: formatting
concept: inline-header-lists
description: "AI outputs bullet lists where every item starts with a bolded label and colon."
tags: formatting, lists, bullets, ai-tells
sources:
  - "references/raw/Wikipedia_Signs_of_AI_writing.pdf"
last_ingested: 2026-04-17
---

## Inline-Header Vertical Lists

AI outputs lists where every item starts with a bolded label and colon. Rewrite as prose or strip the bolded labels.

**Incorrect:**

```markdown
- **Speed:** Code generation is faster.
- **Quality:** Output quality has improved.
- **Adoption:** Usage continues to grow.
```

**Correct:**

```markdown
The update speeds up code generation, improves output quality, and adoption keeps climbing.
```

Rule: if every bullet has the shape `**Word:** sentence`, the list is AI scaffolding. Convert to a sentence. Keep bulleted lists for genuinely enumerable items (steps, options, data).

## Sources

- `references/raw/Wikipedia_Signs_of_AI_writing.pdf` - tell #18 in the formatting tells list
