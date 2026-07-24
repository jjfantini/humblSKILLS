---
title: "Canvas UI: Creative Canvas & WebGL Effects"
context: lib
category: canvas-ui
concept: usage
description: "Canvas UI (canvasui.dev): 27 copy-paste canvas/WebGL effects over real HTML. shadcn install, React/Vue/Svelte/vanilla, experimental-API caveat, license."
tags: canvas-ui, webgl, html-in-canvas, shadcn, particle, hero
sources:
  - "references/raw/canvas-ui.md"
last_ingested: 2026-07-24
---

## Canvas UI: Creative Canvas & WebGL Effects

Canvas UI is a set of ~27 creative canvas + WebGL effects (Particle Reveal,
Liquid, Glass, Shatter, VHS, Blaze, …) that render over real HTML on the GPU.
Most build on the experimental **html-in-canvas** API — the canvas paints your
live DOM as a texture a shader distorts. It is the heaviest tier in this skill;
spend it on ONE signature moment (a hero, a reveal), never scattered.

Install is copy-paste via the shadcn registry (you own the source file, no npm
dependency to update):

```
npx shadcn@latest add @canvas-ui/particle-reveal-react
```

Ships in React, Vue, Svelte, and vanilla TypeScript. Minimal React use:

```tsx
import { ParticleReveal } from "@/components/canvasui/ParticleReveal";

export function Hero() {
  return (
    <ParticleReveal radius={300}>
      <YourContent />
    </ParticleReveal>
  );
}
```

**Incorrect (relies on the experimental API as if it's universally available):**

```tsx
{/* html-in-canvas is an origin-trial Chrome feature; assuming it "just works"
    everywhere means most users get the fallback, not the effect you designed. */}
<Liquid>          {/* html-in-canvas component, no plan for other browsers */}
  <PricingTable />
</Liquid>
```

**Correct (pick a WebGL component for guaranteed support, gate reduced-motion):**

```tsx
import { useReducedMotion } from 'motion/react'; // or a matchMedia helper

function Hero({ children }) {
  const reduce = useReducedMotion();
  // ParticleReveal is WebGL-backed → works cross-browser today.
  // Skip the effect entirely when reduced motion is requested.
  return reduce ? <>{children}</> : <ParticleReveal radius={300}>{children}</ParticleReveal>;
}
```

Key facts to design around:
- **Browser support:** the flagship html-in-canvas mode is experimental
  (Chrome/Edge 140+ behind a flag / origin trial). Components detect support and
  **degrade gracefully** — without it, content renders as plain HTML. **WebGL
  components bypass the requirement and work everywhere today** — prefer them for
  production.
- **Accessibility:** the site says components respect reduced motion, but verify
  per component and be ready to gate manually (see
  `motion/principles/accessibility`) — same discipline as metal-fx.
- **License:** MIT + Commons Clause — free in any personal/commercial app, but
  no reselling/redistributing the components. (The four Antalik libs are pure MIT.)
- Ships an MCP server for AI-assistant integration.

Fit: one high-impact creative element. If you're choosing between two effects,
use one and keep the rest of the page quiet (`motion/principles/design`).

## Sources

- `references/raw/canvas-ui.md` — component list, shadcn install, framework
  matrix, experimental-API/WebGL-fallback behavior, license, MCP server.
