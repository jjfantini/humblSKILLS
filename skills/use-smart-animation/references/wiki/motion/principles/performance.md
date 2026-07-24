---
title: "60fps Motion: Animate Only Compositor Properties"
context: motion
category: principles
concept: performance
description: "Animate transform/opacity only; layout-triggering props reflow every frame and drop below 60fps. Covers will-change, FLIP, the frame budget."
tags: performance, compositor, transform, will-change, flip, 60fps
sources:
  - "references/raw/web-animation-standards.md"
last_ingested: 2026-07-23
---

## 60fps Motion: Animate Only Compositor Properties

The render pipeline is style → layout → paint → composite. Animating a property
that triggers **layout** (`width`, `height`, `top`, `left`, `margin`, `padding`)
forces a reflow every frame and cannot hold 60fps (16.6ms/frame). Animate only
`transform`, `opacity`, and (sparingly) `filter` — these run on the compositor
thread.

**Incorrect (animates layout props — janky, main-thread bound):**

```css
.card {
  transition: left 300ms ease, width 300ms ease; /* both trigger layout */
}
.card:hover { left: 20px; width: 320px; }        /* reflow every frame */
```

**Correct (compositor-only — smooth, off main thread):**

```css
.card {
  transition: transform 300ms cubic-bezier(0.22, 1, 0.36, 1);
  will-change: transform;          /* promote just before animating */
}
.card:hover { transform: translateX(20px) scale(1.06); }
```

Key rules:
- Replace `left/top` with `transform: translate()`; replace `width/height` with
  `transform: scale()` (set `transform-origin`) or use **FLIP** for real layout
  changes: measure First, apply Last DOM, compute the Invert transform, then
  Play it back to zero — turns a layout animation into a transform animation.
- `will-change: transform` promotes a layer but costs memory — add it right
  before animating and remove it after; never blanket-apply it.
- In JS, prefer the Web Animations API (`element.animate`) over hand-rolled
  `requestAnimationFrame` tweens — WAAPI can run on the compositor. Never use
  `setInterval` for animation.
- Budget: 60fps = 16.6ms/frame, 120fps = 8.3ms. Profile with DevTools
  Performance; watch for purple "Layout" bars during animation.

## Sources

- `references/raw/web-animation-standards.md` — compositor property list, FLIP,
  will-change discipline, frame budget.
