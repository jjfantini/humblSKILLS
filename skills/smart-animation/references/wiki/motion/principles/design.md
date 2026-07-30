---
title: "Deliberate Motion: Orchestration Over Scatter"
context: motion
category: principles
concept: design
description: "When and whether to animate at all — concentrate motion on one signature moment; scattered effects read as AI-generated."
tags: motion-design, restraint, orchestration, anti-generic
sources:
  - "references/raw/frontend-design-motion-principles.md"
last_ingested: 2026-07-23
---

## Deliberate Motion: Orchestration Over Scatter

The most common animation mistake is not bad easing — it is animating
*everything*. Excess motion is a documented tell of AI-generated design. Decide
**where and if** motion serves the subject before writing a single keyframe.

The four motion moments worth spending on: an orchestrated **page-load
sequence**, **scroll-triggered reveals**, **hover micro-interactions**, and
subtle **ambient atmosphere**. Pick the one the content actually calls for.

**Incorrect (scatter — every element animates independently):**

```jsx
<Hero className="fade-up" />
<Card className="fade-up bounce" />       {/* different effect */}
<Button className="pulse wobble" />        {/* two effects at once */}
<Stat className="spin-in" />               {/* fifth unrelated effect */}
{/* Nothing is emphasized; the page feels busy and machine-made. */}
```

**Correct (one orchestrated moment, everything else quiet):**

```jsx
{/* One composed entrance: children stagger in, then the page is still. */}
<Section className="reveal-sequence">      {/* staggers its children once */}
  <Hero />
  <Card />
  <Stat />
</Section>
<Button className="hover-lift" />          {/* the ONE interactive response */}
```

Rules distilled from the frontend-design skill:
- **Spend boldness in one place.** One signature motion the page is remembered
  by; keep the rest still.
- **Orchestration beats scatter.** A single composed sequence lands harder than
  many small unrelated effects.
- **Motion serves the subject.** If an animation doesn't reinforce the brief's
  direction, delete it.
- Before shipping, "remove one accessory" — cut the least-necessary animation.

Then verify performance (`motion/principles/performance`) and accessibility
(`motion/principles/accessibility`) — taste without those two is unfinished.

## Sources

- `references/raw/frontend-design-motion-principles.md` — the where/if/restraint
  directives and the over-animation-as-AI-tell rule.
