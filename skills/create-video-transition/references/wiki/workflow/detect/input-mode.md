---
title: "Detect the Input Mode Before Generating Prompts"
context: workflow
category: detect
concept: input-mode
description: "Resolves the user's inputs into one of four modes (neither/start_only/end_only/both), which determines which prompts the skill generates."
tags: mode, detect, input, routing
sources: []
last_ingested: 2026-04-20
---

## Four input modes

The skill's first decision is: which of Prompt A, Prompt B, Prompt C do
we need to write? That depends entirely on what the user already has.

| Mode         | User provides                  | Generate          |
|--------------|--------------------------------|-------------------|
| `neither`    | Description only, no image     | Prompt A + B + C  |
| `start_only` | An image labeled as the start  | Prompt B + C      |
| `end_only`   | An image labeled as the end    | Prompt A + C      |
| `both`       | Two images (start + end)       | Prompt C only     |

Prompt C is always generated.

## Decision signals (in priority order)

1. **Attachment count.** 0 images → `neither`. 2 images → `both`. 1 image → disambiguate.
2. **Explicit labeling.** Words like "starting", "start frame", "beginning" → that image is the start. "Ending", "end frame", "final", "finish" → that image is the end.
3. **Directional language.** "From this to [description]" → the attached image is the start. "To this from [description]" → the attached image is the end.
4. **Ambiguous? Ask one question.** "You uploaded one image — should it be treated as the starting frame or the ending frame?"

## Incorrect (silently guessing)

```
User: "Here's the picture. Add some 'boom' to it."
Agent: [assumes it's the start frame, writes Prompt B + Prompt C, moves on]
```

Ambiguous upload; the agent guessed. If the user meant it as the end
frame, both generated prompts are wrong.

## Correct (resolve mode explicitly)

```
User: "Here's the picture. Add some 'boom' to it."
Agent: "Quick check — should that image be the starting frame (and we
animate it INTO the 'boom' version) or the ending frame (and we animate
FROM a calm predecessor INTO your uploaded image)?"
User: "The calm version. That's where I want to end up."
Agent: [mode = start_only] ...
```

Output of this step: one of `neither | start_only | end_only | both`, plus
any disambiguation answers, passed to
`workflow/elicit/video-length.md` and `workflow/generate/orchestration.md`.

## Sources

- (synthesis) authored from the skill's core requirements.
