---
title: "Sleek Typography Without Web Fonts"
context: html
category: design
concept: typography
description: "System font stacks, scale, and spacing that read as designed — on screen and at print resolution — with zero downloads"
tags: typography, fonts, system-stack, scale, print
sources: []
last_ingested: 2026-07-27
---

## Designed type, zero downloads

Web fonts are the first thing zero-dependency output gives up — and modern
system stacks make that a non-loss. Every OS ships a good grotesque, a good
serif, and a good mono; the craft is in the scale and spacing, not the file.

### Stacks

```css
:root {
  --font-sans: system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  --font-serif: "Iowan Old Style", "Palatino Linotype", Palatino, Georgia, serif;
  --font-mono: ui-monospace, "SF Mono", "Cascadia Code", Menlo, Consolas, monospace;
}
```

Sans for UI/dashboards, serif for document-genre pages (reports, proposals)
where a print feel fits, mono for data/code. Pick per intake genre.

### Scale and rhythm

- A restrained modular scale, declared once:
  `--step--1/-0/+1/+2/+3` ≈ 0.8rem / 1rem / 1.25rem / 1.6rem / 2rem.
  For hybrid screen/print pages, prefer fixed rem steps over `clamp()` so
  the print rendering matches what was reviewed on screen.
- Body: 16px screen; print inherits ~12pt equivalent via the same rem scale
  — don't set a separate pt scale, let US Letter's 8.5in width do the work.
- `line-height: 1.5` body, `1.15` headings; `max-width: 70ch` on prose;
  `letter-spacing: -0.01em` on display sizes only.
- `font-variant-numeric: tabular-nums` on any column of figures — tables
  and invoices read as designed instead of wobbling.
- True quotes/dashes in content (curly quotes, en/em dashes); `&nbsp;`
  between values and units.

### Incorrect

Five font sizes picked ad hoc per section, centered long prose, fake
small-caps via `text-transform` + smaller px, and a Google Fonts `<link>`
"just for the headings".

### Correct

One stack + one scale in `:root`, weights doing the hierarchy work
(400/600/700), and hierarchy audited at both screen size and printed size
before delivery.

## Sources

- (none) — authored from skill design.
