---
title: "HTML Animation: Native Elements That Animate"
context: lang
category: html
concept: animation
description: "Motion from semantic HTML: Popover API, <details>, dialog, and @starting-style hooks — accessible, JS-light interaction states."
tags: html, popover, details, dialog, semantic, starting-style
sources:
  - "references/raw/web-animation-standards.md"
last_ingested: 2026-07-23
---

## HTML Animation: Native Elements That Animate

In a plain-HTML codebase, prefer native interactive elements — they bring
built-in accessibility (focus management, ARIA, keyboard, backdrop) and animate
with a few lines of CSS. Writing a custom `<div>` widget throws all of that
away. The motion itself lives in CSS (see `lang/css/animation`); this file is
about picking the right markup to hang it on.

**Incorrect (div soup — no a11y, needs JS, manual everything):**

```html
<div class="popover-trigger" onclick="togglePopover()">Menu</div>
<div class="popover" id="menu" style="display:none">…</div>
<!-- No focus handling, no Esc-to-close, no ARIA, needs JS to toggle. -->
```

**Correct (native Popover API — accessible, animatable, near-zero JS):**

```html
<button popovertarget="menu">Menu</button>
<div id="menu" popover>…</div>
```

```css
[popover] {
  opacity: 0; transform: translateY(-6px);
  transition: opacity .15s, transform .15s, overlay .15s allow-discrete,
              display .15s allow-discrete;
}
[popover]:popover-open { opacity: 1; transform: translateY(0); }
@starting-style { [popover]:popover-open { opacity: 0; transform: translateY(-6px); } }
@media (prefers-reduced-motion: reduce) { [popover] { transition: none; } }
```

Native elements to reach for:
- **`popover` attribute + `popovertarget`** — tooltips, menus, dropdowns; free
  light-dismiss, focus, and top-layer stacking.
- **`<dialog>`** with `showModal()` — modals with a real backdrop
  (`::backdrop` is animatable) and focus trapping.
- **`<details>/<summary>`** — accordions/disclosures; animate with
  `interpolate-size: allow-keywords` + `transition` on `height` from 0 to `auto`.
- **`@starting-style`** — the entry state for anything appearing in the top
  layer or from `display:none`.

Keep markup semantic first; add `data-*` state hooks only when a native element
genuinely doesn't fit.

## Sources

- `references/raw/web-animation-standards.md` — @starting-style, allow-discrete,
  and the native-element animation model.
