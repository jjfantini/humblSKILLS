---
title: "Headless Chrome PDF Export"
context: pdf
category: export
concept: headless-chrome
description: "The exact chrome --headless --print-to-pdf invocation, binary discovery across macOS/Linux, JS settling with virtual-time-budget, and why header/footer are disabled"
tags: chrome, headless, print-to-pdf, export, cli
sources: []
last_ingested: 2026-07-27
command: scripts/export-pdf.sh
---

## Chrome is the renderer

The PDF must come from the same engine that rendered the review-in-browser
version — headless Google Chrome — so screen sign-off predicts print output.
No wkhtmltopdf, no screenshot-to-PDF, no converter libraries.

### The invocation (wrapped by `scripts/export-pdf.sh`)

```bash
"$CHROME_BIN" \
  --headless \
  --disable-gpu \
  --no-pdf-header-footer \
  --virtual-time-budget=8000 \
  --print-to-pdf="out.pdf" \
  "file:///absolute/path/page.html"
```

- `--headless` — modern headless mode (Chrome 112+; old builds accept
  `--headless=new`).
- `--no-pdf-header-footer` — suppresses Chrome's date/URL/title furniture;
  the document owns its own header/footer as content.
- `--virtual-time-budget=8000` — fast-forwards up to 8s of virtual time so
  inline JS (expanding sections, drawing SVG) settles before printing;
  pages with no JS print immediately.
- Paper size and orientation come from the file's `@page` rule
  (`pdf/export/print-css.md`) — headless CLI has no paper-size flag, which
  is exactly why the contract lives in CSS.
- Use an absolute `file://` URL; relative paths silently 404 inside
  headless Chrome and yield an empty PDF.

### Binary discovery order (what the script does)

1. `$CHROME_BIN` if set
2. macOS: `/Applications/Google Chrome.app/Contents/MacOS/Google Chrome`
   (then Chromium at the same pattern)
3. PATH: `google-chrome`, `google-chrome-stable`, `chromium`,
   `chromium-browser`, `chrome`

### Incorrect

```bash
wkhtmltopdf page.html out.pdf          # different engine, different layout
chrome --headless --screenshot ...     # raster, not selectable/searchable
```

### Correct

```bash
bash scripts/export-pdf.sh page.html landscape out.pdf
```

One command; orientation override injected without touching the source file.

## Sources

- (none) — authored from skill design; flags verified against current Chrome headless.

## Command

```bash
bash scripts/export-pdf.sh <file.html> [portrait|landscape] [out.pdf]
```
