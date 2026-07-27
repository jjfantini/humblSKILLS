---
title: "Intake: What to Ask Before Writing Any HTML"
context: workflow
category: intake
concept: context-gathering
description: "The minimum question set that prevents rebuilding the page: content, audience, structure, orientation, interactivity budget"
tags: intake, requirements, orientation, questions
sources: []
last_ingested: 2026-07-27
---

## Ask first, build once

The most expensive failure mode for this skill is generating a beautiful page
around the wrong content model or the wrong orientation. Collect these five
answers before writing markup. If the user's request already answers one,
don't re-ask it.

### The question set

1. **What is this?** — report, one-pager, dashboard snapshot, resume,
   invoice, proposal, documentation page. The genre picks the layout
   skeleton.
2. **What data/content goes in, and where does it come from?** — pasted
   text, a file in the repo, numbers to visualize. Get the actual content
   (or a pointer to it) before designing around placeholders.
3. **Who reads it, on what?** — screen-first with a print fallback, or
   print-first? A board handout tolerates less interactivity than an
   internal dashboard.
4. **Orientation: portrait or landscape?** — US Letter is fixed; the
   orientation is the user's call and gets baked into `@page`. Wide tables
   and timelines usually want landscape; prose and stacked sections want
   portrait. Recommend one, but let the user decide.
5. **Interactivity budget** — sortable tables, collapsible sections, hover
   detail are fine on screen but must degrade to fully-expanded static
   content in print (see `pdf/export/print-css.md`).

### Incorrect

Jumping straight to markup from "make me a report" — then discovering after
the build that the vendor table is 9 columns wide and needed landscape, or
that half the content was meant to be a chart.

### Correct

One compact intake message (or AskUserQuestion where available) covering
only the unanswered items from the list above, then confirming:
"Portrait US Letter, three sections, inline SVG bar chart for Q3 numbers,
static print layout — building now."

## Sources

- (none) — authored from skill design; update with observed intake failures.
