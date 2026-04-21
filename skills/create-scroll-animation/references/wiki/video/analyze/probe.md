---
title: "Probe the Input MP4 and Pick a Frame Budget"
context: video
category: analyze
concept: probe
description: "Runs ffprobe to read duration/fps/resolution/codec, then computes a target_fps that yields a frame count in the 60-150 range (sweet spot for file size vs. scroll fluidity)."
tags: ffprobe, ffmpeg, video, analysis, frame-budget
sources:
  - "references/raw/scroll-stop-builder-SKILL.md"
last_ingested: 2026-04-20
command: scripts/probe.sh
---

## Probe and Frame Budget

The first step in the pipeline: inspect the video to know what we're
dealing with, then pick a `target_fps` that hits the frame-count sweet
spot. Too few frames (<50) and scroll-scrub feels like a slideshow. Too
many (>200) and the asset bundle blows the 8MB budget.

**Target: 60–150 frames total.** Below 60 = choppy scrub. Above 150 =
diminishing fluidity gains, real bandwidth cost. Default 100 unless the
video duration forces a different call.

**Command invocation:**

```bash
bash scripts/probe.sh path/to/hero.mp4
```

**Output (stdout JSON):**

```json
{
  "path": "path/to/hero.mp4",
  "duration": 6.2,
  "fps": 30,
  "width": 1920,
  "height": 1080,
  "codec": "h264",
  "total_frames": 186
}
```

**Picking `target_fps`:**

```
target_fps = round(target_frame_count / duration)
```

- 6s video, target 100 frames → `target_fps = 17` (round of 16.67)
- 4s video, target 100 frames → `target_fps = 25`
- 10s video, target 100 frames → `target_fps = 10`

If the computed `target_fps` exceeds the source `fps`, cap at source fps
and accept fewer frames (don't interpolate — FFmpeg's `fps` filter will
drop frames, not create new ones, which is correct for our use case).

**Incorrect — hand-picking fps without reading duration:**

```bash
# Output: 180 frames from a 6s video. Way over budget.
ffmpeg -i hero.mp4 -vf "fps=30,scale=1920:-2" -q:v 2 frame_%04d.jpg
```

**Correct — compute fps from duration and target count:**

```bash
bash scripts/probe.sh hero.mp4
# parse duration=6.2 from JSON
# target_fps = round(100 / 6.2) = 16
bash scripts/extract-frames.sh hero.mp4 /tmp/frames-raw 100
#   ^ this script does the math internally from the probe output
```

**Edge cases:**

- **Video shorter than 2s:** target 30–60 frames instead of 100. Scroll
  distance feels too short at 100 frames over <2s of source content.
- **Video longer than 15s:** you probably don't want this technique.
  Push back and suggest trimming before extraction. Frame counts >200
  exceed memory budgets.
- **Variable-framerate (VFR) sources:** `ffprobe` reports an average
  fps. The `fps` filter normalizes to a constant output fps; this is
  what we want.

## Sources

- `references/raw/scroll-stop-builder-SKILL.md` — the "Step 1: Analyze
  the Video" section of the reference skill, which establishes the
  60–150 frame target range and the ffprobe command shape.

## Command

Run the associated script:

```bash
bash scripts/probe.sh path/to/video.mp4
```

Prints compact JSON to stdout. Exits 0 on success; non-zero if the path
doesn't exist, ffprobe is missing, or the file isn't decodable.
