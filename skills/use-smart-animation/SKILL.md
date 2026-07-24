---
name: use-smart-animation
description: >
  Professional, performance-first frontend animation and transition guidance
  that routes by the codebase you are in (HTML, CSS, JavaScript, TypeScript,
  React) and by four ready-made libraries (thinking-orbs, border-beam, metal-fx,
  transitions.dev). Use when adding animations, motion, transitions, page or
  route transitions, scroll reveals, hover micro-interactions, loading/thinking
  indicators, or when the user says "animate this", "add motion", "make it feel
  premium/polished", "page transition", "reveal on scroll", or asks which
  animation library to use. Enforces compositor-only performance and a
  prefers-reduced-motion floor, and favors distinctive, non-generic motion.
  Do NOT use for non-frontend work or for static visual design with no motion.
license: MIT
compatibility: Requires bash, POSIX utilities (awk, sed, find, grep), python3, writable filesystem at the skill target path.
metadata:
  author: jjfantini
  version: "0.1.0"
  category: design
  tags: [animation, transitions, frontend, motion, css, react, performance, accessibility, humblskill]
  platforms: [claude-code, cursor, codex]
  preserve:
    - references/raw/
    - references/wiki/
    - references/decisions.md
    - references/log.md
    - references/patterns.md
---

# Use Smart Animation

Give any frontend the right motion for its stack: read the shared principles,
then load only the file for the language you're in and the library you're
reaching for. Motion is deliberate here — one orchestrated moment beats
scattered effects, everything animates on compositor-only properties, and every
snippet honors `prefers-reduced-motion`.

## Brain Protocol (read BEFORE animating anything)

1. `references/_index.md`    - what this skill knows (map)
2. `references/patterns.md`  - what worked, with numbers
3. `references/decisions.md` - past reasoning, don't repeat mistakes
4. `references/log.md`       - last 5 session entries
5. The relevant `references/wiki/<context>/<category>/` concepts (see routing below)

After completing work, UPDATE the brain:
- New user-provided material -> `references/raw/` (never renamed or edited)
- Distilled insights -> new/updated `references/wiki/<ctx>/<cat>/<concept>.md`,
  citing every raw file used in `sources:`
- Quantified results (fps, bundle size, LCP) -> `patterns.md`
- Non-obvious decisions -> `decisions.md`
- Session summary (always) -> append to `log.md`
- Run `scripts/lint.sh` to regenerate `_index.md` and verify structure

_Full spec: `references/_brain.md`._

## CCCCC Architecture

| Layer        | Role                                | Location                                       |
|--------------|-------------------------------------|------------------------------------------------|
| **Core**     | Root structure of the skill         | `SKILL.md`, `references/`, `scripts/`          |
| **Context**  | Top-level grouping                   | 1st segment under `references/wiki/` (`motion`, `lang`, `lib`) |
| **Category** | Specific topic within a context      | 2nd segment (`principles`; `css`/`react`/…; `orbs`/`beams`/…) |
| **Concept**  | One atomic idea per file             | Filename stem + required frontmatter field     |
| **Command**  | Deterministic executable script      | `scripts/lint.sh`                              |

## When to Use

- Adding or reviewing animations/transitions in a web frontend
- Choosing a motion approach for a specific language (HTML/CSS/JS/TS/React)
- Wiring one of the four bundled libraries (orbs, border-beam, metal-fx, transitions.dev)
- Page/route transitions, scroll reveals, hover micro-interactions, loading states
- Fixing janky animation (dropped frames) or missing reduced-motion handling

## How to Use

**Always read first — the shared principles apply to every task:**
Read `references/wiki/motion/principles/design.md` (where/if to animate),
`references/wiki/motion/principles/performance.md` (compositor-only, 60fps),
and `references/wiki/motion/principles/accessibility.md` (reduced-motion floor + tokens).

**Then route by the codebase language:**

- **CSS-only / static site:** Read `references/wiki/lang/css/animation.md`
- **Plain HTML:** Read `references/wiki/lang/html/animation.md`
- **Vanilla JavaScript:** Read `references/wiki/lang/js/animation.md`
- **TypeScript (non-React):** Read `references/wiki/lang/ts/animation.md`
- **React / Next.js:** Read `references/wiki/lang/react/animation.md`

**And route by the effect / library you need:**

- **Thinking / loading indicator for an AI or agent UI:** Read `references/wiki/lib/orbs/usage.md` (thinking-orbs)
- **Animated glow around a card, button, or input border:** Read `references/wiki/lib/beams/usage.md` (border-beam)
- **Metallic / liquid-metal accent on one premium element:** Read `references/wiki/lib/metal/usage.md` (metal-fx)
- **Standard UI transitions (modals, toasts, tabs) in any stack:** Read `references/wiki/lib/transitions/usage.md` (transitions.dev)

**Live enumeration of everything this skill knows:**
Read `references/_index.md` (auto-regenerated by `scripts/lint.sh`).

**Brain protocol, naming, linking contract, lint checks, `patterns.md` shape:**
Read `references/_brain.md`. **Wiki concept file shape:** Read `references/_template.md`.

### Scripts

- `scripts/lint.sh` — brain health check: regenerates `_index.md`, validates
  every wiki file's path/frontmatter triple and that each cites a real
  `references/raw/` source. Run after editing any wiki concept.

## Examples

### Example 1: "Add a page-load animation to my React landing page"

1. Read the three `motion/principles/*` concepts (deliberate, performant, accessible).
2. Read `lang/react/animation.md` — decision rule says use CSS for the entrance,
   or Motion only if you need staggered orchestration/exit.
3. Implement ONE orchestrated entrance (staggered children), compositor-only
   (`transform`/`opacity`), with a `prefers-reduced-motion` fallback.
4. If they want a premium accent on the CTA, read `lib/metal/usage.md` and gate
   the shader with `paused` under reduced motion.
5. Append a `log.md` entry; if you measured LCP/fps, add a `patterns.md` entry.

### Example 2: "Make my modals and toasts feel polished in a plain HTML/CSS site"

1. Read the `motion/principles/*` concepts.
2. Read `lib/transitions/usage.md` — copy the `t-*` pattern (open + closing
   states, motion tokens) or install the upstream transitions.dev skill.
3. Read `lang/html/animation.md` — prefer `<dialog>`/Popover so the motion hangs
   on accessible native elements.
4. Verify each transition keeps its `@media (prefers-reduced-motion: reduce)` block.

## Troubleshooting

**Animation is janky / drops frames:** you're likely animating a layout
property. Read `motion/principles/performance.md` — move to `transform`/`opacity`
or FLIP.

**`scripts/lint.sh` exits 1 with "path/frontmatter mismatch":** a wiki file's
`context/category/concept` doesn't match its path. Rename the file or fix the
frontmatter, then re-run.

**`scripts/lint.sh` exits 1 with "broken source path":** a wiki concept's
`sources:` points at a missing `references/raw/` file. Fix the path or the raw file.

**Skill never triggers:** the `description:` lacks the phrases a user would type.
Ask "When would you use use-smart-animation?" and adjust until it matches.

## Success Signals

- `scripts/lint.sh` exits 0 after every wiki change (all triples valid, all sources resolve).
- Every animation uses `transform`/`opacity` (no layout-property tweens).
- Every motion-producing snippet has a `prefers-reduced-motion` fallback.
- Motion is concentrated on one signature moment, not scattered across the view.
- The right file loads for the stack: React work reads `lang/react`, a modal reads `lib/transitions`, etc.
