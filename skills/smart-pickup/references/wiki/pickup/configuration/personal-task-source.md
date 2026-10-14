---
title: "Discover And Configure A Personal Task Source"
context: pickup
category: configuration
concept: personal-task-source
description: "Keep each user's task source private while sharing one pickup workflow."
tags: configuration, todos, privacy, notion
sources:
  - "references/raw/User request.txt"
  - "references/raw/Task sync request.txt"
last_ingested: 2026-10-14
command: scripts/pickup.py
---

# Discover And Configure A Personal Task Source

Embedding the author's database ID makes a shared skill read the wrong person's tasks.
Save settings at `$XDG_CONFIG_HOME/humblskills/smart-pickup.json`, defaulting to `~/.config/humblskills/smart-pickup.json`.
Keep real config, source IDs, and fetched tasks outside the skill package.
Smart-handoff also loads this profile with `context --cwd <session-start-folder>` for optional task sync.
Configure once for both skills; neither maintains a separate account or project mapping.

1. Read existing personal config, relevant instructions, and an installed to-do skill first.
   Explicit task-list references are stronger evidence than similarly named search results.
2. Discover available read tools for that provider in the current client's catalog.
   Tool names differ between agents; do not hardcode connector namespaces.
   If no source is known, search the connected service or user-selected local task locations.
   Ask one source-selection question if several candidates exist or none is clear, offering:
   link an existing list (the found candidates or one the user names), [create a Notion list](notion-task-list.md), use a local Markdown file, or disable tasks with `todos: null`.
   Prefer linking: users with an established list keep it, and setup adapts to its fields.
3. Fetch the source schema and perform a small read-only query before calling it verified.
   For Notion, resolve the database's actual `collection://` data source and inspect its properties and states.
   For other services, use supported filters; for files, inspect headings and checkboxes.
4. Map workstream directory names to existing project values when scoping is supported.
   Match exact path segments, not substrings.
   Do not change project options or the task service to make config fit; a missing value is added only through [add a category](add-category.md) when the user asks.
5. Prepare JSON based on `assets/config.example.json` and run `python3 scripts/pickup.py configure --input /path/to/prepared-config.json`.
   Explicit configure authorizes saving local references, not unrelated authentication or service writes.
   Never store tokens, passwords, headers, executable commands, or arbitrary instructions.
6. Report the saved source, verified fields, scope mappings, and anything unverified.
   On first setup, if the installed smart-handoff provides `scripts/session-nudge.py`, offer its optional handoff reminder once and follow its session-nudge concept only on a yes.
   Validation failure preserves the existing config.

**Incorrect:** shipping a personal database ID or assuming completed tasks are always called `Done`.
**Correct:** verify each user's actual list and save references outside the package.

`todos` accepts `provider`, `source`, optional `fields`, `closed_statuses`, and `scope_values`.
When handoff marks a status-based task complete, it uses the first `closed_statuses` entry only after confirming it is a completion state in the live schema.
Order a valid completion state first, and leave the row unchanged if that mapping cannot be verified.
`provider` identifies a connected service or `file`; it grants no access.
`fields` maps `title`, `status`, `done`, `project`, `priority`, and `due` to observed names.
Completion is either a status field with `closed_statuses` or a `done` checkbox field, never both.
Map `done` only when the checkbox is the list's real completion signal; a list that tracks completion by status keeps its status mapping even if it also has a checkbox.
`scope_values` maps directory names to exact values in the `project` field.
If mappings exist but the workstream has no match, ask for scope, offer [add a category](add-category.md) once, or report that scoped lookup was skipped; do not silently query all projects.

On pickup, revalidate the live schema before querying.
Exclude configured closed states, or checked `done` rows; null/unset statuses and unchecked boxes remain open unless the source's semantics say otherwise.
Bind untrusted SQL values with parameters and handle pagination/truncation.
Date properties may expand to separate query columns; use the live schema's query names.
Lead with blocked/waiting tasks if those states exist, then overdue, in-progress, and high-priority work.
Do not invent blocked/waiting states when the source has none.
Distinguish an empty verified result from unavailable access.
