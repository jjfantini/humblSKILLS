---
title: "JavaScript Animation: Web Animations API First"
context: lang
category: js
concept: animation
description: "Vanilla JS motion: use element.animate (WAAPI) over rAF tweens, IntersectionObserver reveals, and startViewTransition. No jQuery."
tags: javascript, waapi, requestanimationframe, intersectionobserver, view-transitions
sources:
  - "references/raw/web-animation-standards.md"
last_ingested: 2026-07-23
---

## JavaScript Animation: Web Animations API First

When motion needs JS (dynamic values, orchestration, gesture response), use the
**Web Animations API** — `element.animate()` — not hand-stepped
`requestAnimationFrame` tweens and never jQuery `.animate()`. WAAPI can run on
the compositor, returns a controllable `Animation`, and its `.finished` is a
Promise. Reserve `requestAnimationFrame` for canvas/physics/measurement loops.

**Incorrect (jQuery / manual rAF tween of a layout prop):**

```js
$('.box').animate({ left: '200px' }, 300);          // jQuery, animates layout
// or a manual rAF loop nudging style.left every frame — main-thread bound.
```

**Correct (WAAPI, compositor-only, reduced-motion aware, awaitable):**

```js
const reduce = matchMedia('(prefers-reduced-motion: reduce)').matches;
const box = document.querySelector('.box');

const anim = box.animate(
  [{ transform: 'translateX(0)' }, { transform: 'translateX(200px)' }],
  { duration: reduce ? 0 : 300, easing: 'cubic-bezier(0.22,1,0.36,1)', fill: 'both' }
);
await anim.finished;        // sequence follow-up work off the Promise
```

Patterns:
- **Reveal-on-scroll:** `IntersectionObserver` adds a class that triggers a CSS
  transition — the JS just toggles state. (In CSS-capable targets prefer
  `animation-timeline: view()`; see `lang/css/animation`.)
- **DOM/route changes:** wrap the mutation in
  `document.startViewTransition(() => updateDOM())` for an automatic crossfade.
- **Control:** `.play() / .pause() / .reverse() / .cancel()`, `.playbackRate`,
  `.currentTime` — everything you'd hand-roll, built in.
- **Cleanup:** cancel animations and disconnect observers on teardown to avoid
  leaks in SPAs.

If you need complex sequenced timelines or scroll-scrubbing beyond native, reach
for GSAP or Motion One (`animate()`), not a custom engine — see the library
landscape in the source.

## Sources

- `references/raw/web-animation-standards.md` — WAAPI surface, rAF guidance,
  IntersectionObserver, View Transitions, library landscape.
