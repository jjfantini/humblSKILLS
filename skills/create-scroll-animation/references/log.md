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

[INGEST 2026-04-20] Scaffolded create-scroll-animation via scripts/scaffold.sh.
  - Directory layout created: references/{wiki,raw}/, brain meta files, templates
  - Awaiting first raw material and wiki concepts

[LINT 2026-04-20] 3 wiki, 4 raw. Hard: 0, Soft: 2. Regenerated _index.md.

[LINT 2026-04-20] 15 wiki, 4 raw. Hard: 0, Soft: 7. Regenerated _index.md.

[INGEST 2026-04-20] 0.1.0 initial release of create-scroll-animation.
  - 15 wiki concepts across 5 contexts: video (3), react (5), nextjs (2), mobile (3), workflow (2)
  - 4 seed raw sources: scroll-stop-builder snapshot, CSS-Tricks Apple technique, Framer Motion docs, iOS Safari memory cap
  - 4 scripts: lint.sh (verbatim from use-smart-skill), probe.sh, extract-frames.sh, to-webp.sh
  - Template: assets/templates/ScrollFrameCanvas.tsx.tmpl — single-file React component with 10 placeholders + 2 conditional blocks
  - Examples: basic-usage.tsx, nextjs-app-router-full.tsx
  - Synthesis concepts (empty sources): 7 — mobile/network, mobile/touch, nextjs/integration (both), react/a11y, react/component/loader-ui, workflow/interview. Acceptable at launch; will collect sources as usage patterns emerge.
  - Architectural decisions: two-phase interview (technical Phase 1 mandatory, aesthetic Phase 2 optional); single-file .tsx output; local public/frames/ default (no hosting adapters in v0.1.0). See decisions.md.
  - Dropped from reference skill: starscape, annotation/snap-stop cards, count-ups, navbar pill, white-first-frame requirement.

[LINT 2026-04-20] 15 wiki, 4 raw. Hard: 0, Soft: 7. Regenerated _index.md.

[LINT 2026-04-20] 15 wiki, 4 raw. Hard: 0, Soft: 7. Regenerated _index.md.
