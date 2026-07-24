---
title: "metal-fx: Liquid-Metal Shader Accent"
context: lib
category: metal
concept: usage
description: "metal-fx (React/WebGL): a liquid-metal shader for ONE premium accent. Install, variants/presets, the reduced-motion gap you must wire manually."
tags: metal-fx, react, webgl, shader, premium, cta
sources:
  - "references/raw/metal-fx.md"
last_ingested: 2026-07-23
---

## metal-fx: Liquid-Metal Shader Accent

`metal-fx` wraps a button/chip/icon in a real-time WebGL "liquid metal" shader.
It is the heaviest tool in this skill — use it for exactly ONE premium accent
(an upgrade CTA, a pro badge, a hero icon), never scattered across a view.

```
npm install metal-fx
```

**Incorrect (WebGL shader on many elements + no reduced-motion handling):**

```jsx
{/* Multiple shader instances + shader runs even for reduced-motion users. */}
{actions.map(a => <MetalFx key={a.id}><button>{a.label}</button></MetalFx>)}
```

**Correct (one accent, reduced-motion wired manually):**

```tsx
import { MetalFx } from 'metal-fx';
import { useState, useEffect } from 'react';

function useReducedMotion() {
  const [r, setR] = useState(false);
  useEffect(() => {
    const mq = matchMedia('(prefers-reduced-motion: reduce)');
    const on = () => setR(mq.matches); on(); mq.addEventListener('change', on);
    return () => mq.removeEventListener('change', on);
  }, []);
  return r;
}

const reduce = useReducedMotion();
<MetalFx variant="button" preset="chromatic" paused={reduce}>
  <button className="upgrade-pill">Upgrade to Pro</button>
</MetalFx>
```

Props: `variant` (`"button"` pill / `"circle"`), `preset` (`"chromatic"`
default / `"silver"` / `"gold"`), `theme` (`auto`/`dark`/`light`), `strength`
(0–1, the one intensity slider), `paused`, `borderRadius`, `reflectionTargets`
(refs for proximity reflection — **dark mode only**).

Performance: single shared WebGL context across all instances, shader compiled
once, one shared rAF loop, `IntersectionObserver` pauses offscreen, buffers
released on last unmount. SSR-safe (transparent placeholder until hydration).

**Accessibility gap — important:** metal-fx does NOT document
`prefers-reduced-motion` handling. You must gate it yourself via `paused` (shown
above). See `motion/principles/accessibility`.

Fit: one high-value moment. If you're tempted to use two, use one and make the
other quiet.

## Sources

- `references/raw/metal-fx.md` — variants, presets, shared-context performance
  model, SSR behavior, and the documented reduced-motion gap.
