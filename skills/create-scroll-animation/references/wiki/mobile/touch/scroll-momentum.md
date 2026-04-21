---
title: "iOS Touch Momentum, Native Scroll, and When Lenis Breaks It"
context: mobile
category: touch
concept: scroll-momentum
description: "iOS native scroll momentum works out-of-the-box with Framer Motion useScroll — no intervention needed. Lenis smooth-scroll can *break* iOS momentum if misconfigured. Recommendation: don't add Lenis by default; if added, test touch fling on a real iPhone."
tags: iOS, touch, momentum, lenis, native-scroll, scroll-behavior
sources: []
last_ingested: 2026-04-20
---

## Native iOS Momentum Works

Framer Motion's `useScroll` reads `window.scrollY` (or a scroll
container's `scrollTop`). iOS Safari's native momentum scroll updates
`scrollY` continuously during the fling. The generated component picks
this up and drives frame draws at momentum speed.

**No extra CSS or JS needed.** Specifically:

- No `overscroll-behavior: contain` required (don't add it — it disables
  bounce).
- No `touch-action` overrides required.
- No `-webkit-overflow-scrolling: touch` required (deprecated; the
  modern iOS default scroll behavior handles this).

If the component works on desktop Chrome, it will work on iOS Safari
with touch momentum. The one thing to test is *fast-fling responsiveness*
— if the frame draw can't keep up, you'll see frame-skip during the fling.
That's a symptom of under-preloading, not of touch handling.

## Where Lenis Helps (and where it hurts)

Lenis smooth-scroll adds a virtual scroll layer on top of native scroll.
The browser's scroll still fires, but Lenis interpolates between the
native scroll position and a smoothed target position.

**Helps:**

- Desktop wheel scroll feels snappier and more refined.
- Anchor-link jumps animate smoothly.
- Page-wide scroll-linked animations share a unified scroll timeline.

**Hurts on iOS if misconfigured:**

- If `smoothTouch: true`, Lenis takes over touch scroll and you lose
  native iOS momentum (it's replaced by Lenis's own interpolation,
  which feels different and often worse on touch).
- **Always set `smoothTouch: false`** or use the default (which is
  `false` in recent Lenis versions, but confirm per version).
- Even with `smoothTouch: false`, Lenis's intervention during touch
  scroll can cause stutter on low-end Android. Test on real devices.

**Minimal safe Lenis config for this use case:**

```tsx
<ReactLenis root options={{ lerp: 0.1, smoothWheel: true }}>
  {children}
</ReactLenis>
```

- `lerp: 0.1` → middle-ground smoothing on wheel
- `smoothWheel: true` → desktop-only enhancement
- `smoothTouch` omitted → defaults to `false`, preserves iOS momentum

## Recommendation for `create-scroll-animation`

**Default: no Lenis.** The effect works fine without it; adding a
smooth-scroll layer spends complexity budget for a polish gain most
users won't notice.

**Opt-in via Phase 2 interview:** if the user explicitly asks for
"smooth" / "silky" / "polish", include the Lenis wrapper template and
configure it conservatively (desktop-wheel only, no touch override).

## Known iOS scroll quirks (unrelated to Lenis but worth knowing)

- **Rubber-band overscroll at the top or bottom of the section** can
  cause `scrollYProgress` to briefly exceed `[0, 1]`. The generated
  component's `frameIndex = useTransform(progress, [0, 1], [0, N-1])`
  clamps this safely — no bug, no fix needed.
- **Scroll snap + our sticky container:** avoid combining CSS scroll
  snap with the sticky inner div. Snap + sticky + momentum = bugs.
- **Address bar show/hide on iOS** changes viewport height mid-scroll.
  The component's `sizeCanvas` on `resize` catches this; frames will
  redraw sharp at the new height. You may see a one-frame flicker during
  the address bar transition — this is unavoidable and Apple does it too.

## Testing

Real iPhone. No substitute.

- Fling scroll hard — does the frame sequence track the fling?
- Scroll slowly with sticky finger — is frame-to-scroll ratio linear?
- Scroll past the end of the section — does the canvas freeze on the
  last frame cleanly?
- Scroll back up — no pop-in?

If any of these fail, the fix is in preload (more critical frames),
not in touch handling.

## Sources

Synthesis concept. The iOS momentum behavior is standard iOS Safari; the
Lenis configuration caveats come from the Lenis docs and community-known
gotchas.
