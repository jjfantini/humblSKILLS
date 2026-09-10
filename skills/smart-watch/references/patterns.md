# Patterns

Performance memory. Each entry records a concrete attempt, its numeric
outcome, and the lesson. Read before every session; append after every
session where quantified results appear.

Entry shape:

```
### <YYYY-MM-DD> | <short title>
- Context: <what was attempted, in one line>
- Approach: <the method used>
- Result: <metrics, numbers, outcomes>
- Worked: <what helped>
- Didn't: <what hurt>
- Lesson: <the rule to apply next time>
```

---

### 2026-09-10 | Vision OCR throughput on a 3440x1440 screen recording
- Context: measure OCR cost per frame to build the indexing cost model.
- Approach: compiled Swift helper, VNRecognizeTextRequest .accurate, DispatchQueue.concurrentPerform over JPEG frames scaled to 1720-2560 px wide.
- Result: 0.55 s per frame single; 0.38 s per frame batched (40 frames in 15.3 s wall at 323% CPU). Yield by width from the same 3 frames: 1280 px = 1768 chars (garbage), 1720 px = 4810 chars, 2400 px = 8382 chars readable.
- Worked: batching across cores; keeping width at or above 1720 px.
- Didn't: downscaling to 1280 px; `swift file.swift` (12 s recompile per run).
- Lesson: budget ~0.4 s per sampled frame (10 min at 1 fps is about 4 min); reduce fps, never resolution.

### 2026-09-10 | SpeechAnalyzer transcription speed
- Context: replace Whisper with the macOS 26 on-device engine.
- Approach: SpeechTranscriber with .audioTimeRange, 16 kHz mono WAV via ffmpeg, spans from first run start to last run end.
- Result: 8.75 s clip in 0.77 s; 70 s clip in 0.82 s (85x realtime). Spoken "eight eight four two" transcribed as "8842". en_* locale assets were preinstalled; the asset check adds under 0.2 s when present.
- Worked: analyzeSequence + finalizeAndFinish; TSV output needs no JSON parser downstream.
- Didn't: `runs.first?.audioTimeRange` alone - 0.24 s span for a 2.2 s sentence.
- Lesson: transcription is effectively free; OCR is the whole budget.

### 2026-09-10 | Frame sampling alignment
- Context: verify sampled-frame timestamps against a burned-in clock (testsrc2) on a VFR clip.
- Approach: `select='isnan(prev_selected_t)+gte(t-prev_selected_t,1)'` + `-fps_mode passthrough` + showinfo pts_time, versus `fps=1` + `-frame_pts 1`.
- Result: select/passthrough matched the burned clock on 20/20 frames (pts_time 0, 1, 2.1, 3.1, 4.2 ...); fps=1 was 0.47 s off on frame 1. mpdecimate kept 116/159 frames of a static recording.
- Worked: showinfo at info loglevel; pairing Nth frame with Nth pts_time line.
- Didn't: `-loglevel error` (swallows showinfo); filename pts (timebase units 1/15360).
- Lesson: never use the fps filter for evidence frames; selftest.sh guards this.

### 2026-09-10 | End-to-end selftest and live recording
- Context: first full run of watch.sh, ask.sh, frame.sh, verify.sh on generated fixtures and a real 4 s capture.
- Approach: selftest.sh generates a 12 s VFR testsrc2 clip muxed with `say` speech into a throwaway XDG_STATE_HOME.
- Result: 9/9 assertions in ~6 s wall (watchkit already built). Live `--record 4` on a 3440x1440 display: 4 frames, 16,064 OCR chars, 9.1 s total including the 4 s capture.
- Worked: throwaway state dir via XDG_STATE_HOME; fixtures generated locally so nothing binary is committed.
- Didn't: `-shortest` when muxing (cut the video to the 5 s speech clip, dropping frames below the assert threshold); `set -e` inherited from lib.sh aborting on a failing `$( )` assignment.
- Lesson: keep the selftest fixture longer than the speech track; capture a failing command's exit with `|| rc=$?`.

### 2026-09-10 | OCR jitter on a static screen vs a changing one
- Context: calibrate the screen-change threshold for brief.sh.
- Approach: 4 s h264 loop of one 3440x1440 screenshot at 1 fps; word-set Jaccard between consecutive OCR results; compared with a 3 s recording of a terminal streaming text and with exact-line Jaccard.
- Result: static screen word similarity 0.75, 0.92, 0.96 (first gap is the keyframe); exact-line similarity on the same frames 0.09-0.22; streaming terminal 0.44-0.56 words.
- Worked: word tokens of >= 3 chars, lowercase; threshold 0.5.
- Didn't: exact line matching (every frame "changed"); the case-insensitive `E[A-Z]{3,}` code pattern (matched "Enterprise", "exercises").
- Lesson: compare OCR output as bags of words, never as lines; keep error-code regexes case-sensitive.
