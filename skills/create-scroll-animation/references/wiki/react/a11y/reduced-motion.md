---
title: "Honor prefers-reduced-motion with a Static Fallback"
context: react
category: a11y
concept: reduced-motion
description: "When the user has prefers-reduced-motion: reduce set, skip the scroll-scrub entirely. Render the first frame as a static img, don't preload the rest, don't bind scroll. Compliance with WCAG 2.3.3 (Animation from Interactions) and respect for vestibular disorders."
tags: a11y, accessibility, reduced-motion, WCAG, vestibular
sources: []
last_ingested: 2026-04-20
---

## Reduced-Motion Fallback

Scroll-driven video is the *definition* of "animation from interactions"
under WCAG 2.3.3. Users with vestibular disorders may get motion sick
from this effect. The fix is straightforward: if
`prefers-reduced-motion: reduce` is set, don't animate.

**Strategy:**

1. No scroll binding (don't call `useScroll`).
2. No preload (don't fetch frames 2..N).
3. Render the first frame as a static `<img>` — still a hero, still
   visually meaningful, zero motion.

**Inline in the component:**

```tsx
'use client';

import { useEffect, useState } from 'react';

export function ScrollFrameCanvas(props: Props) {
  const [prefersReducedMotion, setPrefersReducedMotion] = useState(false);

  useEffect(() => {
    const mq = window.matchMedia('(prefers-reduced-motion: reduce)');
    setPrefersReducedMotion(mq.matches);
    const onChange = (e: MediaQueryListEvent) => setPrefersReducedMotion(e.matches);
    mq.addEventListener('change', onChange);
    return () => mq.removeEventListener('change', onChange);
  }, []);

  if (prefersReducedMotion) {
    const firstFrameUrl = props.frameUrlPattern.replace(
      '{index}',
      String(1).padStart(props.frameDigits ?? 4, '0'),
    );
    return (
      <section style={{ height: '100vh', position: 'relative' }}>
        <img
          src={firstFrameUrl}
          alt=""
          aria-hidden="true"
          style={{
            width: '100%',
            height: '100%',
            objectFit: 'cover',
            display: 'block',
          }}
        />
      </section>
    );
  }

  // ... normal canvas path
}
```

**Why SSR-safe via `useState(false)` + `useEffect`:**

`window.matchMedia` is unavailable during SSR. Initializing state to
`false` means the server renders the animated path by default; the
client hydrates, `useEffect` runs, and if reduced-motion is set the
component swaps to the static `<img>` on the next render. One extra
render on mount for reduced-motion users only.

**Alternative — always SSR the static path:**

Can work if you detect reduced-motion via a cookie set earlier by
client-side JS, but this is complexity you don't need. The one-render
approach above is industry standard.

## Edge cases

- **User toggles the OS setting mid-session.** The `mq.addEventListener('change', ...)`
  handler catches it and the component re-renders.
- **SSR hydration mismatch warnings.** Not an issue with the pattern
  above because the server always renders the animated path; the swap
  happens after hydration, which React handles cleanly.
- **Don't detect at build time.** The value is per-user and per-session,
  not global.

## Why not show an even-simpler CSS-only fallback

We *could* detect `prefers-reduced-motion` in a CSS `@media` block and
display a different image. But:

1. We still need to avoid preloading 100 WebPs on the network.
2. We still need to avoid binding `useScroll` / `useMotionValueEvent`.

CSS can't express "don't run this JS." The JS branch has to exist, so
doing it all in JS (and making the component truly inert for
reduced-motion users) is cleaner than splitting the concern.

## WCAG reference

- **WCAG 2.3.3** (Animation from Interactions, Level AAA): "Motion
  animation triggered by interaction can be disabled, unless the
  animation is essential to the functionality or the information being
  conveyed."
- Scroll-scrubbed video is animation triggered by interaction (scroll).
- It is *not* essential to the information — the hero frame itself
  conveys the same information statically.
- Therefore the accommodation is required.

## Sources

Synthesis concept. The reduced-motion pattern is standard; the
component-level implementation derives from the `useScroll` /
`useMotionValueEvent` contract in
`react/component/scroll-frame-canvas.md`.
