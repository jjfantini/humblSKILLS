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

[INGEST 2026-05-15] Authored use-smart-commit skill.
  - Workflow: inspect pending changes -> bucket by intent -> conventional commit per bucket
  - 3 wiki concepts seeded: workflow/atomic-grouping, messages/conventional-format, anti-patterns/avoid
  - SKILL.md embeds 2 worked examples + DO-NOT list (including skip-CI token rule)

[LINT 2026-05-15] 3 wiki, 0 raw. Hard: 0, Soft: 3. Regenerated _index.md.

[LINT 2026-05-15] 3 wiki, 0 raw. Hard: 0, Soft: 3. Regenerated _index.md.

[LINT 2026-05-15] 3 wiki, 0 raw. Hard: 0, Soft: 3. Regenerated _index.md.
