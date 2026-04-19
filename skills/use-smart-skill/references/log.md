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

[INGEST 2026-04-19] Anthropic "Complete Guide to Building Skills for Claude" PDF.
  - Raw: references/raw/anthropic-skill-building-guide.pdf (561 KB)
  - New context: `anthropic/` with 6 categories and 8 concepts
    - anthropic/frontmatter/{requirements, security}
    - anthropic/description/trigger-design
    - anthropic/structure/{progressive-disclosure, file-layout}
    - anthropic/testing/three-layer-approach
    - anthropic/patterns/five-patterns
    - anthropic/troubleshooting/common-failures
  - All 8 concepts cite the PDF in their `sources:` array
  - decisions.md: added 2026-04-19 entry for context choice + rationale

[QUERY 2026-04-19] Rewrote SKILL.md to Anthropic-compliant shape:
  - humblSKILLS extension fields (version/tags/platforms/preserve) moved into metadata:
  - Added allowed-tools, license, compatibility at top level
  - Reordered sections: Brain Protocol -> Brain Ops -> CCCCC -> When -> How -> Examples -> Troubleshooting -> Success Signals
  - Added Examples (2 concrete user-said blocks) and Success Signals
  - Added Troubleshooting section focused on lint drift and trigger debugging
  - Description rewritten per anthropic/description/trigger-design.md: WHAT + WHEN + trigger phrases + negative trigger + CCCCC architecture note
  - Routed Anthropic-sourced guidance under "How to Use" section
  - decisions.md: 2 new entries (allowed-tools reversal, metadata migration)

[QUERY 2026-04-19] Rewrote scripts/scaffold.sh SKILL_MD_CONTENT template:
  - Frontmatter now Anthropic-compliant (humblSKILLS extensions under metadata:)
  - Added Examples section with 2 TODO:START/END blocks (required, 2 minimum)
  - Added Troubleshooting section with TODO:START/END (optional, delete-if-N/A)
  - Added Success Signals section with seeded bullets
  - Updated "Next steps" trailer to flag every new TODO block

[QUERY 2026-04-19] CLI parser refactored for metadata-nested frontmatter:
  - cli/internal/frontmatter/Frontmatter grew a Metadata sub-struct
  - Accessors (Version, Requires, Platforms, Tags, Preserve) fall back to legacy top-level keys
  - DeprecationWarnings() surfaces unmigrated fields
  - cli/cmd/build-registry prints warnings for unmigrated skills (non-fatal)
  - cli/internal/install/preserve.setPreserveKey writes into metadata: sub-mapping and removes legacy top-level preserve
  - All tests pass; registry.json regen emits zero warnings after all 5 in-repo skills migrated

[LINT 2026-04-19] 19 wiki, 2 raw. Hard: 0, Soft: 12. Regenerated _index.md.

[LINT 2026-04-19] 23 wiki, 2 raw. Hard: 0, Soft: 16. Regenerated _index.md.
