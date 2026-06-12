# Log

Append-only session log. Every session MUST append at least one entry. Never
edit old entries - they are the historical record. Most recent entries appear
at the bottom.

Entry shape:

```markdown
[INGEST|QUERY|LINT <YYYY-MM-DD>] <one-line summary>
  <optional indented detail line(s)>
```

---

[INGEST 2026-06-12] Scaffolded use-worktree-flow.
  - Added brain meta files, raw user request, workflow wiki concepts, and scripts
  - Initial defaults: worktree isolation, Vibe mode on deferral, cleanup enabled

[LINT 2026-06-12] 6 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.
