---
title: "Discover And Synthesize Existing Style"
context: design
category: discovery
concept: synthesize-style
description: "Infer the frontend's real design system from code, assets, copy, and constraints before creating UI"
tags: discovery, synthesis, design-system, frontend, codebase
sources:
  - "references/raw/user-frontend-design-brief.md"
last_ingested: 2026-06-12
---

## Discover And Synthesize Existing Style

After the one design-essence answer, inspect the frontend before designing.
Look for framework, styling system, component primitives, tokens, assets,
screenshots, copy tone, existing layout patterns, accessibility conventions,
and performance constraints.

**Incorrect:**

```markdown
I'll build a polished React card with a modern gradient and rounded corners.
```

That ignores the product's existing language and usually creates portable AI
slop: plausible, context-free, and forgettable.

**Correct:**

```markdown
Evidence found: dense B2B tables, monospace metric labels, muted clay palette,
thin borders, and short operational copy. Direction: industrial editorial UI
with sharp data panels, compressed type, and one high-contrast hazard accent.
```

Synthesis should name the aesthetic thesis before code. Include purpose,
audience, constraints, typography, palette, motion, composition, and the one
detail the user will remember.

## Sources

- `references/raw/user-frontend-design-brief.md` - source for understanding
  context before coding and committing to a cohesive point of view.
