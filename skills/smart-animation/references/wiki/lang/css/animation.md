---
title: "CSS Animation: Native-First Motion"
context: lang
category: css
concept: animation
description: "CSS-only motion best practices: transitions, keyframes, @property, @starting-style, View Transitions, and scroll-driven animations with zero JS."
tags: css, keyframes, view-transitions, scroll-driven, starting-style, property
sources:
  - "references/raw/web-animation-standards.md"
last_ingested: 2026-07-23
---

## CSS Animation: Native-First Motion

For a CSS-only or static site, native CSS now covers most motion without a line
of JavaScript. Reach for a library only when native genuinely falls short. Read
`motion/principles/performance` and `motion/principles/accessibility` first —
they apply to every snippet here.

**Incorrect (JS for what CSS does natively; layout prop; no reduced-motion):**

```js
// Hand-rolled reveal-on-scroll for a job CSS ships natively.
window.addEventListener('scroll', () => {
  document.querySelectorAll('.reveal').forEach(el => {
    if (el.getBoundingClientRect().top < innerHeight) el.style.marginTop = '0';
  });
});
```

**Correct (scroll-driven animation, compositor-only, reduced-motion aware):**

```css
.reveal {
  animation: reveal linear both;
  animation-timeline: view();          /* driven by viewport visibility, no JS */
  animation-range: entry 0% cover 30%;
}
@keyframes reveal {
  from { opacity: 0; transform: translateY(24px); }
  to   { opacity: 1; transform: translateY(0); }
}
@media (prefers-reduced-motion: reduce) {
  .reveal { animation: none; opacity: 1; }
}
```

Native primitives to prefer:
- **`transition`** for state changes; **`@keyframes` + `animation`** for
  looping/multi-step. Use `cubic-bezier()` or `linear()` (springs/bounces) for
  character — the browser default `ease` reads generic.
- **`@property`** to register a typed custom property so gradients, colors, and
  numbers become animatable (the mechanism border-beam relies on).
- **`@starting-style`** (+ `transition-behavior: allow-discrete`) to animate an
  element in from `display:none` / first render.
- **View Transitions:** same-document via JS `startViewTransition`, or full MPA
  navigations with `@view-transition { navigation: auto; }` — a whole-page
  crossfade with zero JS. Give shared elements a `view-transition-name`.
- **Scroll-driven:** `animation-timeline: scroll()` (scroll position) or
  `view()` (element visibility) replace IntersectionObserver for reveals.

Never animate `width/height/top/left`; use `transform`/`opacity`
(see `motion/principles/performance`).

## Sources

- `references/raw/web-animation-standards.md` — CSS primitives, @property,
  @starting-style, View Transitions, scroll-driven timelines.
