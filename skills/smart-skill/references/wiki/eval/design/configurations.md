---
title: "Three Arms - no_skill, flat_skill, smart_skill"
context: eval
category: design
concept: configurations
description: "Why you need three arms (not two) to prove a smart skill compounds; flat_skill isolates brain contribution from skill content."
tags: eval, methodology, arms, humblskills
sources: []
last_ingested: 2026-04-19
---

## Why three arms, not two

Running `with_skill` vs `without_skill` only proves "skill > none". It
doesn't isolate which part of the skill is doing the work: the `SKILL.md`
router, or the compounding brain data. For Smart Skills, the brain is the
claim - the router is just a loader.

The three arms:

| Arm           | What it loads                               | What it isolates                           |
|---------------|---------------------------------------------|--------------------------------------------|
| `no_skill`    | nothing                                     | Agent baseline - how hard is the task?     |
| `flat_skill`  | SKILL.md + scripts/, no wiki/raw/patterns   | Router-only - what does the skeleton buy?  |
| `smart_skill` | Full brain (router + wiki + raw + meta)     | Router + compounding brain                 |

The delta **smart_skill - flat_skill** is the number you want to highlight.
It answers "how much does the brain add on top of what the prose already
tells the agent?" - which is the defensible framing of the compounding
claim.

## Incorrect (two-arm test)

```
without_skill pass_rate 0.25
with_skill    pass_rate 0.65
delta         +0.40
```

Claim: "the skill helped +0.40". True but noise-heavy - the win might be
entirely from the router prose. A generic `SKILL.md` without any brain
would likely close most of the gap.

## Correct (three-arm test)

```
no_skill      pass_rate 0.25
flat_skill    pass_rate 0.60
smart_skill   pass_rate 0.75

smart_vs_flat +0.15   <- the brain's contribution
smart_vs_none +0.50   <- skill + brain vs nothing
```

Claim: "the brain adds +0.15 on top of the router." Stronger, falsifiable,
and correctly attributes where the value came from.

## Flat-skill derivation

`brain.DeriveFlat(srcSkill, dstDir)`:

1. Copy `SKILL.md` verbatim (triggering is identical).
2. Copy `scripts/` (lint.sh and friends still need to work).
3. Keep `references/_brain.md` and `references/_template.md` (structural).
4. Truncate `log.md`, `patterns.md`, `decisions.md`, `_index.md` to their
   header preamble - the format docs, not the entries.
5. Delete `references/wiki/` and `references/raw/` entirely.

Cached by SHA of the source directory so re-running against the same skill
skips the copy.

## Sources

This concept documents the methodology implemented in
[cli/internal/eval/brain/brain.go](../../../../../../cli/internal/eval/brain/brain.go) `DeriveFlat`.
