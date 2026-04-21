---
title: "Build the Scroll-Frame Canvas React Component"
context: react
category: component
concept: scroll-frame-canvas
description: "Core anatomy of the generated .tsx: refs, useScroll with target+offset, useTransform mapping progress to frame index, useMotionValueEvent for non-rerendering draw, DPR scaling, and cover-fit math. All inline in a single file."
tags: react, canvas, framer-motion, useScroll, useMotionValueEvent, DPR, hero
sources:
  - "references/raw/apple-scroll-technique-css-tricks.md"
  - "references/raw/framer-motion-usescroll-docs.md"
last_ingested: 2026-04-20
---

## The Core Component

This is the heart of the skill — the single-file React component that
takes a frame URL array and renders a scroll-scrubbed canvas. Every other
`react/*` concept extends or plugs into this one.

**Non-negotiables:**

1. **`'use client'` at the top** — App Router server components can't hold
   refs, `useScroll` needs `window`, and draw is a DOM side-effect.
2. **Refs, not state, for the canvas.** State triggers React re-renders
   on every change; we need zero re-renders during scroll.
3. **`useMotionValueEvent`, not `useEffect(() => motionValue.on(...))`.**
   The former is the canonical v11+ form and handles cleanup.
4. **DPR scaling.** Canvas internal dimensions = `cssDimensions * devicePixelRatio`;
   CSS dimensions stay unscaled. Without this the canvas renders at
   half resolution on Retina.
5. **Draw only when the rounded frame index changes.** Deduplicate.
6. **Section height drives scroll distance.** 350vh ≈ slow, 250vh ≈ fast.

**Minimal anatomy:**

```tsx
'use client';

import { useEffect, useRef, useState } from 'react';
import { useScroll, useTransform, useMotionValueEvent } from 'framer-motion';

type Props = {
  frameCount: number;
  frameUrlPattern: string;     // e.g. "/frames/frame_{index}.webp"
  frameDigits?: number;        // zero-pad, default 4
  scrollHeightVh?: number;     // default 350
  fit?: 'cover' | 'zoomed-contain'; // default 'cover'
};

export function ScrollFrameCanvas({
  frameCount,
  frameUrlPattern,
  frameDigits = 4,
  scrollHeightVh = 350,
  fit = 'cover',
}: Props) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const canvasRef    = useRef<HTMLCanvasElement | null>(null);
  const framesRef    = useRef<Map<number, HTMLImageElement>>(new Map());
  const lastDrawn    = useRef<number>(-1);

  const [ready, setReady] = useState(false);

  const { scrollYProgress } = useScroll({
    target: containerRef,
    offset: ['start start', 'end end'],
  });
  const frameIndex = useTransform(scrollYProgress, [0, 1], [0, frameCount - 1]);

  // DPR-aware canvas sizing. Re-run on mount and resize.
  useEffect(() => {
    const sizeCanvas = () => {
      const canvas = canvasRef.current;
      if (!canvas) return;
      const dpr = window.devicePixelRatio || 1;
      const w = window.innerWidth;
      const h = window.innerHeight;
      canvas.width  = w * dpr;
      canvas.height = h * dpr;
      canvas.style.width  = `${w}px`;
      canvas.style.height = `${h}px`;
      canvas.getContext('2d')?.setTransform(dpr, 0, 0, dpr, 0, 0);
      if (lastDrawn.current >= 0) drawFrame(lastDrawn.current);
    };
    sizeCanvas();
    window.addEventListener('resize', sizeCanvas);
    return () => window.removeEventListener('resize', sizeCanvas);
  }, []);

  // Draw frame on scroll — no React re-render.
  useMotionValueEvent(frameIndex, 'change', (latest) => {
    const idx = Math.round(latest);
    if (idx === lastDrawn.current) return;
    lastDrawn.current = idx;
    drawFrame(idx);
  });

  function drawFrame(idx: number) {
    const canvas = canvasRef.current;
    const ctx = canvas?.getContext('2d');
    const img = framesRef.current.get(idx);
    if (!canvas || !ctx || !img) return;

    const cw = canvas.width  / (window.devicePixelRatio || 1);
    const ch = canvas.height / (window.devicePixelRatio || 1);
    const iw = img.naturalWidth;
    const ih = img.naturalHeight;

    ctx.clearRect(0, 0, cw, ch);

    if (fit === 'cover') {
      const scale = Math.max(cw / iw, ch / ih);
      const dw = iw * scale;
      const dh = ih * scale;
      ctx.drawImage(img, (cw - dw) / 2, (ch - dh) / 2, dw, dh);
    } else {
      // zoomed-contain: 1.2× contain-fit, centered. Used on mobile.
      const scale = Math.min(cw / iw, ch / ih) * 1.2;
      const dw = iw * scale;
      const dh = ih * scale;
      ctx.drawImage(img, (cw - dw) / 2, (ch - dh) / 2, dw, dh);
    }
  }

  // (Preloading logic lives in preload-strategy.md. Once frame 0 is
  // decoded we call setReady(true) and drawFrame(0).)

  return (
    <section
      ref={containerRef}
      style={{ height: `${scrollHeightVh}vh`, position: 'relative' }}
    >
      <div style={{ position: 'sticky', top: 0, height: '100vh' }}>
        <canvas ref={canvasRef} style={{ display: 'block' }} />
      </div>
    </section>
  );
}
```

**Incorrect — state-driven draw (defeats the pattern):**

```tsx
// BAD: every scroll tick re-renders the whole component
const [frame, setFrame] = useState(0);
useMotionValueEvent(frameIndex, 'change', (v) => setFrame(Math.round(v)));
// somewhere in the render: drawFrame(frame)
```

**Correct — side-effect draw, no state:**

```tsx
const lastDrawn = useRef(-1);
useMotionValueEvent(frameIndex, 'change', (v) => {
  const idx = Math.round(v);
  if (idx === lastDrawn.current) return;
  lastDrawn.current = idx;
  drawFrame(idx);
});
```

## Why `'start start'` to `'end end'` offset

This maps "top of the section just reached the top of the viewport" →
"bottom of the section just reached the bottom of the viewport" to the
0→1 progress range. Combined with a 350vh section and a sticky inner div,
the effect is: canvas stays pinned to viewport while scroll pushes
through 2.5 extra viewport heights of distance, and frame index tracks
that distance linearly.

## Sources

- `references/raw/framer-motion-usescroll-docs.md` — canonical shape of
  `useScroll`, `useTransform`, and `useMotionValueEvent`; the "never
  setState in the callback" rule.
- `references/raw/apple-scroll-technique-css-tricks.md` — the technique
  itself; the cover-fit math; the DPR fix the article misses.
