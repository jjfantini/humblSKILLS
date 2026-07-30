---
name: better-interface
description: >
  Orchestrate a holistic interface review across better-accessibility,
  better-layout, better-writing, better-typography, better-colors, and better-ui.
  Use when explicitly asked for better-interface, a full interface review, a
  holistic UI audit, or a cross-discipline design review. Read-only by default;
  implement only when explicitly requested.
license: MIT
compatibility: "Requires the six named domain skills at runtime for complete coverage, plus bash and python3 for scripts/lint.sh."
upstream:
  name: better-interface
  source: jakubkrehel/skills
  url: https://github.com/jakubkrehel/skills
  fetch: https://raw.githubusercontent.com/jakubkrehel/skills/main/skills/better-interface/SKILL.md
  license: MIT
  preserved: references/raw/
  synced: 2026-07-29
  deltas:
    - "adds progressive-disclosure router, wiki distillation, brain, and lint (upstream files preserved verbatim under references/raw/)"
metadata:
  author: jjfantini
  version: "1.0.0"
  category: design
  tags: [interface-review, design-review, accessibility, layout, ux-writing, typography, color, ui, humblskill]
  platforms: [claude-code, cursor, codex]
  preserve:
    - references/raw/
    - references/wiki/
    - references/decisions.md
    - references/log.md
    - references/patterns.md
---

# Better Interface

Orchestrate one evidence-backed interface review. This skill contains no domain
rules; the six named skills remain the sources of truth.

## Brain Protocol (read BEFORE reviewing)

1. `references/_index.md`
2. `references/patterns.md`
3. `references/decisions.md`
4. Last 5 entries in `references/log.md`
5. Relevant `references/wiki/design/orchestration/` concepts

After work, append to `references/log.md`; record measured outcomes or
non-obvious orchestration choices when present; run `bash scripts/lint.sh`
after wiki changes. Keep `references/raw/` unchanged.

## Resolve Mode

- `full` is the default: review the requested scope and report at most 15 findings.
- `quick`: inspect all available domains, report only `HIGH` and `MEDIUM`, cap at 5.
- Stay read-only unless the user explicitly requests implementation.

## Required Owner Order

1. `better-accessibility`
2. `better-layout`
3. `better-writing`
4. `better-typography`
5. `better-colors`
6. `better-ui`

Load each owner by that exact skill name. Never recreate, duplicate, or override
its rules. If a runtime owner is missing, mark its domain `Not reviewed`, name
the missing skill, and continue.

## Consolidation Contract

- One root cause becomes one finding, owned by one domain.
- Require exact source or rendered evidence; verification gaps are not findings.
- Rank by severity, reach, and leverage.
- Include `LOW` only in full mode.
- Record real candidates considered but rejected.
- End with `Block`, `Needs changes`, or `Approve` based on remaining findings
  and verified coverage.

## How to Use

- `Owner Routing`: `references/wiki/design/orchestration/owner-routing.md`
- `Mode And Caps`: `references/wiki/design/orchestration/mode-and-caps.md`
- `Evidence And Verdict`: `references/wiki/design/orchestration/evidence-and-verdict.md`

Read `references/raw/SKILL.md` for the unchanged upstream orchestration source.

## Implementation Boundary

This is only an orchestrator. If implementation was requested, use the
consolidated report as change scope, invoke the relevant owners, complement
frontend construction with `smart-frontend-design`, and re-run verification.

## Success Signals

- Owners run in the required order.
- Quick mode has at most 5 HIGH/MEDIUM findings.
- Full mode defaults automatically and has at most 15 findings.
- Missing owners are marked `Not reviewed` without stopping other reviews.
- No domain rules are duplicated in this skill.
- `bash scripts/lint.sh` exits 0 and `SKILL.md` stays under 200 lines.
