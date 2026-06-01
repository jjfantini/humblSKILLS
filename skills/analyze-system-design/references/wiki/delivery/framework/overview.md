---
title: "The 6-Step Delivery Framework"
context: delivery
category: framework
concept: overview
description: "The sequence and timing (requirements -> entities -> API -> data flow -> high-level -> deep dives) that delivers a complete, simple system in 45-50 min."
tags: framework, delivery, ooda, timing
sources:
  - "references/raw/learn/system-design/in-a-hurry/delivery.md"
  - "references/raw/learn/system-design/in-a-hurry/introduction.md"
  - "references/raw/learn/system-design/in-a-hurry/how-to-prepare.md"
  - "references/raw/whiteboards/framework_ooda.png"
last_ingested: 2026-06-01
---

## The 6-Step Delivery Framework

The single biggest failure mode is not delivering a working system, usually from over-engineering early. The framework gives you a linear track so you never get stuck, and a fallback when nerves hit.

**Incorrect (no structure, scope creep):**

```text
Jump straight to drawing boxes, add a cache and a queue and 3 shards
before any requirement demands them, run out of time, no complete system.
```

**Correct (linear build, simple first):**

| # | Step | Time | Output |
|---|------|------|--------|
| 1 | Requirements | ~5 min | 3-5 functional + 3-5 non-functional |
| 2 | Core entities | ~2 min | Bulleted noun list |
| 3 | API / interface | ~5 min | The contract |
| 4 | Data flow (optional) | ~5 min | Numbered transform sequence |
| 5 | High-level design | ~10-15 min | Boxes and arrows satisfying the API |
| 6 | Deep dives | ~10 min | Hardening for non-functional requirements |

Steps 1-5 satisfy **functional** requirements. Step 6 satisfies **non-functional** requirements.

Mental model (OODA): **1** observe, **2-4** orient, **5** output, **6** optimize. See the framework whiteboard.

Seniority calibration: mid-level candidates cover the basics well; senior/staff move fast through 1-5 and lead the deep dives proactively. Do not monologue; leave the interviewer room to probe.

## Sources

- `references/raw/learn/system-design/in-a-hurry/delivery.md` - the step sequence, timings, and anti-patterns.
- `references/raw/learn/system-design/in-a-hurry/introduction.md` - rubric (problem navigation, solution design, technical excellence, communication).
- `references/raw/learn/system-design/in-a-hurry/how-to-prepare.md` - prep workflow (pick a framework, practice problems) that frames why the framework matters.
- `references/raw/whiteboards/framework_ooda.png` - the observe/orient/output/optimize overlay on the 6 steps.
