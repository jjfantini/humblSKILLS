---
title: "Commercial-Director Style: Naming Lens, Camera, Light, Motion"
context: prompt
category: video
concept: director-style
description: "The vocabulary Prompt C blocks pull from. Specificity — named lenses, cameras, lighting, motion verbs — is what separates commercial-grade output from generic video."
tags: director, style, commercial, cinematography, vocabulary
sources: []
last_ingested: 2026-04-20
---

## Why specificity matters

Video models trained on commercial cinematography respond to trade
language. "The camera moves in" is interpreted generically. "Slow 35mm
push-in, 4-second travel" is interpreted as a specific shot type with
known conventions. This concept lists the vocabulary to use in every
block of Prompt C.

## Lenses (pick one, stay consistent across A, B, C)

| Lens               | Look                                            |
|--------------------|-------------------------------------------------|
| 35mm prime         | slight wide, minor distortion, documentary feel |
| 50mm prime         | natural perspective, default "human eye"         |
| 85mm portrait      | flattering compression, shallow DOF              |
| 120mm macro        | extreme close, razor DOF — product photography   |
| 85mm anamorphic    | cinematic wide, horizontal lens flare           |
| 16mm wide          | dramatic perspective, exaggerated depth          |

## Cameras / sensors (names that read as premium)

- ARRI Alexa 35 — cinema gold standard
- Phantom Flex 4K — high-speed / slow-motion
- Phase One IQ4 150MP — commercial product photography
- Sony Venice 2 — modern cinema workhorse
- RED V-Raptor — sharp, vivid

## Lighting vocabulary

- Key: primary directional light, state position ("upper-left") and color temp ("5500K")
- Fill: secondary soft light, usually opposite the key
- Rim / edge: behind-subject light to separate from background
- Practical: in-frame light source (lamp, window, LED strip)
- Softbox / hard source / bounce / flag: name the modifier
- Kelvin ranges: 3200K (tungsten / warm), 4300K (neutral), 5500K (daylight), 6500K (cool)

## Motion verbs (use these, not "moves")

| Verb           | Meaning                                               |
|----------------|-------------------------------------------------------|
| push-in        | camera physically dollies toward subject              |
| pull-out       | camera physically dollies away                        |
| dolly (left/right) | lateral travel                                    |
| truck          | synonym for dolly laterally                           |
| orbit          | camera circles the subject on a fixed radius           |
| rack focus     | focus shifts between two focal planes                 |
| whip pan       | rapid angular pan                                      |
| zoom in/out    | focal length change (not camera movement)             |
| crane up/down  | vertical travel                                       |
| locked-off     | camera does not move                                   |
| handheld       | camera floats / shakes slightly (avoid unless requested) |

## Style shorthand

- "Commercial advertising quality" — premium polish
- "Apple product aesthetic" — clean, white, macro-sharp
- "Freeze-frame 1/10000s" — high-speed photography
- "Slow-motion 120fps" — super-smooth motion
- "ASMR-like mechanical precision" — satisfying, deliberate motion

## Incorrect (generic)

```
Block 1: the camera moves in on the smoothie.
```

No lens, no speed, no focal behavior, no lighting note.

## Correct (directorial)

```
Block 1: slow 35mm push-in on the smoothie over 2s, focus held on the
front edge of the glass, softbox key light from upper-left intensifies
by 10% to pick out the first micro-fracture.
```

Named lens, named move, named speed, named focus, named light.

## Sources

- (synthesis) adapted from commercial cinematography trade conventions.
