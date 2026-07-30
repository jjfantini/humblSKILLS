# Motion for React (formerly Framer Motion)

Source: https://motion.dev/docs/react-quick-start,
https://motion.dev/docs/react-reduce-bundle-size,
https://motion.dev/docs/react-use-reduced-motion
Author: Matt Perry. License: MIT. Researched 2026-07-30.

## Naming: `motion`, not `framer-motion`

The library was renamed. `framer-motion` is the legacy package name and still
resolves, but the current package is **`motion`** and the React entry point is
**`motion/react`**. New code should install `motion` and import from
`motion/react`. Docs, changelog, and version numbers all track the `motion`
package now.

## Install
```
npm install motion
```

```tsx
import { motion } from "motion/react"
```

## What it is
A hybrid animation engine for React. Animations run **natively in the browser
via the Web Animations API and ScrollTimeline** where possible (off the main
thread, up to 120fps), and fall back to a JavaScript engine where the browser
cannot express the effect — spring physics, interruptible keyframes, gesture
tracking, independent transforms.

That hybrid design is the reason it isn't simply "CSS with extra steps": the
things it does that CSS cannot are interruption, spring physics, layout
animation, and exit animation.

## Core API surface
- `<motion.div>` — any element, with animation props
- `initial` / `animate` — state-driven animation, declarative
- `exit` + `<AnimatePresence>` — animate elements as they unmount (CSS cannot)
- `layout` prop — automatic FLIP layout animation when position/size changes
- `whileHover`, `whileTap`, `whileInView` — gesture and viewport props
- `variants` — named states that cascade to children, enabling stagger
- SVG support including `pathLength`
- `useAnimate` — imperative animation hook

## Minimal usage
```tsx
import { motion } from "motion/react"

export function Card() {
  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3, ease: [0.22, 1, 0.36, 1] }}
    />
  )
}
```

## Staggered children via variants
```tsx
const list = {
  hidden: {},
  show: { transition: { staggerChildren: 0.06 } },
}
const item = {
  hidden: { opacity: 0, y: 12 },
  show:   { opacity: 1, y: 0 },
}

<motion.ul variants={list} initial="hidden" animate="show">
  {items.map(i => <motion.li key={i.id} variants={item} />)}
</motion.ul>
```

## Bundle size (Rollup numbers; Webpack is slightly larger)
| Import | Size |
| --- | --- |
| `motion` component (full) | **34kb** — cannot be tree-shaken smaller; the props-driven declarative API means the bundler can't know which features you use |
| `m` component + LazyMotion | **4.6kb** initial render |
| `domAnimation` feature package | **+15kb** — animation, variants, exit, gestures |
| `domMax` feature package | **+25kb** — everything in domAnimation plus pan/drag and layout animations |
| `useAnimate` mini | **2.3kb** — WAAPI only, hardware accelerated |
| `useAnimate` hybrid | **17kb** — sequences, motion values, independent transforms |

### LazyMotion pattern
```tsx
import * as m from "motion/react-m"
import { LazyMotion, domAnimation } from "motion/react"

<LazyMotion features={domAnimation}>
  <m.div animate={{ opacity: 1 }} />
</LazyMotion>
```
Features can be bundled or lazy-loaded (`features={() => import("./features")}`).
Note the swap: `m.div` instead of `motion.div`.

## Reduced motion
```tsx
import { useReducedMotion } from "motion/react"

const shouldReduceMotion = useReducedMotion()
const closedX = shouldReduceMotion ? 0 : "-100%"
```

The hook **actively responds to changes** and re-renders when the OS setting
flips. Motion's own documented guidance:
1. Replace potentially motion-sickness-inducing `x`/`y` animations with `opacity`
2. Disable autoplay on background video
3. Remove parallax
4. Adapt dynamically rather than deciding once at mount

Note the emphasis: **replace**, not delete. An opacity crossfade still
communicates the state change; `transition: none` often just makes the UI feel
broken.

## When CSS is the right answer instead
Motion's own docs: "for simple, self-contained effects (like a color change on
hover) a standard CSS transition is a lightweight solution." A 34kb dependency
for hover states is not a trade worth making. The threshold is whether you need
one of: exit animation, layout animation, spring physics, interruption, or
orchestrated stagger. Those are the four things CSS genuinely cannot do well.
