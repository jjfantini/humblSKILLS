---
title: "OCR Resolution Is the Deliverable, Dedupe on Text Not Pixels"
context: perception
category: ocr
concept: vision-resolution
description: "Vision OCR yield vs frame width (1280px garbage, 2400px readable) and why consecutive frames are deduped by OCR text hash instead of mpdecimate"
tags: vision, ocr, VNRecognizeTextRequest, resolution, dedupe, mpdecimate, performance
sources:
  - "references/raw/probe-findings-2026-09-10.md"
last_ingested: 2026-09-10
command: scripts/watch.sh
---

## Keep the Pixels, Dedupe the Text

`watchkit ocr` runs Vision's `VNRecognizeTextRequest` (`.accurate`, language correction on) across
frames with `DispatchQueue.concurrentPerform`. Measured on a 3440x1440 screen recording:
0.55 s per frame alone, 0.38 s per frame batched (40 frames in 15.3 s wall, 323% CPU).

**Incorrect (looks like a free speed-up):**

```bash
-vf "select=...,scale=1280:-2"
# same frames yielded 1768 chars, mostly garbage: "fuck-serged sale Sata dewlap"
```

**Correct:**

```bash
-vf "select=...,scale='min(iw\,2560)':-2"
# 1720px -> 4810 chars; 2400px -> 8382 readable chars from the same 3 frames
```

Small UI text is the whole point of a screen recording, and it vanishes below roughly 1720 px wide.
Width is capped at 2560 only to bound OCR time on 5K displays; never add a smaller scale. The
budget lever is the sample rate (`--fps 0.5`), not the resolution.

Dedupe: `mpdecimate` kept 116 of 159 frames of a **static** desktop recording (73%) at defaults and
125 with loosened thresholds; it is tuned for film grain, not a blinking cursor. `watch.sh` instead
hashes each frame's OCR text and skips a frame whose text equals the previous one. The frame file is
kept on disk so `frame.sh` can still return it; only the index row is skipped.

Cost model to tell the user: 10 min at 1 fps is about 600 frames x 0.4 s = roughly 4 minutes.

## Sources

- `references/raw/probe-findings-2026-09-10.md` - timings, resolution yield table, mpdecimate counts.
