---
title: "Gear-Icon Settings Modal for Live $VIDEO_* Swaps"
context: output
category: html
concept: settings-modal
description: "Spec for the gear-icon modal that lets the user live-swap $VIDEO_ASPECT_RATIO, $VIDEO_AUDIO, $VIDEO_FPS in the rendered prompts. Length is read-only here."
tags: html, modal, settings, gear-icon, live-swap
sources: []
last_ingested: 2026-04-20
---

## Purpose

The HTML page has a gear icon top-right. Clicking it opens a modal that
exposes the four video settings as editable fields. "Apply" performs an
in-DOM find-and-replace on the `$VIDEO_*` tokens across all three prompt
panels. "Reset" restores the original generated text.

## Fields

| Field                  | Type                          | Default                            | Effect                                              |
|------------------------|-------------------------------|------------------------------------|-----------------------------------------------------|
| `VIDEO_LENGTH`         | number (seconds)              | `{{DEFAULT_VIDEO_LENGTH}}`         | Replaces literal `$VIDEO_LENGTH`. Shows warning (see below). |
| `VIDEO_ASPECT_RATIO`   | select: 16:9 / 9:16 / 1:1 / 4:3 / 21:9 | `{{DEFAULT_VIDEO_ASPECT_RATIO}}` | Replaces literal `$VIDEO_ASPECT_RATIO`.         |
| `VIDEO_AUDIO`          | toggle (true/false)           | `{{DEFAULT_VIDEO_AUDIO}}`          | Replaces literal `$VIDEO_AUDIO`.                    |
| `VIDEO_FPS`            | select: 24 / 30 / 60          | `{{DEFAULT_VIDEO_FPS}}`            | Replaces literal `$VIDEO_FPS`.                      |

## VIDEO_LENGTH warning banner

Changing `VIDEO_LENGTH` in the modal does NOT recompute the 2-second
block count in Prompt C — only the skill can do that. Show a warning:

> Changing the video length here will substitute the token but will not
> add or remove 2-second blocks in Prompt C. For a new block count,
> regenerate via the skill.

Banner shows on any `VIDEO_LENGTH` modal interaction, stays visible until
modal is closed.

## Apply / Reset / Close behavior

- **Apply**: take the current form values, perform `innerHTML` or
  `textContent` replace on each `.prompt-text` element targeting the four
  `$VIDEO_*` tokens. Do not re-escape — the tokens are plain strings
  already in escaped text. Persist the applied settings in an in-memory
  state object so future Apply clicks diff against it.
- **Reset**: read each `.prompt-text`'s `data-original` attribute
  (set once at page render time) and restore its content. Reset the form
  inputs to their `{{DEFAULT_*}}` values.
- **Close**: hide the modal, leave any applied changes in place.

## Interaction with per-tab copy button

Copy reads `.prompt-text.textContent` at click time. That means copy
always reflects the latest applied settings — which is exactly what the
user wants. No extra wiring needed.

## Rendering requirements (for the template fork)

- Gear icon positioned top-right of `.header` (absolute or within
  `.panel-header-right` — both acceptable). Use an inline SVG gear glyph.
- Modal uses the VoltFlow aesthetic: dark glass card, lime accent border,
  backdrop-blur, space-grotesk heading, JetBrains-Mono for numeric
  inputs, Archivo for labels.
- Modal is hidden by `display: none` plus a CSS transition on `.open`
  class for fade-in.
- Form labels must identify the variable as a `$VIDEO_*` token so power
  users connect UI → prompt directly.

## Incorrect (modal does not change prompt text)

```javascript
function applySettings() {
  console.log("Applied");  // does nothing
}
```

The button moves but prompts do not update. User thinks it is broken.

## Correct

```javascript
function applySettings() {
  const tokens = {
    $VIDEO_LENGTH: document.getElementById("in-video-length").value,
    $VIDEO_ASPECT_RATIO: document.getElementById("in-video-ratio").value,
    $VIDEO_AUDIO: document.getElementById("in-video-audio").checked ? "true" : "false",
    $VIDEO_FPS: document.getElementById("in-video-fps").value,
  };
  document.querySelectorAll(".prompt-text").forEach(node => {
    let text = node.dataset.original || node.textContent;
    Object.entries(tokens).forEach(([k, v]) => {
      text = text.split(k).join(v);
    });
    node.textContent = text;
  });
}
```

`data-original` is set on first render so Reset can roll back cleanly.

## Sources

- (synthesis) authored from the settings-modal UX requirement.
