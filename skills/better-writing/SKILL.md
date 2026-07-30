---
name: better-writing
description: >
  Review user-facing interface copy for clarity, consistency, actionability, localization safety, and appropriate tone. Use when the user asks about UX writing, microcopy, interface copy, button labels, error messages, empty states, placeholders, settings labels, or voice and tone.
  This skill owns domain review and is read-only by default; edit or implement
  only when explicitly requested. For full frontend construction use
  smart-frontend-design.
license: MIT
compatibility: "Requires bash and python3 for scripts/lint.sh; review guidance is framework-agnostic."
upstream:
  name: better-writing
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
  tags: [ux-writing, microcopy, content-design, localization, error-messages, design-review, humblskill]
  platforms: [claude-code, cursor, codex]
  preserve:
    - references/raw/
    - references/wiki/
    - references/decisions.md
    - references/log.md
    - references/patterns.md
---

# Better Writing

Review user-facing interface copy for clarity, consistency, actionability, localization safety, and appropriate tone. It complements `smart-frontend-design`: this skill owns
review evidence and domain-specific findings, while that skill owns holistic
frontend construction.

## Brain Protocol (read BEFORE reviewing)

1. `references/_index.md` - current concept map
2. `references/patterns.md` - measured outcomes
3. `references/decisions.md` - prior scope decisions
4. Last 5 entries in `references/log.md`
5. Relevant `references/wiki/design/writing/` concepts

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
  `smart-frontend-design` and use this skill as the writing review owner.

## How to Use

Read `references/_index.md`, then load only the concepts relevant to the request:

- `Voice And Tone`: `references/wiki/design/writing/voice-and-tone.md`
- `Actions And Navigation`: `references/wiki/design/writing/actions-and-navigation.md`
- `Errors And Empty States`: `references/wiki/design/writing/errors-and-empty-states.md`
- `Localization Safe Copy`: `references/wiki/design/writing/localization-safe-copy.md`

For complete upstream detail, read the cited file under `references/raw/`.

## Motion Boundary

Motion details outside this domain remain owned by `better-ui`; dedicated animation design and implementation remain owned by `use-smart-animation`.

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
