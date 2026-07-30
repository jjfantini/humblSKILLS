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

[INGEST 2026-07-27] Scaffolded create-smart-html via scripts/scaffold.sh.
  - Directory layout created: references/{wiki,raw}/, brain meta files, templates
  - Awaiting first raw material and wiki concepts

[LINT 2026-07-27] 5 wiki, 0 raw. Hard: 0, Soft: 5. Regenerated _index.md.

[SESSION 2026-07-27] Authored the skill's initial concept set.
  - 5 wiki concepts: workflow/intake/context-gathering, html/design/{single-file,typography}, pdf/export/{print-css,headless-chrome}
  - scripts/export-pdf.sh (headless Chrome, US Letter, portrait/landscape override via injected @page) + lint.sh
  - Verified end-to-end on macOS: portrait MediaBox 612x792, landscape 792x612, JS content rendered, temp copy cleaned up

[LINT 2026-07-30] 5 wiki, 0 raw. Hard: 0, Soft: 5. Regenerated _index.md.

[LINT 2026-07-30] 5 wiki, 0 raw. Hard: 0, Soft: 5. Regenerated _index.md.
