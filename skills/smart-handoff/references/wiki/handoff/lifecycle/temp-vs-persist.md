---
title: "Keep Handoffs In Their Starting Workstream Folder"
context: handoff
category: lifecycle
concept: temp-vs-persist
description: "Use the same starting-folder contract for producing and consuming handoffs."
tags: lifecycle, paths, workstream, persist, cleanup
sources:
  - "references/raw/Tandem request.txt"
last_ingested: 2026-10-14
command: scripts/preflight.sh
---

## Resolve The Starting Folder First

A Git root and a session workstream can differ.
Saving under the Git root makes a handoff invisible to pickup launched in a nested workstream.
Capture the session's starting folder from the user/session context before changing directories.
Persist there by default, unless the user supplies another destination or requests a temporary handoff.

| Mode | Path | Lifecycle |
|------|------|-----------|
| **persist** (default) | `<session-start-folder>/.humblskills/handoffs/<slug>-handoff-<date>.md` | durable workstream record |
| **temp** (explicit) | `/tmp/.humblskills/handoffs/<slug>-handoff-<date>.md` | explicit-path transfer |

Run the skill's `scripts/preflight.sh --cwd <session-start-folder> --slug <name>`.
`SESSION_START_DIR`, `PERSIST_DIR`, and `PERSIST_FILE` identify the workstream; `REPO_ROOT` describes Git state separately.
Passing `--cwd` keeps the destination stable when the agent has moved into a subject repository or the skill directory.
Normal preflight lists existing handoffs only from the workstream's own folder, not a global temporary directory.

## Persisted Handoffs And Pickup

Write to the resolved `PERSIST_FILE` and include the session-start-folder row in the document.
Use absolute artifact paths when work spans repositories.
Normal smart-pickup in the same starting folder selects that folder's newest matching handoff.
It does not automatically select from ancestors, siblings, descendants, or temporary folders.
If the local folder has no match, pickup asks rather than silently switching workstreams.
Do not move or rewrite old misplaced handoffs; link a relevant predecessor explicitly when needed.

Do not change `.gitignore` automatically.
Preflight reports whether the actual workstream handoff directory is ignored; otherwise report that the file will appear in Git status.
Persisted handoffs have no delete-after-use banner.

## Explicit Temporary Transfer

Only use temporary storage when requested.
Include the exact file path in the receiving prompt because normal pickup never scans `/tmp`.
If smart-pickup is installed, the receiving agent can invoke it with that explicit path.
Do not delete the document during orientation; cleanup occurs after the requested continuation finishes if authorized by the temporary lifecycle.

## Non-Git Starting Folders

The folder contract is identical without Git.
`REPO_ROOT=none` does not invalidate the starting folder or require another lifecycle question.
Use absolute paths in the document.

**Incorrect:** starting in a nested billing workstream but writing its handoff at the repository root.
**Correct:** persist in billing's own `.humblskills/handoffs/`, then launch pickup in billing.
