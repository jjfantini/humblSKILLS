---
title: "iOS Safari Canvas Memory Cap and Resolution Downshift"
context: mobile
category: memory
concept: ios-safari-cap
description: "iOS Safari caps canvas-backing-store memory at ~384MB per origin. At 1920x1080 RGBA, each decoded frame is ~8.3MB — 50 frames blows the cap. Downshift mobile to 1280x720 (or 960x540 on low-memory) to survive with a margin. Without this: tab reloads mid-scroll."
tags: iOS, safari, mobile, memory, canvas, downshift, resolution
sources:
  - "references/raw/ios-safari-canvas-memory.md"
last_ingested: 2026-04-20
---

## The iOS Memory Problem

WebKit caps total canvas-backing-store memory per origin at roughly
**384 MB** (lower on older / low-RAM iPhones, closer to 224–288 MB).
Exceeding the cap doesn't throw an error — it silently clears the
canvas, reloads the tab, or crashes Safari.

**The math that bites you:**

```
1920 × 1080 × 4 bytes (RGBA) = 8,294,400 bytes ≈ 8.3 MB per decoded frame
```

Multiply by frame count:

| Frame count | Memory at 1920×1080 | iOS verdict       |
|-------------|---------------------|-------------------|
| 50          | ~415 MB             | Over cap — crash  |
| 100         | ~830 MB             | Guaranteed crash  |
| 150         | ~1.24 GB            | Instant crash     |

**Desktop users never see this.** Desktop Chrome/Safari caps are
effectively uncapped for practical purposes (GB+). So "works on my
laptop" isn't evidence. Test on a real iPhone.

## Mitigations (in order of preference)

### 1. Downshift resolution on mobile (do this first)

At 1280×720 RGBA each frame is ~3.7 MB. 100 frames = ~370 MB — still
close, but survivable on most iPhones with a margin.

At 960×540 RGBA each frame is ~2.1 MB. 100 frames = ~210 MB — safe on
all but the oldest supported devices.

**Implement this at extraction time, not component time.** Keep a
separate mobile frame set in `public/frames-mobile/` and pick the URL
pattern based on viewport width.

### 2. Reduce mobile frame count

Instead of 100 frames at 720p, try 60 frames at 720p. Sixty frames is
still smooth enough to read as "video" at scroll speed and cuts memory
40% further.

### 3. Sliding-window decoding (advanced)

Keep only ~20 decoded frames around the current scroll position;
release the rest by setting `img.src = ''` on out-of-window frames. More
code, more bugs, more risk. Only reach for this if (1) and (2) aren't
enough.

## Detection

There is no reliable "is iOS" API. Use a combination:

```ts
function isLikelyIOS(): boolean {
  return /iPad|iPhone|iPod/.test(navigator.userAgent) && !(window as any).MSStream;
}
```

`navigator.deviceMemory` **does not work on Safari** as of 2026 — it's
`undefined`. Don't rely on it for iOS gating.

The viewport-width heuristic in the generated component (`matchMedia('(max-width: 768px)')`)
is a good proxy for "mobile-class memory envelope" and doesn't require
UA sniffing. But it misses iPads in landscape (which are mobile-memory
but wide-viewport). If you need stricter gating, combine UA + viewport.

## URL pattern wiring

The default URL pattern in the component can point at a multi-resolution
set:

```tsx
const urlPattern = window.matchMedia('(max-width: 768px)').matches
  ? '/frames-mobile/frame_{index}.webp'
  : '/frames/frame_{index}.webp';

<ScrollFrameCanvas frameUrlPattern={urlPattern} />
```

This requires extracting twice (once at 1920, once at 720 or 540). The
`to-webp.sh` script already accepts an `$OUTDIR` — call it twice with
different sources/outputs.

## Testing

- **Real iPhone, not the Simulator.** The Simulator uses the host Mac's
  memory budget and won't reproduce the cap.
- **Scroll fast, then hard-reload, then scroll again.** Memory behavior
  differs between cold and warm caches.
- **Watch for tab reloads.** Safari reloading the tab silently is the
  symptom. Check Safari's remote debugger console for
  `WebContent process crashed` messages.

## Sources

- `references/raw/ios-safari-canvas-memory.md` — the ~384MB cap, the
  decoded-frame math, the public bug-tracker citations, and the
  `navigator.deviceMemory` caveat on Safari.
