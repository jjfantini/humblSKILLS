---
title: "Fill the HTML Template: Placeholders and Escaping"
context: output
category: html
concept: template-fill
description: "Placeholder list + HTML-escaping rules for producing the prompts.html deliverable from assets/prompt-page-template.html."
tags: html, template, fill, placeholder, escaping
sources: []
last_ingested: 2026-04-20
---

## Placeholder inventory

Read `assets/prompt-page-template.html`, then substitute these tokens in
order. Always HTML-escape prompt text before insertion (step 3 below).

| Placeholder                       | Source / example                                 |
|-----------------------------------|---------------------------------------------------|
| `{{OBJECT_NAME}}`                 | short noun phrase, page title, e.g. "Smoothie Explosion" |
| `{{HEADING_LINE1}}`               | first word of heading, e.g. "SMOOTHIE"            |
| `{{HEADING_LINE2}}`               | second word, faded in UI, e.g. "EXPLOSION"        |
| `{{TAB_A_NAME}}`                  | `Start Frame`                                     |
| `{{TAB_A_SHORT}}`                 | `Start` (mobile)                                  |
| `{{TAB_B_NAME}}`                  | `End Frame`                                       |
| `{{TAB_B_SHORT}}`                 | `End`                                             |
| `{{PROMPT_C}}`                    | full text of Prompt C (escaped). Used twice: once in `data-original` attribute, once as inner text. |
| `{{CONTENT_A_BODY}}`              | inner HTML for tab A: either `<div class="prompt-text" data-original="...">...</div>` when generated, or a `<div class="prompt-placeholder">...</div>` card when provided |
| `{{CONTENT_B_BODY}}`              | inner HTML for tab B (same shape as A)            |
| `{{PROMPT_A_MODE}}`               | `generated` or `provided`                         |
| `{{PROMPT_B_MODE}}`               | `generated` or `provided`                         |
| `{{DEFAULT_VIDEO_LENGTH}}`        | integer seconds, e.g. `8`                         |
| `{{DEFAULT_VIDEO_ASPECT_RATIO}}`  | e.g. `16:9`                                       |
| `{{DEFAULT_VIDEO_AUDIO}}`         | `true` or `false`                                 |
| `{{DEFAULT_VIDEO_FPS}}`           | `24`, `30`, or `60`                               |

## Three-step fill procedure

1. **Compute mode tokens.** From the resolved input mode:
   - `start_only` → `PROMPT_A_MODE=provided`, `PROMPT_B_MODE=generated`
   - `end_only`   → `PROMPT_A_MODE=generated`, `PROMPT_B_MODE=provided`
   - `both`       → both `provided`
   - `neither`    → both `generated`

2. **HTML-escape prompt text.** Replace, in order: `&` → `&amp;`, `<` → `&lt;`, `>` → `&gt;`. Apply to each prompt BEFORE inserting into the generated-body shape or into `{{PROMPT_C}}`. Apply escaping to both the `data-original` attribute value AND the inner-text slot (use the same escaped string for both).

3. **Build `CONTENT_A_BODY` / `CONTENT_B_BODY`:**
   - If mode for this tab = `generated`:
     `<div class="prompt-text" data-original="<ESCAPED_PROMPT>"><ESCAPED_PROMPT></div>`
   - If mode for this tab = `provided`:
     ```
     <div class="prompt-placeholder">
       <div class="prompt-placeholder-icon">
         <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21.44 11.05l-9.19 9.19a6 6 0 01-8.49-8.49l9.19-9.19a4 4 0 015.66 5.66l-9.2 9.19a2 2 0 01-2.83-2.83l8.49-8.48"/></svg>
       </div>
       <h3>Provided by user</h3>
       <p>You uploaded this as the <START_OR_END> frame — no prompt generated.</p>
     </div>
     ```
     Fill `<START_OR_END>` with `start` for tab A, `end` for tab B.

4. **Substitute all placeholders.** Simple string replace. All placeholders MUST be replaced — any literal `{{...}}` left in the file is a fill bug.

## Critical: leave video tokens as literal text

The prompt text contains the literal strings `$VIDEO_LENGTH`,
`$VIDEO_ASPECT_RATIO`, `$VIDEO_AUDIO`, `$VIDEO_FPS`. **Do not substitute
these during template fill.** They must survive into the rendered HTML so
the settings modal can swap them live. The defaults go into the modal via
the `{{DEFAULT_VIDEO_*}}` placeholders — that is how the user sees a
concrete starting value while still being able to retarget.

## Incorrect (premature substitution)

```
PROMPT C ... Block 1 (0-2s): ... at 30 fps, 16:9 aspect ratio ...
```

The video tokens are now baked in. The modal has nothing to swap, so
changing the aspect ratio in the UI does nothing.

## Correct

```
PROMPT C ... Block 1 (0-2s): ... at $VIDEO_FPS fps, $VIDEO_ASPECT_RATIO ...
```

Modal defaults read from `{{DEFAULT_VIDEO_FPS}}` and
`{{DEFAULT_VIDEO_ASPECT_RATIO}}` so the UI shows `30` and `16:9` on load,
and "Apply" replaces the tokens in-place.

## Sources

- (synthesis) derived from the HTML template contract.
