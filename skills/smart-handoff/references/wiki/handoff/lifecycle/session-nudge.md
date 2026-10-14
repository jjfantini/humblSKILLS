---
title: "Remind The Agent To Offer A Handoff"
context: handoff
category: lifecycle
concept: session-nudge
description: "Optional Stop hook that offers a handoff when a session has unsaved work, without nagging every reply."
tags: lifecycle, hooks, reminder, onboarding
sources:
  - "references/raw/Nudge request.txt"
last_ingested: 2026-10-14
command: scripts/session-nudge.py
---

# Remind The Agent To Offer A Handoff

Sessions that end without a handoff lose their state; a reminder at the right moment fixes that.
Claude Code and Codex run Stop hooks after every reply, not once at session end, so a naive reminder nags constantly.
`scripts/session-nudge.py` is opt-in and stays silent unless this session has unsaved work.

## When It Speaks

- Unsaved work means this session edited files under the workstream root after its last handoff.
  Edits come from the session's own transcript: Claude Edit/Write calls, Codex `apply_patch`, and shell `>`, `>>`, or `tee` targets.
  Other sessions in the same folder never trigger it; writes made by scripts or subagents are not seen.
- Writing any `.humblskills/handoffs/*.md` resets it until new work is edited.
- After one nudge it stays quiet until both 45 minutes (`--minutes`) have passed and new edits exist.
- The reminder asks the agent to offer a handoff in one or two sentences and to accept a no.

## Offer It Once During Onboarding

Offer it when the user first sets up handoffs or pickup, or asks to be reminded; never install it silently.
Ask which agents and scope: every session, or one project.
Then run the installer by absolute path:

```
python3 <skill-dir>/scripts/session-nudge.py install --agent claude              # every Claude Code session
python3 <skill-dir>/scripts/session-nudge.py install --agent claude --project DIR # one project, .claude/settings.local.json
python3 <skill-dir>/scripts/session-nudge.py install --agent codex --project DIR  # one project, .codex/hooks.json
```

Installing again replaces only this hook and keeps every other setting and hook.
`uninstall` with the same flags removes it.
Codex installs are per project and ask the user to trust the new hook on first run.
Project installs use the local, uncommitted Claude settings file so a personal reminder does not land in a shared repository.

**Incorrect:** a reminder after every reply, or one triggered by another session's edits.
**Correct:** one offer after real work, repeated only after a quiet period with new edits.
