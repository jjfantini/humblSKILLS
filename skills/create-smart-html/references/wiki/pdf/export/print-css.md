---
title: "US Letter Print CSS: @page, Breaks, Exact Color"
context: pdf
category: export
concept: print-css
description: "The print stylesheet contract every generated file ships: @page size letter portrait/landscape, page-break control, print-color-adjust, and print-mode overrides for interactive elements"
tags: print, us-letter, page, landscape, portrait, breaks, color
sources: []
last_ingested: 2026-07-27
---

## The page IS the spec

Chrome's PDF export honors CSS `@page`. That makes the HTML file itself the
single source of truth for paper size and orientation — the export script
only overrides orientation when asked.

### The contract (ships in every generated file)

```css
/* Orientation decided at intake — `letter` or `letter landscape` */
@page {
  size: letter;          /* US Letter: 8.5in x 11in */
  margin: 0.6in 0.7in;
}

@media print {
  html { -webkit-print-color-adjust: exact; print-color-adjust: exact; }

  /* Interactive chrome disappears; content it controlled renders expanded */
  .no-print, nav.toolbar, button { display: none !important; }
  details { display: block; }
  details > * { display: block; }

  /* Break control */
  h1, h2, h3 { break-after: avoid; }
  table, figure, .card { break-inside: avoid; }
  p { orphans: 3; widows: 3; }
  thead { display: table-header-group; }  /* repeat header on each page */
}
```

### Orientation

- Portrait: `size: letter;` — prose, stacked sections, resumes.
- Landscape: `size: letter landscape;` — wide tables, timelines, dashboards.
- Bake the intake decision into the file; `scripts/export-pdf.sh <file>
  portrait|landscape` can override at export time by injecting a trailing
  `@page` rule, so one HTML file can produce both PDFs.

### Design within the page budget

Content must be composed for the printable area (~7.1in x 9.8in portrait
at the margins above). Full-bleed backgrounds, fixed headers, and
100vh hero sections are screen furniture — cap them or hide them in print.
For a "one-pager", verify the PDF is actually one page and tighten spacing
at the print breakpoint if not.

### Incorrect

Relying on the export tool's paper flags while the CSS says nothing —
different Chrome versions and margins produce different layouts, and the
on-screen review no longer predicts the PDF.

### Correct

`@page` + `@media print` in the file (above), so opening the browser's
print preview shows exactly what `export-pdf.sh` will emit.

## Sources

- (none) — authored from skill design; verified against Chrome headless print-to-pdf behavior.
