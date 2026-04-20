---
title: "Prompt C Template: Video Transition in Exact 2-Second Blocks"
context: prompt
category: video
concept: block-timeline
description: "Authoritative spec for Prompt C. Each 2-second block is a single dense sentence describing camera, subject, lighting, and focus. Block count is derived from $VIDEO_LENGTH."
tags: prompt-c, video, timeline, 2-second-blocks, transition
sources: []
last_ingested: 2026-04-20
---

## Purpose

Prompt C tells the video model how to transition from the start frame to
the end frame. It is decomposed into **exact 2-second blocks** so the
user can inspect and reason about the per-second motion instead of
trusting a generic "make it smooth" directive.

Prompt C is always generated, regardless of mode.

## Block count rule

```
full_blocks = floor(VIDEO_LENGTH / 2)
tail_seconds = VIDEO_LENGTH mod 2
total_blocks = full_blocks + (1 if tail_seconds > 0 else 0)
```

| VIDEO_LENGTH | Full 2s blocks | Tail block | Total |
|--------------|----------------|------------|-------|
| 4s           | 2              | —          | 2     |
| 5s           | 2              | 1s         | 3     |
| 6s           | 3              | —          | 3     |
| 7s           | 3              | 1s         | 4     |
| 8s           | 4              | —          | 4     |
| 10s          | 5              | —          | 5     |

## Authoritative output shape

```
PROMPT C - VIDEO TRANSITION

START FRAME: <one-sentence anchor matching Prompt A's output or the
uploaded start image>.
END FRAME: <one-sentence anchor matching Prompt B's output or the
uploaded end image>.

TIMELINE (total duration $VIDEO_LENGTH seconds at $VIDEO_FPS fps,
$VIDEO_ASPECT_RATIO):

Block 1 (0-2s): <camera move + subject motion + lighting shift + focal
concern, one dense sentence>.
Block 2 (2-4s): <continues from Block 1 — no jump cut>.
Block 3 (4-6s): <...>.
Block N (last — tail length noted if less than 2s): <...>.

AUDIO: $VIDEO_AUDIO.
STYLE: photorealistic, commercial-grade, <lens/sensor cue consistent with
Prompt A>. Camera: <locked-off | handheld | dolly | push-in>, no unplanned
movement.
CAMERA CONTINUITY: same focal length, same framing anchor as Prompt A,
same color science as Prompt A and B.
QUALITY: high fidelity, smooth at $VIDEO_FPS fps, no artifacts, no
morph-y in-betweens.
```

## Five rules for each block

1. **One dense sentence.** No lists, no sub-bullets inside a block.
2. **Name the motion.** Use a verb from `wiki/prompt/video/director-style.md` — push-in, dolly-back, rack focus, orbit, whip pan, etc. Never just "the camera moves".
3. **Cumulative motion.** Block N+1 picks up where Block N left off. No cuts. No teleporting.
4. **Name the focus.** Say what the viewer's eye is drawn to in that block — the foreground element, the shattering edge, the rising dust.
5. **Include lighting shift if any.** If the lighting changes between start and end, distribute the change across blocks rather than snapping.

## Worked example (4 blocks, 8s, smoothie explosion)

```
PROMPT C - VIDEO TRANSITION

START FRAME: tropical smoothie in a frosted glass on a white gradient,
centered 3/4-front medium-close with soft upper-left key light.
END FRAME: the same glass shattered into ~30 curved shards, liquid
erupting in splashing sheets, ingredients separated along an upward axis,
all frozen mid-explosion.

TIMELINE (total duration $VIDEO_LENGTH seconds at $VIDEO_FPS fps,
$VIDEO_ASPECT_RATIO):

Block 1 (0-2s): Camera holds locked-off on the intact smoothie, subtle
condensation droplet rolls down the glass, key light flickers imperceptibly
brighter as the first micro-fracture appears at the rim and viewer focus
tightens onto that fracture line.

Block 2 (2-4s): The glass begins a slow outward fracture from the rim
downward in a spiral pattern, individual shards start to separate with
0.5-unit motion per frame, liquid bulges at the surface without yet
breaking, and the rim-light intensifies to pick out each new shard edge.

Block 3 (4-6s): Shards lift clear of the glass silhouette in a
cascading outward push, liquid ruptures into sheeting splashes along
the explosion axis, ingredients (mint, pineapple, ice, mango) begin to
separate and lift, camera remains locked-off while rim light sharpens.

Block 4 (6-8s): All shards reach their peak outward travel and slow to
a freeze, the liquid sheets hold at full extension, a sugar-dust cloud
forms at upper center, ingredients lock into the final separated
positions, and the key light pulses once to punctuate the freeze-frame.

AUDIO: $VIDEO_AUDIO.
STYLE: photorealistic, commercial-grade, Phase One IQ4 120mm macro cue
consistent with Prompt A. Camera: locked-off tripod, no handheld drift.
CAMERA CONTINUITY: same focal length, same 3/4-front framing anchor,
same color science as Prompt A and Prompt B.
QUALITY: high fidelity, smooth at $VIDEO_FPS fps, no morph artifacts.
```

## Rules recap

- `$VIDEO_LENGTH`, `$VIDEO_ASPECT_RATIO`, `$VIDEO_AUDIO`, `$VIDEO_FPS` stay as LITERAL tokens in the output — the HTML settings modal substitutes them live.
- Block count MUST match the rule above. No fudging.
- Blocks must be narratively cumulative — no jump cuts.

## Sources

- (synthesis) authored from the skill's 2-second-block requirement.
