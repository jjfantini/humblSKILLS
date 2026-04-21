---
title: "Render a Provided-by-User Placeholder Tab"
context: output
category: html
concept: placeholder-tab
description: "When a prompt was skipped because the user uploaded that image, render the tab body as a placeholder card instead of a prompt body."
tags: html, placeholder, provided, tab, ux
sources: []
last_ingested: 2026-04-20
---

## Purpose

The HTML page always shows three tabs (A / B / C) for layout consistency.
When the user uploaded the start or end image, the matching prompt was
not generated. Instead of leaving the tab empty or hiding it, render a
"Provided by user" card in that tab's body. This keeps the navigation
steady and tells the user exactly what happened.

Prompt C is always generated, so tab C never renders as a placeholder.

## When to render the placeholder

| `{{PROMPT_A_MODE}}` | `{{CONTENT_A_BODY}}` contains                    |
|---------------------|--------------------------------------------------|
| `generated`         | `<div class="prompt-text" ...>ESCAPED PROMPT</div>` |
| `provided`          | `<div class="prompt-placeholder">...card...</div>`  |

Same rule for B. The template-fill step (see `template-fill.md`) builds
the full inner HTML for each tab, including the placeholder card markup,
and substitutes it into the `{{CONTENT_A_BODY}}` / `{{CONTENT_B_BODY}}`
slots. The JS `switchTab` logic reads `data-mode="generated|provided"` on
the content container to decide whether to show the copy button and
which panel header style to render.

## Card contents

Replace the `.prompt-text` container on that tab with a
`.prompt-placeholder` container holding:

- Paperclip SVG glyph (or clapperboard — anything that reads "uploaded file")
- Heading: `Provided by user`
- Subtext: `You uploaded this as the <start|end> frame — no prompt generated.`
- The copy button is hidden for this tab.

## Fill-time handling

Since the placeholder is rendered server-side (by the template-fill
step), detect the `__PROVIDED__` sentinel during substitution and emit
the placeholder HTML instead of wrapping it in `.prompt-text`.

Example fill logic (conceptual):

```
if prompt_a_text == "__PROVIDED__":
    contentA_inner = render_placeholder_card(frame="start")
    panel_header_a_hide_copy_btn = True
else:
    contentA_inner = f'<div class="prompt-text" data-original="{esc}">{esc}</div>'
    panel_header_a_hide_copy_btn = False
```

The JS that switches tabs must also hide the copy button when the active
tab is a placeholder (or it should always check `.prompt-text.active` is
present before enabling copy).

## Styling

Keep the placeholder card centered vertically within the panel body.
Use `.text-secondary` color for the subtext, lime for the heading, and a
soft glass background to stay on-brand. Aspect-ratio instruction tags on
the panel header can be hidden too (no image will be generated from a
provided upload).

## Incorrect (empty or hidden tab)

- Hiding tab A entirely → three-tab grid collapses to two, user may wonder what happened
- Rendering an empty `.prompt-text` → looks like a bug
- Rendering "TODO" → unprofessional

## Correct

A styled placeholder card that matches the rest of the page and makes
the state obvious. Copy button hidden because there is nothing to copy.

## Sources

- (synthesis) authored from the skill's placeholder-tab UX requirement.
