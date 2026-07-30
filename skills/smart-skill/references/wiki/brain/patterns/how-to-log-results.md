---
title: "The Shape of a patterns.md Entry with a Worked Example"
context: brain
category: patterns
concept: how-to-log-results
description: "Compounding performance memory; read every session"
tags: patterns, logging, metrics, performance, brain
sources: []
last_ingested: 2026-04-16
---

## patterns.md Entry Shape

`patterns.md` is where quantified outcomes accumulate. After 20 entries,
the agent stops guessing and starts predicting what works.

**Incorrect (vague, unmeasured, unactionable):**

```markdown
### 2026-04-12 | Content Post
Wrote a LinkedIn post. It did pretty well.
What worked: good hook.
```

No numbers, no hook described, no replicable rule. Future agents can't
learn from it.

**Correct (quantified, specific, action-generating):**

```markdown
### 2026-04-12 | LinkedIn Post: Security Checks Lead Magnet
- Context: test disaster-story hook on a high-stakes security topic
- Approach: disaster story + $87,500 fraud hook + lead magnet CTA (SHIP keyword)
- Result:
  - Impressions: 75,613
  - Comments: 117 (100+ from decision-makers)
  - Engagement rate: 0.31%
- Worked:
  - $87,500 fraud hook stopped the scroll (highest first-3-lines retention in 30 days)
  - SHIP keyword CTA converted 100+ comments with minimal friction
- Didn't:
  - 0.31% engagement is below average; massive reach, most lurked
  - No follow-up DM automation; most commenters weren't contacted for 48h
- Lesson: DISASTER STORY + DOLLAR AMOUNT + LEAD MAGNET CTA = viral formula.
  Replicate hook structure. Add DM automation to capture lurkers.
```

## Required Fields per Entry

| Field      | Required | Notes                                               |
|------------|----------|-----------------------------------------------------|
| Title line | Yes      | `### <YYYY-MM-DD> \| <short description>`           |
| Context    | Yes      | What was attempted, in one sentence                 |
| Approach   | Yes      | The method / structure / specific choices           |
| Result     | Yes      | Numeric outcomes; include units and baselines       |
| Worked     | Yes      | What specifically helped (evidence-backed)          |
| Didn't     | Yes      | What hurt or under-performed                        |
| Lesson     | Yes      | The rule to apply next time, stated as an imperative|

## When to Skip

If the session produced no measurable outcome, don't invent one. Write
the session to `log.md` and skip `patterns.md`. Patterns are for actual
evidence.

## When to Revise an Entry

Never edit old entries. If a later result contradicts an earlier one,
add a new entry and reference the older one:

```markdown
### 2026-05-20 | Disaster-hook formula: diminishing returns
- Context: third replication of the 2026-04-12 formula
- Result: impressions down 62% vs first run
- Lesson: formula is fatiguing to the audience. Rotate hooks after 2 uses.
  See 2026-04-12 entry for original result.
```

## Reading patterns.md

Every session starts with a full read of `patterns.md`. If the file
grows past ~100 entries, consider summarising older entries into a
wiki concept under `brain/patterns/` and keeping only the last 30
entries in the live file. Lint can flag this threshold.

## Sources

- (none)
