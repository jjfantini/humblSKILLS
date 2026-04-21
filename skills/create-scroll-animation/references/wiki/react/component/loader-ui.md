---
title: "Percent-Done Loader with Scroll Gating"
context: react
category: component
concept: loader-ui
description: "Full-viewport loader overlay that shows percent of critical frames decoded, disables body scroll until ready, and fades out in 400ms. Prevents the jarring experience of scrolling into a blank canvas while frames are still decoding."
tags: loader, scroll-gating, overlay, progress, UX
sources: []
last_ingested: 2026-04-20
---

## Loader UI

The loader has two jobs:

1. **Tell the user something is happening.** Hero canvas without a loader
   looks like a broken page until frames decode.
2. **Prevent early scrolling.** If the user scrolls before the critical
   batch is ready, the canvas shows blank or stuck-at-frame-0. We gate
   scroll with `document.documentElement.style.overflow = 'hidden'`
   until `ready`.

**Inline in the component:**

```tsx
const [progress, setProgress] = useState(0);
const [ready, setReady] = useState(false);

// Gate body scroll while loading.
useEffect(() => {
  if (ready) return;
  const prev = document.documentElement.style.overflow;
  document.documentElement.style.overflow = 'hidden';
  return () => {
    document.documentElement.style.overflow = prev;
  };
}, [ready]);

return (
  <>
    <section ref={containerRef} style={{ height: `${scrollHeightVh}vh` }}>
      <div style={{ position: 'sticky', top: 0, height: '100vh' }}>
        <canvas ref={canvasRef} style={{ display: 'block' }} />
      </div>
    </section>

    {!ready && (
      <div
        aria-busy="true"
        aria-label="Loading scroll animation"
        style={{
          position: 'fixed', inset: 0, zIndex: 50,
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          backgroundColor: 'var(--brand-bg, #000)',
          transition: 'opacity 400ms ease',
          opacity: ready ? 0 : 1,
          pointerEvents: ready ? 'none' : 'auto',
        }}
      >
        <div style={{ width: 240, textAlign: 'center', color: '#fff' }}>
          <div style={{ fontSize: 14, opacity: 0.6, marginBottom: 12 }}>
            Loading
          </div>
          <div
            style={{
              height: 2, background: 'rgba(255,255,255,0.12)', borderRadius: 1,
              overflow: 'hidden',
            }}
          >
            <div
              style={{
                height: '100%',
                width: `${Math.round(progress * 100)}%`,
                background: 'var(--brand-accent, #fff)',
                transition: 'width 150ms ease',
              }}
            />
          </div>
          <div style={{ fontSize: 11, opacity: 0.4, marginTop: 10 }}>
            {Math.round(progress * 100)}%
          </div>
        </div>
      </div>
    )}
  </>
);
```

## Why 400ms fade-out (not 0ms, not 1000ms)

- **0ms (instant):** jarring. The user's eye hasn't latched onto the
  canvas before the overlay vanishes.
- **400ms:** the eye tracks the canvas appearing; the transition reads
  as intentional.
- **1000ms+:** feels like the loader hung. You've lost trust.

400ms is the default for Material and Apple motion systems. Stick to it
unless there's a brand reason to deviate.

## Scroll gating nuances

- Setting `document.documentElement.style.overflow = 'hidden'` disables
  scroll on both desktop and mobile.
- On iOS Safari, add `position: fixed; width: 100%` to `<body>` if you
  notice background scroll bleed-through on modal overlays. Not
  typically needed for our case since the loader is pointer-events:auto
  and covers the whole viewport.
- **Restore on unmount.** The cleanup function restores the previous
  overflow value. Without this, unmounting the component while the
  loader is still up leaves scroll disabled.

## Reduced motion

If `prefers-reduced-motion: reduce` is set, don't render the loader at
all — just show frame 0 as a static `<img>`. See `react/a11y/reduced-motion.md`.

## Brand hookup

`--brand-bg` and `--brand-accent` are CSS custom properties. The template
substitutes them inline from the Phase 2 interview (if the user opted
into brand chrome) or falls back to `#000` / `#fff`. This keeps the
component visually neutral by default.

## Sources

Synthesis concept. The gating + fade-out timing is standard UX; the
scroll-lock via `overflow: hidden` on `<html>` is the standard approach
for full-viewport modal overlays.
