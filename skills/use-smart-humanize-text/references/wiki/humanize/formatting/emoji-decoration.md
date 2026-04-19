---
title: "Emoji Decoration in Headings and Bullets"
context: humanize
category: formatting
concept: emoji-decoration
description: "AI puts emoji in front of headings and bullet points as decoration."
tags: formatting, emoji, decoration, ai-tells
sources:
  - "references/raw/Wikipedia_Signs_of_AI_writing.pdf"
last_ingested: 2026-04-17
---

## Emoji Decoration

AI puts emoji in front of headings and bullet points. Strip them.

**Incorrect:**

```markdown
🚀 **Launch Phase:** The product launches in Q3
```

**Correct:**

```markdown
The product launches in Q3.
```

Rule: no decorative emoji in structural elements (headings, list markers, section intros). The exception: emoji that carry information (reaction emoji in a chat log, emoji that are the subject of the text).

## Sources

- `references/raw/Wikipedia_Signs_of_AI_writing.pdf` - tell #19 in the formatting tells list
