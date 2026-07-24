# transitions.dev (the "transitions" library / skill)

Source: https://transitions.dev/ , https://transitions.dev/skill.html ,
https://github.com/Jakubantalik/transitions.dev
Author: Jakub Antalik. License: MIT. Researched 2026-07-23.

## What it is
A collection of ~27 production-ready, **framework-agnostic CSS** UI transitions
(copy-paste), plus a Pro tier and an installable agent skill. The free tier is
copy-paste CSS + an agent skill, NOT an npm import library.

## Install (verbatim from skill.html)
- Main skill (27+ transitions):  `npx skills add Jakubantalik/transitions.dev`
- Polish add-on skill:           `npx skills add Jakubantalik/transitions.dev -s transitions-polish`
- Pro (subscription):            `npx transitions-pro skill`
- Pro CLI helpers:               `npx transitions-pro add card-resize` (free),
  `npx transitions-pro login`, `npx transitions-pro add --pro`,
  `npx transitions-refine live`
- Related npm packages (CLIs, not the transition code): `transitions-pro`,
  `transitions-refine`.

## How the skill is structured (what we fold in)
Repo path `skills/transitions-dev/`:
- `SKILL.md` — name `transitions-dev`, description "Production-ready CSS
  transitions for web apps."
- Per-transition markdown files (e.g. `01-card-resize.md` … `18-texts-reveal.md`;
  27 total advertised).
- `_root.css` — the universal install block with shared motion tokens.

Each per-transition file: (1) usage guidance / when to use, (2) HTML markup with
state classes (e.g. `<div class="t-modal" role="dialog">` toggled with
`.is-open` / `.is-closing`), (3) tunable CSS custom properties, (4) CSS
implementation with default/open/closing states **plus a mandatory
`@media (prefers-reduced-motion: reduce)` block**, (5) JS orchestration
(`openModal()`/`closeModal()` toggling classes with `setTimeout` cleanup
timing). All selectors namespaced `t-*`.

## Transition catalog (27)
card resize, number pop-in, notification badge, text states swap, menu
dropdown, modal open/close, panel reveal, page side-by-side, icon swap, success
check, avatar group hover, error state shake, input clear/dissolve, skeleton
loader & reveal, shimmer text, tabs sliding, tooltip open/close, texts reveal,
card hover tilt, plus-to-menu morph, accordion expand, toast open/close, like
button, learn-more hover, checkbox check, spinning counter, toggle.

## Agent commands the skill adds
- `transitions reveal`  — list all transitions (numbered)
- `transitions review`  — audit project for ad-hoc animations, suggest replacements
- `transitions apply`   — auto-detect + install best-fit transition
- `transitions refine`  — find hardcoded durations, suggest motion tokens
- `transitions polish`  — (polish add-on) layered rules for open/close, hover, stagger
Works with Claude Code, Cursor, GitHub Copilot, Codex, Gemini CLI.

## Motion tokens (`_root.css`) — the values to reuse
Durations:
```
--duration-stagger:   40ms
--duration-micro:      80ms
--duration-quick:     150ms
--duration-fast:      250ms
--duration-medium:    350ms
--duration-slow:      400ms
--duration-very-slow: 500ms
```
Easings:
```
--ease-smooth-out:   cubic-bezier(0.22, 1, 0.36, 1)   /* modals/panels */
--ease-in-out
--ease-out
--ease-linear                                         /* spinners */
--ease-bounce
--ease-bounce-strong
```
Distances: `--distance-micro: 4px` … `--distance-large: 30px`.
Scales: `--scale-tiny: 0.99` … `--scale-large: 0.96`.
Blur: `--blur-small: 2px` … `--blur-large: 8px`.
Plus per-component token groups (badge, dropdown, modal, panel, page, icon,
check, avatar, shake, skeleton, shimmer, tabs, tooltip, tilt, morph, accordion,
toast, like, checkbox, counter, toggle).

## Accessibility guidance the skill hammers on
- Always preserve the `@media (prefers-reduced-motion: reduce)` block — every
  snippet ships one; removing it fails a11y audits.
- Common mistakes it warns against: stripping close-state cleanup, forgetting
  reflow timing, hardcoding `stroke-dasharray`, mixing error classes.

## Fit
Best for: standard UI micro-transitions (modals, toasts, tabs, accordions) in
ANY stack — it is plain CSS. Two integration paths: install the upstream skill
(`npx skills add Jakubantalik/transitions.dev`), or copy the `t-*` pattern +
`_root.css` tokens directly.

## Verification note
Live sites are JS-rendered; snippets above are reproduced from the GitHub
READMEs/skill files. Prop tables, package names, and install commands are
reliable; for pixel-exact CSS pull the raw files from
`raw.githubusercontent.com/Jakubantalik/transitions.dev/main/skills/transitions-dev/`.
