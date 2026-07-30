---
title: "Responsive Fit: Cover Desktop, Zoomed-Contain Mobile"
context: react
category: component
concept: responsive-fit
description: "Desktop uses cover-fit (frame fills edge-to-edge, crops overflow). Mobile uses zoomed-contain (1.2× contain-fit, centered) so the subject stays visible on narrow portrait screens. Breakpoint: 768px. All math runs on every draw so it handles orientation changes."
tags: responsive, cover, contain, mobile, canvas, fit
sources:
  - "references/raw/scroll-stop-builder-SKILL.md"
last_ingested: 2026-04-20
---

## Cover vs. Zoomed-Contain

Different screens demand different framing strategies:

- **Desktop (landscape, wide):** cover-fit. The frame fills the entire
  viewport edge-to-edge. Any overflow is cropped. This is what makes
  the Apple hero feel cinematic.
- **Mobile (portrait, narrow):** cover-fit would crop so aggressively
  that the subject falls off-screen. Use a 1.2× zoomed-contain instead
  — contain-fit ensures the subject stays fully visible, and the 1.2×
  scale avoids the letterboxed look that contain-fit alone would give.

**Breakpoint:** 768px (standard tablet boundary). Below → zoomed-contain,
above → cover.

**Implementation (runs on every `drawFrame` call):**

```tsx
function pickFit(): 'cover' | 'zoomed-contain' {
  return window.matchMedia('(max-width: 768px)').matches
    ? 'zoomed-contain'
    : 'cover';
}

function drawFrame(idx: number) {
  const canvas = canvasRef.current;
  const ctx = canvas?.getContext('2d');
  const img = framesRef.current.get(idx);
  if (!canvas || !ctx || !img) return;

  const dpr = window.devicePixelRatio || 1;
  const cw = canvas.width  / dpr;   // CSS pixels
  const ch = canvas.height / dpr;
  const iw = img.naturalWidth;
  const ih = img.naturalHeight;

  ctx.clearRect(0, 0, cw, ch);

  const mode = pickFit();
  let scale: number;
  if (mode === 'cover') {
    // Fill — larger of the two scale factors
    scale = Math.max(cw / iw, ch / ih);
  } else {
    // zoomed-contain: smaller scale * 1.2
    scale = Math.min(cw / iw, ch / ih) * 1.2;
  }

  const dw = iw * scale;
  const dh = ih * scale;
  const dx = (cw - dw) / 2;
  const dy = (ch - dh) / 2;
  ctx.drawImage(img, dx, dy, dw, dh);
}
```

**Why check `matchMedia` on every draw instead of caching:**

The draw cost of `matchMedia` is microseconds — negligible next to
`drawImage`. Rechecking means orientation changes (portrait → landscape
on tablet) take effect on the very next scroll tick, with zero extra
logic. Simpler than wiring a resize listener to re-cache the mode.

**Alternative (cached with resize listener) — use only if profiling
shows `matchMedia` as a hot spot:**

```tsx
const fitRef = useRef<'cover' | 'zoomed-contain'>('cover');
useEffect(() => {
  const mq = window.matchMedia('(max-width: 768px)');
  const update = () => { fitRef.current = mq.matches ? 'zoomed-contain' : 'cover'; };
  update();
  mq.addEventListener('change', update);
  return () => mq.removeEventListener('change', update);
}, []);
```

This is a micro-optimization; don't reach for it unless you've measured.

## Scroll-section height per breakpoint

Slower scroll on desktop (more frames per pixel); faster on mobile so
the interaction feels responsive on shorter touch scrolls.

| Viewport | Section height |
|----------|----------------|
| Desktop (>1024px) | 350vh |
| Tablet (768–1024) | 300vh |
| Phone   (<768px)  | 250vh |

Exposed as the `scrollHeightVh` prop so users can tune. The template's
default is 350; the orchestration concept picks the right default based
on target device.

## Why 1.2× for zoomed-contain (not 1.1× or 1.5×)

- **1.0× (pure contain):** letterboxed on tall portrait screens; looks
  broken.
- **1.1×:** barely fills the edge; still visible bars in the worst
  aspect-ratio mismatch.
- **1.2×:** sweet spot. Fills aggressively without cropping the subject
  on typical 9:16 or 9:19.5 portrait screens.
- **1.5×+:** starts cropping the subject. Defeats the point of
  contain-fit.

## Sources

- `references/raw/scroll-stop-builder-SKILL.md` — the reference skill
  establishes the cover/zoomed-contain dichotomy and the 350/300/250vh
  section-height scale. We use the same.
