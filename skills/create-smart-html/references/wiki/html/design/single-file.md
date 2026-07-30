---
title: "Zero-Dependency Single-File HTML"
context: html
category: design
concept: single-file
description: "Rules that keep the output one portable file: inline everything, no network references, semantic skeleton, interactivity that degrades to static print"
tags: single-file, zero-dependency, portable, inline, offline
sources: []
last_ingested: 2026-07-27
---

## One file, works anywhere

The deliverable must open from `file://`, from an email attachment, or from
an air-gapped machine and look identical. That means the file carries
everything it needs and references nothing it doesn't.

### Rules

- **All CSS in one `<style>` block** in `<head>`. No `<link rel="stylesheet">`.
- **All JS in one `<script>` block** before `</body>`. No `src=` scripts,
  no import maps, no CDNs — vanilla JS only, and only where it earns its place.
- **No external assets.** Images are inline SVG or `data:` URIs; icons are
  inline SVG paths, not icon fonts. Charts are hand-built inline SVG —
  no charting libraries.
- **No web fonts by default.** Use system font stacks
  (`html/design/typography.md`). If the user insists on a specific face,
  embed it as a `data:` URI `@font-face` and warn about file size.
- **Semantic skeleton**: `header/main/section/footer`, one `h1`, heading
  levels in order — screen readers and Chrome's PDF outline both benefit.
- **Modern CSS, no build step**: custom properties for the palette, CSS
  grid/flex for layout, `prefers-color-scheme` for dark mode on screen
  (print always renders the light scheme).
- **Interactivity degrades**: anything collapsible/sortable/hoverable must
  render fully expanded and readable with JS disabled — that's what the
  print pipeline sees.

### Verification (run before delivering)

```bash
grep -nE 'https?://' page.html   # expect: no asset/script/style hits (bare text links in content are fine)
```

Open the file directly (`open page.html`) — not through a dev server — and
confirm it renders complete.

### Incorrect

```html
<link href="https://fonts.googleapis.com/css2?family=Inter" rel="stylesheet">
<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
```

Breaks offline, breaks air-gapped, and the PDF export races the network.

### Correct

```html
<style>
  :root { --ink: #1a1d23; --accent: #2563eb; }
  body { font-family: system-ui, -apple-system, "Segoe UI", Roboto, sans-serif; }
</style>
<svg viewBox="0 0 480 200" role="img" aria-label="Q3 revenue by region">…</svg>
```

## Sources

- (none) — authored from skill design; add raw sources as real pages are built.
