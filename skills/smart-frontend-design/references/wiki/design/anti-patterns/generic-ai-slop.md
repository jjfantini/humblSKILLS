---
title: "Reject Generic AI Frontend Slop"
context: design
category: anti-patterns
concept: generic-ai-slop
description: "The three looks AI-generated design currently clusters around, why they are defaults rather than choices, and the one case where using them is correct"
tags: anti-patterns, ai-slop, calibration, typography, layouts, frontend
sources:
  - "references/raw/frontend-design-SKILL.md"
last_ingested: 2026-07-30
---

## Reject Generic AI Frontend Slop

AI-generated design currently clusters around **three specific looks**. Knowing
them by name is the calibration — they are hard to spot from the inside, because
each one is genuinely attractive.

1. **Warm cream + serif + terracotta** — a near-`#F4F1EA` background, a
   high-contrast serif display face, a terracotta accent.
2. **Near-black + one acid accent** — a near-black background with a single
   bright acid-green or vermilion accent.
3. **Broadsheet** — hairline rules, zero border-radius, dense newspaper-like
   columns.

All three are legitimate for *some* briefs. The problem is that they show up
**regardless of subject**, which is what makes them defaults rather than choices.

**Incorrect (a free axis spent on a default):**

```markdown
Brief: a dashboard for freight dispatchers tracking cold-chain temperature.
Direction: cream background, high-contrast serif headline, terracotta accent.
```

Nothing in cold-chain freight suggests cream and terracotta. The brief left the
palette free, and the free axis went straight to cluster #1.

**Correct (the same free axis, spent on the subject's own world):**

```markdown
Brief: a dashboard for freight dispatchers tracking cold-chain temperature.
Direction: refrigerated-trailer instrumentation — condensation-grey panels,
amber and hard-cyan threshold states borrowed from reefer gauges, stencilled
unit numbers, one oversized temperature readout as the anchor.
```

### The exception that matters

**Where the brief pins down a visual direction, follow it exactly — the brief's
own words always win, including when it asks for one of these three looks.**
Overriding a requested aesthetic to prove independence is its own failure. The
rule applies only to axes the brief leaves free.

### Do not converge across generations

Vary light and dark themes, font systems, density, layout, and decorative
language based on the actual context. Reaching for the same "safe" answer every
time is the tell, even when the answer is a good one — Space Grotesk as a
repeated fallback is the classic example.

There is a real balance here, the same one a hired designer navigates: doing
what you are good at, versus treating each project as a chance to experiment.
Not taking a risk is itself a risk.

## Sources

- `references/raw/frontend-design-SKILL.md` — the three-cluster calibration, the
  brief's-words-win exception, and the ban on cross-generation convergence.
