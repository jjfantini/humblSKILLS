---
title: "Verify Frontend Design Quality"
context: design
category: verification
concept: review-checklist
description: "Check functionality, accessibility, responsiveness, craft, and distinctive identity before claiming design work is done"
tags: verification, checklist, accessibility, responsive, quality
sources:
  - "references/raw/user-frontend-design-brief.md"
last_ingested: 2026-06-12
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

## Sources

- `references/raw/user-frontend-design-brief.md` - source for production-grade,
  functional, cohesive, memorable, and meticulously refined output.
