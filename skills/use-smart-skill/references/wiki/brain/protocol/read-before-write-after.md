---
title: "The Non-Negotiable Read-Before / Write-After Brain Protocol"
context: brain
category: protocol
concept: read-before-write-after
description: "Without this block, agents read brain 0% of the time"
tags: protocol, brain, read-before, write-after, non-negotiable
sources: []
last_ingested: 2026-04-16
---

## The Read-Before / Write-After Protocol

Every Smart Skill ships a non-negotiable block at the top of its
`SKILL.md` body. Without it, agents generate from scratch every session
and the brain stays empty forever.

**Incorrect (protocol absent or buried):**

```markdown
# My Skill

## When to Use
- ...

## How to Use
- ...

## Brain Protocol            <!-- too late; agent has already planned -->
```

The agent has already decided what to do by the time it reaches the
brain section. Memory is never consulted.

**Correct (protocol FIRST, order matters):**

```markdown
# My Skill

[one-sentence description]

## Brain Protocol (read BEFORE creating anything)

1. `references/_index.md`       - what this skill knows (map)
2. `references/patterns.md`     - what worked, with numbers
3. `references/decisions.md`    - past reasoning, don't repeat mistakes
4. `references/log.md`          - last 5 session entries
5. Relevant `references/wiki/<context>/<category>/` concepts per task

After completing work, UPDATE the brain:
- New user-provided material lives in `references/raw/` (LLM never renames or edits it)
- Distilled insights -> new/updated `references/wiki/<context>/<category>/<concept>.md`
  - Cite every raw file you used in the `sources:` frontmatter array
- Performance data (if reported) -> `patterns.md`
- Non-obvious decisions -> `decisions.md`
- Session summary (always) -> append to `log.md`
- Run `scripts/lint.sh` to regenerate `_index.md` and verify structure

Every 2 weeks OR when requested: run `scripts/lint.sh` for contradictions,
stale data, orphan wiki pages, orphan raw files, and broken `sources:` paths.

No data, no improvement.

## When to Use
- ...
```

## Order Rules

1. Brain Protocol block comes BEFORE When-to-Use
2. Read operations (1-5) come BEFORE write operations
3. `log.md` write is ALWAYS on (every session appends at least one line)
4. `patterns.md` / `decisions.md` writes are conditional on what the
   session produced

## What Counts as "Session"

A session is any agent turn that invokes this skill. Even a simple
"what concepts exist?" query is a session - the agent reads the brain
and appends a QUERY entry to `log.md`. Over time the log becomes the
full audit trail.

## Copy-Paste Block

`scripts/scaffold.sh` injects the protocol block verbatim into every
new `SKILL.md`. If you migrate an existing skill, copy the "Correct"
example above and paste it at the top of the body. Do not paraphrase -
the exact wording is what the agent matches on.

## Sources

- (none)
