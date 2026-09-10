---
name: smart-watch
description: >
  Drop a screen recording, video, or audio file and get a timestamped brief -
  errors on screen, UI/screen changes, pinned moments, transcript, and a
  contact sheet - built with native macOS tooling (ffmpeg frames, Vision OCR,
  on-device SpeechAnalyzer, sqlite FTS5). Then answer "what happened at 3:12?"
  from the index, and prove finished work with a frozen verification contract
  (file hashes, JSON values, SQL results, HTTP responses, exit codes) that
  returns VERIFIED/FAILED, not opinion.
  Use when the user says "watch this recording", "index this video", "what
  happened at", "record my screen for N seconds", "transcribe this screen
  recording", "find where X appears in the video", "verify the work",
  "write a contract for this", or "prove it". Do NOT use for producing or
  editing video (use create-smart-video-transition), for storing files (use
  smart-upload-thing), or for live browser interaction (use the Chrome tools).
license: MIT
compatibility: >
  macOS 13+ with Xcode Command Line Tools (swiftc) and Homebrew ffmpeg. Native:
  /usr/sbin/screencapture, Vision OCR, sqlite3 FTS5, shasum, curl. macOS 26+
  transcribes on-device via SpeechAnalyzer; older macOS falls back to brew
  whisper-cpp. Optional: jq (json_path checks), yt-dlp (URL sources).
  scripts/doctor.sh reports what is missing and installs only with approval.
allowed-tools: "Bash(bash:*) Bash(ffmpeg:*) Bash(ffprobe:*) Bash(sqlite3:*) Bash(screencapture:*) Bash(open:*) Read Write Glob Grep"
metadata:
  author: jjfantini
  version: "0.2.0"
  category: development
  tags: [video, screen-recording, ocr, transcription, evidence, verification, ffmpeg, vision, speech, humblskill]
  platforms: [claude-code, cursor, codex]
  dependencies:
    - { name: ffmpeg,        kind: brew,   required: true,  provides: "frame + audio extraction (also ffprobe)" }
    - { name: swiftc,        kind: native, required: true,  provides: "builds the OCR/transcribe helper", install: "xcode-select --install" }
    - { name: screencapture, kind: native, required: true,  provides: "--record live screen capture" }
    - { name: sqlite3,       kind: native, required: true,  provides: "FTS5 evidence index" }
    - { name: python3,       kind: native, required: true,  provides: "brief.sh digest (ships with Command Line Tools)" }
    - { name: jq,            kind: brew,   required: false, provides: "verify.sh json_path checks and contract parsing" }
    - { name: yt-dlp,        kind: brew,   required: false, provides: "URL sources for watch.sh" }
    - { name: whisper-cpp,   kind: brew,   required: false, provides: "transcription on macOS < 26", when: "macos_lt_26" }
  preserve:
    - references/raw/
    - references/wiki/
    - references/decisions.md
    - references/log.md
    - references/patterns.md
---

# Smart Watch

Perception and proof for agents, built from what macOS already ships. Drop a
recording: `watch.sh` indexes frames + OCR + transcript with true timestamps and
prints a brief (errors, screen changes, pins, transcript, contact sheet);
`ask.sh` searches it; `verify.sh` evaluates a frozen contract and returns a
verdict. No server - the agent is the query engine.

## Dependencies and approval

Every external tool is declared in `metadata.dependencies` above and detected by
`scripts/doctor.sh`. **Nothing is installed silently.** On first use run the
doctor; if anything is `MISSING`, tell the user which tools, what each one
provides, and the exact `brew install` line, and only after they agree run
`bash scripts/doctor.sh --install --yes`. Required today: Homebrew `ffmpeg`
and the Xcode Command Line Tools. Everything else is optional or native.

## Brain Protocol (read BEFORE creating anything)

1. `references/_index.md`       - what this skill knows (map)
2. `references/patterns.md`     - what worked, with numbers
3. `references/decisions.md`    - past reasoning, don't repeat mistakes
4. `references/log.md`          - last 5 session entries
5. Relevant `references/wiki/<context>/<category>/` concepts per task

After completing work, UPDATE the brain:
- New user-provided material lives in `references/raw/` (LLM never renames or edits it)
- Distilled insights -> new/updated `references/wiki/<context>/<category>/<concept>.md`
  - Cite every raw file you used in the `sources:` frontmatter array
- Performance data (if reported) -> `patterns.md`
- Non-obvious decisions -> `decisions.md`
- Session summary (always) -> append to `log.md`
- Run `scripts/lint.sh` to regenerate `_index.md` and verify structure

No data, no improvement. Full spec: `references/_brain.md`.

## When to Use

- The user hands over a screen recording, meeting video, or audio file and wants to know what happened, when
- The agent needs evidence of its own work: `--record N` captures the screen while it drives an app
- Searching for a word, error string, or spoken phrase across one or many recordings
- Proving a task is done with checks a model cannot rationalise: hashes, JSON values, SQL counts, HTTP status, exit codes

## How to Use

**Ask two things before indexing** (one AskUserQuestion, recommended defaults first):

1. Detail level - `auto` (recommended: <=3 min 1 s, <=15 min 3 s, <=45 min 5 s, else 10 s), `fine` (every 1 s, fine-tooth comb, ~0.4 s OCR per frame so a 10-min video costs ~4 min), `normal` (3 s), `coarse` (10 s, long recordings or "just the gist").
2. Timestamps to pin - moments the user already knows matter (`1:23, 4:56`). Each gets an exact frame plus one 2 s before and after, never deduped, listed first in the brief.

Skip the question only when the user already said (e.g. "quick pass", "look at 2:10").

Then:

1. **Doctor** (once per machine): `bash scripts/doctor.sh`. Approval rule above.
2. **Watch**: `bash scripts/watch.sh <file|url> --detail auto --at 1:23,4:56` or `--record 30 [--audio]`. Prints `brief.md` then the source id. Read the brief - it is the deliverable: errors/warnings with frame paths, pinned frames, screen changes (+new / -gone text), transcript, and `sheet.jpg` (change-point frames tiled in time order - `Read` it to see the whole recording at once).
3. **Dig**: `bash scripts/ask.sh <id|all> "<query>"` (FTS5; natural questions are rewritten automatically; `deploy*` for prefixes; `--timeline --from S --to S` for a slice). `bash scripts/frame.sh <id> <hh:mm:ss>` returns the nearest frame to `Read`. More pins later: `watch.sh <file> --at 7:10` adds them without re-indexing.
4. **Verify** (when proving work): write a contract *before* the work (`assets/contract.example.json`), run `bash scripts/verify.sh contract.json` after. Exit 0 only when every check is VERIFIED. `UNVERIFIED` = could not run; never report it as a pass.
5. **Clean up** when a recording captured something sensitive: `bash scripts/forget.sh <id>` (everything is stored in plain text under `~/.local/state/smart-watch`).

Concepts, one per file:
- Capture: `references/wiki/perception/capture/screencapture.md`
- Frame timing: `references/wiki/perception/frames/pts-preserving-sampling.md`
- OCR resolution and dedupe: `references/wiki/perception/ocr/vision-resolution.md`
- Transcription engines: `references/wiki/perception/transcript/speech-analyzer.md`
- Index schema and queries: `references/wiki/perception/index/fts5-schema.md`
- Brief: change detection and error patterns: `references/wiki/perception/brief/change-points.md`
- Contract schema and verdicts: `references/wiki/verify/contracts/schema.md`
- Dependency policy: `references/wiki/setup/deps/native-first.md`

## Examples

### Example 1: Drop a recording, get the context

User says: "Here's a recording of the bug ~/Desktop/repro.mov, figure out what went wrong"

Actions:
1. Ask: detail `auto` (2 min video -> every 1 s) and any timestamps to pin? User: "around 1:20"
2. `bash scripts/watch.sh ~/Desktop/repro.mov --at 1:20` -> brief printed, id `a1b2c3d4e5f6`
3. Read `## Errors and warnings` (first hit `00:01:18.00 TypeError: cannot read properties of undefined`), the pinned frames at 1:18/1:20/1:22, and `Read` `sheet.jpg` for the flow
4. `bash scripts/frame.sh a1b2c3d4e5f6 1:18` and `Read` the JPEG to see the stack trace

Result: "The error appears at 00:01:18 right after clicking Save (screen change at 00:01:17: form -> spinner). Stack trace names `orders.map` in `OrderList.tsx`; the transcript at 01:15 says 'and now I hit save'."

### Example 2: Record own work, then prove it

User says: "Fix the failing test, record yourself doing it, and prove it passes"

Actions:
1. Write `contract.json` with a `command` check (`npm test`, `expect_exit: 0`)
2. `bash scripts/watch.sh --record 120 --no-brief &` then do the work
3. `bash scripts/verify.sh contract.json`
4. `bash scripts/brief.sh <id>` and `bash scripts/ask.sh <id> "PASS"` for the timestamp the test went green

Result: verdict table with VERIFIED, plus a timestamped frame showing the passing run.

## Troubleshooting

**`--record` produces no file / macOS asks for Screen Recording permission**
Cause: first use of `screencapture -v` from this terminal.
Fix: grant Screen Recording to the terminal app in System Settings > Privacy & Security, re-run. Do not replace screencapture with a custom binary - it would lose the grant on every rebuild.

**Timestamps are off by up to a second**
Cause: someone replaced the `select` filter with `-vf fps=N`, which retimes VFR input onto a grid.
Fix: keep `select='isnan(prev_selected_t)+gte(t-prev_selected_t,INTERVAL)'` with `-fps_mode passthrough`. See `perception/frames/pts-preserving-sampling.md`.

**OCR text is garbage**
Cause: frames were downscaled below ~1720 px wide.
Fix: `watch.sh` only caps width at 2560; do not add a smaller `scale`. See `perception/ocr/vision-resolution.md`.

**`downloading speech assets…` on first transcribe**
Cause: SpeechAnalyzer fetches the locale model once per machine. Harmless.

**Screen changes lists every frame as changed / lists nothing**
Cause: `brief.sh --threshold` (default 0.5 on OCR word-set similarity; a static screen re-OCRs at 0.75-0.96).
Fix: lower it for fewer, bigger changes; raise toward 0.7 for subtle ones. Re-run `bash scripts/brief.sh <id> --threshold 0.4`.

**Video shows as `video=0` / no frames**
Cause: the only video stream is an attached cover image (mp3/m4a), or ffprobe could not read the file.
Fix: audio-only sources still transcribe; for a real video check `ffprobe <file>`.

**Verdict `UNVERIFIED`**
Cause: the check could not execute (jq/sqlite3 missing, file absent, host down). It is not FAILED and not VERIFIED.
Fix: resolve the tooling or reachability, re-run. Never report UNVERIFIED as a pass.

## Success Signals

- `bash scripts/selftest.sh` passes (14 checks): PTS alignment, transcript, search, pins, brief sections, contact sheet, query rewrite, verify verdicts
- Every `ask.sh` hit has a timestamp that matches the frame when opened
- `verify.sh` exits 0 only on all-VERIFIED
- `scripts/lint.sh` exits 0; `log.md` grows by one entry per session
