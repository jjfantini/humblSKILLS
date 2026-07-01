---
name: smart-frontend-design
description: >
  Create distinctive, production-grade frontend interfaces with a synthesized
  design system and one style-defining intake question. Use when the user asks
  to build or redesign web components, pages, apps, dashboards, landing pages,
  UI flows, React/Vue/HTML/CSS interfaces, or "make this look better". Avoids
  generic AI aesthetics by discovering the existing frontend style before
  coding. Do NOT use for backend-only work, copy-only edits, or pure design
  critique where no frontend implementation is requested.
license: MIT
compatibility: "Requires python3 only for scripts/lint.sh. Frontend implementation uses whatever stack exists in the target repo."
metadata:
  author: jjfantini
  version: "1.0.3"
  category: design
  tags: [frontend, design, ui, ux, react, css, humblskill]
  platforms: [claude-code, cursor, codex]
  preserve:
    - references/raw/
    - references/wiki/
    - references/decisions.md
    - references/log.md
    - references/patterns.md
---

# Smart Frontend Design

Create frontend code that feels designed, not generated: one sharp design-intent
question, real codebase discovery, a synthesized aesthetic thesis, then working
UI with production-grade craft.

## Brain Protocol (read BEFORE creating anything)

1. `references/_index.md` - what this skill knows (map)
2. `references/patterns.md` - what worked, with numbers
3. `references/decisions.md` - past reasoning, don't repeat mistakes
4. `references/log.md` - last 5 session entries
5. Relevant `references/wiki/design/<category>/` concepts per task

After completing work, UPDATE the brain:
- New user-provided material lives in `references/raw/` (LLM never renames or edits it)
- Distilled insights -> new/updated `references/wiki/<context>/<category>/<concept>.md`
  - Cite every raw file you used in the `sources:` frontmatter array
- Performance data (if reported) -> `patterns.md`
- Non-obvious decisions -> `decisions.md`
- Session summary (always) -> append to `log.md`
- Run `scripts/lint.sh` to regenerate `_index.md` and verify structure

No data, no improvement.

_Full spec: `references/_brain.md`._

## CCCCC Architecture

| Layer | Role | Location |
|---|---|---|
| Core | Root structure of the skill | `SKILL.md`, `references/`, `scripts/` |
| Context | Top-level taxonomy grouping | First segment under `references/wiki/` |
| Category | Specific topic within a context | Second segment under `references/wiki/` |
| Concept | One atomic idea per file | Filename stem and frontmatter field |
| Command | Deterministic executable script | `scripts/<command>.sh\|.py` |

## Mandatory Step 0: Ask One Question

Before frontend implementation, ask exactly one pushback or context-gathering
question about the essence or style of the design. If the user already supplied
a style cue, brand reference, screenshot, audience, or emotional target, count
that as the answer and do not ask again.

Default question:

```markdown
What should this feel like in one sentence: brutal, luxe, playful, editorial,
raw, calm, weird, or something else?
```

## When to Use

- Building or redesigning frontend components, pages, dashboards, landing pages, apps, or UI flows
- Turning rough requirements into polished React, Vue, HTML/CSS, or framework-native UI
- Improving visual design quality while keeping implementation production-ready
- Avoiding generic AI UI patterns by discovering and extending the product's real design system

## How to Use

**Live enumeration of categories and concepts:**
Read `references/_index.md` after running `scripts/lint.sh`.

**Ask one design-intent question:**
Read `references/wiki/design/intake/one-question.md`.

**Discover the existing frontend style:**
Read `references/wiki/design/discovery/synthesize-style.md`.

**Commit to the aesthetic thesis:**
Read `references/wiki/design/direction/bold-aesthetic.md`.

**Translate the thesis into visual craft:**
Read `references/wiki/design/aesthetics/typography-color-motion.md` and
`references/wiki/design/anti-patterns/generic-ai-slop.md`.

**Implement and verify production UI:**
Read `references/wiki/design/implementation/production-code.md` and
`references/wiki/design/verification/review-checklist.md`.

## Examples

### Example 1: New Component

User says: "Build a pricing card component for this app."

Actions:
1. Ask the one style question unless the repo or user already supplies style.
2. Discover framework, styling system, existing components, tokens, and copy tone.
3. Synthesize a design thesis, then implement working UI in the repo's stack.

Result: A functional pricing card that fits the app, has a specific visual point
of view, and avoids stock SaaS design.

### Example 2: Visual Upgrade

User says: "Make this dashboard look less generic."

Actions:
1. Treat "less generic" as the trigger and ask for the desired feeling in one sentence.
2. Inspect current dashboard density, colors, typography, states, and data hierarchy.
3. Replace generic choices with a coherent aesthetic direction and verify behavior.

Result: The dashboard keeps its functionality but gains a recognizable design
system, responsive polish, and one memorable visual detail.

## Troubleshooting

**Output still feels generic**
Cause: The aesthetic thesis was too vague or discovery stopped at surface-level files.
Fix: Re-read discovery, direction, aesthetics, and anti-pattern concepts. Name the
specific thesis before touching code.

**User asks several design questions back**
Cause: The skill forgot the one-question contract.
Fix: Ask only the single highest-leverage essence/style question. Infer the rest
from repo discovery and implementation constraints.

## Success Signals

- Exactly one essence/style question is asked unless user context already answers it.
- The agent states a specific aesthetic thesis before implementation.
- Typography, color, motion, layout, and details support the same thesis.
- The implementation is real working code in the target stack, not a mock.
- Verification covers behavior, accessibility, responsiveness, and non-generic design quality.
- `scripts/lint.sh` exits 0 and `_index.md` matches the wiki taxonomy.
