---
title: "Preload Frames with a Concurrency Semaphore"
context: react
category: component
concept: preload-strategy
description: "Batched preload: first N frames (default 10) are awaited as critical — scroll gates until they're decoded. Remaining frames stream in with a concurrency semaphore (6 desktop / 3 mobile) to avoid browser decoder throttling and mobile memory spikes."
tags: preload, decode, semaphore, concurrency, critical, streaming
sources:
  - "references/raw/apple-scroll-technique-css-tricks.md"
last_ingested: 2026-04-20
---

## Preload Strategy

Inline inside the single-file component. Two phases:

1. **Critical batch** — first N frames (default 10). Awaited with
   `img.decode()`. Scroll is gated (`setReady(false)` until done).
2. **Streaming batch** — remaining frames, fired with a concurrency
   semaphore. Users can start scrolling as soon as the critical batch
   is ready; the rest fill in behind the scroll position.

**Why a semaphore?**

- HTTP/2 multiplexes well but the browser's `ImageDecoder` (the thing
  behind `img.decode()`) throttles past ~6–8 in-flight decodes.
- On mobile, peak-decoded-image memory during a burst decode is the
  single biggest cause of iOS Safari canvas-memory crashes. Concurrency
  3 keeps the burst small.

**Inline implementation:**

```tsx
async function preloadAll(
  frameCount: number,
  frameUrl: (idx: number) => string,
  criticalCount: number,
  concurrency: number,
  framesRef: React.MutableRefObject<Map<number, HTMLImageElement>>,
  onCriticalReady: () => void,
  onProgress: (loaded: number) => void,
) {
  let loaded = 0;
  const loadOne = async (idx: number) => {
    const img = new Image();
    img.decoding = 'async';
    img.fetchPriority = idx < criticalCount ? 'high' : 'auto';
    img.src = frameUrl(idx);
    await img.decode().catch(() => {
      /* decode can throw on decode errors; swallow and move on */
    });
    framesRef.current.set(idx, img);
    loaded += 1;
    onProgress(loaded);
  };

  // Phase 1: critical batch, fully awaited.
  await Promise.all(
    Array.from({ length: Math.min(criticalCount, frameCount) }, (_, i) => loadOne(i)),
  );
  onCriticalReady();

  // Phase 2: streamed rest, bounded by the semaphore.
  const queue: number[] = [];
  for (let i = criticalCount; i < frameCount; i++) queue.push(i);

  const workers = Array.from({ length: concurrency }, async () => {
    while (queue.length > 0) {
      const next = queue.shift();
      if (next === undefined) return;
      await loadOne(next);
    }
  });
  await Promise.all(workers);
}
```

**Wired into the component:**

```tsx
useEffect(() => {
  const isMobile = window.matchMedia('(max-width: 768px)').matches;
  const concurrency = isMobile ? 3 : 6;
  const criticalCount = Math.min(10, frameCount);

  const frameUrl = (idx: number) =>
    frameUrlPattern.replace('{index}', String(idx + 1).padStart(frameDigits, '0'));

  preloadAll(
    frameCount,
    frameUrl,
    criticalCount,
    concurrency,
    framesRef,
    () => {
      setReady(true);
      drawFrame(0);
    },
    (n) => setProgress(n / frameCount),
  );
}, [frameCount, frameUrlPattern, frameDigits]);
```

**Incorrect — fire all preloads at once:**

```tsx
// BAD on mobile: 100 in-flight decodes peak memory over the iOS cap.
await Promise.all(
  Array.from({ length: frameCount }, (_, i) => loadOne(i)),
);
```

**Correct — critical first, then throttled stream:**

(see inline implementation above)

## Tradeoffs

- **`fetchpriority='high'` on critical frames** tells the browser to
  prioritize these image fetches ahead of other page resources. Browser
  support is universal in 2026.
- **`decoding='async'`** hints that decoding can happen off the main
  thread. Doesn't guarantee it but gives the browser the option.
- **`img.decode()` vs. `img.onload`:** `decode()` returns a promise that
  resolves after the image is fully decoded and ready to paint without
  blocking. `onload` resolves earlier and can cause a stutter on first
  draw if decode lags.

## What criticalCount should be

Default 10. Tune based on:

- Scroll speed — if users typically scroll fast, raise to 20 so the
  early frames don't pop-in.
- Video content — if the first second of video has the money shot,
  raise so nothing in that range pops.
- Total frame count — 10 of 100 (10%) is reasonable; 10 of 40 (25%) is
  heavy. On short sequences, drop to 5.

## Sources

- `references/raw/apple-scroll-technique-css-tricks.md` — establishes
  the "preload all frames before showing" principle. This concept
  refines it to critical + streaming.
