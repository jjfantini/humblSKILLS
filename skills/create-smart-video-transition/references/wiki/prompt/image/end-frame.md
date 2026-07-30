---
title: "Prompt B Template: The Ending Frame"
context: prompt
category: image
concept: end-frame
description: "Guidance for Prompt B (the final frame of the transition). Describes the delta from Prompt A (or the uploaded reference) and preserves everything else. Raw template at assets/templates/prompt-b-end-frame.tmpl."
tags: prompt-b, end-frame, image-generation, delta, continuity
sources: []
last_ingested: 2026-04-21
---

## Purpose

Prompt B produces the ending image. Its job is to stay visually coherent
with the starting frame so the video model has a clean interpolation
target. Describe the **delta** (what changed) and preserve **everything
else** (what must stay identical).

Skip this concept when `mode = end_only` or `mode = both`.

When `mode = start_only` (user uploaded the start), also read
`wiki/prompt/image/reference-fidelity.md` — the generated end frame must
use the upload as a visual reference.

## Template

The raw fill-in-the-blank template lives at:

```
assets/templates/prompt-b-end-frame.tmpl
```

Read the template file, fill the "Preserved" list exhaustively (don't
paraphrase "same as Prompt A" — enumerate), and list the deltas in
order of visual impact.

## Rules

1. **Preserve list must be explicit.** "Same as Prompt A" is not enough
   — enumerate every surface, light, material, and framing attribute
   the image model needs to keep identical.
2. **Deltas must be enumerated in order of visual impact.** Big changes
   first.
3. **If count matters** (shards, ingredients, droplets), state it
   numerically.
4. **Never rewrite the subject** — swapping the subject breaks
   transition continuity.

## Worked example (smoothie explosion)

```
PROMPT B - ENDING FRAME

Reference: follows from Prompt A's output (tropical smoothie in frosted
glass). Use Prompt A's output as a visual reference — preserve
every attribute below except the explicit deltas.

Preserved from the reference:
- Subject: same smoothie, same glass, same garnish
- Composition: same 3/4-front medium-close framing, same centering
- Camera: Phase One IQ4 150MP, 120mm macro, f/4.5
- Lighting: softbox key upper-left 5500K, white right fill, rim from behind
- Materials: same frosted glass, same fruit-puree color gradient
- Background: same soft white gradient floor

Deltas:
- The glass has shattered into ~30 curved shards frozen mid-flight,
  retaining the glass's silhouette in negative space
- The liquid has exploded outward in splashing sheets and droplets,
  frozen at peak energy
- Ingredients (mint leaf, pineapple wedge, two ice cubes, mango
  chunks, passion-fruit seeds) are flying outward along the same
  explosion axis, visibly separated
- One sugar-dust cloud is suspended at the upper center

Aspect ratio: $VIDEO_ASPECT_RATIO. Photorealistic, same color science,
1/10000s freeze-frame feel.
```

## Incorrect (drifting from the reference)

```
A smoothie glass shattering with liquid everywhere, vibrant and exciting.
```

No preservation contract. Model will re-roll the glass, angle, lighting,
materials — guaranteed drift.

## Correct

See the worked example above, and always pair with
`wiki/prompt/image/reference-fidelity.md` when one image was uploaded.

## Sources

- (synthesis) adapted from commercial product photography conventions.
