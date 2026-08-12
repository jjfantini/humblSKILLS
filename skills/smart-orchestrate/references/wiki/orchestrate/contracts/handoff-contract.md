---
title: "The Worker-to-Parent Handoff Contract"
context: orchestrate
category: contracts
concept: handoff-contract
description: "A fixed return shape is the only worker context worth keeping, and the cheapest to verify"
tags: handoff, return, contract, verification, escalation
sources:
  - "references/raw/orchestrate-SKILL.md"
last_ingested: 2026-08-12
---

## The Contract

Every worker must return exactly this shape. Incomplete = failed gate.

```text
Status: done | blocked | needs-escalation
Files touched:
- …
Commands run:
- …
Verify result: pass | fail — <one line>
Summary: <2-4 sentences of what changed>
Risks / open questions:
- …
Escalate?: no | yes — <why>
```

The parent treats this as the **only** worker context worth keeping — not the full
worker transcript.

## Why a Free-Form Reply Is a Failed Gate

**Incorrect (accepted prose):**

```
Worker: "Done! I added the limiter and it seems to work. I also cleaned up a
         couple of things in the router while I was in there."
```

Three unanswerable questions: which files, was verify run, is the router change in
scope. The parent must now diff and re-derive everything the contract would have
stated — the delegation saved nothing.

**Correct (contract return):**

```text
Status: done
Files touched:
- internal/middleware/ratelimit.go
- internal/middleware/chain.go
Commands run:
- go build ./...
- go vet ./internal/middleware/...
Verify result: pass — build and vet clean
Summary: Added a per-key token-bucket limiter using golang.org/x/time/rate and
  wired it into the middleware chain ahead of auth. No new dependencies. Limits
  are read from the existing config struct; no config surface was added.
Risks / open questions:
- Limiter map is unbounded; a key-eviction policy is not in scope here.
Escalate?: no
```

## Handling Each Status

| Status | Parent action |
|---|---|
| `done` | Verify the diff against the plan, then close the gate with `smart-commit` |
| `blocked` | Read the blocker; fix the brief or the environment, re-dispatch |
| `needs-escalation` | Re-route per `references/wiki/orchestrate/routing/model-selection.md` — usually to a stronger worker or the parent |

A `Verify result: fail` with `Status: done` is a contradiction — treat it as
`blocked` and re-dispatch.

## Sources

- `references/raw/orchestrate-SKILL.md`
