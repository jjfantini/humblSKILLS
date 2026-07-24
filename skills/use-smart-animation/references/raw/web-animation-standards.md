# Web animation standards & performance reference

Reference notes for native web animation across HTML/CSS/JS/TS/React.
Compiled 2026-07-23 from established platform documentation (MDN, web.dev,
CSSWG specs). Ground truth for the native-first best practices.

## The performance rule (compositor-only properties)
The browser renders in stages: style → layout → paint → composite. Animating a
property that triggers **layout** (`width`, `height`, `top`, `left`, `margin`)
forces reflow every frame and drops below 60fps. Animate ONLY:
- `transform` (translate / scale / rotate / skew)
- `opacity`
- `filter` (GPU-composited, but expensive; use sparingly)
These are handled on the compositor thread and can run off the main thread.

Corollaries:
- Replace `left/top` motion with `transform: translate()`.
- Replace `width/height` animation with `transform: scale()` (+ correct
  transform-origin), or the FLIP technique for layout changes.
- `will-change: transform` promotes a layer, but overuse costs memory — add it
  just before animating, remove it after. Do not put it on everything.
- 60fps = 16.6ms per frame budget. 120fps displays = 8.3ms.

## FLIP (animate layout changes cheaply)
First-Last-Invert-Play: measure First position, apply the Last DOM state,
compute the Invert transform (delta from first to last), then Play by
transitioning the transform back to zero. Turns an expensive layout animation
into a compositor-only transform animation.

## CSS primitives
- `transition: <prop> <duration> <easing> <delay>` — for state changes.
- `@keyframes name { from/to | 0%..100% }` + `animation:` — for looping /
  multi-step.
- `cubic-bezier(x1,y1,x2,y2)` — custom easing. `linear()` — multi-point easing
  (springs, bounces) in modern browsers.
- `@property --x { syntax; inherits; initial-value; }` — register a custom
  property so gradients/colors/numbers can be animated (typed custom props).
- `@starting-style { }` — define the pre-open state so elements can animate on
  first render / when moving from `display:none` (pairs with
  `transition-behavior: allow-discrete`).

## View Transitions API
- Same-document: `document.startViewTransition(() => updateDOM())` crossfades old
  and new DOM states; customize via `::view-transition-*` pseudo-elements and
  `view-transition-name`.
- Cross-document (MPA): opt in with `@view-transition { navigation: auto; }` —
  animates full page navigations with no JS.

## Scroll-driven animations (CSS, no JS)
- `animation-timeline: scroll()` — drive a keyframe animation by scroll position.
- `animation-timeline: view()` — drive by an element's visibility in the
  viewport (reveal-on-scroll with zero JS and no IntersectionObserver).
- `animation-range` — set start/end of the timeline mapping.

## Web Animations API (JS)
- `element.animate(keyframes, options)` — imperative equivalent of CSS
  keyframes; returns an `Animation` with `.play()/.pause()/.reverse()/.finished`
  (a Promise) / `.cancel()`.
- Prefer WAAPI over manually stepping `requestAnimationFrame` for property
  tweens — it runs on the compositor where possible and stays off the main
  thread.
- Use `requestAnimationFrame` only for physics/canvas/measurement loops, never
  `setInterval` for animation.

## IntersectionObserver (reveal-on-scroll, JS fallback)
Observe elements; add a class when they enter the viewport to trigger a CSS
transition. The modern zero-JS alternative is `animation-timeline: view()`.

## Accessibility
- `@media (prefers-reduced-motion: reduce) { ... }` — remove or drastically
  reduce non-essential motion. Keep essential feedback (e.g. a short opacity
  fade) but cut parallax, large movement, autoplay, spin.
- JS: `window.matchMedia('(prefers-reduced-motion: reduce)')` — gate imperative
  animations; listen for changes.
- Never remove focus outlines when adding motion. Never trap users in an
  animation. Avoid flashing (>3/s = seizure risk).

## TypeScript notes
- WAAPI types: `Keyframe[] | PropertyIndexedKeyframes`,
  `KeyframeAnimationOptions`, return type `Animation`.
- `matchMedia` returns `MediaQueryList`; guard `prefers-reduced-motion` behind a
  typed helper.

## Library landscape (native-first, reach for these only when native falls short)
- **Motion** (`motion`, formerly Framer Motion) — React declarative animation,
  layout animations, gestures, `useReducedMotion()` hook. Also ships a vanilla
  `animate()` (Motion One) for non-React.
- **GSAP** — imperative timeline library, framework-agnostic, strong for complex
  sequenced/scroll (ScrollTrigger) work. Ships TS types.
- **anime.js**, **auto-animate** — lighter options for specific needs.
Rule: CSS/WAAPI first; add a library only when the animation is genuinely beyond
native (complex orchestrated timelines, layout/shared-element transitions in
React, gesture physics).
