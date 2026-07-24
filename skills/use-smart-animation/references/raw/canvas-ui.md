# Canvas UI (canvasui.dev)

Source: https://canvasui.dev/ , https://canvasui.dev/docs ,
https://github.com/DavidHDev/canvas-ui
Author: David Haz (David HD / DavidHDev — also the author of react-bits).
License: MIT + Commons Clause (free for personal & commercial apps; reselling /
redistributing the components is prohibited). Researched 2026-07-24.

## What it is
A library of creative canvas + WebGL effects that run over real HTML. Most
components build on the experimental **html-in-canvas** API: a `<canvas>` lays
out and paints live DOM content so your component becomes a texture a shader can
sample and distort. GPU-accelerated, animates outside React's render cycle.

## Components (~27)
Asciify, Bend, Blaze, Bubble, Cloth, Clouds, Dithered Object, Droplets, Glass,
Glass Object, Glitch, Grid, Hex Float, Laser, Liquid, Magnify, Particle Object,
Particle Reveal, Particle Scroll, Peel, Retro Dither, Ripple, Shatter, VHS
(more planned).

## Install — shadcn registry, copy-paste (NOT an npm dependency)
Each component is a single standalone file dropped into your project via the
shadcn CLI; you own and version the source, there is no package to update.
Example (framework suffix on the registry name):
```
npx shadcn@latest add @canvas-ui/particle-reveal-react
```

## Framework support
Framework-agnostic engine shipped in four flavors — React, Vue, Svelte, and
vanilla TypeScript — each a single file built on the same engine.

## Minimal usage (React)
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

## Browser support / graceful degradation — IMPORTANT
- The **html-in-canvas** API (`drawElement` / `chrome://flags/#canvas-draw-element`)
  is Chromium-only, experimental (origin trial); broadly it needs Chrome/Edge
  140+ with `#enable-experimental-web-platform-features`.
- **Safari/WebKit and Firefox do NOT support html-in-canvas** (no announced
  implementation as of 2026-07). On Safari the flagship effect will not render at
  full fidelity — observed in practice as noticeably reduced clarity vs a
  Chromium browser.
- **SVG `<foreignObject>` is the cross-browser workaround** for rendering live
  HTML into a canvas without the html-in-canvas API: serialize the HTML → wrap in
  `<svg><foreignObject>` → Blob/data-URL → `<img>` → `ctx.drawImage()`. It works
  in Safari but imperfectly: external stylesheets/fonts may not load, some CSS
  properties and pseudo-elements render differently or not at all, cross-origin
  images taint the canvas (CORS), and Safari has a known quirk where the first
  render can silently fail (render twice, or preload assets as data URIs).
- Components **detect support at runtime and degrade gracefully**: without the
  API, content renders as normal HTML and the parts of the effect that can still
  run, still do.
- **WebGL-based components bypass the html-in-canvas requirement and work
  everywhere today** (including Safari). Prefer these for production/Safari where
  the flagship html-in-canvas experience isn't guaranteed.

## Accessibility
The site states components respect reduced-motion preferences. Because these are
WebGL/canvas effects (heaviest tier), still verify per component and be ready to
gate manually behind `prefers-reduced-motion` — same discipline as metal-fx.

## Extras
Ships an MCP server for AI-assistant integration.

## Fit
Best for: ONE high-impact creative moment — a hero, a signature reveal, an
atmospheric background — not scattered across a view. Heaviest tool in the skill
(WebGL + experimental API). For guaranteed cross-browser support, choose a WebGL
component over an html-in-canvas one.
