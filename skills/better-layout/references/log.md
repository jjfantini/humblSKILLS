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

[INGEST 2026-07-29] Scaffolded better-layout via scripts/scaffold.sh.
  - Directory layout created: references/{wiki,raw}/, brain meta files, templates
  - Awaiting first raw material and wiki concepts

[INGEST 2026-07-29] Ported better-layout from jakubkrehel/skills.
  - Preserved upstream Markdown unchanged in references/raw/
  - Added progressive-disclosure wiki concepts, brain files, attribution, and lint

[LINT 2026-07-29] 3 wiki, 3 raw. Hard: 0, Soft: 0. Regenerated _index.md.
