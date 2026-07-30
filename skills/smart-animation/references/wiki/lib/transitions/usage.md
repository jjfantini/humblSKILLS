---
title: "transitions.dev: Framework-Agnostic CSS Transitions"
context: lib
category: transitions
concept: usage
description: "transitions.dev: 27 copy-paste CSS UI transitions for any stack. Two integration paths (install the skill or copy the t-* pattern) plus motion tokens."
tags: transitions-dev, css, framework-agnostic, modal, toast, motion-tokens
sources:
  - "references/raw/transitions-dev.md"
last_ingested: 2026-07-23
---

## transitions.dev: Framework-Agnostic CSS Transitions

transitions.dev is a set of ~27 production-ready, **plain-CSS** UI transitions
(modals, toasts, tabs, accordions, badges, checks, …). Because it's CSS, it
works in any stack — HTML, vanilla JS, React, anything. Author: Jakub Antalik,
MIT.

Two integration paths — pick one:

1. **Install the upstream agent skill** (pulls the full catalog + `_root.css`):
   ```
   npx skills add Jakubantalik/transitions.dev
   ```
   Then its agent commands are available: `transitions reveal` (list),
   `transitions apply` (auto-fit), `transitions review`/`refine` (audit
   ad-hoc animations and hardcoded durations).

2. **Copy the `t-*` pattern directly** — for a small need, lift one transition
   plus the shared motion tokens. This is the portable route with no extra tool.

**Incorrect (ad-hoc modal: hardcoded timings, no close state, no reduced-motion):**

```css
.modal { transition: all 0.3s; }         /* animates layout, magic number */
.modal.open { display: block; }          /* no exit animation, no cleanup */
```

**Correct (the `t-*` pattern: tokens, open + closing states, reduced-motion):**

```css
:root {
  --duration-fast: 250ms;
  --ease-smooth-out: cubic-bezier(0.22, 1, 0.36, 1);
}
.t-modal {
  opacity: 0; transform: scale(0.96);
  transition: opacity var(--duration-fast) var(--ease-smooth-out),
              transform var(--duration-fast) var(--ease-smooth-out);
}
.t-modal.is-open    { opacity: 1; transform: scale(1); }
.t-modal.is-closing { opacity: 0; transform: scale(0.96); }
@media (prefers-reduced-motion: reduce) {
  .t-modal, .t-modal.is-open, .t-modal.is-closing { transition: none; }
}
```

```js
const open  = m => m.classList.add('is-open');
const close = m => { m.classList.add('is-closing');
  setTimeout(() => m.classList.remove('is-open', 'is-closing'), 250); }; // match duration
```

Conventions: all selectors namespaced `t-*`; state via `.is-open`/`.is-closing`;
timing/distance/scale/blur pulled from `_root.css` custom properties; **every
snippet ships a `@media (prefers-reduced-motion: reduce)` block — keep it**.
Common mistakes it warns against: dropping the closing-state cleanup, forgetting
the reflow timing (the `setTimeout` must match the CSS duration), hardcoding
`stroke-dasharray`, mixing error classes.

Catalog (27): card resize, number pop-in, notification badge, text swap, menu
dropdown, modal, panel reveal, page side-by-side, icon swap, success check,
avatar hover, error shake, input clear, skeleton, shimmer text, sliding tabs,
tooltip, texts reveal, card tilt, plus-to-menu morph, accordion, toast, like
button, learn-more hover, checkbox, spinning counter, toggle.

Motion tokens to reuse live in `motion/principles/accessibility`.

## Sources

- `references/raw/transitions-dev.md` — install paths, the `t-*` pattern,
  `_root.css` tokens, catalog, agent commands, and accessibility rules.
