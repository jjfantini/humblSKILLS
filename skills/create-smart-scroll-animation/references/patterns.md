# Patterns

Performance memory. Each entry records a concrete attempt, its numeric
outcome, and the lesson. Read before every session; append after every
session where quantified results appear.

Entry shape (see `wiki/brain/patterns/how-to-log-results.md` for the full
worked example):

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

### 2026-04-20 | Initial-release dry run against synthesized 3s test MP4
- Context: validate probe/extract/to-webp pipeline end-to-end before first commit. Input was a lavfi-synthesized 3-second 30fps 640×360 H.264 test video.
- Approach: `probe.sh` → `extract-frames.sh (target=30)` → `to-webp.sh (q=80, cwebp)`.
- Result: probe returned valid JSON (duration=3.0, fps=30.0, total_frames=90). extract-frames computed target_fps=10 and emitted exactly 30 PNGs. to-webp produced 30 WebPs totaling 0.47MB (avg 15.9KB/frame).
- Worked: the `target_fps = round(target_count / duration)` formula landed on the budget exactly (30 → 30). cwebp was on PATH and the primary encoder path ran.
- Didn't: the test resolution was 640×360, not 1920×1080, so the per-frame size (15.9KB) is artificially small. Real-world hero videos at 1920×1080 will be closer to the 40KB/frame target — validate with a real asset during first production use.
- Lesson: the pipeline's math is correct; the size budget needs a real-world test at full resolution before we can claim the <4MB-for-100-frames number with confidence. Flag for first production session.
