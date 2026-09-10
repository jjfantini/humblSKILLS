---
title: "Sample Frames Without Losing True Timestamps"
context: perception
category: frames
concept: pts-preserving-sampling
description: "select on the real timeline plus -fps_mode passthrough keeps frame times exact; the fps filter was measured 0.47 s off"
tags: ffmpeg, pts, vfr, select, showinfo, timestamps, frames
sources:
  - "references/raw/probe-findings-2026-09-10.md"
last_ingested: 2026-09-10
command: scripts/watch.sh
---

## Sample on the Real Timeline, Not a Retimed Grid

Every hit `ask.sh` returns is only as good as the timestamp on its frame. Screen recordings are VFR,
and the obvious sampling filter silently rewrites time.

**Incorrect (0.47 s off, verified against a burned-in clock):**

```bash
ffmpeg -i rec.mov -vf fps=1 -frame_pts 1 frames/%d.jpg
# fps= drops the source onto a 1 Hz grid; frame "1" showed a burned clock of 00:00:01.467
```

**Correct (exact across 20 sampled frames):**

```bash
ffmpeg -hide_banner -loglevel info -nostats -i rec.mov \
  -vf "select='isnan(prev_selected_t)+gte(t-prev_selected_t\,1)',scale='min(iw\,2560)':-2,showinfo" \
  -fps_mode passthrough -q:v 3 frames/%06d.jpg 2> showinfo.log
grep -o 'pts_time:[0-9.]*' showinfo.log | cut -d: -f2   # Nth line = Nth frame's true time
```

Why it works: `select` picks the first source frame at least INTERVAL after the previous pick, on the
source's own clock; `-fps_mode passthrough` forbids retiming; `showinfo` prints each kept frame's
`pts_time`. Reported times were `0, 1, 2.1, 3.1, 4.2` and matched the burned-in clock exactly.

Gotchas:
- `showinfo` logs at **info** level. `-loglevel error` swallows it and the pairing silently breaks; `watch.sh` dies with "no pts_time lines" instead.
- `-frame_pts 1` names files in stream timebase units (1/15360 here), not seconds. Use showinfo, not filenames.
- Homebrew ffmpeg is built without libfreetype: **no `drawtext`**. Don't plan on burning overlays.
- `selftest.sh` asserts this: every OCR row's `t_start` second must appear in that frame's burned-in clock.

## Sources

- `references/raw/probe-findings-2026-09-10.md` - the 0.47 s measurement and the 20-frame alignment check.
