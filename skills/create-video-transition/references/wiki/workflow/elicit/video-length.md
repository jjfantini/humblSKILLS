---
title: "Elicit VIDEO_LENGTH Before Writing Prompt C"
context: workflow
category: elicit
concept: video-length
description: "Prompt C's block count is derived from $VIDEO_LENGTH, so the length MUST be fixed before Prompt C is written. Other video settings have safe defaults."
tags: elicit, video-length, settings, variables, blocks
sources: []
last_ingested: 2026-04-20
---

## Why length must be fixed up front

Prompt C is broken into **exact 2-second blocks**. The number of blocks
is:

```
blocks = floor(VIDEO_LENGTH / 2)  full 2s blocks
       + 1 tail block of (VIDEO_LENGTH mod 2) seconds if remainder > 0
```

Changing `$VIDEO_LENGTH` after Prompt C is written would invalidate the
block timeline. The HTML settings modal cannot recompute blocks — only
the skill can. So the length is elicited BEFORE Prompt C is written, not
left to the modal.

The other variables (`$VIDEO_ASPECT_RATIO`, `$VIDEO_AUDIO`, `$VIDEO_FPS`)
are safe to swap after generation because they do not change the prompt's
structure — they are literal tokens interpolated into the text.

## When length is already stated

If the user's message contains a length ("4 seconds", "make it 8s", "10
second transition"), echo it back and proceed:

> "Got it — 8-second transition, so Prompt C will have 4 blocks (0-2s, 2-4s, 4-6s, 6-8s). Moving on."

## When length is missing

Ask once, with concrete defaults:

> "How long should the final video be? Common choices: **4s** (2 blocks), **6s** (3 blocks), **8s** (4 blocks), **10s** (5 blocks)."

Default if the user does not answer: `6s` (3 blocks). Note this explicitly
in the log and on the HTML page.

## Other settings (gather defaults, let the HTML modal swap)

| Variable               | Default | Modal lets user change? |
|------------------------|---------|-------------------------|
| `$VIDEO_LENGTH`        | 6       | No (blocks don't recompute) |
| `$VIDEO_ASPECT_RATIO`  | 16:9    | Yes                     |
| `$VIDEO_AUDIO`         | false   | Yes                     |
| `$VIDEO_FPS`           | 30      | Yes                     |

Accept user overrides at elicitation time if volunteered, otherwise use
the defaults and let the HTML modal handle tweaks.

## Incorrect (writing Prompt C before length is known)

```
Agent: [writes a vague Prompt C with "approximately 5-second transition, divided into beats"]
```

The block timeline lies. Block count is unverifiable. The deliverable
promises 2-second precision but cannot honor it.

## Correct

```
Agent: "Before I write Prompt C — how long is this video? 6s (3 blocks) is a common default."
User: "8 seconds."
Agent: [writes Prompt C with exactly 4 blocks: 0-2s, 2-4s, 4-6s, 6-8s]
```

## Sources

- (synthesis) authored from the skill's core requirements.
