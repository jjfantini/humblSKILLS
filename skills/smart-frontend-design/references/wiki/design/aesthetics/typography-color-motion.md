---
title: "Use Distinctive Typography Color And Motion"
context: design
category: aesthetics
concept: typography-color-motion
description: "Translate the aesthetic thesis into type, theme, motion, space, and atmospheric details"
tags: typography, color, motion, layout, atmosphere
sources:
  - "references/raw/user-frontend-design-brief.md"
last_ingested: 2026-06-12
---

## Use Distinctive Typography Color And Motion

Design details should support the thesis. Choose characterful fonts, dominant
colors with sharp accents, motion with a few high-impact moments, spatial
composition with deliberate tension, and backgrounds that create atmosphere.

**Incorrect:**

```css
:root {
  --font-sans: Inter, system-ui, sans-serif;
  --primary: #7c3aed;
}
```

This starts from defaults rather than the product's emotional and visual
identity.

**Correct:**

```css
:root {
  --font-display: "Fraunces", Georgia, serif;
  --font-body: "IBM Plex Sans Condensed", sans-serif;
  --ink: #16130f;
  --paper: #f4ead8;
  --accent: #d4471f;
}
```

Use CSS variables for consistency. Prefer one orchestrated page-load reveal,
strong hover or scroll surprises, and purposeful textures such as grain,
gradient mesh, geometric patterning, layered transparency, dramatic shadows,
decorative borders, or custom cursors when they fit the aesthetic.

## Sources

- `references/raw/user-frontend-design-brief.md` - source for typography,
  color, motion, spatial composition, and visual-detail guidance.
