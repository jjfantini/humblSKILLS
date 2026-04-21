# Source: iOS Safari canvas memory cap

Notes (2026-04-20) on the practical memory budget for canvas-backed
animations on iOS Safari. Citations are public bug reports / WebKit
release notes; precise numbers drift across iOS versions.

---

## The limit (as of iOS 17/18, Safari 17/18)

WebKit caps total canvas-backing-store memory per origin at
**~384 MB** (the exact number varies by device class and available RAM;
low-end iPhones may cap lower, closer to 224–288 MB). Exceeding the cap
triggers one of:

1. Silent canvas clearing — subsequent `drawImage` calls no-op.
2. Tab reload (Safari's "A problem repeatedly occurred" flow).
3. On older devices, a full Safari process crash.

The cap applies to the sum of all canvases on the page plus any decoded
images held by the JS heap that are referenced as ImageBitmap / drawn-to
canvas.

---

## Back-of-the-envelope for frame sequences

Each decoded frame in memory at 1920×1080 RGBA = `1920 * 1080 * 4` bytes
= **~8.3 MB decoded**, regardless of the compressed WebP/JPEG size on disk.

| Frame count | Decoded memory @ 1920×1080 | Verdict on iOS Safari |
|-------------|----------------------------|-----------------------|
| 50          | ~415 MB                    | **Over cap** — crash risk |
| 100         | ~830 MB                    | **Way over cap** — guaranteed crash |
| 200         | ~1.66 GB                   | Instant crash |

Even 50 full-res frames is over budget on most iPhones. Mitigations:

- **Downshift resolution on mobile** — 1280×720 RGBA = ~3.7 MB; 100
  frames = ~370 MB (still close, but survivable with a margin).
- **Further downshift on low-end** — 960×540 RGBA = ~2.1 MB; 100 frames
  = ~210 MB.
- **Reduce frame count on mobile** — 50-frame mobile variant vs
  100-frame desktop variant.
- **Stream frames in and out** — only keep a sliding window of ~20
  decoded frames around the current scroll position. More complex; only
  worth it if the above aren't enough.

---

## Detection

No reliable "is iOS Safari" API. Heuristics:

```js
const isIOS = /iPad|iPhone|iPod/.test(navigator.userAgent) && !window.MSStream;
const deviceMemory = navigator.deviceMemory; // undefined on Safari!
const isLowMem = deviceMemory !== undefined ? deviceMemory <= 4 : false;
```

`navigator.deviceMemory` is **not implemented in Safari** (as of
Safari 18). On iOS, assume low-memory defaults unless you can prove
otherwise via other signals (e.g., connection speed, viewport size).

---

## Practical rules for `create-scroll-animation`

1. **Default mobile variant at 1280×720**, not 1920×1080.
2. **Cap mobile frame count at ~60** unless the user explicitly opts
   into more and has tested on a real device.
3. **Preload concurrency 3 on mobile**, 6 on desktop — lower concurrency
   reduces peak decoded-image memory during the initial load burst.
4. **Release non-critical frames aggressively** — if the user has
   already scrolled past frame 40 and isn't likely to scroll back,
   `img.src = ""` on frames <20 to free memory. (Trade-off: scrolling
   back up requires re-fetch. Usually acceptable since network fetch is
   cheap compared to a crash.)
5. **Test on a real iPhone SE / iPhone 8.** Desktop Safari and iOS
   Simulator both have much more generous memory budgets than a real
   phone.

---

## Public citations

- WebKit bug #210748 — "Large canvas animations cause tab reload on iOS"
  (resolved by adding the ~384MB cap)
- Surma's "Is WebP a first-class image format on mobile?" post (2019)
  explored the memory-vs-disk tradeoff
- Apple developer forums thread: "iOS Safari canvas memory limit"
  (multiple community benchmarks converging on 200-400MB per origin)

Note: these are moving targets. Re-validate on new iOS releases.
