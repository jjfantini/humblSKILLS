---
title: "Verify Frontend Design Quality"
context: design
category: verification
concept: review-checklist
description: "Check functionality, accessibility, responsiveness, craft, and distinctive identity before claiming design work is done"
tags: verification, checklist, accessibility, responsive, quality
sources:
  - "references/raw/frontend-design-SKILL.md"
last_ingested: 2026-07-30
---

## Verify Frontend Design Quality

Before completion, verify both engineering behavior and design quality. The UI
should work, fit the stack, respond across breakpoints, respect accessibility,
and feel like it came from a specific design system.

**Incorrect:**

```markdown
Looks good. The component was added.
```

That checks existence, not quality.

**Correct:**

```markdown
Verified: lint and tests pass; keyboard focus is visible; reduced motion is
handled; mobile and desktop layouts preserve hierarchy; typography, palette,
motion, and visual details all reinforce the "industrial field manual" thesis.
```

Use local evidence where possible: tests, lints, browser inspection,
screenshots, accessibility checks, or responsive snapshots. If a verification
cannot be run, say exactly what was not verified.

### Critique while building, not only at the end

Take screenshots as you go if the environment supports it — a picture is worth
1000 tokens, and most craft failures are visible instantly and invisible in
code review.

### The quality floor, built without announcing it

Responsive down to mobile. Visible keyboard focus. Reduced motion respected.
These are not features to report; they are the floor. Motion implementation
detail lives outside this skill — see the animation skill for the
reduced-motion substitutions themselves.

### Remove one accessory

Spend boldness in **one** place: let the signature element be the one memorable
thing and keep everything around it quiet and disciplined. Then cut any
decoration that does not serve the brief.

Chanel's rule, applied literally before you call it done: **look in the mirror
and take one thing off.** Name the thing you removed.

**Incorrect (every idea survived to production):**

```markdown
Shipped: grain overlay, custom cursor, parallax hero, marquee ticker, gradient
mesh background, hover tilt on all cards, and the scroll-linked callout lines.
```

Seven signatures means none of them is the signature.

**Correct:**

```markdown
Shipped: scroll-linked callout lines as the single signature; everything else
quiet. Removed the custom cursor — it competed with the callouts for the same
attention and served no part of the brief.
```

If you have somewhere to jot notes about what you have tried, use it — human
designers accumulate memory across projects and deliberately avoid repeating
themselves. In this skill that place is `references/patterns.md` and
`references/log.md`.

## Sources

- `references/raw/frontend-design-SKILL.md` - source for production-grade,
  cohesive, memorable output, the unannounced quality floor, critique-while-
  building with screenshots, spending boldness in one place, the remove-one-
  accessory rule, and keeping notes for future passes.
