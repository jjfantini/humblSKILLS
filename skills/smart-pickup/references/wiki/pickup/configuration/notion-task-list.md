---
title: "Create A Notion Task List During Setup"
context: pickup
category: configuration
concept: notion-task-list
description: "Provision a new user's private task list through their connected Notion tool."
tags: notion, setup, provisioning, todos
sources:
  - "references/raw/User request.txt"
last_ingested: 2026-10-14
command: scripts/pickup.py
---

# Create A Notion Task List During Setup

If a user has no task list, configure can create one instead of requiring manual setup.
Offer reuse, create in Notion, or local-file/disabled lookup.
Only call the creation tool after the user chooses creation; automatic discovery does not authorize provisioning.
For an existing working configuration, reuse it unless the user asks for another list.

Run `python3 scripts/pickup.py init-notion` to prepare the creation request.
It returns JSON for the agent's connected tool; a local script cannot invoke a host's MCP tools itself.
Use `--title` to set the requested name or `--parent` for a user-selected parent page ID.
With no parent, create as a private workspace-level page.
Read the connected tool's schema/access first; adapt the JSON to that client's equivalent database-creation tool.
If Notion is not connected, report the missing connection; do not collect API tokens or create a lookalike local file.

The supplied DDL creates Task name, Status, Priority, Project, and Due date.
Its statuses include Not started, In progress, Blocked, Waiting, and Done.
No task rows, assignees, sharing changes, or arbitrary project labels are seeded.

After creation, capture the returned database URL and `collection://` data source ID.
Fetch its schema and run a read-only open-task query, then use the personal-task-source flow to save the real ID and verified fields.
Set `closed_statuses` to `["Done"]` only after the returned schema confirms that state.
Configure workstream/project mappings with the user; do not assume the database title is its ID.
If creation succeeds but later configuration fails, retain and report the returned URL so setup can reconnect without creating duplicates.
After a timeout or unknown creation outcome, inspect the target before retrying; never blindly create another database.

**Incorrect:** showing a prepared request as a created task list, or making a new database on every pickup.
**Correct:** user chooses creation, the connected tool returns the real source, and configure saves it for later reuse.
