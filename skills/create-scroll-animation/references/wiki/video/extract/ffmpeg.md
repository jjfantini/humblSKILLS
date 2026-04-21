---
title: "Extract Numbered PNG Frames with FFmpeg"
context: video
category: extract
concept: ffmpeg
description: "Runs FFmpeg with fps + scale filters to emit a numbered PNG sequence at target_fps. PNGs are an intermediate; to-webp.sh converts them to WebP downstream. Scales to 1920 wide with -2 height for codec-safe even dimensions."
tags: ffmpeg, extraction, png, frame-sequence, scale
sources:
  - "references/raw/scroll-stop-builder-SKILL.md"
last_ingested: 2026-04-20
command: scripts/extract-frames.sh
---

## Extract Frames

Take a probed MP4, compute `target_fps`, and dump numbered PNG frames to
a temp directory. PNG is deliberately chosen as the intermediate format
(losslessness for the WebP conversion step), **not** as the delivered
format. The delivered format is WebP — see
`video/optimize/webp-encoding.md`.

**Command invocation:**

```bash
bash scripts/extract-frames.sh path/to/hero.mp4 /tmp/frames-raw 100
```

- `$1` = video path
- `$2` = output directory (will be created)
- `$3` = target frame count (default 100 if omitted)

**What the script runs internally:**

```bash
mkdir -p "$OUTDIR"
ffmpeg -v error -i "$VIDEO" \
  -vf "fps=${TARGET_FPS},scale=1920:-2" \
  "$OUTDIR/frame_%04d.png"
```

**Output:** `frame_0001.png` … `frame_NNNN.png` in `$OUTDIR`, plus a
`manifest.json` at the same path listing the frame count, source
duration, and computed `target_fps`.

**Filter explained:**

| Filter                | Purpose                                                  |
|-----------------------|----------------------------------------------------------|
| `fps=${TARGET_FPS}`   | Resample to our target frame rate; drops / picks frames  |
| `scale=1920:-2`       | Scale width to 1920, height auto with `-2` (force even)  |

The `-2` in scale forces the computed height to the nearest even
integer. Odd dimensions break some WebP/H.264 codec paths, so this is
defensive even though we're emitting PNG.

**Incorrect — extracts every source frame, ignores target count:**

```bash
ffmpeg -i hero.mp4 -vf scale=1920:-2 frame_%04d.png
# A 6s 30fps video emits 180 frames. Blows the budget.
```

**Correct — fps filter enforces the budget:**

```bash
# For a 6.2s video with target 100 frames:
# TARGET_FPS = round(100 / 6.2) = 16
ffmpeg -v error -i hero.mp4 \
  -vf "fps=16,scale=1920:-2" \
  /tmp/frames-raw/frame_%04d.png
# Emits ~99 frames (rounding).
```

**Why PNG as the intermediate (not JPEG):**

- WebP encoding from PNG preserves full color fidelity → we can hit a
  smaller final WebP at equivalent visual quality.
- The reference skill emits JPEG directly because it doesn't have a
  WebP conversion pass. We do, so PNG → WebP is the higher-quality path.
- Intermediate PNGs live in `/tmp/frames-raw` (not committed anywhere);
  disk space during the pipeline is not a concern.

**Edge cases:**

- **`target_fps` exceeds source fps:** FFmpeg duplicates frames to hit
  the rate. We explicitly don't want this. The `probe.sh` / calling
  code must clamp `target_fps` to `min(computed, source_fps)`.
- **Output dir is non-empty:** the script `rm -rf`s and recreates to
  avoid half-stale manifests. Confirm with the user before overwriting
  anything outside `/tmp/*`.
- **Videos with audio / multiple streams:** `-vn` (no video) isn't
  needed — we're not muxing an output file, we're decoding to image
  sequence. FFmpeg handles this correctly by default.

## Sources

- `references/raw/scroll-stop-builder-SKILL.md` — the "Step 2: Extract
  Frames" section of the reference skill, which establishes the
  `fps=X,scale=1920:-2` filter shape.

## Command

Run the associated script:

```bash
bash scripts/extract-frames.sh <video> <outdir> [target_count]
```

Emits PNGs + a `manifest.json` into `<outdir>`.
