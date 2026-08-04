---
title: "Temp vs Persist - The One Question You Must Ask"
context: handoff
category: lifecycle
concept: temp-vs-persist
description: "Most handoffs are disposable; writing them into a repo by default pollutes history with stale context nobody deletes"
tags: lifecycle, paths, temp, persist, cleanup, gitignore
sources: []
last_ingested: 2026-08-03
command: scripts/preflight.sh
---

## Ask Before Writing

A handoff has exactly two lifecycles, and the user picks. Ask once, up front,
unless the invocation already answered it:

| Mode | Path | Lifecycle |
|------|------|-----------|
| **temp** (default) | `/tmp/.humblskills/handoffs/<slug>-handoff-<date>.md` | created, consumed, deleted |
| **persist** | `<repo>/.humblskills/handoffs/<slug>-handoff-<date>.md` | kept as a durable record |

`scripts/preflight.sh --slug <name>` prints both resolved paths as
`TEMP_FILE=` / `PERSIST_FILE=` so there is nothing to assemble by hand.

Signals that answer the question without asking: "temporary", "throwaway",
"just for this handoff", "delete after" → temp. "keep this", "commit it",
"save it in the repo", "for the team" → persist.

## Temp Mode

```bash
mkdir -p /tmp/.humblskills/handoffs
```

`/tmp` survives across processes but not reboots, which is the correct TTL for
a document meant to be read once. Keep the lifecycle banner in the doc so the
receiving agent knows to delete it:

```markdown
> **Lifecycle:** temporary — delete this file once you have picked up the work.
```

Deletion is the receiving agent's job, not a background timer. Say so in the
Pickup Prompt:

```
Read /tmp/.humblskills/handoffs/oauth-token-refresh-handoff-2026-08-03.md,
continue from "Next Steps", then delete the file.
```

## Persist Mode

```bash
mkdir -p "$REPO_ROOT/.humblskills/handoffs"
```

Two things to get right:

1. **Do not touch `.gitignore` on your own.** `preflight.sh` reports
   `PERSIST_DIR_GITIGNORED=yes|no`. If it is `no`, tell the user the file will
   show up in `git status` and let them decide whether to commit it or ignore
   it. Silently adding an ignore rule hides a document they asked to keep.
2. **Drop the delete-after-use banner.** Replace it with the reason it is being
   kept, so a reader six weeks later knows whether it is still live:
   ```markdown
   > **Lifecycle:** persisted — record of the OAuth migration handoff to Codex.
   ```

## Non-Git Directories

If `preflight.sh` reports `REPO_ROOT=none`, persist mode falls back to
`$PWD/.humblskills/handoffs`. Confirm that is what the user meant before
writing, and use absolute paths throughout the Artifacts section — relative
paths are meaningless to an agent started from a different directory.

### Incorrect

```markdown
| Spec | `./docs/spec.md` | the contract |
```

### Correct

```markdown
| Spec | `/Users/dev/projects/api/docs/spec.md` | the contract |
```

## Sources

- (none) - authored from the smart-handoff design session.
