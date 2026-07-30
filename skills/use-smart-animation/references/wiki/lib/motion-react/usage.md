---
title: "Motion for React: Springs, Layout, and Exit Animations"
context: lib
category: motion-react
concept: usage
description: "Motion (formerly Framer Motion): the four things CSS can't do (exit, layout, spring, interruption), the 34kb vs 4.6kb LazyMotion trade, and useReducedMotion."
tags: motion, framer-motion, react, spring, layout-animation, exit-animation, lazymotion, bundle-size
sources:
  - "references/raw/motion-react.md"
last_ingested: 2026-07-30
---

## Motion for React: Springs, Layout, and Exit Animations

Motion is the React animation engine formerly published as `framer-motion`.
**The package is now `motion` and the import is `motion/react`** — `framer-motion`
still resolves but is the legacy name; write new code against `motion`.

```
npm install motion
```

It runs a **hybrid engine**: animations go through the Web Animations API and
ScrollTimeline natively where the browser can express them (off main thread, up
to 120fps), and fall back to JS for spring physics, interruptible keyframes, and
gesture tracking.

### Reach for it only for the four things CSS cannot do

This skill is native-first, and Motion's own docs agree — "for simple,
self-contained effects (like a color change on hover) a standard CSS transition
is a lightweight solution." A 34kb dependency for hover states is a bad trade.
The threshold is whether you need one of:

1. **Exit animation** — animating an element as it unmounts (`AnimatePresence`)
2. **Layout animation** — FLIP on position/size change (the `layout` prop)
3. **Spring physics** — velocity-aware motion, not a fixed duration
4. **Interruption** — a new gesture retargeting a running animation mid-flight

Need none of those? Read `lang/css/animation.md` or `lang/react/animation.md`
instead and close this file.

**Incorrect (Motion for what CSS already does — 34kb for a hover):**

```tsx
import { motion } from "motion/react"

<motion.button whileHover={{ scale: 1.03 }} />   // a CSS :hover transform
```

**Correct (Motion earning its weight — exit + stagger, which CSS can't do):**

```tsx
import { motion, AnimatePresence, useReducedMotion } from "motion/react"

const list = { hidden: {}, show: { transition: { staggerChildren: 0.06 } } }

export function Results({ items }) {
  const reduce = useReducedMotion()
  const item = {
    hidden: { opacity: 0, y: reduce ? 0 : 12 },   // replace travel, keep the fade
    show:   { opacity: 1, y: 0 },
    exit:   { opacity: 0 },
  }
  return (
    <motion.ul variants={list} initial="hidden" animate="show">
      <AnimatePresence>
        {items.map(i => (
          <motion.li key={i.id} variants={item} exit="exit" layout />
        ))}
      </AnimatePresence>
    </motion.ul>
  )
}
```

### Bundle size — know what you're importing

| Import | Size |
| --- | --- |
| `motion` component (full) | **34kb** — un-tree-shakeable; the props-driven API hides which features you use from the bundler |
| `m` + `LazyMotion` | **4.6kb** initial |
| `+ domAnimation` | +15kb — animations, variants, exit, gestures |
| `+ domMax` | +25kb — the above plus drag/pan and layout animations |
| `useAnimate` mini | 2.3kb — WAAPI only |
| `useAnimate` hybrid | 17kb — sequences, motion values, independent transforms |

If Motion is in the bundle for one or two components, use LazyMotion — note the
component swap from `motion.div` to `m.div`:

```tsx
import * as m from "motion/react-m"
import { LazyMotion, domAnimation } from "motion/react"

<LazyMotion features={domAnimation}>
  <m.div animate={{ opacity: 1 }} />
</LazyMotion>
```

`features` also accepts a loader (`() => import("./features")`) to keep it out
of the main bundle entirely.

### Reduced motion: replace, don't delete

`useReducedMotion()` re-renders when the OS setting changes, so gate values
rather than branching once at mount. Motion's documented guidance: swap
motion-sickness-inducing `x`/`y` for `opacity`, disable video autoplay, drop
parallax. Deleting the transition outright usually reads as broken — the
crossfade still communicates the state change. See
`motion/principles/accessibility.md`.

### Note on stacking

React Bits components frequently bundle `motion` themselves (see
`lib/react-bits/usage.md`). If both are in play, you are paying for one copy —
check before adding a second animation library on top.

## Sources

- `references/raw/motion-react.md` — package rename, hybrid engine, full API
  surface, verbatim bundle-size table, LazyMotion pattern, and the
  `useReducedMotion` guidance.
