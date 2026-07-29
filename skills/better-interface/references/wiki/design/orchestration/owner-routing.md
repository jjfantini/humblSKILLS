---
title: "Owner Routing"
context: design
category: orchestration
concept: owner-routing
description: "Run the six domain owners in the required order and continue with an explicit Not reviewed result when an owner is unavailable."
tags: orchestration, review, interface-design
sources:
  - "references/raw/SKILL.md"
last_ingested: 2026-07-29
---

## Owner Routing

Run the six domain owners in the required order and continue with an explicit Not reviewed result when an owner is unavailable.

### Review lens

- Load better-accessibility, better-layout, better-writing, better-typography, better-colors, then better-ui.
- Treat every available owner as authoritative for its domain.
- For each unavailable owner, write Not reviewed, identify the missing skill, and continue.

### Source of truth

Read `references/raw/SKILL.md` before making findings or implementing a requested
change. The raw upstream source is authoritative; this concept is a routing and
progressive-disclosure summary.
