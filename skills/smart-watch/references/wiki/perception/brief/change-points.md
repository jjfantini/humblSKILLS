---
title: "The Brief: Screen Changes, Error Hits, Pins, Contact Sheet"
context: perception
category: brief
concept: change-points
description: "How brief.sh turns the index into the deliverable - word-set Jaccard change detection tuned against measured OCR jitter, an error regex, pinned frames, and a tiled contact sheet"
tags: brief, change-detection, jaccard, errors, contact-sheet, pins, digest
sources:
  - "references/raw/probe-findings-2026-09-10.md"
last_ingested: 2026-09-10
command: scripts/brief.sh
---

## One Read Gives the Whole Recording

`watch.sh` ends by running `brief.sh`, which writes `brief.md` and `sheet.jpg` next to the frames
and prints the brief. Sections, in the order an agent needs them:

1. **Errors and warnings** - every OCR line or speech segment matching the error regex, with
   timestamp and frame path, de-duplicated by text.
2. **Pinned frames** - the `--at` moments: exact frame plus one 2 s before and after, never deduped.
3. **Screen changes** - frames whose visible words differ from the previous frame, with `+` lines
   that appeared and `-` lines that vanished.
4. **Transcript** - every speech segment with its span.
5. **Contact sheet** - up to 12 change-point frames tiled 4 wide in time order; the timestamps are in
   the brief because brew ffmpeg has no `drawtext`.

**Incorrect (exact-line similarity):**

```text
00:00:01.01  changed 92%   # same static desktop; OCR read "To Automation:" then "Automation."
```

**Correct (word-set Jaccard, threshold 0.5):**

```python
words = set(re.findall(r"[a-z0-9]{3,}", text.lower()))
sim = len(prev & cur) / len(prev | cur)     # static screen measured 0.75-0.96
novel = [l for l in lines if len(words(l) - prev) >= 2]   # +/- lines with >= 2 new words
```

Measured on a 4 s video of one screenshot: consecutive-frame word similarity 0.75, 0.92, 0.96 - the
0.75 is the h264 keyframe. A terminal streaming text scored 0.44-0.56. So 0.5 splits jitter from
real change; `--threshold` moves it (lower = fewer, bigger changes).

Error regex (case-insensitive except the `E[A-Z]{3,}` code form, which must stay case-sensitive or
"Enterprise" and "exercises" match): error, exception, traceback, failed, fatal, panic, warning,
undefined, cannot read, not found, denied, forbidden, unauthorized, timeout, refused, crash,
segfault, stack trace, unhandled, rejected, invalid, `ECONNREFUSED`-style codes, `4xx/5xx <phrase>`.

Why python3: set arithmetic and regex over thousands of rows is a few lines in Python and a page of
awk; python3 ships with the Command Line Tools that `swiftc` already requires, so it adds no install.

## Sources

- `references/raw/probe-findings-2026-09-10.md` - OCR jitter and mpdecimate measurements that motivated text-based change detection.
