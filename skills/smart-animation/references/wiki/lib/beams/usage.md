---
title: "border-beam: Animated Border Glow"
context: lib
category: beams
concept: usage
description: "border-beam (React): a traveling/breathing glow around an element's border. Install, size/color presets, props, gating with active, vanilla port."
tags: border-beam, react, border-glow, css-property, cta, attention
sources:
  - "references/raw/border-beam.md"
last_ingested: 2026-07-23
---

## border-beam: Animated Border Glow

`border-beam` wraps an element and animates a traveling or breathing glow around
its border. React 18+; requires CSS `@property` (Chrome 85+, Safari 15.4+,
Firefox 128+).

> Naming note: the library is **border-beam** (site `beam.jakubantalik.com`). It
> is a border glow, not free-floating light beams. A third-party vanilla port,
> `border-beam-vanilla`, exists for non-React codebases.

```
npm install border-beam
```

**Incorrect (always-on glow on everything — pure noise):**

```jsx
{/* Every card glowing forever = nothing stands out, and it never rests. */}
{cards.map(c => <BorderBeam key={c.id}><Card {...c} /></BorderBeam>)}
```

**Correct (one element, gated to a real state via `active`):**

```tsx
import { BorderBeam } from 'border-beam';

// Glow the input only while the agent is processing it.
<BorderBeam size="line" colorVariant="ocean" active={isProcessing}>
  <input aria-label="Ask anything" />
</BorderBeam>
```

Presets:
- `size`: `'md'` (default, full border) · `'sm'` · `'line'` (bottom-only) ·
  `'pulse-inner'` (contained breathing) · `'pulse-outside'` (halo)
- `colorVariant`: `'colorful'` (default) · `'mono'` · `'ocean'` · `'sunset'`

Other props: `theme` (`dark`/`light`/`auto`), `strength` (default 1),
`duration`, `active` (default true — gate it!), `borderRadius` (auto),
`brightness`, `saturation`, `hueRange`, `staticColors`, `className`, `style`,
`onActivate`, `onDeactivate`.

Performance/a11y: rotate/line use GPU CSS animation; pulse types use a
~30fps-capped rAF that pauses when inactive/offscreen; layers are
`pointer-events: none`. Pulse types include `prefers-reduced-motion` support —
for rotate/line, gate `active` off yourself when reduced motion is preferred
(see `motion/principles/accessibility`).

Fit: draw attention to ONE thing — an active input, a primary CTA, an
"AI is working" card. Pair the glow with a real state change, never as decoration.

## Sources

- `references/raw/border-beam.md` — presets, full prop table, browser
  requirement, vanilla port, performance/a11y notes.
