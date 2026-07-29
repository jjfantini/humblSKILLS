---
title: "Mode And Caps"
context: design
category: orchestration
concept: mode-and-caps
description: "Default to full mode with at most 15 findings; quick mode reports only HIGH and MEDIUM findings and caps the report at 5."
tags: orchestration, review, interface-design
sources:
  - "references/raw/SKILL.md"
last_ingested: 2026-07-29
---

## Mode And Caps

Default to full mode with at most 15 findings; quick mode reports only HIGH and MEDIUM findings and caps the report at 5.

### Review lens

- Use full when no mode is supplied and cap the consolidated report at 15.
- In quick mode inspect all available owners, include only HIGH and MEDIUM, and cap at 5.
- Never pad a report to reach its cap and never imply uninspected scope was reviewed.

### Source of truth

Read `references/raw/SKILL.md` before making findings or implementing a requested
change. The raw upstream source is authoritative; this concept is a routing and
progressive-disclosure summary.
