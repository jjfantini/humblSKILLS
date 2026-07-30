# Source: Framer Motion — useScroll + useMotionValueEvent

Distilled notes (2026-04-20) from Framer Motion v11+ docs:
- https://motion.dev/docs/react-use-scroll
- https://motion.dev/docs/react-use-motion-value-event
- https://motion.dev/docs/react-use-transform

Applies to Framer Motion v11+ (same API shape in v12). Works identically
in Next.js App Router client components.

---

## `useScroll`

Returns four motion values:

| Value             | Range    | Meaning                                      |
|-------------------|----------|----------------------------------------------|
| `scrollX`         | px       | Absolute horizontal scroll position          |
| `scrollY`         | px       | Absolute vertical scroll position            |
| `scrollXProgress` | 0 → 1    | Normalized horizontal progress               |
| `scrollYProgress` | 0 → 1    | Normalized vertical progress                 |

### Common call shapes

```jsx
// Track full page scroll
const { scrollYProgress } = useScroll();

// Track a specific element's scroll into viewport
const ref = useRef(null);
const { scrollYProgress } = useScroll({
  target: ref,
  offset: ["start start", "end end"], // progress runs 0→1 from "ref top hits viewport top" to "ref bottom hits viewport bottom"
});

// Track a scrollable container
const { scrollYProgress } = useScroll({ container: containerRef });
```

### `offset` grammar

`[startOffset, endOffset]`. Each offset is `"<target-edge> <container-edge>"`:

- `"start start"` — target's start edge meets container's start edge
- `"end end"`     — target's end edge meets container's end edge
- `"start end"`   — target's start edge meets container's end edge (just entering viewport)
- `"end start"`   — target's end edge meets container's start edge (just exiting viewport)

For the Apple-style scroll-video effect, the idiomatic choice is
`offset: ["start start", "end end"]`, where the canvas section occupies
something like `350vh` and progress runs 0→1 over that distance.

---

## `useTransform`

Maps a motion value into another range:

```jsx
const frameIndex = useTransform(
  scrollYProgress,
  [0, 1],
  [0, frameCount - 1]
);
```

- Input range `[0, 1]`
- Output range `[0, frameCount - 1]`
- Output is itself a motion value, not a number; read via `.get()` or
  bind with `useMotionValueEvent`.
- Supports non-linear mapping with more breakpoints:
  `useTransform(scrollYProgress, [0, 0.2, 1], [0, 10, frameCount - 1])`.

---

## `useMotionValueEvent` (the critical hook for canvas)

Subscribes to a motion value without triggering React re-renders.

```jsx
useMotionValueEvent(frameIndex, "change", (latest) => {
  const idx = Math.round(latest);
  drawFrame(idx);
});
```

- Event names: `"change"`, `"animationStart"`, `"animationComplete"`,
  `"animationCancel"`
- The callback runs inside Framer's own rAF loop — **do not** wrap it in
  your own `requestAnimationFrame`.
- **Never** call React state setters inside this callback. That defeats
  the purpose — you want the draw to happen without rendering.
- Read stale-free values via closure or `frameIndex.get()` inside the
  callback.

### Why not `useEffect(() => frameIndex.on("change", ...))`

Historically people wired `.on("change", fn)` in `useEffect`. That works
but `useMotionValueEvent` does the mount/unmount cleanup for you and is
the current canonical form.

---

## Typical component skeleton

```jsx
"use client";
import { useRef } from "react";
import { useScroll, useTransform, useMotionValueEvent } from "framer-motion";

export function ScrollFrameCanvas({ frames }) {
  const containerRef = useRef(null);
  const canvasRef    = useRef(null);

  const { scrollYProgress } = useScroll({
    target: containerRef,
    offset: ["start start", "end end"],
  });
  const frameIndex = useTransform(scrollYProgress, [0, 1], [0, frames.length - 1]);

  useMotionValueEvent(frameIndex, "change", (latest) => {
    const ctx = canvasRef.current?.getContext("2d");
    const img = frames[Math.round(latest)];
    if (ctx && img) ctx.drawImage(img, 0, 0);
  });

  return (
    <section ref={containerRef} style={{ height: "350vh" }}>
      <canvas ref={canvasRef} style={{ position: "sticky", top: 0 }} />
    </section>
  );
}
```

This is the shape the generated component template substitutes into.

---

## Gotchas

- **SSR**: `useScroll` depends on `window`. Any component using it must
  be a client component (`"use client"` at the top in App Router).
- **No nested containers**: if you wrap the section in a custom scroll
  container (e.g. a Lenis root), pass `container: lenisRef` or drop the
  custom container entirely — conflicts between the browser's scroll and
  a custom one will give you mismatched progress values.
- **Offsets are relative to target+container, not viewport**: easy to
  confuse. Test by logging `scrollYProgress.get()` at known scroll
  positions.
- **`useTransform` output is a motion value**: don't render it directly
  with `{frameIndex}` — JSX will stringify the motion value object.
  Always read via `.get()` or subscribe.
