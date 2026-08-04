---
title: "Capture Session Context Without Re-Deriving It"
context: handoff
category: capture
concept: session-context
description: "The receiving agent starts productive instead of re-investigating what this session already learned"
tags: capture, context, git, transcript, recall
sources: []
last_ingested: 2026-08-03
command: scripts/preflight.sh
---

## The Problem

A handoff written from memory alone drifts: the agent recalls intent but
misremembers branch names, file paths, and which commands actually passed.
A handoff written from `git log` alone is worse — it records what landed and
loses every reason, every rejected approach, and every open question.

Both halves are required. The live conversation supplies **why**; git supplies
**where**.

## Default Capture: Live Context + Git

Run `bash scripts/preflight.sh --slug <descriptive-name>` first. It returns
the date, both destination directories, branch, HEAD, dirty-file count,
unpushed-commit count, status, diffstat, last 10 commits, which instruction
files exist, and the installed skill list — one read-only call, no writes.

Then fill the human half from your own context:

| Doc section | Source |
|-------------|--------|
| Objective | invocation arguments, else the user's stated goal this session |
| What Is Already Done | your recall, **labelled `[verified]` / `[assumed]`** |
| Next Steps | your plan, in order, each naming a file |
| Environment & Commands | commands actually run this session (not guessed) |
| Artifacts | paths and URLs surfaced this session |
| Gotchas & Dead Ends | approaches tried and abandoned |
| Open Questions | decisions you deferred to the user |

## The `[verified]` / `[assumed]` Rule

Every completion claim carries a label. `[verified]` means a command was run
and its output observed — cite it. `[assumed]` means the code looks right but
was never exercised. Unlabelled claims are how a receiving agent builds on a
foundation that was never tested.

### Incorrect

```markdown
## What Is Already Done
- Added the retry wrapper to the fetch client
- Migration applied
- Tests pass
```

### Correct

```markdown
## What Is Already Done
- **[verified]** Retry wrapper added at `src/http/client.ts:88` —
  `pnpm test src/http` passes, 14/14
- **[verified]** Migration `0042_add_retry_log` applied to the local DB —
  `pnpm db:status` shows it at head
- **[assumed]** Retry honours `Retry-After` on 429 — code path written,
  never exercised against a real 429
```

## Transcript Digging Is Opt-In, Not Default

Do not parse session transcripts to write a routine handoff. Reach for them
only when the user explicitly asks for a deep or exhaustive capture, or when
this session was compacted and your recall of early decisions is thin. Then
locate the newest transcript for the current project and read it for the
decisions you lost:

```bash
ls -t ~/.claude/projects/*"$(basename "$PWD")"*/*.jsonl 2>/dev/null | head -3
```

Claude Code only. Codex and Cursor store history differently — in those
harnesses, ask the user to paste what matters rather than guessing a path.

## Do Not Duplicate Existing Artifacts

If a spec, plan, ADR, issue, PR, or diff already states something, link it and
move on. Copying it in makes two sources of truth, and the copy is stale the
moment the original changes. See
[../document/structure.md](../document/structure.md).

## Sources

- (none) - authored from the smart-handoff design session.
