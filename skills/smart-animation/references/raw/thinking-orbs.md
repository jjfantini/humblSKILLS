# thinking-orbs (the "orbs" library)

Source: https://orbs.jakubantalik.com/ and https://github.com/Jakubantalik/thinking-orbs
Author: Jakub Antalik. License: MIT. Researched 2026-07-23.

## What it is
Dotted "thinking" / loading orb indicators designed for AI & agent UIs. Six
hand-tuned animated states, monochrome, rendered on a **2D canvas** (no WebGL).
Zero dependencies.

## Install
```
npm install thinking-orbs
```

## Framework
React component (`<ThinkingOrb>`) with a vanilla 2D-canvas engine underneath.

## Minimal usage
```tsx
import { ThinkingOrb } from 'thinking-orbs';

function Status() {
  return <ThinkingOrb state="searching" size={64} />;
}
```

## The six states ("personalities") — selected via the `state` prop
- `working`   — particles on tilted orbits
- `searching` — scan meridian sweeps a dotted globe
- `solving`   — bands scramble, then click back solved
- `listening` — waveform rolls through the rings
- `composing` — undulating multi-band sash
- `shaping`   — dotted outline morphs circle → triangle → square

## Props
- `state` (required) — one of the six above
- `size` — `64` (chat-avatar scale) or `20` (inline scale). Different
  designs/dot-counts per size, not just a scale factor.
- `speed` — playback multiplier
- `paused` — boolean, freeze on current frame
- `theme` — `"auto"` (default) / `"dark"` / `"light"`. `auto` detects an
  ancestor `data-theme`, a `dark`/`light` class, or `prefers-color-scheme`.
- `aria-label` — accessible label
- pass-through canvas props: `className`, `style`, `data-*`

## Performance / accessibility
- Renders a **static frame** under `prefers-reduced-motion: reduce`.
- Auto-pauses offscreen via `IntersectionObserver`.
- 2D canvas only (no WebGL, no CSS filters).
- device-pixel-ratio capped at 2.

## Fit
Best for: agent/chat "thinking" indicators, tool-call spinners, streaming
status. Not a general-purpose loader for e-commerce/marketing — it reads as an
AI product signal.
