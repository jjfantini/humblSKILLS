---
title: "Two-Phase Interview: Technical First, Aesthetic Optional"
context: workflow
category: interview
concept: phase-order
description: "Phase 1 (mandatory, technical): video path, output frame directory, frame-count budget. Phase 2 (optional, aesthetic): brand colors, loader vibe. Skipped entirely if the user only wants the canvas primitive. Separates must-have inputs from taste calls."
tags: interview, workflow, phase, technical, aesthetic
sources: []
last_ingested: 2026-04-20
---

## Two-Phase Interview

The reference skill (`scroll-stop-builder`) mixes technical inputs
(video path, FFmpeg prereqs) with aesthetic inputs (brand, colors, vibe)
in one mandatory questionnaire. That's wrong for our case: most users
just want the canvas primitive with neutral styling.

Split the interview into two phases. **Phase 1 runs every time. Phase 2
runs only if the user asks for styled chrome.**

---

## Phase 1 — Technical (mandatory)

Questions to ask, in order. If the user's initial message already
answers a question, skip it.

### 1. Video path

> "Where is the MP4 you want to scrub?"

Look for a path in the initial message. If not provided, ask. If
provided, confirm it exists before running `probe.sh`.

### 2. Output frame directory

> "Where should the extracted frames go?"

Default: `public/frames/` if the current working directory contains a
`next.config.*` file or an `app/` directory. Otherwise ask the user.

### 3. Frame count budget

> "Default is 100 frames. Want to override?"

Default 100. Accept anything 30–200. Push back on anything outside that
range:

- <30 frames: too choppy. Ask if they're sure.
- >200 frames: asset bundle too large; memory risk on iOS. Ask if they
  can trim the video instead.

### 4. FFmpeg availability (check, don't ask)

Run `ffmpeg -version` silently. If it fails:

> "FFmpeg isn't installed. Run `brew install ffmpeg` (macOS) or
> `apt install ffmpeg` (Linux) first, then re-invoke."

Halt. Do not auto-install.

### 5. Target framework (implicit)

Assume Next.js App Router unless the user says otherwise. If they say
"plain React" or "Vite" or "CRA", note it — the generated component is
framework-agnostic but the integration concept differs.

---

## Decision point: run Phase 2?

**Run Phase 2 if any of these are true:**

- User mentioned colors, brand, or styling in their initial message.
- User explicitly asked for "styled", "branded", "polished", or "with
  loader design".
- User's project has design tokens visible in the cwd (tailwind config
  with custom colors, design-system imports, etc.) and they've
  implicitly delegated design calls to you.

**Skip Phase 2 if:**

- User said "just the canvas" / "minimal" / "I'll style it myself".
- User hasn't mentioned aesthetics at all and their project doesn't
  signal otherwise.

When in doubt: **skip**. The component looks fine with neutral defaults
(`#000` bg, `#fff` accent in the loader). Users can override via CSS
custom properties.

---

## Phase 2 — Aesthetic (optional)

Only these questions. No logo, no typography, no design system — those
are out of scope for this skill.

### 1. Loader background color

> "What background color for the loading overlay? (hex like `#0A1A2F`,
> or skip for black)"

### 2. Loader accent color

> "What accent color for the progress bar? (hex like `#D4AF37`, or skip
> for white)"

That's it. Two questions. Takes 30 seconds.

---

## What NOT to ask

These belong to other skills or to the user's own design system:

- Logo file (we don't render it)
- Typography / fonts (user's layout owns these)
- Navbar, footer, CTA copy (we produce the hero only)
- Vibe / mood / tone (too subjective; user interprets the component in
  their own brand context)
- Full website content (scope creep — that's `scroll-stop-builder`)

If the user asks for any of these, redirect:

> "This skill generates just the scroll-scrubbed hero component. For a
> full website build, see `scroll-stop-builder` (HTML/JS) or combine
> this component with your existing page builder."

## Sources

Synthesis concept. The two-phase structure is a design call documented
in `decisions.md` under the "Split technical and aesthetic interview"
entry (written after the initial clarification round with the user).
