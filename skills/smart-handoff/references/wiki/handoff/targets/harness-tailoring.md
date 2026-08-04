---
title: "Tailor the Doc to the Receiving Harness"
context: handoff
category: targets
concept: harness-tailoring
description: "A doc full of Claude-specific slash commands is noise to Codex, and the receiving agent silently skips the steps it cannot parse"
tags: targets, harness, claude, codex, cursor, instructions
sources: []
last_ingested: 2026-08-03
---

## Infer First, Ask Only If Blank

Read the target off the invocation before asking. "hand this to codex",
"moving to cursor", "picking this up in Claude Code" all answer the question.
Ask exactly one question only when there is no signal at all:

> Which agent is picking this up — Claude Code, Codex, Cursor, or something else?

If the user shrugs or says "any", write the neutral form below rather than
asking twice.

## What Actually Differs

| Aspect | Claude Code | Codex | Cursor | Unknown / other |
|--------|-------------|-------|--------|-----------------|
| Instruction file | `CLAUDE.md` | `AGENTS.md` | `.cursor/rules`, `.cursorrules` | "the repo's agent instruction file" |
| Skill invocation | `/smart-commit` or "use the smart-commit skill" | prose: "follow the smart-commit skill at `<path>`" | prose + `@`-file references | prose only |
| Skill install path | `~/.claude/skills/` | `~/.agents/skills/` (**not** `~/.codex/`) | `~/.cursor/skills/` | `humblskills install <name>` and let the CLI resolve it |
| Sub-agents / tasks | available | plain execution | plain execution | do not assume |
| Tool naming | Read / Edit / Bash | shell-first | editor-first | describe the action, not the tool |

Name the instruction file the target actually reads, and check it exists —
`preflight.sh` prints an `instruction-files` block.

The SKILL.md format itself is identical across all three (the humblskills
adapters are `transform: passthrough`), so a suggested skill is portable. Only
the *path* and the *way you reference it* change.

## Claude Desktop and claude.ai Are the Exception

They cannot read skills off the filesystem — skills arrive as an account-level
zip upload. An `humblskills install` line is useless to them. When the target
is Claude Desktop or claude.ai, write the install column as:

```markdown
| `smart-commit` | step 3 | `humblskills export desktop smart-commit`, then upload the zip in Settings → Capabilities → Skills |
```

Also drop every filesystem path from Next Steps — that agent has no shell and
no repo.

## Write Actions, Not Tool Calls

The one rule that survives every harness: describe **what must happen**, not
which tool does it.

### Incorrect

```markdown
1. Use the Read tool on src/auth.ts, then Edit line 88 to add the retry guard.
2. Invoke /smart-commit to land it.
```

### Correct (Codex target)

```markdown
1. Add the retry guard at `src/auth.ts:88` — wrap `fetchToken()` so a 429
   retries once after honouring `Retry-After`.
2. Commit it following the smart-commit skill
   (`~/.agents/skills/smart-commit/SKILL.md`, install:
   `humblskills install smart-commit`) — one atomic conventional commit.
```

## The Neutral Form

When the target is unknown, one line replaces every harness-specific
reference:

```markdown
Read this repo's agent instruction file (`CLAUDE.md`, `AGENTS.md`, or
`.cursor/rules` — whichever exists) before making changes.
```

## Note the Origin Too

Record which harness and model wrote the doc in the header. It tells the
receiving agent how much to trust the `[assumed]` lines and where the
transcript lives if a deeper dig is needed.

```markdown
> **From:** Claude Code (Opus 5) → **To:** Codex
```

## Sources

- (none) - authored from the smart-handoff design session.
