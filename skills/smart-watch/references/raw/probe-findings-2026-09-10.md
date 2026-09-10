# smart-watch viability probe — measured 2026-09-10

Host: macOS 26.5.2 (25F84), Apple Silicon, Swift 6.3.3, ffmpeg 8.1.2 (Homebrew), sqlite3 w/ FTS5.

## Native stack that replaces the upstream Python engine

| Upstream piece | Native replacement | Measured |
| --- | --- | --- |
| capture | `/usr/sbin/screencapture -v -V <s> -D <n> -k [-g]` | 3s @ 3440x1440 h264, 1.59 MB, no new TCC prompt |
| frame extract | ffmpeg `select` + `-fps_mode passthrough` | 0.19s for 3s clip; videotoolbox hwaccel available |
| OCR | Vision `VNRecognizeTextRequest`, compiled Swift helper | 0.55 s/frame single; 0.38 s/frame batched (40 frames, 15.3 s wall, 323% CPU) |
| transcription | `SpeechAnalyzer` + `SpeechTranscriber` (macOS 26) | 70 s audio -> 0.82 s = **85x realtime**, on-device, word-level `audioTimeRange`; en_* models preinstalled |
| index / search | `sqlite3` FTS5 (ships with macOS) | built in |
| query engine | the agent itself | n/a |

No Python, no Whisper, no tesseract, no MCP server.

## Gotchas found by measurement

1. **Timestamp drift (the load-bearing one).** `screencapture -v` writes VFR
   (`r_frame_rate=600/1` is a container artifact). Sampling with `-vf fps=1` and
   naming via `-frame_pts 1` produced frame `pts=1` whose burned-in content clock
   read `00:00:01.467` — **0.47 s off**, because the `fps` filter retimes onto a grid
   and discards true PTS.
   Correct recipe (verified exact against a burned-in counter across 20 samples):
   ```
   -vf "select='isnan(prev_selected_t)+gte(t-prev_selected_t\,1)',showinfo" \
   -fps_mode passthrough -frame_pts 1
   ```
   `showinfo` gave `pts_time: 0, 1, 2.1, 3.1, 4.2` matching burned timecodes exactly.
   `showinfo`/`metadata=print` log at info level — `-v error` silently swallows them.

2. **OCR resolution is the deliverable, not a perf knob.** Same 3440x1440 recording,
   same frames: 1280px wide -> 1768 chars of mostly garbage ("fuck-serged sale Sata
   dewlap"); 1720px -> 4810 chars; 2400px -> 8382 chars, readable. Downscaling to save
   OCR time destroys small UI text.

3. **`mpdecimate` is a weak gate on screen content.** On a *static* desktop recording it
   still kept 116 of 159 frames (73%) at defaults; loosened thresholds kept 125. Tuned
   for film grain, not for a cursor blinking. Dedupe on the OCR **text hash** instead.

4. Transcript spans: `r.text.runs.first?.audioTimeRange` gives only the first run
   (0.24 s for a 2.2 s sentence). Use `runs.first.start` -> `runs.last.end`.

5. Homebrew ffmpeg is built without libfreetype: **no `drawtext` filter**. Bites any
   burned-in overlay/annotation plan.

6. `swift file.swift` recompiles every run (~12 s). Compile once (`swiftc -O`, 5.9 s),
   then 0.27 s startup.

7. `screencapture -g` records the default **input device** only. System audio needs
   BlackHole 16ch + a Multi-Output Device (both already set up on this machine).

8. Capture must stay `/usr/sbin/screencapture` (Apple-signed, inherits the terminal's
   Screen Recording grant). A homegrown ScreenCaptureKit binary would be ad-hoc signed
   and lose its TCC grant on every rebuild — already recorded in CLAUDE.md.

## Cost model

10-minute screen recording, 1 fps, OCR at ~1720-2400px: ~600 frames x ~0.4 s = **~4 min**
of indexing, plus ~7 s for transcription. Transcription is free; OCR is the whole budget.
Levers: lower sample rate (0.5 fps), scene-gate before OCR, or OCR lazily on the window
the agent asks about.

## Scope verdicts

- **Perception** (frames + OCR + transcript + FTS5 index): viable, fully native, fast.
- **Verification / contracts**: `shasum`, `jq`, `sqlite3`, `curl` — trivially native. Thin
  `verify.sh` over a frozen contract file is enough; the 4-verdict taxonomy and subprocess
  sandboxing are upstream distrusting its own harness.
- **MCP server (39 tools), web workspace, DeepSeek bundle**: skip. Claude Code already has
  Bash and sqlite3; upstream needed 39 tools because its host could not shell out.

## Probe artifacts in this directory

`ocr.swift`/`ocrbin` (single-frame OCR), `ocrbatch.swift`/`ocrbatchbin` (parallel batch OCR),
`sp.swift` (locale probe), `tr.swift`/`trbin` (SpeechAnalyzer transcription with spans),
`rec.mov` (real 3 s screen recording), `cfr.mp4`/`vfr.mp4` (burned-counter alignment fixtures).
