---
title: "Five Skill Patterns from Anthropic Early Adopters"
context: anthropic
category: patterns
concept: five-patterns
description: "Sequential workflow, multi-MCP coordination, iterative refinement, context-aware tool selection, domain intelligence - with when to use each."
tags: patterns, workflows, mcp, anthropic
sources:
  - "references/raw/anthropic-skill-building-guide.pdf"
last_ingested: 2026-04-19
---

## Choosing a pattern

Most skills lean either problem-first (user describes outcome, skill orchestrates the tools) or tool-first (user has tools connected, skill teaches the workflow). Knowing which framing applies helps you pick the right structure.

## Pattern 1 - Sequential workflow orchestration

Use when: users need a multi-step process in a specific order, and each step may depend on previous steps' output.

Example (customer onboarding):

```
Step 1: create_customer(name, email, company)
Step 2: setup_payment_method() - wait for verification
Step 3: create_subscription(plan_id, customer_id from Step 1)
Step 4: send_email(template=welcome_email_template)
```

Key techniques:
- Explicit step ordering with data passing between steps
- Validation at each stage before advancing
- Rollback instructions for failures mid-flow

## Pattern 2 - Multi-MCP coordination

Use when: the workflow spans multiple services and each is already exposed via its own MCP.

Example (design-to-dev handoff):

```
Phase 1 (Figma MCP):   export assets, generate specs
Phase 2 (Drive MCP):   store assets, generate share links
Phase 3 (Linear MCP):  create dev tasks, attach asset links
Phase 4 (Slack MCP):   post handoff summary to #engineering
```

Key techniques:
- Clear phase separation (readable as a pipeline)
- Validation gates between phases (don't post to Slack if Linear failed)
- Centralised error handling at the phase boundary

## Pattern 3 - Iterative refinement

Use when: output quality improves measurably with iteration (drafts, reports, designs).

Example (report generation):

```
Initial draft -> fetch data, generate draft, save to temp
Quality check -> run validation script, collect issues
Refinement loop -> address each issue, regenerate affected sections, re-validate
Finalisation -> apply formatting, save final
```

Key techniques:
- Explicit quality criteria (machine-checkable where possible)
- Validation scripts rather than language instructions
- A stop condition (quality threshold OR max iterations)

## Pattern 4 - Context-aware tool selection

Use when: the same user outcome can be achieved by different tools depending on input characteristics.

Example (file storage decision tree):

```
Large files (>10MB)    -> cloud storage MCP
Collaborative docs     -> Notion/Docs MCP
Code files             -> GitHub MCP
Temporary files        -> local storage
```

Key techniques:
- Decision criteria explicit at the top of the skill
- Fallback path when the preferred tool is unavailable
- Tell the user which choice was made and why (transparency)

## Pattern 5 - Domain-specific intelligence

Use when: the skill adds specialised knowledge that goes beyond just orchestrating tools (compliance, financial rules, medical guidelines).

Example (payment processing with compliance):

```
Before processing:
  1. Fetch transaction
  2. Check sanctions lists, jurisdiction rules, risk level
  3. Document compliance decision

Processing:
  IF compliance passed  -> call payment MCP, apply fraud checks
  ELSE                  -> flag for review, create compliance case

Audit trail:
  - Log all compliance checks
  - Record processing decisions
  - Generate audit report
```

Key techniques:
- Domain expertise embedded in the skill's logic, not the user's prompt
- Gate actions on compliance, not the other way around
- Comprehensive audit logging for governance

## Pattern selection matrix

| If you need...                                      | Use              |
|-----------------------------------------------------|------------------|
| A fixed ordered pipeline                            | Pattern 1        |
| Coordination across services                        | Pattern 2        |
| Quality that improves with iteration                | Pattern 3        |
| Different tools for different inputs                | Pattern 4        |
| Domain rules that gate tool use                     | Pattern 5        |

Combine freely. Many real skills are Pattern 2 (multi-MCP) wrapped in Pattern 5 (domain rules) with Pattern 3 (refinement) on the output.

## Sources

- `references/raw/anthropic-skill-building-guide.pdf` - Chapter 5 "Patterns and troubleshooting / Patterns 1-5"
