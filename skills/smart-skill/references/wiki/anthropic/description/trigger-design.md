---
title: "Description Field: Trigger-Phrase Design"
context: anthropic
category: description
concept: trigger-design
description: "How to write the description field so Claude loads the skill on the right queries and leaves it alone on the rest. WHAT + WHEN + trigger phrases + negative triggers."
tags: description, triggers, activation, anthropic
sources:
  - "references/raw/anthropic-skill-building-guide.pdf"
last_ingested: 2026-04-19
---

## The description field IS the trigger

The `description:` is the ONLY part of your skill that's always in Claude's context. If the description doesn't match how a user actually talks, the skill never loads. Over-match, it loads on unrelated queries and pollutes context.

### Required structure

```
[What it does] + [When to use it] + [Key capabilities / trigger phrases]
```

Good descriptions include:

- The outcome / deliverable (not the architecture)
- Concrete trigger phrases users would type
- File types, tools, or domain terms where relevant
- A negative trigger when over-matching is plausible

### Good examples (Anthropic-provided)

```yaml
# Specific, actionable, trigger-rich
description: Analyzes Figma design files and generates developer handoff
  documentation. Use when user uploads .fig files, asks for "design specs",
  "component documentation", or "design-to-code handoff".
```

```yaml
# Clear value proposition + exact phrases
description: End-to-end customer onboarding workflow for PayFlow. Handles
  account creation, payment setup, and subscription management. Use when
  user says "onboard new customer", "set up subscription", or "create
  PayFlow account".
```

### Bad examples

```yaml
# Too vague - matches everything and nothing
description: Helps with projects.

# Missing triggers - Claude has no phrase to match against
description: Creates sophisticated multi-page documentation systems.

# Implementation detail, not user-facing outcome
description: Implements the Project entity model with hierarchical relationships.
```

### Negative triggers (under-disclosed but critical)

When your skill is specific, tell Claude what it is NOT for. This prevents over-triggering.

```yaml
description: Advanced data analysis for CSV files. Use for statistical
  modeling, regression, clustering. Do NOT use for simple data exploration
  (use data-viz skill instead).
```

```yaml
description: PayFlow payment processing for e-commerce. Use specifically
  for online payment workflows. Do NOT use for general financial queries.
```

### Debugging triggering

Ask Claude directly: `"When would you use the [skill-name] skill?"` Claude will quote the description back. If it can't articulate the trigger conditions, users won't either - tighten the description.

### Iterating based on telemetry

| Symptom                              | Likely cause              | Fix                                          |
|--------------------------------------|---------------------------|----------------------------------------------|
| Users manually enable the skill      | Under-triggering           | Add concrete trigger phrases, technical terms |
| Skill loads on unrelated queries     | Over-triggering            | Add negative triggers, narrow scope          |
| Support tickets "when do I use X?"   | Description too vague      | Lead with outcome + concrete user phrases    |
| Skill triggers but does wrong thing  | Wrong trigger terms        | Remove ambiguous phrases, check overlap with other skills |

## Sources

- `references/raw/anthropic-skill-building-guide.pdf` - Chapter 2 "Writing effective skills / The description field" and Chapter 5 "Skill doesn't trigger" / "Skill triggers too often"
