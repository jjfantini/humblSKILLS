---
title: "Select The Handoff For This Workstream"
context: pickup
category: resume
concept: workstream-selection
description: "Resume the current directory's work without mixing sibling projects or temporary handoffs."
tags: handoff, workstream, selection, scope
sources:
  - "references/raw/User request.txt"
  - "references/raw/Tandem request.txt"
last_ingested: 2026-10-14
command: scripts/pickup.py
---

# Select The Handoff For This Workstream

The newest handoff anywhere may belong to an unrelated project.
Default to `.humblskills/handoffs/` in the session's starting directory itself.
Never automatically select from ancestors, descendants, or temporary directories.
Smart-handoff uses the same folder for persisted output by default and records it as `SESSION_START_DIR`.
The Git root remains a separate repository fact; it does not override the workstream folder.

Run `python3 scripts/pickup.py context` there, optionally with `--topic billing`.
Choose the newest matching Markdown file by modification time, with filename as a tie-breaker.
Read another same-day document only when its name/title or the selected handoff establishes the same topic.
A shared date alone does not make two handoffs related.

If the local folder is missing, empty, or has no topic match, the helper lists the newest matching handoff per workstream under configured `project_roots`, up to four project-directory levels deep.
Fallback candidates are never automatically selected.
Ask which to resume and stop until selected.
With no roots configured, ask for the intended directory or document path.

Explicit Markdown paths override discovery, including temporary smart-handoff files.
Pass `--handoff /path/to/session-handoff.md`; never glob or scan `/tmp`.
Selection authorizes reading the file, not performing every recorded action.

**Incorrect:** launching in a nested workstream and taking a newer parent-project handoff.
**Correct:** use the nested workstream's own folder, or ask when it has none.

Check the handoff's repository, branch, and commit against the checkout with read-only Git commands when available.
Report mismatches; never silently switch directories, reset, stash, or check out a branch.
Read applicable instruction files and suggested skills before proposing work.
Do not delete persisted or temporary handoffs during orientation.
