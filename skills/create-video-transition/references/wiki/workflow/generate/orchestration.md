---
title: "End-to-End Orchestration: Mode -> Elicit -> Generate -> HTML -> Open"
context: workflow
category: generate
concept: orchestration
description: "The canonical eight-step recipe the skill follows from user request to HTML deliverable. Every session executes these steps in order."
tags: orchestration, workflow, recipe, steps, end-to-end
sources: []
last_ingested: 2026-04-20
---

## The eight steps (run in order, every time)

1. **Detect mode.** Read `wiki/workflow/detect/input-mode.md` → resolve to `neither | start_only | end_only | both`.
2. **Elicit settings.** Read `wiki/workflow/elicit/video-length.md` → confirm `$VIDEO_LENGTH`, collect (or default) `$VIDEO_ASPECT_RATIO`, `$VIDEO_AUDIO`, `$VIDEO_FPS`.
3. **Generate image prompts** per mode:
   - `neither` → Prompt A (`wiki/prompt/image/start-frame.md`) + Prompt B (`wiki/prompt/image/end-frame.md`)
   - `start_only` → Prompt B only; read `wiki/prompt/image/end-frame.md` + `wiki/prompt/image/reference-fidelity.md`
   - `end_only` → Prompt A only; read `wiki/prompt/image/start-frame.md` + `wiki/prompt/image/reference-fidelity.md`
   - `both` → skip image prompts
4. **Generate Prompt C.** Read `wiki/prompt/video/block-timeline.md` + `wiki/prompt/video/director-style.md` → write blocks; keep `$VIDEO_LENGTH`, `$VIDEO_ASPECT_RATIO`, `$VIDEO_AUDIO`, `$VIDEO_FPS` as literal tokens (not substituted) so the HTML modal can swap them.
5. **Build HTML.** Read `wiki/output/html/template-fill.md` + `settings-modal.md` + `placeholder-tab.md` → substitute placeholders (HTML-escape first); set `__PROVIDED__` sentinel for any skipped prompt; write defaults into the modal.
6. **Deliver.** Read `wiki/output/html/deliver.md` → write `prompts.html` to the user's cwd; `open prompts.html` (macOS) / `xdg-open prompts.html` (Linux); on failure print absolute path.
7. **Echo a compact fallback in chat.** Short summary of mode, length, block count, and where the HTML was written.
8. **Append to `log.md`.** One `[QUERY <date>]` entry recording mode, length, and any decisions made.

## Why this order

- Mode must be known before step 2 (no point asking length if mode was `both` and the user did not want Prompt C either — but Prompt C is always generated, so this is a safety net rather than a hard branch).
- Length must be fixed before step 4 (blocks depend on it).
- Image prompts must be written before HTML fill (HTML needs their text).
- HTML write must precede open (obvious).
- Log is last so it records the full outcome.

## Incorrect (skipping orchestration)

```
User: "Make a transition from this smoothie to the explosion."
Agent: [writes all three prompts immediately with a guessed 5s length, builds HTML]
```

The length was never confirmed. Prompt C claims 5s but the block timing
does not match any 2s grid. User gets a page that looks polished but is
structurally wrong.

## Correct

```
User: "Make a transition from this smoothie to the explosion."
Agent: [step 1: mode = start_only]
Agent: [step 2: "How long is the video? 6s default"]
User: "8 seconds"
Agent: [steps 3-8 in sequence]
```

## Sources

- (synthesis) authored from the skill's core requirements.
