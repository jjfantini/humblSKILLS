---
title: "Deliver prompts.html: Write and Open"
context: output
category: html
concept: deliver
description: "Write the filled template to prompts.html in the user's cwd and open it in the default browser. Fall back to printing the absolute path."
tags: html, deliver, open, cwd, fallback
sources: []
last_ingested: 2026-04-20
---

## The delivery contract

The HTML page is the primary deliverable. Chat output is a fallback.
Once the template is filled:

1. Write to `prompts.html` in the user's current working directory.
2. Open the file in the default browser.
3. On open failure, print the absolute path so the user can click/open manually.

## Where to write

```
<cwd>/prompts.html
```

If `prompts.html` already exists at the target path, overwrite it — the
user just asked for new prompts; the old file is stale.

## How to open

| OS       | Command                   |
|----------|---------------------------|
| macOS    | `open prompts.html`       |
| Linux    | `xdg-open prompts.html`   |
| Windows  | `start prompts.html`      |

Detect by checking the environment. If the detection is unreliable, try
`open` first, then `xdg-open`, then fail with the path printout.

## Chat fallback (always, regardless of open success)

After the HTML deliverable is produced, echo a compact summary in chat:

```
Done. Wrote prompts.html to <absolute-path>.

  Mode: <neither | start_only | end_only | both>
  Video: <VIDEO_LENGTH>s at <FPS>fps, <ASPECT_RATIO>, audio=<true|false>
  Block count: <N> (Prompt C)

  Paste targets:
  - Prompt A -> image generator (if generated)
  - Prompt B -> image generator (if generated)
  - Prompt C -> video model (start frame = Prompt A, end frame = Prompt B)

  Gear icon in the top-right of the page lets you retarget aspect/audio/fps live.
```

Keep under 15 lines. The page is the detail; chat is the signpost.

## Incorrect (writing nowhere, chat-only delivery)

```
Agent: [prints prompt text in chat, does not create HTML]
```

User asked for an HTML deliverable. This breaks the contract.

## Correct

```
Agent: [writes prompts.html, runs `open prompts.html`, echoes compact summary]
```

Page is primary; chat is secondary.

## Sources

- (synthesis) authored from the skill's HTML-primary delivery contract.
