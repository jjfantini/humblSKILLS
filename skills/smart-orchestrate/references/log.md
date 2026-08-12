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

[INGEST 2026-08-12] Scaffolded smart-orchestrate via scripts/scaffold.sh.
  - Directory layout created: references/{wiki,raw}/, brain meta files, templates
  - Awaiting first raw material and wiki concepts

[INGEST 2026-08-12] Migrated the flat `orchestrate` skill into smart-orchestrate.
  - Raw: references/raw/orchestrate-SKILL.md (163-line flat SKILL.md, verbatim)
  - New context `orchestrate/` with 7 categories and 9 concepts:
    - roles/{parent-orchestrator, worker-agent}
    - isolation/worktree-first
    - routing/model-selection
    - loop/session-loop
    - contracts/{brief-template, handoff-contract}
    - closeout/commit-and-ship
    - anti-patterns/avoid
  - All 9 concepts cite the raw file in `sources:`
  - SKILL.md rewritten as a router: Brain Protocol, CCCCC table, When/How to Use,
    2 Examples, Troubleshooting, Success Signals
  - Preserved upstream `disable-model-invocation: true`; see decisions.md
  - humblSKILLS extension fields nested under `metadata:` per repo convention

[LINT 2026-08-12] 9 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.
