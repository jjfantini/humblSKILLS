---
title: "React Animation: CSS First, Motion When It Earns It"
context: lang
category: react
concept: animation
description: "React motion decision rule: CSS/native for state and reveals; reach for Motion (framer-motion) only for layout, presence, and gesture animations."
tags: react, motion, framer-motion, animatepresence, usereducedmotion, layout
sources:
  - "references/raw/web-animation-standards.md"
  - "references/raw/frontend-design-motion-principles.md"
last_ingested: 2026-07-23
---

## React Animation: CSS First, Motion When It Earns It

Do not import an animation library for what CSS does for free. The decision
rule: **use CSS/native** for hover, focus, state transitions, and
scroll-reveals; **reach for Motion** (`motion`, formerly Framer Motion) only for
the three things CSS can't do cleanly in React — enter/exit of unmounting
components, automatic layout (shared-element) animation, and gesture/drag
physics.

**Incorrect (heavy lib for a job CSS handles; no reduced-motion):**

```jsx
import { motion } from 'motion/react';
// A hover lift does not need JS, a component, or a bundle cost.
<motion.button whileHover={{ scale: 1.05 }}>Save</motion.button>
```

```jsx
// Better: plain CSS.
<button className="btn-lift">Save</button>
/* .btn-lift { transition: transform 150ms var(--ease-smooth-out); }
   .btn-lift:hover { transform: scale(1.05); }
   @media (prefers-reduced-motion: reduce) { .btn-lift { transition: none; } } */
```

**Correct (Motion for exit animation — the case CSS can't do on unmount):**

```jsx
import { AnimatePresence, motion, useReducedMotion } from 'motion/react';

function Toast({ open, children }) {
  const reduce = useReducedMotion();
  return (
    <AnimatePresence>
      {open && (
        <motion.div
          initial={reduce ? false : { opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          exit={reduce ? { opacity: 0 } : { opacity: 0, y: 12 }}
          transition={{ duration: 0.25, ease: [0.22, 1, 0.36, 1] }}
        >
          {children}
        </motion.div>
      )}
    </AnimatePresence>
  );
}
```

Guidance:
- Always branch on `useReducedMotion()` (Motion) or `matchMedia` — degrade to
  opacity-only or no motion.
- `layout` prop / `<AnimatePresence>` are the real reasons to use Motion:
  auto-animated reordering and unmount transitions. `layoutId` gives
  shared-element transitions between components.
- React 19 + the View Transitions API can crossfade route/state changes with no
  library — consider it before adding Motion for page transitions.
- Apply `motion/principles/design`: one orchestrated entrance (staggered
  children via a parent variant) beats per-component effects.

## Sources

- `references/raw/web-animation-standards.md` — Motion/GSAP landscape, View
  Transitions, the native-first rule.
- `references/raw/frontend-design-motion-principles.md` — orchestration and
  restraint applied to component motion.
