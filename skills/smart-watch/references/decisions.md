# Decisions

Reasoning memory. Each entry records a non-obvious choice: the context,
the options considered, what was chosen, why, and the observed result.
Never delete entries - if a decision is reversed, add a new entry that
references the old one.

Entry shape:

```
### <YYYY-MM-DD> | <short title>
- Context: <the situation that required a choice>
- Options: (A) <opt>, (B) <opt>, (C) <opt>
- Chose: <letter and name>
- Why: <the rationale, ideally citing evidence>
- Result: <what happened after, or "TBD">
```

---

### 2026-09-10 | Native macOS stack instead of porting the Python engine
- Context: upstream watch-skill needs Python 3.11+, Node 22+, Whisper, tesseract and a 39-tool MCP server. The user wanted native/internal tooling.
- Options: (A) pip-install upstream and wrap it, (B) rewrite the perception pipeline on screencapture + ffmpeg + Vision + SpeechAnalyzer + sqlite3 FTS5, (C) perception only, drop verification.
- Chose: B.
- Why: every upstream component has a measured native equivalent on this machine (see raw/probe-findings-2026-09-10.md): Vision OCR 0.38 s/frame batched, SpeechAnalyzer 85x realtime on-device, FTS5 ships in /usr/bin/sqlite3. The 39 MCP tools exist because upstream's host cannot shell out; Claude Code can. Verification is the cheap half (shasum/jq/sqlite3/curl), so dropping it saved nothing.
- Result: zero Python/Node; only Homebrew ffmpeg is required. selftest.sh passes end to end in ~6 s.

### 2026-09-10 | Declare tool deps in `metadata.dependencies`, not `metadata.requires`
- Context: the user asked for a frontmatter field listing tools that doctor.sh may install with approval. The agentskills spec has `compatibility` (prose, 500 chars) and humblSKILLS has `metadata.requires`.
- Options: (A) `metadata.requires: [ffmpeg, jq]`, (B) prose only in `compatibility`, (C) new `metadata.dependencies` list of flow mappings plus a prose `compatibility` summary.
- Chose: C.
- Why: `requires` is skill-to-skill dependency resolution, validated against KnownSkills in cli/internal/frontmatter/validate.go:195 - `ffmpeg` there fails `make registry` with `unknown dep`. Prose alone cannot drive a detector. Unknown metadata keys are tolerated by the parser (frontmatter_test.go:159) and ignored by the registry, so a structured list costs nothing and doctor.sh reads it directly, keeping one source of truth.
- Result: doctor.sh renders a table from the block and never drifts from SKILL.md.

### 2026-09-10 | `select` + `-fps_mode passthrough` over `-vf fps=N` for frame sampling
- Context: screencapture output is VFR; sampled frames need true timestamps for ask.sh hits to be trustworthy.
- Options: (A) `fps=1` + `-frame_pts 1`, (B) `select='isnan(prev_selected_t)+gte(t-prev_selected_t,1)'` + passthrough + showinfo, (C) extract every frame and pick in shell.
- Chose: B.
- Why: A retimes onto a grid; a frame named pts=1 showed a burned-in clock of 00:00:01.467 (0.47 s off). B reported pts_time 0, 1, 2.1, 3.1, 4.2 matching the burned clock exactly across 20 frames. C decodes and writes 30x more JPEGs for nothing.
- Result: selftest.sh asserts alignment on every indexed frame; passes with misaligned=0.

### 2026-09-10 | Dedupe on OCR text hash, not mpdecimate; cap width at 2560, never lower
- Context: OCR is the entire time budget (~0.4 s/frame). Two obvious savings were on the table.
- Options: (A) `mpdecimate` to drop near-identical frames before OCR, (B) downscale frames to ~1280 px, (C) OCR at up to 2560 px and skip index rows whose text equals the previous frame's.
- Chose: C.
- Why: mpdecimate kept 116/159 frames (73%) of a static desktop recording - it is tuned for film grain, not a blinking cursor. Downscaling to 1280 px turned the same frames into 1768 chars of garbage vs 8382 readable chars at 2400 px. The user cares about UI text, so resolution is the deliverable; `--fps 0.5` is the budget lever.
- Result: wiki/perception/ocr/vision-resolution.md records the yield table.

### 2026-09-10 | unicode61 tokenizer, no porter stemming
- Context: first FTS5 index used `porter unicode61`.
- Options: (A) keep porter, (B) `unicode61 remove_diacritics 2` and rely on `prefix*`, (C) trigram tokenizer for substring search.
- Chose: B.
- Why: with porter, "deployment" indexed as "deploy" while the query "deploy" stemmed to "deploi" - zero hits for the most natural query. Trigram needs 3+ char terms and would drop short OCR tokens like "30". Whole-word tokens plus `deploy*` are predictable and documented in ask.sh.
- Result: `ask.sh <id> "deploy*"` matches "deployment"; `8842` matches exactly.

### 2026-09-10 | Compiled Swift helper cached in ~/.cache, keyed by source hash
- Context: Vision and SpeechAnalyzer have no CLI; `swift file.swift` recompiles every run (~12 s).
- Options: (A) `swift` script each run, (B) compile into the skill dir, (C) `swiftc -O` into `~/.cache/smart-watch/bin`, rebuild when `shasum` of the source changes.
- Chose: C.
- Why: A costs 12 s per call. B is wiped by `humblskills update` unless listed in `preserve`, and a preserved stale binary would silently run old code after a skill update. C survives updates and self-invalidates. TCC is not a concern - neither Vision nor SpeechAnalyzer needs a grant, so the ad-hoc signature changing on rebuild is harmless (unlike screen capture, which stays on Apple's binary).
- Result: 6 s one-time build, 0.27 s startup afterwards.

### 2026-09-10 | Add forget.sh after the first live recording captured a secret
- Context: the first `--record` test captured display 1, which was showing an `.env.local` with a live service key; its OCR text landed in plain text in the index.
- Options: (A) leave cleanup to `rm -rf` + manual SQL, (B) `forget.sh <id>` that removes files and rows, (C) redact secrets automatically during OCR.
- Chose: B.
- Why: recordings will keep capturing whatever is on screen; a one-command purge is the minimum. Automatic redaction (C) is a heuristic that would create false confidence.
- Result: the offending source was purged; SKILL.md and the capture concept warn about plain-text storage.

### 2026-09-10 | Brief is the deliverable; change detection on OCR words, not lines or pixels
- Context: v0.1 returned an id and left the agent to query. "Drop a video, get the context" needs a digest. First attempt compared exact OCR lines between frames.
- Options: (A) exact-line Jaccard, (B) word-set Jaccard with a jitter-calibrated threshold, (C) pixel diff (ffmpeg scdet/mpdecimate).
- Chose: B.
- Why: A read a static desktop as "changed 92%" because Vision re-reads the same pixels slightly differently ("To Automation:" vs "Automation."). C is blind to text and kept 73% of static frames earlier. Measured word similarity: static screen 0.75-0.96, streaming terminal 0.44-0.56; threshold 0.5 separates them. `+`/`-` lines require >= 2 novel words so jitter does not surface.
- Result: brief.sh emits errors, pins, changes, transcript, contact sheet; watch.sh prints it. python3 (Command Line Tools) accepted as a native dependency for it.

### 2026-09-10 | Detail level is a question to the user, with `auto` by duration as the default
- Context: OCR cost is linear in sampled frames (~0.4 s each). A 1 s cadence on an hour-long video is 24 min; a 10 s cadence on a 2 min clip misses things.
- Options: (A) fixed 1 fps, (B) always ask, (C) `--detail auto|fine|normal|coarse` where auto maps duration to 1/3/5/10 s, and SKILL.md tells the agent to ask once, recommending auto, alongside "any timestamps to pin?".
- Chose: C.
- Why: the user asked for exactly this choice plus the ability to name timestamps. `--at` pins give precision where it matters (exact frame +/-2 s) without paying for it everywhere; pins can be added later without re-indexing (`.done` marker distinguishes a finished index from a half-built dir, which is redone).
- Result: `watch.sh --detail auto --at 1:23` in the default flow; selftest covers pins, pins-only re-run, and out-of-range pins.
