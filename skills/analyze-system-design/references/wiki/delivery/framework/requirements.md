---
title: "Gathering Functional and Non-Functional Requirements"
context: delivery
category: framework
concept: requirements
description: "Step 1: scope the system into a prioritized top 3-5 functional features and top 3-5 quantified non-functional qualities."
tags: requirements, functional, non-functional, scoping
sources:
  - "references/raw/learn/system-design/in-a-hurry/delivery.md"
last_ingested: 2026-06-01
---

## Gathering Requirements

Step 1 (~5 min). The whole rest of the interview exists to satisfy the requirements you set here, so prioritization is graded directly at FAANG.

**Incorrect (long unprioritized dump):**

```text
Functional: post, follow, feed, DMs, search, trends, notifications,
verified badges, lists, bookmarks, analytics, ads...
```

A long list hurts you more than it helps.

**Correct (top 3-5, PM-style conversation):**

```text
Functional ("Users should be able to..."):
- post tweets
- follow other users
- see a feed of tweets from people they follow
```

Drive it like talking to a product manager: "does the system need X?", "what happens if Y?" Arrive at a prioritized short list.

**Non-functional ("The system should be..."):** the top 3-5 system qualities, quantified and in context. "Low latency" is meaningless; "feed renders under 200ms p99" is useful because it names the operation and a target. Use the checklist in `nonfunctional-checklist`.

Order matters: discuss CAP (consistency vs availability) first because it shapes every later choice.

## Sources

- `references/raw/learn/system-design/in-a-hurry/delivery.md` - functional vs non-functional split, Twitter example, prioritization grading.
