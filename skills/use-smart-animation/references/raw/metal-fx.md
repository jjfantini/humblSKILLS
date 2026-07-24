# metal-fx (the "metal" library)

Source: https://metal.jakubantalik.com/ and https://github.com/Jakubantalik/metal-fx
Author: Jakub Antalik. License: MIT. Researched 2026-07-23.

## What it is
Animated **WebGL "liquid metal" shader** effect for React. Wrap a button, chip,
or icon → real-time metal ring, with optional proximity reflection on
neighboring elements.

## Install
```
npm install metal-fx
```

## Framework
React component with a WebGL/shader backend.

## Minimal usage
```tsx
import { MetalFx } from 'metal-fx';

function App() {
  return (
    <MetalFx variant="button">
      <button className="upgrade-pill">Upgrade to Pro</button>
    </MetalFx>
  );
}
```

## Props
- `variant` — `"button"` (pill silhouette, 1px ring, scale 1.6) or `"circle"`
  (compact, 2px ring, scale 1.3)
- `preset` — `"chromatic"` (iridescent rainbow, default), `"silver"` (cool
  steel), `"gold"` (warm gold)
- `theme` — `"auto"` (default), `"dark"`, `"light"`
- `strength` — 0–1 intensity (default 1). The single advertised "one slider".
- `paused` — freeze shader on current frame
- `borderRadius` — numeric override of child's computed radius
- `reflectionTargets` — array of refs to neighboring elements that receive
  proximity reflection (**dark mode only**)

## Performance
- Single **shared WebGL context** across all instances; shader compiled once;
  one shared requestAnimationFrame loop.
- `IntersectionObserver` pauses offscreen copies; GL render skips when all
  offscreen.
- `ResizeObserver` debounced via rAF.
- context/program/buffers released on last unmount.
- SSR: renders a transparent placeholder during SSR, mounts the WebGL pipeline
  only after client hydration.

## Accessibility — IMPORTANT GAP
Reduced-motion is **not documented** in the README. Assume metal-fx does NOT
auto-honor `prefers-reduced-motion`. Wire it manually: read the media query and
pass `paused` when the user prefers reduced motion.

## Fit
Best for: ONE premium accent — an upgrade CTA, a pro badge, a hero icon. WebGL
is the heaviest tool in this skill; never scatter multiple metal instances
across a view.
