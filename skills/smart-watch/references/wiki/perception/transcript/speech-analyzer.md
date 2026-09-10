---
title: "On-Device Transcription: SpeechAnalyzer First, whisper-cpp Fallback"
context: perception
category: transcript
concept: speech-analyzer
description: "macOS 26 SpeechAnalyzer transcribes 85x realtime with word timings and no install; older macOS uses brew whisper-cpp with approval"
tags: speech, transcription, SpeechAnalyzer, SpeechTranscriber, whisper-cpp, macos-26, audio
sources:
  - "references/raw/probe-findings-2026-09-10.md"
last_ingested: 2026-09-10
command: scripts/watch.sh
---

## Two Engines, One TSV

Audio is extracted once as 16 kHz mono WAV (`ffmpeg -vn -ac 1 -ar 16000`), the format both engines
want. `watch.sh` picks the engine from `sw_vers`:

| macOS | Engine | Install | Measured |
| --- | --- | --- | --- |
| 26+ | `watchkit transcribe` -> `SpeechAnalyzer` + `SpeechTranscriber` | none (en_* models preinstalled; other locales download once via `AssetInventory`) | 70 s of audio in 0.82 s = 85x realtime, on-device |
| < 26 | `whisper-cli -m ggml-base.en.bin -ocsv` | `brew install whisper-cpp` + ~150 MB model, **only after the user approves** via `doctor.sh --install` | not measured here |
| < 26, no whisper | none | - | frames + OCR only, warning logged |

Both paths end in `transcript.tsv`: `start<TAB>end<TAB>text` in seconds.

**Incorrect (spans cover one word):**

```swift
let range = result.text.runs.first?.audioTimeRange   // 0.24 s for a 2.2 s sentence
```

**Correct:**

```swift
let ranges = result.text.runs.compactMap { $0.audioTimeRange }
let start = ranges.first?.start.seconds, end = ranges.last?.end.seconds
```

Build notes for `scripts/swift/watchkit.swift`:
- `SpeechAnalyzer` only exists in the macOS 26 SDK. The file guards it with `#if compiler(>=6.2)`
  (Swift 6.2 shipped with Xcode 26) so the same source still builds OCR-only on older toolchains.
- Must compile with `-parse-as-library` because of `@main`; `lib.sh::ensure_watchkit` does this and
  caches the binary in `~/.cache/smart-watch/bin` keyed by source hash (~6 s build, 0.27 s startup).
- "checking speech assets..." on stderr is normal; the request returns quickly when assets exist.

## Sources

- `references/raw/probe-findings-2026-09-10.md` - realtime factor, span bug, locale probe.
