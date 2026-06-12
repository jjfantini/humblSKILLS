---
title: "Reject Generic AI Frontend Slop"
context: design
category: anti-patterns
concept: generic-ai-slop
description: "Avoid overused fonts, purple gradients, predictable layouts, and context-free component patterns"
tags: anti-patterns, ai-slop, typography, layouts, frontend
sources:
  - "references/raw/user-frontend-design-brief.md"
last_ingested: 2026-06-12
---

## Reject Generic AI Frontend Slop

Never default to the recurring AI frontend look: Inter or system fonts,
purple gradients on white, centered hero plus cards, evenly distributed
palettes, vague glassmorphism, and components that could belong to any app.

**Incorrect:**

```markdown
Use a modern SaaS hero with purple gradient blobs, rounded cards, Inter, and
subtle shadows.
```

It is technically acceptable but visually interchangeable.

**Correct:**

```markdown
Use weathered shipping-label typography, off-register stamp accents, cramped
manifest grids, and one oversized customs seal as the interaction anchor.
```

Do not converge on the same "safe" choices across generations. Vary light and
dark themes, font systems, density, layout, and decorative language based on
the actual context. Avoid Space Grotesk as a repeated fallback.

## Sources

- `references/raw/user-frontend-design-brief.md` - source for explicit bans on
  generic AI aesthetics and repeated default choices.
