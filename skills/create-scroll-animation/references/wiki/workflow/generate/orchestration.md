---
title: "End-to-End Orchestration: Interview to Written Component"
context: workflow
category: generate
concept: orchestration
description: "The canonical recipe: interview → probe → extract → webp → read template + placeholders → render → write .tsx → log. Claude runs every step itself. The user provides the MP4 and gets a ready-to-import component file back."
tags: orchestration, workflow, pipeline, end-to-end
sources:
  - "references/raw/scroll-stop-builder-SKILL.md"
last_ingested: 2026-04-20
---

## Orchestration Recipe

Claude runs all ten steps. Do not hand the user a "here's how to run
this" instruction list — run it.

### Step 1 — Interview

Read `wiki/workflow/interview/phase-order.md`. Run Phase 1. Decide on
Phase 2. Collect:

- `VIDEO_PATH` (absolute path, validated to exist)
- `OUTPUT_DIR` (e.g., `public/frames/`)
- `TARGET_FRAME_COUNT` (default 100)
- Optionally: `BRAND_BG`, `BRAND_ACCENT`

### Step 2 — Probe

```bash
bash scripts/probe.sh "$VIDEO_PATH"
```

Parse the JSON. Extract `duration`, `fps`, `width`, `height`.

### Step 3 — Decide `TARGET_FPS` (done inside extract-frames.sh)

```
target_fps = round(TARGET_FRAME_COUNT / duration)
# clamp to source fps to avoid frame duplication
target_fps = min(target_fps, source_fps)
```

### Step 4 — Extract frames

```bash
bash scripts/extract-frames.sh "$VIDEO_PATH" /tmp/frames-raw "$TARGET_FRAME_COUNT"
```

Parse the emitted `manifest.json` to confirm the actual frame count (may
differ from target by ±1 due to rounding).

### Step 5 — Convert to WebP

```bash
bash scripts/to-webp.sh /tmp/frames-raw "$OUTPUT_DIR" 80
```

Read the size report. If total >8MB, drop `-q` to 70 and re-run. If
still >8MB, recommend resolution downshift in the extraction step.

### Step 6 — Read the wiki concepts

Before generating code, re-read:

- `wiki/react/component/scroll-frame-canvas.md` (core anatomy)
- `wiki/react/component/preload-strategy.md` (semaphore)
- `wiki/react/component/loader-ui.md` (if loader included)
- `wiki/react/component/responsive-fit.md` (fit math)
- `wiki/react/a11y/reduced-motion.md` (if fallback included)
- `wiki/nextjs/integration/app-router-setup.md` (client boundary)

### Step 7 — Read the template

```
assets/templates/ScrollFrameCanvas.tsx.tmpl
```

### Step 8 — Substitute placeholders

Text replacements:

| Placeholder                | Value                                    |
|----------------------------|------------------------------------------|
| `{{COMPONENT_NAME}}`       | `ScrollFrameCanvas` (or user override)   |
| `{{FRAME_COUNT}}`          | actual frame count from manifest         |
| `{{FRAME_URL_PATTERN}}`    | e.g. `/frames/frame_{index}.webp`        |
| `{{FRAME_DIGITS}}`         | `4`                                      |
| `{{SCROLL_HEIGHT_VH}}`     | `350` (desktop-default)                  |
| `{{CRITICAL_PRELOAD_COUNT}}` | `10`                                   |
| `{{BRAND_BG}}`             | from Phase 2, or `#000000`               |
| `{{BRAND_ACCENT}}`         | from Phase 2, or `#ffffff`               |

Conditional block removal:

- If `INCLUDE_REDUCED_MOTION_FALLBACK = false`: delete both
  `// ==== BEGIN INCLUDE_REDUCED_MOTION_FALLBACK` and
  `// ==== END INCLUDE_REDUCED_MOTION_FALLBACK` lines AND everything
  between them. Default: `true`.
- If `INCLUDE_LOADER_UI = false`: same treatment for all the
  `INCLUDE_LOADER_UI` markers. There are three pairs in the template
  (state, progress update, JSX). Delete all three. Default: `true`.

After substitution, verify:

- No `{{...}}` remains in the file.
- No `// ==== BEGIN/END` comment markers remain.
- `'use client'` is on line 1 (after substitution).

### Step 9 — Write the component

Default target path: `components/ScrollFrameCanvas.tsx` (or
`app/_components/ScrollFrameCanvas.tsx` for App Router users).
Ask the user if the default is OK or pick their own.

```
Write <absolute-path-to-target>/ScrollFrameCanvas.tsx
```

Confirm the file count (frames, component, manifest) in a short message
to the user:

> "Done. Wrote components/ScrollFrameCanvas.tsx, 100 WebP frames
> (3.8MB total) to public/frames/. Import and render <ScrollFrameCanvas />
> wherever you want the hero. Add framer-motion: `npm i framer-motion`."

### Step 10 — Append to `log.md`

Format:

```
[INGEST 2026-04-20] Generated ScrollFrameCanvas from hero.mp4.
  - 100 frames @ 1920x1080 WebP q=80, total 3.8 MB
  - Phase 2 interview skipped (no brand chrome requested)
  - Target: components/ScrollFrameCanvas.tsx
```

If any metrics emerged (Lighthouse scores if the user tested, wall-clock
times, unusual frame counts) add a `patterns.md` entry.

---

## Variations

### User already has frames extracted

Skip steps 2–5. Confirm the URL pattern they want and jump to step 6.

### User wants mobile frame set too

Run steps 4–5 twice:
- Once at 1920 wide → `public/frames/`
- Once at 1280 wide → `public/frames-mobile/`

Then the component consumer code picks the pattern at runtime (see
`mobile/memory/ios-safari-cap.md`).

### User's project isn't Next.js

Skip `wiki/nextjs/*` concepts. The component still works — just don't
mention App Router setup. Mention that `'use client'` is a Next.js
directive; in Vite/CRA/etc. it's a harmless comment that gets stripped.

---

## Failure modes and halts

| Situation                              | Response                                                 |
|----------------------------------------|----------------------------------------------------------|
| `ffmpeg` missing                       | Halt, print install command, wait for user               |
| `VIDEO_PATH` doesn't exist             | Halt, ask for correct path                               |
| `probe.sh` can't decode the file       | Halt, print the error, suggest re-export                 |
| `to-webp.sh` bundle >8MB after `-q 70` | Continue, warn user, suggest resolution downshift        |
| Target write directory doesn't exist   | Create it (with `mkdir -p`)                              |
| Target component file already exists   | Confirm with user before overwriting                     |

## Sources

- `references/raw/scroll-stop-builder-SKILL.md` — the reference skill's
  build-process structure (probe → extract → build) is the shape we
  follow. We add the interview-phase split and the template-render step.
