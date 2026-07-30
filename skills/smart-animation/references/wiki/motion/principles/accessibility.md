---
title: "Reduced Motion and the Accessibility Floor"
context: motion
category: principles
concept: accessibility
description: "prefers-reduced-motion is non-negotiable; keep essential feedback, cut parallax/spin/large movement. Includes shared motion tokens."
tags: accessibility, prefers-reduced-motion, vestibular, focus, motion-tokens
sources:
  - "references/raw/web-animation-standards.md"
  - "references/raw/transitions-dev.md"
last_ingested: 2026-07-23
---

## Reduced Motion and the Accessibility Floor

Every motion-producing file must honor `prefers-reduced-motion: reduce`. This is
not optional polish — large movement, parallax, and autoplay trigger vestibular
disorders, and missing reduced-motion handling fails accessibility audits. Keep
essential feedback (a short opacity fade, a state change) but cut motion that
moves the viewport or an element a long distance.

**Incorrect (no reduced-motion guard — moves regardless of user setting):**

```css
.reveal { animation: slide-up 500ms both; }
@keyframes slide-up { from { transform: translateY(40px); opacity: 0; } }
/* A user with vestibular sensitivity gets 40px of forced movement. */
```

**Correct (essential fade kept, large movement removed on request):**

```css
.reveal { animation: slide-up 500ms both; }
@keyframes slide-up { from { transform: translateY(40px); opacity: 0; } }

@media (prefers-reduced-motion: reduce) {
  .reveal { animation: fade-in 200ms both; }   /* fade only, no translate */
}
@keyframes fade-in { from { opacity: 0; } }
```

JS equivalent — gate imperative animation:

```js
const reduce = window.matchMedia('(prefers-reduced-motion: reduce)');
if (!reduce.matches) el.animate(keyframes, options);
```

Other floor requirements: never remove focus outlines when adding motion; never
trap the user in an animation; avoid flashing >3 times/second (seizure risk).

### Shared motion tokens (from transitions.dev `_root.css`)

Reuse these instead of hardcoding durations/easings — consistent tokens are what
make a set of animations feel like one system.

```css
:root {
  --duration-micro: 80ms;  --duration-quick: 150ms; --duration-fast: 250ms;
  --duration-medium: 350ms; --duration-slow: 400ms; --duration-stagger: 40ms;
  --ease-smooth-out: cubic-bezier(0.22, 1, 0.36, 1); /* modals, panels, reveals */
  --ease-linear: linear;                              /* spinners */
}
```

## Sources

- `references/raw/web-animation-standards.md` — media query, JS gating, seizure
  and focus rules.
- `references/raw/transitions-dev.md` — the mandatory reduced-motion block per
  snippet and the motion-token values.
