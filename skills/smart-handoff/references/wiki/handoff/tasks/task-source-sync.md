---
title: "Sync Handoff Tasks For The Next Pickup"
context: handoff
category: tasks
concept: task-source-sync
description: "Create and complete relevant tasks using the same private configuration that pickup reads."
tags: tasks, pickup, configuration, continuity
sources:
  - "references/raw/Task sync request.txt"
last_ingested: 2026-10-14
---

# Sync Handoff Tasks For The Next Pickup

Handoff writes agreed task changes; pickup reads the current list.
Both use smart-pickup's external personal profile and the same starting-folder scope.
An invoked handoff includes this sync when configured, unless the user requests a read-only handoff or no task changes.
This is agent-mediated work through the current client's connected tool, not background synchronization.

1. Resolve the installed smart-pickup directory from the client skill catalog or preflight inventory.
   Read its personal-task-source concept and run its `scripts/pickup.py context --cwd <session-start-folder>` by absolute path.
   Honor its config-path handling, including `XDG_CONFIG_HOME`; do not keep a second task-source configuration in smart-handoff.
   If the skill or profile is missing, save the handoff and report task sync unavailable with a configure suggestion.
   Explicit `todos: null` disables sync without falling back to another list.
2. Use the returned source, field mappings, closed states, and `todo_scope` for both reads and writes.
   If `todo_scope_unmapped` is true, report the missing mapping and skip task writes.
   Never replace a scoped lookup with an all-project lookup.
   Discover provider tools from the current catalog; connector names vary between clients.
3. Fetch the live schema before writing and inspect existing tasks in that scope, including closed tasks when checking duplicates or completion.
   For Notion, use its returned data-source ID and exact writable property names and status values.
   For a configured local list, preserve unrelated text and headings while adding or checking only matching scoped items.
   Missing write access or unsupported mappings leave sync unavailable; never create a replacement database automatically.
4. Add only agreed human-owned Next Steps or Open Questions that need tracking.
   Reuse a matching task by its link or clearly matching intent and scope; repeated handoffs must not create duplicates.
   If a matching row is already closed, do not reopen it merely because an older handoff repeats it.
   Keep titles short; put handoff links and file/run evidence in the task body.
   Set the resolved project value and a verified open state; leave unknown owners, priorities, and dates unset.
5. Complete only matching tasks this session actually finished, with verified evidence.
   For status-based sources, use the first configured `closed_statuses` entry only after verifying it is a completion state in the live schema.
   If it is not a valid completion state or no completion state is configured, report the missing mapping and leave the row unchanged.
   Assumed progress, proposed work, and partially completed tasks stay open.
   Append completion evidence without replacing existing task content.
6. Verify the resulting rows and record their links under the handoff's Done or Next Steps section, with a short task-sync result.
   After an uncertain write, inspect current state before retrying to avoid duplicates or repeated changes.
   Partial failure is reported as partial; pickup can still read the handoff and recheck live task state.
   Keep source IDs, task rows, and personal paths out of the shareable skill brain.

**Incorrect:** every handoff creates another copy of its next steps or closes tasks described only as assumed.
**Correct:** the handoff completes a verified task and links one new open task; pickup reads that same scoped source and shows the open task.
