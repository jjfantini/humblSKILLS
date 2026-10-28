---
title: "Add A Category For A Workstream"
context: pickup
category: configuration
concept: add-category
description: "Create a missing project value in the task source and map a workstream folder to it."
tags: configuration, todos, scope, notion
sources:
  - "references/raw/Category request.txt"
last_ingested: 2026-10-14
command: scripts/pickup.py
---

# Add A Category For A Workstream

A workstream without a mapped project value is unscoped, so pickup cannot filter its tasks and handoff skips task writes.
Adding a category fixes both at once: one value in the task source, one `scope_values` entry in the personal profile.
Offer this once when `context` reports `todo_scope_unmapped`, or run it when the user asks for a new category.

1. Run `context --cwd <workstream-folder>` and read the configured `project` field.
   The mapping key is the folder's own directory name; nested folders inherit their nearest mapped ancestor.
2. Fetch the live schema of that field and look for an existing value first.
   Compare case-insensitively, but map the exact stored spelling.
   An existing match needs only step 4; never add a near-duplicate value.
3. If no value exists, confirm the exact new name with the user; the request authorizes this one schema change only.
   For free-text project fields, no schema change is needed.
   For Notion select or multi-select fields, the option-setting statement replaces the whole option list.
   Restate every existing option with its current color, add the new one, and never recolor or drop existing options.
   Re-fetch afterward and confirm the existing options are unchanged and the new value is present.
   Notion may rename a database page template to `Untitled` during this change; check the template names and restore any that changed.
   If the provider or access cannot add values, report that and stop; do not create views, rows, or a replacement source.
4. Add the folder-to-value entry to the full current config and save it with `configure --input`, preserving every other setting.
   Validation failure keeps the previous profile.
5. Re-run `context` in the folder and report the resolved `todo_scope`.

**Incorrect:** adding `Acme` when `ACME` already exists, or writing only the profile entry so pickup filters on a value no task can hold.
**Correct:** the value exists in the live schema, the folder maps to its exact spelling, and `context` resolves it.
