---
title: "Prompt A Template: The Starting Frame"
context: prompt
category: image
concept: start-frame
description: "Commercial-director-grade prompt template for the first frame of the video. Emphasizes reproducibility so Prompt B can reference it exactly."
tags: prompt-a, start-frame, image-generation, commercial-director
sources: []
last_ingested: 2026-04-20
---

## Purpose

Prompt A produces the starting image for the transition. This is the
anchor that Prompt B will be the visual delta from, and that Prompt C
will name as its START FRAME. It must be unambiguously reproducible:
every lighting, framing, and material choice has to be explicit so a
second pass can match.

Skip this concept when `mode = start_only` or `mode = both`.

## Template

```
PROMPT A - STARTING FRAME

Subject: <specific noun phrase — the object / scene / character>.
Composition: <angle, framing, distance>. <subject-in-frame description>.
Environment: <background / setting>, <color palette>, <mood>.

Camera: <sensor, lens, f-stop, focus>. <locked-off or handheld>.
Lighting: <key + fill + rim>, <direction>, <color temperature in Kelvin>,
<soft/hard>. Shadows: <hard/soft, direction>.
Materials: <per-surface specificity: metal finish, fabric type, glass,
liquid state>. Textures: <sharpness, grain>.

Style: photorealistic, commercial-grade, <lens-brand cue if relevant>.
Aspect ratio: $VIDEO_ASPECT_RATIO. Ultra-sharp detail, natural color
science, no text overlays, no logos.
```

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

## Rules

1. **Name real things.** "Phase One IQ4 150MP" > "high-end camera". "5500K daylight" > "soft light".
2. **No vague adjectives.** Remove "beautiful", "stunning", "amazing" — they are noise to the image model.
3. **Lock the composition** so Prompt B can match: same angle, same distance, same framing, same background.
4. **Leave `$VIDEO_ASPECT_RATIO` as a literal token** — the HTML settings modal substitutes it live.

## Incorrect (vague, unreproducible)

```
A beautiful smoothie on a white background, looking fresh and tropical.
```

## Correct

See the worked example above.

## Sources

- (synthesis) adapted from commercial product photography conventions.
