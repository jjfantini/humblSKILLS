# Source: scroll-stop-builder (Cursor skill)

Snapshot (2026-04-20) of the relevant portions of
`/Users/jjfantini/.cursor/skills/scroll-stop-builder/SKILL.md` and
`references/sections-guide.md`. Captured only the canvas + frame extraction
+ responsive-fit + best-practices content. The skill's other features
(starscape, annotation cards, count-ups, navbar pill) are intentionally
out of scope for `create-scroll-animation` — do not port them.

---

## Technique: Frame Sequence + Canvas

The most reliable approach for scroll-driven video:

1. Extract frames from the video using FFmpeg.
2. Preload all frames as images with a loading indicator.
3. Draw frames to a canvas based on scroll position.
4. The scroll position maps to a frame index — scrolling forward advances
   the video, scrolling backward reverses it.

Why not `<video>` with `currentTime`? Browser video decoders aren't
optimized for seeking on every scroll event. Canvas + pre-extracted frames
is buttery smooth and gives frame-perfect control.

---

## Step 1: Analyze the Video (ffprobe)

```bash
ffprobe -v quiet -print_format json -show_streams -show_format "{VIDEO_PATH}"
```

Extract duration, fps, resolution, total frame count. **Target 60-150
frames total.**

---

## Step 2: Extract Frames (FFmpeg)

```bash
mkdir -p "{OUTPUT_DIR}/frames"
ffmpeg -i "{VIDEO_PATH}" \
  -vf "fps={TARGET_FPS},scale=1920:-2" \
  -q:v 2 \
  "{OUTPUT_DIR}/frames/frame_%04d.jpg"
```

- `fps={TARGET_FPS}` where `TARGET_FPS = round(total_target_frames / duration_seconds)`
- `scale=1920:-2` → width 1920, height auto-computed preserving aspect (the
  `-2` ensures an even number for codec compatibility)
- `-q:v 2` → high quality JPEG

Reference skill uses JPEG. `create-scroll-animation` overrides this with
WebP (user decision 2026-04-20; see `video/optimize/webp-encoding.md`).

---

## Canvas rendering with Retina support

```javascript
canvas.width  = window.innerWidth  * window.devicePixelRatio;
canvas.height = window.innerHeight * window.devicePixelRatio;
canvas.style.width  = window.innerWidth  + 'px';
canvas.style.height = window.innerHeight + 'px';
```

The DPR scaling is non-negotiable — missing this is the #1 cause of
blurry canvas rendering on Retina / 2x / 3x displays.

---

## Cover-fit (desktop) vs zoomed contain-fit (mobile)

- **Desktop:** cover-fit — frame fills the viewport edge-to-edge; overflow
  is cropped. Use this when the subject is centered and edge detail is
  decorative.
- **Mobile:** a slightly zoomed contain-fit — the subject stays centered
  and visible, with ~1.2× zoom so the framing feels intentional rather
  than letterboxed.

Canvas scroll-animation heights control perceived speed:

| Viewport | Height |
|----------|--------|
| Desktop  | 350vh  |
| Tablet   | 300vh  |
| Phone    | 250vh  |

---

## Best practices (canvas-only subset)

1. **`requestAnimationFrame` for drawing.** Never draw directly in the
   scroll handler.
2. **`{ passive: true }` on scroll listeners.** Enables scroll optimizations.
3. **Canvas with `devicePixelRatio`.** Crisp on Retina.
4. **Preload all frames before showing.** No pop-in during scroll.
5. **Frame deduplication.** Only call `drawFrame` when the frame index
   actually changed.
6. **No `scroll-behavior: smooth`.** Interferes with frame-accurate
   scroll mapping.
7. **Sticky canvas.** `position: sticky` keeps the canvas viewport-fixed
   while the scroll container moves past it.

---

## Error recovery (canvas-only subset)

| Issue | Solution |
|-------|----------|
| Frames don't load | Check paths; a local server is required (file:// won't work) |
| Animation is choppy | Reduce frame count; ensure JPEG/WebP not PNG; each <100KB |
| Canvas is blurry | Apply `devicePixelRatio` scaling |
| Scroll feels too fast/slow | Adjust scroll-animation height (200vh fast → 500vh slow → 800vh cinematic) |

---

## Features dropped for `create-scroll-animation`

Intentionally out of scope. Do not port:

- Interview brand/colors/vibe beyond an optional Phase 2 color pair
- Animated starscape background
- Annotation cards with JS-based snap-stop scroll
- Count-up animations on specs
- Navbar scroll-to-pill transform
- Glass-morphism design system tokens
- Three.js card scanner
- White-first-frame requirement (dropped 2026-04-20 per user)
