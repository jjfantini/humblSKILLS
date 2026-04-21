# Source: Apple-style scroll-controlled video (CSS-Tricks)

Distilled notes (2026-04-20) from the CSS-Tricks article on replicating
the Apple product-page scroll-video effect. Original article:
"Lets Make One of Those Fancy Scrolling Animations Used on Apple Product Pages"
(https://css-tricks.com/lets-make-one-of-those-fancy-scrolling-animations-used-on-apple-product-pages/).

---

## Core technique (unchanged since 2019)

1. Take a short video (ideally <10s).
2. Extract every Nth frame as an image (JPEG or WebP).
3. Preload all frames as `Image()` objects into an in-memory array.
4. On scroll, compute `progress = scrollY / maxScroll` (0–1) and
   `frameIndex = Math.round(progress * (frameCount - 1))`.
5. Draw `frames[frameIndex]` to a sticky/fixed canvas.

Apple literally uses this technique on iPad Pro, AirPods, iPhone pages.
Viewing source reveals a large image array (`seq = [...]`) and a scroll
handler that maps scroll to frame index.

---

## Why canvas, not `<video currentTime>`

The naive approach is to use an HTML `<video>` element and set
`currentTime` based on scroll. **This is janky in practice.** Video
decoders aren't designed for random seek on every scroll event. You get:

- Stutters on scroll direction change
- Frame-drop when scrolling fast
- Different behavior across codec/container combinations
- Memory/battery spikes on mobile

Pre-extracted frames eliminate the decoder from the hot path. Every
frame is already a decoded bitmap by the time scroll fires.

---

## Preloading strategy (from the article)

The article's basic version preloads all frames on mount and waits until
all images emit `load` before activating scroll. In practice:

- 100-frame sequence × ~40KB WebP = ~4MB (fine on 4g, slow on 3g)
- Use `Promise.all()` with an `Image().decode()` per frame
- For production: split into critical (first ~10) and streamed (rest);
  gate scroll until critical batch is decoded

---

## The scroll math (Framer Motion adaptation)

Article uses raw scroll listener. Framer Motion idiom:

```jsx
const { scrollYProgress } = useScroll({ target: ref, offset: ["start", "end"] });
const frameIndex = useTransform(scrollYProgress, [0, 1], [0, frameCount - 1]);
useMotionValueEvent(frameIndex, "change", (idx) => {
  drawFrame(Math.round(idx));
});
```

This avoids the article's `React.setState(frame)` anti-pattern (React
re-render on every scroll tick kills perf).

---

## Canvas dimensions + DPR

Article's example misses `devicePixelRatio`; production code must add it:

```jsx
const dpr = window.devicePixelRatio || 1;
canvas.width  = cssWidth  * dpr;
canvas.height = cssHeight * dpr;
canvas.style.width  = cssWidth  + "px";
canvas.style.height = cssHeight + "px";
ctx.scale(dpr, dpr);
```

Without this, Retina/2x displays render a half-resolution blur.

---

## Reference implementations

- Apple product pages (view source on iPad Pro launch page)
- CSS-Tricks demo code (vanilla JS, single-file HTML)
- Josh Comeau's Framer Motion scroll demos (modern React idiom)

---

## Known limitations / not solved by the technique

- **Memory on iOS Safari** — full-res frame sequences can blow the
  ~384MB canvas memory budget (not discussed in the article). Downshift
  resolution on mobile.
- **Initial bundle size** — 4-8MB of frames is a real cost; the article
  hand-waves it. In 2026, HTTP/2 + CDN + service-worker caching
  mitigate but don't eliminate.
- **No accessibility discussion** — `prefers-reduced-motion` must be
  honored; article predates universal awareness.
