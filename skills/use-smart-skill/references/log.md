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

[INGEST 2026-04-16] 1.0.0 initial release of use-smart-skill.
  - Smart Skill pattern (CCCCC: Core, Context, Category, Concept, Command)
  - Brain subsystem: raw/ + wiki/ + index/patterns/decisions/log
  - scripts/scaffold.sh and scripts/lint.sh

[LINT 2026-04-16] 10 wiki, 0 raw. Hard: 0, Soft: 12. Regenerated index.md.

[LINT 2026-04-17] 10 wiki, 0 raw. Hard: 0, Soft: 12. Regenerated index.md.

[QUERY 2026-04-17] Renamed required wiki frontmatter key `impactDescription` to `description` (lint.sh REQUIRED list, `_template.md`, `_brain.md`, all wiki concepts). Lint OK.

[INGEST 2026-04-17] Captured agentskills.io SKILL.md frontmatter spec.
  - Raw: references/raw/agentskills-spec.md
  - New concept: wiki/smart/spec/skill-frontmatter.md (new spec/ category)
  - Propagated compatibility field: SKILL.md, scaffold.sh template, workflow.md, validation-checklist.md, migrate/workflow.md
  - Decision: allowed-tools skipped (experimental, not applicable to meta-skill) - see decisions.md 2026-04-17 entry

[LINT 2026-04-17] 11 wiki, 1 raw. Hard: 0, Soft: 12. Regenerated index.md.

[LINT 2026-04-17] 11 wiki, 1 raw. Hard: 0, Soft: 12. Regenerated _index.md.
