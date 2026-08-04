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

[INGEST 2026-08-03] Scaffolded smart-handoff via scripts/scaffold.sh.
  - Directory layout created: references/{wiki,raw}/, brain meta files, templates
  - Awaiting first raw material and wiki concepts

[QUERY 2026-08-03] Authored smart-handoff end to end.
  - 6 wiki concepts under new `handoff/` context: capture/session-context,
    document/{structure,suggested-skills}, security/secret-handling,
    lifecycle/temp-vs-persist, targets/harness-tailoring
  - scripts/preflight.sh (read-only snapshot: paths, git, instruction files,
    installed skills), scripts/scan-secrets.sh (masked FAIL/WARN sweep),
    scripts/lint.sh (copied from smart-commit, skill-agnostic)
  - assets/handoff-template.md: 12-section skeleton
  - tests/run.sh: 22 assertions, all passing under /bin/bash 3.2
  - decisions.md: 3 entries (live-context default, bash masking over sed,
    EXAMPLE stays safe-listed)

[LINT 2026-08-03] 6 wiki, 0 raw. Hard: 0, Soft: 6. Regenerated _index.md.

[LINT 2026-08-03] 6 wiki, 0 raw. Hard: 0, Soft: 6. Regenerated _index.md.
