---
title: "Discover And Synthesize Existing Style"
context: design
category: discovery
concept: synthesize-style
description: "Infer the frontend's real design system from code, assets, copy, and constraints before creating UI"
tags: discovery, synthesis, design-system, frontend, codebase
sources:
  - "references/raw/frontend-design-SKILL.md"
last_ingested: 2026-07-30
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

### Ground it in the subject

If the brief does not pin down what the product or subject actually is, **pin it
yourself before designing**: name one concrete subject, its audience, and the
page's single job, then state that choice out loud so it can be corrected.

Then mine the subject itself. **The subject's own world — its materials,
instruments, artifacts, and vernacular — is where distinctive choices come
from.** A design for a cold-chain freight tool should be drawing on reefer
gauges and chain-of-custody forms, not on whatever palette is in season. This is
also the reliable antidote to the three default looks in
`anti-patterns/generic-ai-slop`: a free axis spent on the subject's world cannot
land on a generic default by accident.

Build with the brief's real content and subject matter throughout, not lorem
placeholders that let generic layout decisions hide.

### Use what you already know about this user

If there is anything in memory about the user's preferences, context on what
they are building, or designs made for them before, treat it as a hint and use
it. Repeating a question they have already answered — in this session or a
previous one — spends the one-question budget on nothing.

## Sources

- `references/raw/frontend-design-SKILL.md` - source for understanding
  context before coding, grounding the design in the subject's own world and
  vernacular, using memory of the user's preferences, and committing to a
  cohesive point of view.
