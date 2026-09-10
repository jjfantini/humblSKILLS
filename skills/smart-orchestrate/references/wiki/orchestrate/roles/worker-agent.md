---
title: "The Worker Agent: One Gated Subtask, One Contract Back"
context: orchestrate
category: roles
concept: worker-agent
description: "Bounded worker authority is what makes cheap models safe to dispatch to"
tags: worker, subagent, scope, cost, delegation
sources:
  - "references/raw/orchestrate-SKILL.md"
  - "references/raw/orchestrator-model-sweep-2026-09-10.md"
  - "references/raw/cursor-cli-model-sweep-2026-08-13.md"
last_ingested: 2026-09-10
---

## A Worker Executes One Brief and Returns One Contract

Prefer the smallest, fastest **`(model, effort)` pair** that can finish the brief
cleanly. On the Cursor CLI these must be measured IDs, not plausible ones:

| Slot | Pick | Evidence |
|---|---|---|
| Cheapest reliable | `gpt-5.3-codex-low-fast` | 21/21 across all sweeps |
| Fastest | `gpt-5.6-luna-high` | 3/3, 5s |
| Hard brief | `claude-opus-5-thinking-high` | 3/3, 9s |
| Mid-tier, non-Cursor | Fable 5.1 at `effort=low` or `medium` | Anthropic: at `low` "often competitive with Claude Opus and Claude Sonnet models on cost per task while scoring higher" |
| Non-Cursor, needs autonomy | `gpt-6-astra` at `low` via `codex -m` | 4/4, 6-7s |

**Do not dispatch to Grok 4.6 or Composer 2.5.** An earlier version of this
concept recommended both as cheap worker slots; the measurements in
`references/wiki/orchestrate/routing/cursor-cli-models.md` contradict that
outright — Grok 4.6 scores ~3/24 across all eight tiers and `composer-2.5` 1/4.
`cursor-grok-4.5-low-fast` (9/9) is the only safe Grok ID of fourteen tested, and
it sits inside an otherwise broken family, so prefer a GPT ID for anything
load-bearing.

Reserve the *strong* worker slots for briefs with real blast radius, and set
effort explicitly — a mechanical brief inherits the harness default (usually
`high`) and quietly pays frontier thinking for a rename.

**Incorrect (unbounded worker):**

```
Parent: "Take the rate limiting work."
Worker: refactors the router, adds a config system, commits three times,
        opens a PR, and replies with a 4,000-token transcript.
```

The parent now owns a design it never approved, in commits it did not write, and
must read a transcript to find out what happened.

**Correct (bounded worker):**

Owns exactly three things:

1. Execute one gated subtask from the parent's brief, inside the assigned worktree
2. Stay inside the given scope and files
3. Return the handoff contract — summary only, not a full transcript

Workers do **not** re-plan the whole feature, expand scope, commit, open PRs, or
touch the main checkout without escalating.

## The Worker's Default Failure Is Vendor-Specific

Bounded authority is necessary but not sufficient: each backend under-delivers
in its own documented way, and the fix is a short appendix on the brief. Claude
Fable 5.1 ends a turn describing what it would do next; GPT-6 Astra stops to ask
a question nobody is there to answer. Pick the appendix from
`references/wiki/orchestrate/prompting/brief-prompting-by-vendor.md` — the two
vendors need *opposite* delegation nudges, so one shared appendix is wrong for
at least one of them.

## Why the Limits Are the Point

Cheap models are safe to dispatch to precisely because their authority is small.
Every constraint removed from a worker is a constraint the parent must re-verify
by hand, which cancels the cost saving that motivated delegation.

## Sources

- `references/raw/orchestrate-SKILL.md` — worker authority and the three things
  a worker owns.
- `references/raw/cursor-cli-model-sweep-2026-08-13.md` — the Grok and Composer
  scores that retire the old recommendation.
- `references/raw/orchestrator-model-sweep-2026-09-10.md` — the GPT-6 Astra
  worker slot.
