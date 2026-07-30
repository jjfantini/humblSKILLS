---
title: "Implement Production-Grade Frontend Code"
context: design
category: implementation
concept: production-code
description: "Build real working UI in the detected stack with code complexity matched to the aesthetic vision"
tags: implementation, production, react, css, frontend
sources:
  - "references/raw/frontend-design-SKILL.md"
last_ingested: 2026-07-30
---

## Implement Production-Grade Frontend Code

The deliverable is working code, not a mock or prompt. Use the repo's existing
framework, styling primitives, import conventions, accessibility patterns, and
test setup unless the user explicitly asks for a standalone artifact.

**Incorrect:**

```markdown
Here is a design direction you can implement later.
```

The skill exists to ship usable frontend code with high craft.

**Correct:**

```markdown
Implement the page using the existing Next.js route, tokenized CSS variables,
semantic landmarks, responsive states, reduced-motion handling, and focused
component tests where the behavior is non-trivial.
```

Match code complexity to the aesthetic. Maximalist concepts may need layered
backgrounds, staggered animations, masking, or scroll choreography. Refined
minimal concepts need tighter spacing, type scale, contrast, and state polish,
not extra decoration. Elegance is executing the chosen vision well, not
executing a restrained vision.

Build from the plan produced in `process/two-pass-plan` — follow the revised
plan exactly and derive every color and type decision from it, rather than
re-deciding in the editor.

### Watch CSS selector specificity

Be careful how you structure selector specificity. It is easy to generate
classes that silently cancel each other out — especially a type-based selector
like `.section` colliding with an element-based one like `.cta`. Spacing is
where this bites most often: paddings and margins between sections.

**Incorrect (two rules, one silently wins):**

```css
.section { padding-block: 6rem; }
.section .cta { margin-block: 4rem; }
.cta { margin-block: 0; }        /* loses to the line above; the gap stays */
```

**Correct (one owner per property, no cross-cutting override):**

```css
.section        { padding-block: var(--space-section); }
.section > * + * { margin-block-start: var(--space-flow); }
.cta            { /* owns its own internals only, never sibling spacing */ }
```

Decide once which selector owns spacing between sections, and keep every other
rule out of that property.

## Sources

- `references/raw/frontend-design-SKILL.md` - source for real working code,
  production quality, matching implementation complexity to vision, deriving
  code from the revised plan, and the CSS selector-specificity warning.
