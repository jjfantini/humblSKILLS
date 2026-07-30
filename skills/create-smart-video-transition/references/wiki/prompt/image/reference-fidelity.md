---
title: "Reference Fidelity: Keep Generated Image True to the Uploaded Reference"
context: prompt
category: image
concept: reference-fidelity
description: "The preservation contract used when one image was uploaded. Forces the generated counterpart to stay visually coherent with the reference."
tags: reference, fidelity, preservation, continuity, upload
sources: []
last_ingested: 2026-04-20
---

## When this applies

Only when `mode = start_only` or `mode = end_only` (exactly one image was
uploaded). In these modes, one image prompt (A or B) must direct the
image model to treat the upload as a visual reference and drift only in
the explicit directions the user requested.

Skip this concept when `mode = neither` or `mode = both`.

## The preservation contract

Every reference-based prompt (A when end was uploaded, B when start was
uploaded) MUST include these clauses:

```
Use the uploaded image as the visual reference. Preserve, identically:
- Subject identity (same object, same person, same scene)
- Materials, finishes, and surface textures
- Colors, color palette, and color temperature
- Camera angle, focal length, subject distance, framing
- Lighting direction, intensity, and color temperature
- Shadow direction and softness
- Background, surface, and any environmental elements
- Aspect ratio and composition

Apply these deltas ONLY:
- <delta 1>
- <delta 2>
- ...
```

Language must be imperative and explicit. Image models reward directness.

## Five rules

1. **Name the upload explicitly** — "the uploaded image" / "the provided reference". Do not let the model interpret this as optional.
2. **Enumerate what stays, not just what changes.** Models drift on anything unlisted.
3. **Bound the deltas.** Every change should be phrased as "A becomes B" or "add C at position D" — no open-ended "make it different".
4. **Match the reference's color science.** Add "same color temperature", "same contrast curve", "same color grade as the uploaded image".
5. **If using a separate reference-aware pipeline** (e.g. Midjourney `--cref`, Higgsfield reference, DALL-E edit), note that explicitly so the user knows to enable it.

## Incorrect (soft reference)

```
Take the uploaded image and make it more dramatic.
```

"Take" is weak. "More dramatic" is unbounded. Model will regenerate
everything and lose the reference.

## Correct (hard reference)

```
Use the uploaded image as the visual reference. Preserve, identically:
the subject (same glass, same smoothie, same garnish), the 3/4-front
medium-close framing, the 120mm macro lens look, the softbox key light
from upper-left at 5500K, the soft right fill, the soft white gradient
background, and the same color science.

Apply these deltas only:
- The glass shatters into ~30 curved shards frozen mid-flight
- The liquid explodes outward into splashing sheets
- Ingredients (mint, pineapple, ice, mango chunks, passion-fruit
  seeds) are visibly separated along the explosion axis
```

## Sources

- (synthesis) authored from the skill's reference-preservation requirement.
