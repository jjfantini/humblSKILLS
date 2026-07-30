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

[INGEST 2026-06-12] Scaffolded smart-frontend-design via scripts/scaffold.sh.
  - Directory layout created: references/{wiki,raw}/, brain meta files, templates
  - Awaiting first raw material and wiki concepts

[LINT 2026-06-12] 7 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[LINT 2026-07-30] 10 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[RESYNC 2026-07-30] Re-synced against current upstream frontend-design; 1.0.3 -> 1.1.0.
  - Preserved copy replaced (45 -> 55 lines) and renamed user-frontend-design-brief.md
    -> frontend-design-SKILL.md; all 7 concepts repointed to the new sources path.
  - REWROTE anti-patterns/generic-ai-slop: three current clusters (cream+serif+terracotta,
    near-black+acid, broadsheet) replacing the stale Inter/purple-gradient tells, plus the
    brief's-words-win exception.
  - NEW: aesthetics/hero-thesis, aesthetics/structural-honesty, process/two-pass-plan.
  - UPDATED: direction/bold-aesthetic (one justifiable risk), discovery/synthesize-style
    (ground in the subject's world; use memory), implementation/production-code (build from
    the revised plan; CSS specificity gotcha), verification/review-checklist (critique while
    building, quality floor, remove one accessory).
  - Added upstream: block + derived ATTRIBUTION.md; router gained the 3 new concepts and
    handoff pointers to the animation skill (motion implementation) and better-writing (copy).
  - 2 decisions recorded: the stale-mirror re-sync, and upstream-block-as-source-of-truth.
  - Deferred: lint.sh generating ATTRIBUTION.md + staleness warning; an "update mirrors"
    command to re-sync every mirrored skill from its upstream.

[LINT 2026-07-30] 10 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[LINT 2026-07-30] 10 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[LINT 2026-07-30] 10 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.
