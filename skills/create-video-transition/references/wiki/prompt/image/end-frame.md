---
title: "Prompt B Template: The Ending Frame"
context: prompt
category: image
concept: end-frame
description: "Template for the final frame of the transition. Describes the delta from Prompt A (or from the uploaded reference) and preserves everything else."
tags: prompt-b, end-frame, image-generation, delta, continuity
sources: []
last_ingested: 2026-04-20
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

```
PROMPT B - ENDING FRAME

Reference: <either "follows from Prompt A's output with the changes below"
OR "use the uploaded image as the visual reference; preserve every
attribute below except the explicit deltas">.

Preserved from the reference (must be identical):
- Subject: <same noun>, same position in frame
- Composition: same angle, same framing, same distance
- Camera: same sensor, same lens, same f-stop
- Lighting: same direction, same color temperature, same intensity
- Materials: same finishes, same colors, same textures
- Background: same surface, same gradient, same mood

Deltas (what changes):
- <explicit change 1 — new position / new state / new color of one specific element>
- <explicit change 2>
- <explicit change N>

Aspect ratio: $VIDEO_ASPECT_RATIO. Photorealistic, same color science as
reference, no drift.
```

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

## Rules

1. **Preserve list must be explicit.** "Same as Prompt A" is not enough — enumerate.
2. **Deltas must be enumerated in order of visual impact.** Big changes first.
3. **If count matters** (shards, ingredients, droplets), state it numerically.
4. **Never rewrite the subject** — swapping the subject breaks transition continuity.

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
