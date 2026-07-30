---
title: "Testing a Skill: Trigger, Functional, Performance"
context: anthropic
category: testing
concept: three-layer-approach
description: "The three test layers Anthropic recommends for every skill: triggering (does it load?), functional (does it work?), and performance (is it better than no skill?)."
tags: testing, validation, anthropic
sources:
  - "references/raw/anthropic-skill-building-guide.pdf"
last_ingested: 2026-04-19
---

## The three test layers

Test every skill at three levels before shipping. Each answers a different question.

### 1. Triggering tests - "does it load on the right queries?"

Goal: 90%+ load rate on relevant queries, near 0 on unrelated ones.

Run 10-20 prompts. Track what auto-loads vs what needs explicit invocation.

```
Should trigger:
- "Help me set up a new ProjectHub workspace"
- "I need to create a project in ProjectHub"
- "Initialize a ProjectHub project for Q4 planning"

Should NOT trigger:
- "What's the weather in San Francisco?"
- "Help me write Python code"
- "Create a spreadsheet" (unless this skill handles sheets)
```

If the skill does not trigger on obvious queries: tighten the `description:` field. Add concrete user phrases. See `anthropic/description/trigger-design.md`.

If it triggers on unrelated queries: add negative triggers, narrow the scope.

### 2. Functional tests - "does it produce correct outputs?"

Goal: valid outputs, no API failures, edge cases handled.

```
Test: Create project with 5 tasks
Given: Project name "Q4 Planning", 5 task descriptions
When:  Skill executes workflow
Then:
  - Project created
  - 5 tasks created with correct properties
  - All tasks linked to project
  - No API errors
```

Anthropic's pro tip: iterate on ONE challenging task until Claude succeeds, then extract the winning approach into the skill. Faster signal than broad testing, leverages in-context learning.

### 3. Performance comparison - "is the skill actually helping?"

Goal: prove the skill beats the no-skill baseline on real work.

| Metric              | Without skill | With skill |
|---------------------|---------------|------------|
| Back-and-forth msgs | 15            | 2          |
| Failed API calls    | 3             | 0          |
| Tokens consumed     | 12,000        | 6,000      |
| User corrections    | 5             | 0          |

If the skill does not reduce messages, errors, or tokens vs a no-skill baseline - the skill is not worth the complexity. Delete or rewrite.

## Quantitative targets (aspirational)

- Skill triggers on 90%+ of relevant queries
- Completes workflow in N tool calls (baseline - N improvement)
- 0 failed API calls per workflow run

## Qualitative checks

- Users do not need to prompt Claude about next steps mid-workflow
- Workflows complete without user correction
- New users succeed on first try with minimal guidance

## Testing surfaces

| Surface                           | Use for                                         |
|-----------------------------------|-------------------------------------------------|
| Manual in Claude.ai / Claude Code | Fast iteration while developing                  |
| Scripted in Claude Code           | Repeatable automated cases across changes       |
| Programmatic via Skills API       | Full evaluation suites, CI, regression tracking |

Pick the surface that matches your rigor needs. An internal team skill has different testing requirements than one shipped to thousands of users.

## Sources

- `references/raw/anthropic-skill-building-guide.pdf` - Chapter 3 "Testing and iteration / Recommended Testing Approach"
