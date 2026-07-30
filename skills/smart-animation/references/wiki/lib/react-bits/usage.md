---
title: "React Bits: 140+ Copy-Paste Animated React Components"
context: lib
category: react-bits
concept: usage
description: "React Bits: 140+ animated React components via shadcn/jsrepo in four JS/TS x CSS/Tailwind variants. Per-component dependency cost (gsap/motion/three/ogl) and the missing reduced-motion gate."
tags: react-bits, shadcn, jsrepo, text-animation, animated-background, gsap, three, react-19
sources:
  - "references/raw/react-bits.md"
last_ingested: 2026-07-30
---

## React Bits: 140+ Copy-Paste Animated React Components

React Bits (reactbits.dev) is a catalog of **140+ finished animated components** —
text effects, animated backgrounds, UI elements, shapes — by David Haz. Official
Vue and Svelte ports exist. Requires **React 19**.

It is a catalog of *set-pieces*, not an engine. Reach for it when you want a
specific finished effect (split-text reveal, aurora background, magnet button);
reach for `lib/motion-react/usage.md` when you need a primitive to build with.

### Install: the variant is encoded in the name

Every component ships in four combinations — **JS+CSS, JS+Tailwind, TS+CSS,
TS+Tailwind** — chosen at install time via the suffix:

```
npx shadcn@latest add @react-bits/BlurText-TS-TW
```

Also available via `jsrepo` or plain copy-paste. You own the resulting file;
there is no runtime dependency on React Bits itself to keep updated.

### The dependency cost is per-component, and it is not small

The site says "minimal dependencies … lightweight and tree-shakeable." That is
true *per component* and misleading in aggregate — the project pulls `gsap`,
`motion`, `three`, `ogl`, `matter-js`, `@react-three/fiber`, and
`@react-three/drei`. A text effect might pull `gsap`; a 3D background pulls the
entire three.js stack.

**The imports at the top of the copied file are the dependency contract. Read
them before you commit to the component.**

**Incorrect (installing on vibes, then discovering the bundle):**

```
npx shadcn@latest add @react-bits/Aurora-TS-TW
# ...quietly requires three + @react-three/fiber + drei
```

**Correct (check the contract, then gate the motion you now own):**

```tsx
// 1. Read the imports first — this one is gsap, not three. Acceptable.
import gsap from "gsap"
import { useEffect, useRef } from "react"

// 2. You own this file now, so add the gate React Bits omits.
export function BlurText({ text }) {
  const ref = useRef(null)
  useEffect(() => {
    const reduce = window.matchMedia("(prefers-reduced-motion: reduce)").matches
    if (reduce) return                       // render final state, skip the tween
    gsap.from(ref.current.children, { opacity: 0, y: 12, stagger: 0.04 })
  }, [])
  return <span ref={ref}>{/* … */}</span>
}
```

### Accessibility is your job here

The components ship **no `prefers-reduced-motion` handling**. Scroll-driven text
reveals and particle/shader backgrounds are precisely the category that triggers
vestibular discomfort. Because copy-paste means you own the source, add the
guard *in* the file rather than wrapping it from outside. See
`motion/principles/accessibility.md` for what to substitute rather than simply
deleting.

### License

**MIT + Commons Clause** — free for personal and commercial use inside your own
product, but you may not sell the components themselves as a component library.
Same posture as Canvas UI (`lib/canvas-ui/usage.md`). The repo's `package.json`
is private with no `license` field; the grant is in the site and the repo
LICENSE file.

### Budget

Animated backgrounds and 3D effects are the heaviest tier here. Per
`motion/principles/design.md`, spend them on **one** signature moment — a hero —
and use `lang/css/animation.md` techniques for the rest of the page.

## Sources

- `references/raw/react-bits.md` — install paths and variant suffixes, category
  breakdown, the verbatim dependency table from the repo `package.json`, license
  posture, and the accessibility gap.
