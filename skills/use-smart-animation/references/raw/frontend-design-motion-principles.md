# Frontend-design motion & anti-generic principles

Distilled from the Claude "frontend-design" plugin skill, researched 2026-07-23.
This is the taste layer: WHERE and IF motion should exist, and how to avoid
motion that reads as AI-generated slop.

## Core directive: leverage motion deliberately
Think about *where and if* animation serves the subject. Named motion moments:
- **Page-load sequence** — an orchestrated entrance.
- **Scroll-triggered reveal** — content animating in as the reader scrolls.
- **Hover micro-interactions** — small responses on interactive elements.
- **Ambient atmosphere** — subtle continuous background motion.

## The judgments it insists on
- **Orchestration beats scatter.** One deliberate, composed motion sequence
  lands harder than many small unrelated effects sprinkled across the page.
- **Less is often more — over-animation reads as AI-generated.** Excess motion
  is explicitly a tell of generic AI output. Restraint signals craft.
- **Motion must serve the subject**, not exist for its own sake. Chosen only
  where it reinforces the brief's direction.
- **Spend your boldness in one place.** Let the signature element be the one
  memorable thing; keep everything around it quiet. Applied to motion:
  concentrate it, don't diffuse it.

## Quality floor (never simplify away)
Respect `prefers-reduced-motion`. Responsive to mobile. Visible keyboard focus.
Motion is never allowed to break the accessibility baseline.

## Anti-generic design (the three AI defaults to avoid)
Current AI-generated design falls into three clusters — deliberately avoid them
where the brief leaves the axis free (if the brief pins one, follow it exactly):
1. Warm cream background (~`#F4F1EA`) + high-contrast serif display + terracotta
   accent.
2. Near-black background + a single bright acid-green or vermilion accent.
3. Broadsheet layout with hairline rules, zero border-radius, dense newspaper
   columns.
These are "defaults rather than choices" that appear regardless of subject.

## Adjacent taste notes (why designs feel templated)
- **Typography carries personality.** Pair display + body faces deliberately —
  not the same families you'd reach for on any other project. Intentional type
  scale, weights, widths, spacing.
- **Color:** 4–6 named hex values in a token system, every choice derived from
  the plan, not ad hoc.
- **Structure as information:** numbering / eyebrows / dividers must encode
  something true. Numbered markers (01/02/03) only if the content is genuinely a
  sequence.
- **Process:** brainstorm a compact token system (color, type, layout, ONE
  signature element), then critique it against the brief — if any part reads
  like the generic default, revise it. Only then write code. "Remove one
  accessory" before shipping.
