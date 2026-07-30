---
title: "Two Passes: Plan A Token System, Then Critique It Before Building"
context: design
category: process
concept: two-pass-plan
description: "Brainstorm a compact color/type/layout/signature token system, then review it against the brief for genericness and revise before writing any code"
tags: process, planning, token-system, critique, wireframe, frontend
sources:
  - "references/raw/frontend-design-SKILL.md"
last_ingested: 2026-07-30
---

## Two Passes: Plan A Token System, Then Critique It Before Building

Work in two passes. Code comes only after the second one.

### Pass 1 — a compact token system

Brainstorm a short design plan from the brief, with exactly four parts:

- **Color** — the palette as **4–6 named hex values**, not adjectives.
- **Type** — typefaces for **2+ roles**: a characterful display face used with
  restraint, a complementary body face, and a utility face for captions or data
  if the content needs one.
- **Layout** — a layout concept, expressed as one-sentence prose plus **ASCII
  wireframes**, so alternatives can be compared cheaply before committing.
- **Signature** — the single unique element this page will be remembered by, and
  how it embodies the brief.

### Pass 2 — critique the plan against the brief

Review the plan before building. The test is concrete: **work through a similar
prompt and see whether you arrive somewhere similar.** Any part that reads like
the generic default you would produce for any comparable page gets revised — and
you say what you changed and why.

Only after confirming the plan is specific to *this* brief do you write code,
following the revised plan exactly and deriving every color and type decision
from it.

**Incorrect (planning in adjectives, then improvising in the editor):**

```markdown
Plan: modern and clean, with a nice sans-serif and a blue accent. Let's start
building and see how it feels.
```

Nothing here can be critiqued, so pass 2 has nothing to bite on and every
decision gets made ad hoc in CSS.

**Correct (a plan concrete enough to be wrong):**

```markdown
Color:  --slate #1B1F23  --bone #E8E3D9  --rust #A6431F
        --signal #F2C744  --hair #3A4046
Type:   display "格" GT Sectra Display (restraint: headline + signature only)
        body    Söhne
        utility Söhne Mono (spec tables, part numbers)
Layout: full-bleed part diagram left, spec column right, no centered content.
        +----------------------+-------------+
        |   exploded diagram   | spec table  |
        |   (sticky, 60vw)     | (scrolls)   |
        +----------------------+-------------+
Signature: the diagram's callout lines redraw as the spec column scrolls.

Pass 2: the bone/rust pairing is one hue away from the cream+terracotta default
(see anti-patterns/generic-ai-slop). Revised rust -> #A6431F oxide and dropped
the warm background to --slate, so the palette reads as machine shop rather than
artisanal. Kept the signature; it is specific to exploded diagrams.
```

Do most of this planning and iteration **in thinking**. Show ideas to the user
only once confidence is high enough that they will land.

## Sources

- `references/raw/frontend-design-SKILL.md` — the two-pass structure, the
  four-part token system with its 4–6 hex / 2+ typeface specifics, the
  ASCII-wireframe instruction, the self-similarity test, and the direction to
  plan in thinking.
