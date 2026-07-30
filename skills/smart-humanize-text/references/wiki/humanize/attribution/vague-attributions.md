---
title: "Vague Attributions: Experts Argue / Observers Note"
context: humanize
category: attribution
concept: vague-attributions
description: "AI writes 'experts argue,' 'observers note,' 'industry reports suggest' without naming anyone."
tags: attribution, sourcing, hedging, ai-tells
sources:
  - "references/raw/Wikipedia_Signs_of_AI_writing.pdf"
last_ingested: 2026-04-17
---

## Vague Attributions

AI writes "experts argue," "observers note," "industry reports suggest" without naming anyone. Name a source or state the claim directly.

**Incorrect:**

```markdown
Experts argue that this approach is more effective.
```

**Correct:**

Name the expert, or state the claim directly.

```markdown
Andrej Karpathy argues that this approach is more effective.
```

Or:

```markdown
This approach is more effective because it ships 3x faster.
```

Rule: if you can't name the source, don't invoke one.

## Sources

- `references/raw/Wikipedia_Signs_of_AI_writing.pdf` - tell #9 in the structural tells list
