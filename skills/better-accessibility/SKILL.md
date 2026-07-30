---
name: better-accessibility
description: >
  Review product interfaces for keyboard, semantic, form, screen-reader, target-size, motion, and zoom accessibility. Use when the user asks about accessibility, a11y, WCAG, ARIA, keyboard navigation, focus, forms, screen readers, hit areas, reduced motion, or zoom behavior.
  This skill owns domain review and is read-only by default; edit or implement
  only when explicitly requested. For full frontend construction use
  smart-frontend-design.
license: MIT
compatibility: "Requires bash and python3 for scripts/lint.sh; review guidance is framework-agnostic."
upstream:
  name: better-accessibility
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
  tags: [accessibility, a11y, wcag, aria, keyboard, screen-readers, design-review, humblskill]
  platforms: [claude-code, cursor, codex]
  preserve:
    - references/raw/
    - references/wiki/
    - references/decisions.md
    - references/log.md
    - references/patterns.md
---

# Better Accessibility

Review product interfaces for keyboard, semantic, form, screen-reader, target-size, motion, and zoom accessibility. It complements `smart-frontend-design`: this skill owns
review evidence and domain-specific findings, while that skill owns holistic
frontend construction.

## Brain Protocol (read BEFORE reviewing)

1. `references/_index.md` - current concept map
2. `references/patterns.md` - measured outcomes
3. `references/decisions.md` - prior scope decisions
4. Last 5 entries in `references/log.md`
5. Relevant `references/wiki/design/accessibility/` concepts

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
  `smart-frontend-design` and use this skill as the accessibility review owner.

## How to Use

Read `references/_index.md`, then load only the concepts relevant to the request:

- `Focus And Keyboard`: `references/wiki/design/accessibility/focus-and-keyboard.md`
- `Semantics And Aria`: `references/wiki/design/accessibility/semantics-and-aria.md`
- `Forms`: `references/wiki/design/accessibility/forms.md`
- `Screen Readers`: `references/wiki/design/accessibility/screen-readers.md`
- `Hit Areas`: `references/wiki/design/accessibility/hit-areas.md`
- `Motion And Zoom`: `references/wiki/design/accessibility/motion-and-zoom.md`

For complete upstream detail, read the cited file under `references/raw/`.

## Motion Boundary

Motion details outside this domain remain owned by `better-ui`; dedicated animation design and implementation remain owned by `smart-animation`.

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
