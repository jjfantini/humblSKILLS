# border-beam (the "beams" library)

Source: https://beam.jakubantalik.com/ and https://github.com/Jakubantalik/border-beam
Author: Jakub Antalik. License: MIT. Researched 2026-07-23.

NOTE: the URL sometimes cited as `beams.jakubantalik.com` does not resolve.
The real Antalik library is **border-beam** (singular `beam` subdomain). It is a
**traveling / breathing glow around an element's border**, NOT free-floating
light beams.

## What it is
A lightweight React component that wraps any element and adds a traveling or
breathing glow around its border (cards, buttons, inputs, search bars). Uses CSS
gradients + masking + `@property`-driven animation.

## Install
```
npm install border-beam
```

## Framework / requirements
React 18+. Requires CSS `@property` support (Chrome 85+, Safari 15.4+,
Firefox 128+). A separate third-party vanilla port exists:
`jqueryscript/border-beam-vanilla` (NOT by Antalik) for non-React codebases.

## Minimal usage
```tsx
import { BorderBeam } from 'border-beam';

function App() {
  return (
    <BorderBeam>
      <div style={{ padding: 32, borderRadius: 16, background: '#1d1d1d' }}>
        Your content here
      </div>
    </BorderBeam>
  );
}
```

## `size` presets
- `'md'`  (default) — full border glow
- `'sm'`  — compact
- `'line'` — bottom-only traveling glow
- `'pulse-inner'` — contained breathing glow
- `'pulse-outside'` — outward halo

## `colorVariant`
`'colorful'` (default, rainbow), `'mono'`, `'ocean'`, `'sunset'`.

## Full props
`children`; `size`; `colorVariant`; `theme` (`'dark'` default / `'light'` /
`'auto'`); `strength` (num, default 1); `duration` (num, per-type default
~1.96 / 3.1 / 2.3s); `active` (bool, default true); `borderRadius`
(auto-detected); `brightness` (per-type, ~1.3); `saturation` (1.2); `hueRange`
(30); `staticColors` (bool, false); `className`; `style`; `onActivate`;
`onDeactivate`.

## Performance / accessibility
- rotate/line types use GPU-accelerated CSS animation.
- pulse types use a **~30fps-capped requestAnimationFrame** loop that pauses
  when inactive/offscreen.
- Effect layers are decorative: `pointer-events: none`.
- Pulse types include built-in `prefers-reduced-motion: reduce` support.

## Fit
Best for: drawing attention to one CTA, an active input, an "AI is working"
card. Use `active` to gate it to a real state change; a permanently glowing
border is noise.
