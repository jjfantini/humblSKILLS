---
title: "Common Skill Failures and Fixes"
context: anthropic
category: troubleshooting
concept: common-failures
description: "Upload rejected, skill doesn't trigger, triggers too often, instructions not followed, large-context degradation - with concrete fixes."
tags: troubleshooting, debugging, anthropic
sources:
  - "references/raw/anthropic-skill-building-guide.pdf"
last_ingested: 2026-04-19
---

## Skill won't upload

**Error: "Could not find SKILL.md in uploaded folder"**
Cause: file not named exactly `SKILL.md` (case-sensitive).
Fix: rename to `SKILL.md`. Verify with `ls -la` before zipping.

**Error: "Invalid frontmatter"**
Cause: YAML formatting issue. Common mistakes:

```yaml
# Wrong - missing delimiters
name: my-skill
description: Does things

# Wrong - unclosed quotes
---
name: my-skill
description: "Does things
---

# Correct
---
name: my-skill
description: Does things
---
```

**Error: "Invalid skill name"**
Cause: name has spaces, capitals, underscores, or uses reserved prefix `claude`/`anthropic`.
Fix: kebab-case, no reserved prefixes.

## Skill doesn't trigger

Symptom: the skill never loads automatically.

Debugging steps:

1. Ask Claude: `"When would you use the [skill-name] skill?"` - it will quote the description back.
2. If the description is too generic ("Helps with projects"), tighten it.
3. Add concrete trigger phrases users would actually type.
4. Mention relevant file types when applicable (`.csv`, `.fig`, `.pdf`).
5. Test a paraphrased query, not just the exact phrase in the description.

See `anthropic/description/trigger-design.md` for the full trigger-design playbook.

## Skill triggers too often

Symptom: the skill loads for queries unrelated to its purpose.

Fixes (in priority order):

1. Add negative triggers:
   ```yaml
   description: Advanced data analysis for CSV files. Use for statistical
     modeling. Do NOT use for simple data exploration (use data-viz skill).
   ```
2. Narrow the description - remove generic words like "processes", "helps", "handles".
3. Scope the domain explicitly ("PayFlow payment processing for e-commerce, not general financial queries").

## MCP connection failures

Symptom: the skill loads but MCP tool calls fail.

Checklist:

1. Verify the MCP server is connected (Settings > Extensions > [your service]).
2. Check authentication: API keys valid, permissions/scopes granted, OAuth tokens not expired.
3. Test the MCP independently of the skill: `"Use [Service] MCP to fetch my projects"`. If this fails, the issue is the MCP, not the skill.
4. Verify tool names in the skill match the MCP exactly (case-sensitive).

## Instructions not followed

Symptom: the skill loads but Claude doesn't follow its instructions.

Common causes:

1. Instructions too verbose. Keep bullets, lists, crisp prose. Move detail to `references/`.
2. Instructions buried. Put critical instructions at the TOP of SKILL.md. Use `## Important` or `## Critical` headers.
3. Ambiguous language:
   ```
   # Bad
   Make sure to validate things properly.

   # Good
   CRITICAL: Before calling create_project, verify:
   - Project name is non-empty
   - At least one team member assigned
   - Start date is not in the past
   ```
4. Model "laziness" on long sessions. Add explicit encouragement in the user prompt (more effective than in SKILL.md):
   ```
   Take your time to do this thoroughly. Quality is more important than speed.
   Do not skip validation steps.
   ```

**Advanced: when language instructions are too flaky, bundle a script.** Code is deterministic; natural language is not. See Anthropic's Office skills for examples - they use validation scripts rather than trusting instruction-following for critical checks.

## Large context / slow responses

Symptom: responses degrade, latency rises, or outputs lose quality over a long session.

Causes and fixes:

1. SKILL.md too large -> move detail to `references/`, keep SKILL.md under 5,000 words (~200 lines is a good target).
2. Too many skills enabled -> audit which skills are active. Anthropic suggests evaluating if you have 20-50+ skills simultaneously.
3. All content eagerly loaded -> verify you're using progressive disclosure. Body should link to references, not inline them.

## Quick diagnostic sequence

When something is wrong, run these in order:

1. Can Claude describe when to use the skill? (triggering / description quality)
2. Does the skill auto-load on an obvious query? (frontmatter)
3. Does the first MCP call succeed directly without the skill? (MCP / auth)
4. Does the skill produce correct output when manually invoked? (body / instructions)
5. Is the output consistent across runs? (determinism - consider bundling a script)

## Sources

- `references/raw/anthropic-skill-building-guide.pdf` - Chapter 5 "Patterns and troubleshooting / Troubleshooting" and Chapter 3 "Iteration based on feedback"
