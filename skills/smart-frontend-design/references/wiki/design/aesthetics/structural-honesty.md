---
title: "Structure Is Information"
context: design
category: aesthetics
concept: structural-honesty
description: "Numbering, eyebrows, dividers, and labels must encode something true about the content instead of decorating it"
tags: structure, numbering, hierarchy, labels, dividers, frontend
sources:
  - "references/raw/frontend-design-SKILL.md"
last_ingested: 2026-07-30
---

## Structure Is Information

Structural devices — numbering, eyebrows, dividers, labels — should **encode
something true about the content**, not decorate it. Every one of them makes a
claim to the reader. If the claim is false, the design is lying quietly.

Numbered markers (`01 / 02 / 03`) are the most common offender. They assert
*sequence*. That is only appropriate when the content genuinely is a sequence —
a real process, a typed timeline, ordered steps where order carries information
the reader needs.

**Incorrect (numbering three peer features that have no order):**

```html
<section>
  <span class="eyebrow">01</span><h3>Fast</h3>
  <span class="eyebrow">02</span><h3>Secure</h3>
  <span class="eyebrow">03</span><h3>Affordable</h3>
</section>
<!-- Nothing is first. The numbers add rhythm and assert a lie. -->
```

**Correct (numbering something that is actually ordered):**

```html
<ol class="process">
  <li><span class="step">01</span><h3>Sample collected on site</h3></li>
  <li><span class="step">02</span><h3>Chain-of-custody sealed</h3></li>
  <li><span class="step">03</span><h3>Lab confirms within 24h</h3></li>
</ol>
<!-- Order is the information. The markup is <ol> because the content is ordered. -->
```

The same test applies to the rest: an eyebrow should name a real category, a
divider should separate things that genuinely differ in kind, a label should
label. Before adding a structural device, ask what it asserts and whether that
assertion is true. If it is there for rhythm, find rhythm in spacing or type
instead.

## Sources

- `references/raw/frontend-design-SKILL.md` — "structure is information", the
  `01 / 02 / 03` example, and the instruction to question structural devices
  before incorporating them.
