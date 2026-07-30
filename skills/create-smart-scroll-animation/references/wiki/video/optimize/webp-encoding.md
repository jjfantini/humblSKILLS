---
title: "Batch Convert PNG Frames to WebP (JPEG Fallback)"
context: video
category: optimize
concept: webp-encoding
description: "Converts the PNG intermediate frames to WebP at q=80 targeting <40KB per frame @ 1920x1080. Uses cwebp if present; falls through to ffmpeg libwebp if missing. JPEG fallback branch documented for browsers that don't support WebP (vanishingly rare in 2026)."
tags: webp, cwebp, ffmpeg, jpeg, compression, optimization
sources:
  - "references/raw/scroll-stop-builder-SKILL.md"
  - "references/raw/apple-scroll-technique-css-tricks.md"
last_ingested: 2026-04-20
command: scripts/to-webp.sh
---

## WebP Encoding

WebP is the default delivery format. It compresses ~30% smaller than JPEG
at equivalent visual quality and is supported in every browser that
matters in 2026 (Chrome, Safari 14+, Firefox, Edge, all modern mobile).

**Targets:**

| Metric                 | Target                                    |
|------------------------|-------------------------------------------|
| Per-frame size         | <40 KB at 1920×1080                       |
| Bundle size (100 fr)   | <4 MB                                     |
| Bundle size (150 fr)   | <6 MB                                     |
| Quality (cwebp `-q`)   | 80 (good visual quality, aggressive size) |

**Command invocation:**

```bash
bash scripts/to-webp.sh /tmp/frames-raw public/frames 80
```

- `$1` = input directory containing PNGs
- `$2` = output directory (created if missing)
- `$3` = quality 0–100 (default 80)

**Primary path (cwebp, preferred):**

```bash
for png in "$INDIR"/frame_*.png; do
  out="$OUTDIR/$(basename "${png%.png}.webp")"
  cwebp -q "$QUALITY" -quiet "$png" -o "$out"
done
```

- `cwebp` is the reference encoder from Google (ships with `libwebp`).
- Best quality-per-byte ratio.
- Install: `brew install webp` (macOS) or `apt install webp` (Debian).

**Fallback path (ffmpeg libwebp, if cwebp missing):**

```bash
ffmpeg -v error -i "$png" -c:v libwebp -quality "$QUALITY" "$out"
```

- Works but produces ~15–20% larger files than `cwebp` at the same `-q`.
- Good enough; avoids making cwebp a hard install requirement.
- The script auto-selects: uses cwebp if on PATH, else falls through.

**Size report:**

The script prints a final report:

```
to-webp.sh: 100 frames, 3.8 MB total (avg 38 KB/frame)
```

If the total exceeds 8 MB for 100 frames, drop quality to 70 and re-run.
If it exceeds 8 MB at q=70, the source resolution is too high — consider
scaling to 1280 wide in the extraction step instead of 1920.

**Incorrect — using PNG or default-quality JPEG for delivery:**

```bash
# PNG frames at 1920x1080 are 400-800 KB each. 100 frames = 40-80 MB.
# Guaranteed to miss LCP budget.
cp /tmp/frames-raw/*.png public/frames/
```

**Correct — WebP q=80 as the delivery format:**

```bash
bash scripts/to-webp.sh /tmp/frames-raw public/frames 80
# 100 frames → ~3.8 MB total → LCP-safe
```

## JPEG fallback branch

Ship JPEGs alongside WebPs only if the user explicitly targets legacy
browsers (<1% of traffic in 2026). The technique:

1. Generate JPEGs in a parallel step:
   ```bash
   for png in /tmp/frames-raw/frame_*.png; do
     out="public/frames/$(basename "${png%.png}.jpg")"
     ffmpeg -v error -i "$png" -q:v 2 "$out"
   done
   ```
2. In the generated component, use a `<picture>` / `srcset` pattern to
   serve WebP first, JPEG as fallback.

For v0.1.0 of this skill: do **not** ship the JPEG branch by default.
WebP support is universal enough that the bytes and the code complexity
aren't justified. Only add if the user asks.

## Tradeoff table (q vs. size, 1920×1080)

| Quality | Avg KB/frame | Visual |
|---------|--------------|--------|
| 95      | ~80          | Indistinguishable from source |
| 85      | ~50          | Excellent, default for hero imagery |
| **80**  | **~38**      | **Default — our pick** |
| 75      | ~30          | Slight softening on gradients |
| 70      | ~25          | Acceptable for secondary imagery, not hero |

Numbers approximate; real sizes depend heavily on content (photographic
vs. flat-color imagery). The script's size report is the source of truth.

## Sources

- `references/raw/scroll-stop-builder-SKILL.md` — the reference skill
  uses JPEG `-q:v 2`. We improve on this with WebP as primary.
- `references/raw/apple-scroll-technique-css-tricks.md` — the article
  discusses frame-sequence bandwidth cost. WebP directly attacks that cost.

## Command

Run the associated script:

```bash
bash scripts/to-webp.sh <indir> <outdir> [quality]
```

Prints a size report on completion.
