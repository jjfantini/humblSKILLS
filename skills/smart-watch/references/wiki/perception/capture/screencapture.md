---
title: "Record the Screen with Apple's screencapture, Never a Custom Binary"
context: perception
category: capture
concept: screencapture
description: "Why watch.sh --record shells out to /usr/sbin/screencapture -v and what its flags do and do not capture"
tags: screencapture, recording, tcc, screen-recording-permission, audio
sources:
  - "references/raw/probe-findings-2026-09-10.md"
last_ingested: 2026-09-10
command: scripts/watch.sh
---

## Record with `/usr/sbin/screencapture -v`

Live capture is one Apple-signed binary: `screencapture -v -V <seconds> -D <display> -k [-g] out.mov`.
A 3 s capture of a 3440x1440 display produced a 1.6 MB h264 `.mov` with no new permission prompt
because the binary inherits the terminal's Screen Recording grant.

**Incorrect (tempting, breaks silently):**

```bash
# A homegrown ScreenCaptureKit helper compiled with swiftc
swiftc -O capture.swift -o capture && ./capture 10 out.mov
# ad-hoc signature = new cdhash on every rebuild = TCC grant revoked, prompt again
```

**Correct:**

```bash
/usr/sbin/screencapture -v -V 10 -D 1 -k out.mov   # -k shows clicks
```

Flag notes:
- `-D 1` is the main display; the numbering follows System Settings, not window position. Check with the Tester of what is on that display before recording.
- `-g` records the **default input device** (microphone), not system audio. System audio needs a loopback device (BlackHole) inside a Multi-Output Device selected as system output.
- Output is **variable frame rate**; `r_frame_rate=600/1` in ffprobe is a container artifact. This is why frame sampling must preserve PTS (see `perception/frames/pts-preserving-sampling.md`).
- Everything on screen is captured, secrets included, and OCR text is stored in plain text under `~/.local/state/smart-watch`. `scripts/forget.sh <id>` removes a source completely.

## Sources

- `references/raw/probe-findings-2026-09-10.md` - measured capture size, VFR observation, TCC reasoning.

## Command

```bash
bash scripts/watch.sh --record 30 [--audio] [--display 2]
```
