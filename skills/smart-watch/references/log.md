# Log

Append-only session log. Every session MUST append at least one entry.
Never edit old entries - they are the historical record. Most recent
entries appear at the bottom.

Entry shape:

```
[INGEST|QUERY|LINT <YYYY-MM-DD>] <one-line summary>
  <optional indented detail line(s)>
```

---

[INGEST 2026-09-10] Scaffolded smart-watch via smart-skill scaffold.sh and built v0.1.0.
  - Probed native replacements for oxbshw/watch-skill on macOS 26.5; findings in raw/probe-findings-2026-09-10.md
  - Wrote scripts: lib.sh, doctor.sh, watch.sh, ask.sh, frame.sh, verify.sh, forget.sh, selftest.sh, swift/watchkit.swift
  - Seeded 7 wiki concepts (perception/{capture,frames,ocr,transcript,index}, verify/contracts, setup/deps), 7 decisions, 4 patterns
  - selftest.sh 9/9 green; live --record verified; first recording purged after it captured a secret (see decisions)

[LINT 2026-09-10] 7 wiki, 1 raw. Hard: 0, Soft: 1. Regenerated _index.md.

[LINT 2026-09-10] 8 wiki, 1 raw. Hard: 0, Soft: 1. Regenerated _index.md.

[INGEST 2026-09-10] v0.2.0 - brief.sh, --detail, --at pins, query rewrite, robustness pass.
  - Holes closed: half-built dir treated as indexed (.done marker), pins-only re-run, attached-pic "video" streams, FTS5 syntax errors on natural questions, pipefail in selftest, stale schema (pinned column migration)
  - brief.sh: errors/pins/screen-changes/transcript/contact sheet; word-set Jaccard calibrated on measured jitter
  - selftest 14/14; new wiki concept perception/brief/change-points; SKILL.md now asks detail level + pins before indexing

[LINT 2026-09-10] 8 wiki, 1 raw. Hard: 0, Soft: 1. Regenerated _index.md.
