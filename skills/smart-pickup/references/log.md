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

[INGEST 2026-10-14] Scaffolded smart-pickup via scripts/scaffold.sh.
  - Directory layout created: references/{wiki,raw}/, brain meta files, templates
  - Awaiting first raw material and wiki concepts

[LINT 2026-10-14] 4 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[QUERY 2026-10-14] Completed configurable pickup with read-only discovery, external personal settings, and agent-mediated Notion setup; 15 offline tests and public-package gate pass.

[LINT 2026-10-14] 4 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[QUERY 2026-10-14] Live Notion creation, schema/query, isolated configuration, and pickup loading verified; main config unchanged. Strict starting-folder handoff integration passes 19 pickup tests and 22 handoff checks.

[LINT 2026-10-14] 4 wiki, 2 raw. Hard: 0, Soft: 0. Regenerated _index.md.
