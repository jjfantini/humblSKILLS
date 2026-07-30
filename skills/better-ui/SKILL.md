---
name: better-ui
description: >
  Review interface polish across surfaces, icons, interaction states, animation details, and rendering performance. Use when the user asks about UI polish, visual details, border radius, shadows, icons, hover states, press feedback, micro-interactions, or motion restraint.
  This skill owns domain review and is read-only by default; edit or implement
  only when explicitly requested. For full frontend construction use
  smart-frontend-design.
license: MIT
compatibility: "Requires bash and python3 for scripts/lint.sh; review guidance is framework-agnostic."
upstream:
  name: better-ui
  source: jakubkrehel/skills
  url: https://github.com/jakubkrehel/skills
  license: MIT
  preserved: references/raw/
  synced: 2026-07-29
  deltas:
    - "adds progressive-disclosure router, wiki distillation, brain, and lint (upstream files preserved verbatim under references/raw/)"
metadata:
  author: jjfantini
  version: "1.0.0"
  category: design
  tags: [ui, visual-design, polish, icons, surfaces, micro-interactions, performance, design-review, humblskill]
  platforms: [claude-code, cursor, codex]
  preserve:
    - references/raw/
    - references/wiki/
    - references/decisions.md
    - references/log.md
    - references/patterns.md
---

# Better UI

Review interface polish across surfaces, icons, interaction states, animation details, and rendering performance. It complements `smart-frontend-design`: this skill owns
review evidence and domain-specific findings, while that skill owns holistic
frontend construction.

## Brain Protocol (read BEFORE reviewing)

1. `references/_index.md` - current concept map
2. `references/patterns.md` - measured outcomes
3. `references/decisions.md` - prior scope decisions
4. Last 5 entries in `references/log.md`
5. Relevant `references/wiki/design/ui/` concepts

After work, append a session entry to `references/log.md`; add measured outcomes
to `patterns.md`, non-obvious choices to `decisions.md`, and run
`bash scripts/lint.sh` after wiki changes. Keep `references/raw/` unchanged.

## Operating Contract

- Review is read-only unless the user explicitly requests implementation.
- Inspect the repository’s existing system before judging or proposing changes.
- Cite exact evidence and distinguish confirmed findings from unverified concerns.
- Consolidate repeated symptoms under one root cause.
- When orchestrated by `better-interface`, provide domain evidence and findings;
  let the orchestrator own severity, caps, consolidation, verdict, and output.
- For building or redesigning a complete frontend, route to
  `smart-frontend-design` and use this skill as the ui review owner.

## How to Use

Read `references/_index.md`, then load only the concepts relevant to the request:

- `Surfaces`: `references/wiki/design/ui/surfaces.md`
- `Icons`: `references/wiki/design/ui/icons.md`
- `Animations`: `references/wiki/design/ui/animations.md`
- `Performance`: `references/wiki/design/ui/performance.md`

For complete upstream detail, read the cited file under `references/raw/`.

## Motion Boundary

`better-ui` reviews whether motion details fit the interface and flags concrete polish or performance issues. Route dedicated animation design or implementation to `use-smart-animation`; do not duplicate its performance-first animation system.

## Review Result

For a standalone review, report prioritized findings with severity, exact
location, current evidence, actionable change, verification, and verdict.
Report no finding when evidence is insufficient. When implementation is
requested, preserve the report as scope and re-run relevant verification.

## Success Signals

- `bash scripts/lint.sh` exits 0.
- `SKILL.md` remains under 200 lines.
- Every wiki concept cites an unchanged upstream raw source.
- Review requests do not mutate code unless implementation was requested.
- Findings are evidence-backed and assigned to this domain’s owning rule.
