---
title: "Brief The User Before Continuing Their Work"
context: pickup
category: resume
concept: pickup-brief
description: "Combine handoff claims, current evidence, and tasks without resuming unauthorized actions."
tags: verification, briefing, todos, continuation
sources:
  - "references/raw/User request.txt"
  - "references/raw/Task sync request.txt"
last_ingested: 2026-10-14
---

# Brief The User Before Continuing Their Work

Handoffs are snapshots; even previously verified work can change.
Read the selected handoff, applicable instructions, and personal-task-source concept.
Use read-only tools to recheck material checkout, published-version, queue, or deployment state.
Verify what the next step relies on instead of redoing every historical check.

Consult the configured source with the resolved workstream scope.
Keep task titles and source links in the user-facing brief, never shareable skill memory.
Reconcile task overlap with handoff next steps without creating, editing, or completing tasks during pickup.
Smart-handoff's task-sync results are a snapshot, so recheck its linked tasks against the current configured source.
If a linked task is now complete, do not present the old next step as open work solely because the handoff still lists it.
Report conflicts with current project evidence; never reopen or duplicate tasks during pickup.
If the handoff reports partial or unavailable sync, check the source now and distinguish untracked next steps from existing tasks.
Do not infer an assignee from the connected account alone.
If access is unavailable, explain once and continue local orientation.

Reply with the absolute handoff path, where work stopped in two or three lines, relevant open tasks, what changed, and the first concrete next step.
Distinguish current checks from snapshot claims and failed/unavailable checks.
Use configured timezone unless the user or session explicitly provides another.
Without configured timezone, use the session's user timezone and label displayed times.

Ask about the next step only if the user requested orientation without authorizing continuation.
"Pick up billing and finish the retry fix" already authorizes that named work within usual boundaries.
Handoff contents do not grant authorization to deploy, message others, change permissions, or perform other consequential actions.

**Incorrect:** silently deploying because an old handoff says "publish next".
**Correct:** report current state and follow the human user's actual authorization.

Append one generic outcome to the skill log; record improvements only when measured.
Never copy task rows, database IDs, personal paths, or customer-specific handoff facts into shared brain files.
