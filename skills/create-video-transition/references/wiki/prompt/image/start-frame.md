---
title: "Prompt A Template: The Starting Frame"
context: prompt
category: image
concept: start-frame
description: "Guidance for Prompt A (the first frame of the video). Rules + worked example here; raw template at assets/templates/prompt-a-start-frame.tmpl."
tags: prompt-a, start-frame, image-generation, commercial-director
sources: []
last_ingested: 2026-04-21
---

## Purpose

Prompt A produces the starting image for the transition. This is the
anchor that Prompt B will be the visual delta from, and that Prompt C
will name as its START FRAME. It must be unambiguously reproducible:
every lighting, framing, and material choice has to be explicit so a
second pass can match.

Skip this concept when `mode = start_only` or `mode = both`.

## Template

The raw fill-in-the-blank template lives at:

```
assets/templates/prompt-a-start-frame.tmpl
```

Read the template file, substitute every `<...>` placeholder with a
specific, reproducible value (see rules below), and keep
`$VIDEO_ASPECT_RATIO` as a literal token so the HTML settings modal can
swap it live.

## Rules

1. **Name real things.** "Phase One IQ4 150MP" > "high-end camera".
   "5500K daylight" > "soft light".
2. **No vague adjectives.** Remove "beautiful", "stunning", "amazing" —
   they are noise to the image model.
3. **Lock the composition** so Prompt B can match: same angle, same
   distance, same framing, same background.
4. **Leave `$VIDEO_ASPECT_RATIO` as a literal token** — the HTML settings
   modal substitutes it live.

## Worked example (smoothie)

```
PROMPT A - STARTING FRAME

Subject: A tall, frosted glass filled with a tropical smoothie — bright
mango-orange gradient, condensation beading on the glass, a sprig of mint
and a thin pineapple wedge as garnish. Centered, 3/4-front angle.

Composition: medium-close shot, product centered, slight low angle so
the drink reads premium. Clean white surface with soft gradient floor.

Camera: Phase One IQ4 150MP, 120mm macro prime, f/4.5, focused on the
front edge of the glass. Locked-off tripod.

Lighting: large softbox key from upper left (5500K), white bounce fill
from the right, subtle edge rim from behind. Shadows: soft, raking right.

Materials: frosted glass with visible condensation droplets, viscous
fruit-puree liquid with subtle striations, fresh mint leaf with pronounced
veining, pineapple wedge with fibrous texture.

Style: photorealistic, commercial advertising quality, Apple-product
aesthetic. Aspect ratio: $VIDEO_ASPECT_RATIO. Ultra-sharp, natural color,
no text.
```

## Incorrect (vague, unreproducible)

```
A beautiful smoothie on a white background, looking fresh and tropical.
```

No specificity. Next run will diverge. Prompt B will not be able to
match — it has no lighting direction, no camera system, no materials to
preserve.

## Correct

See the worked example above. Every surface, every light, every camera
parameter has a concrete value.

## Sources

- (synthesis) adapted from commercial product photography conventions.
