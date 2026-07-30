---
title: "thinking-orbs: Agent Status Indicators"
context: lib
category: orbs
concept: usage
description: "thinking-orbs (React/2D-canvas): six stateful 'thinking' indicators for AI/agent UIs. Install, the six states, props, and when to use it."
tags: thinking-orbs, react, canvas, ai-ui, loading, status
sources:
  - "references/raw/thinking-orbs.md"
last_ingested: 2026-07-23
---

## thinking-orbs: Agent Status Indicators

`thinking-orbs` renders dotted "thinking"/loading orbs on a 2D canvas (no
WebGL), purpose-built for AI/agent UIs. Six hand-tuned states, each with a
distinct personality, chosen via the `state` prop. React, zero dependencies.

```
npm install thinking-orbs
```

**Incorrect (generic spinner for an agent that has real sub-states):**

```jsx
{/* A single spinner tells the user nothing about what the agent is doing. */}
{loading && <Spinner />}
```

**Correct (state maps to what the agent is actually doing):**

```tsx
import { ThinkingOrb } from 'thinking-orbs';

const STATE = {
  retrieval: 'searching',   // scanning docs
  reasoning: 'solving',     // working a problem
  writing:   'composing',   // generating output
} as const;

<ThinkingOrb state={STATE[phase]} size={64} aria-label={`Agent ${phase}`} />
```

The six states (pick the one that matches the phase, don't just default to one):
- `working` — particles on tilted orbits (general processing)
- `searching` — scan meridian sweeps a globe (retrieval/lookup)
- `solving` — bands scramble then resolve (reasoning/computation)
- `listening` — waveform through the rings (awaiting/receiving input)
- `composing` — undulating sash (generating output)
- `shaping` — outline morphs circle→triangle→square (transforming/formatting)

Props: `state` (required), `size` (`64` avatar / `20` inline — distinct designs,
not just scaled), `speed`, `paused`, `theme` (`auto`/`dark`/`light`),
`aria-label`, plus `className`/`style`/`data-*`.

Built-in performance/a11y (nothing to wire): static frame under
`prefers-reduced-motion`, auto-pause offscreen via `IntersectionObserver`, DPR
capped at 2. Always pass `aria-label`.

Fit: agent/chat thinking indicators and tool-call status. It reads as an "AI
product" signal — don't use it as a generic loader on marketing/e-commerce UI.

## Sources

- `references/raw/thinking-orbs.md` — states, props, install, performance notes.
