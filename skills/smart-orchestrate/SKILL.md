---
name: smart-orchestrate
description: >
  Run a frontier-parent / cost-optimized-worker orchestration pattern across the
  Claude, Codex, and Cursor CLIs: plan on a strong model, isolate in a worktree,
  dispatch bounded briefs to cheaper worker agents, and verify the whole. Use when
  the user says "orchestrate this", "farm this out to subagents", "plan then gate
  execution", "parent orchestrator", "multi-agent this feature", or wants planning
  split from implementation across model tiers. Do NOT use for a single-file edit,
  for authoring commits (use smart-commit), or for the worktree/PR mechanics
  themselves (use smart-worktree-flow).
license: MIT
compatibility: "Requires git 2.5+ for worktree isolation and at least one agent CLI (claude, codex, or cursor-agent) to dispatch workers to. Network access for model calls."
metadata:
  author: jjfantini
  version: "1.2.0"
  category: development
  tags: [orchestration, multi-agent, subagents, planning, worktree, routing, humblskill]
  platforms: [claude-code, cursor, codex]
  preserve:
    - references/raw/
    - references/wiki/
    - references/decisions.md
    - references/log.md
    - references/patterns.md
---

# Orchestrate

Act as (or under) a **parent orchestrator**: plan on a frontier model, execute on
cheaper models, verify the whole. Prefer **Claude** as the entry point; **Codex**
or **Cursor** can also be the parent, and any parent may farm work to agents
running through Claude, the Cursor CLI, or the Codex plugin.

This skill is model-invocable: the description's trigger phrases ("orchestrate
this", "farm this out to subagents", "parent orchestrator", etc.) let the
model start a run without the user typing `/smart-orchestrate`. A run still
spends frontier tokens and creates worktrees, so match the request to a real
multi-phase or multi-agent need before invoking - don't reach for this on a
single-file edit.

## Brain Protocol (read BEFORE orchestrating anything)

1. `references/_index.md` - what this skill knows (map)
2. `references/patterns.md` - which routing choices worked, with numbers
3. `references/decisions.md` - past reasoning, don't repeat mistakes
4. `references/log.md` - last 5 session entries
5. Relevant `references/wiki/orchestrate/<category>/` concepts per task

After completing work, UPDATE the brain:
- Routing outcomes with numbers (model tier, retries, wall clock) -> `patterns.md`
- Non-obvious calls (why a task stayed on the parent) -> `decisions.md`
- Session summary (always) -> append to `log.md`
- Run `bash scripts/lint.sh` to regenerate `_index.md` and verify structure

_Full spec: `references/_brain.md`._

## CCCCC Architecture

| Layer | Role | Location |
|---|---|---|
| Core | Root structure of the skill | `SKILL.md`, `references/`, `scripts/` |
| Context | Top-level taxonomy grouping | First segment under `references/wiki/` |
| Category | Specific topic within a context | Second segment under `references/wiki/` |
| Concept | One atomic idea per file | Filename stem and frontmatter field |
| Command | Deterministic executable script | `scripts/<command>.sh` |

## When to Use

- A feature or refactor is large enough to split into gated phases
- Planning deserves a frontier model but implementation does not
- Work should be farmed to subagents or to another CLI (Codex, Cursor)
- Parallel agents need non-colliding file scopes inside an isolated worktree
- A long-running goal or loop needs a single owner holding context across workers

## How to Use

**Live enumeration of categories and concepts:**
Read `references/_index.md` after running `bash scripts/lint.sh`.

**What the parent owns and which model tier to run it on:**
Read `references/wiki/orchestrate/roles/parent-orchestrator.md`.

**What a worker owns, and the hard limits on worker authority:**
Read `references/wiki/orchestrate/roles/worker-agent.md`.

**Isolate before dispatch:**
Read `references/wiki/orchestrate/isolation/worktree-first.md`, then follow
`smart-worktree-flow` to create the paired worktree and branch.

**Pick a model tier for a subtask, or decide to escalate:**
Read `references/wiki/orchestrate/routing/model-selection.md`.

**Dispatch to a Cursor CLI worker — which `--model` IDs actually work:**
Read `references/wiki/orchestrate/routing/cursor-cli-models.md`, then dispatch
with `scripts/dispatch-cursor-worker.sh`. Never `--model auto` (measured 0/12).

**Run the end-to-end orchestration loop:**
Read `references/wiki/orchestrate/loop/session-loop.md`.

**Write the brief you hand a worker:**
Read `references/wiki/orchestrate/contracts/brief-template.md`.

**Accept or reject what a worker returns:**
Read `references/wiki/orchestrate/contracts/handoff-contract.md`.

**Close a gate and ship:**
Read `references/wiki/orchestrate/closeout/commit-and-ship.md`, then run
`smart-commit` per gate and `smart-worktree-flow` to ship.

**Check yourself against known failure modes:**
Read `references/wiki/orchestrate/anti-patterns/avoid.md`.

## Examples

### Example 1: Feature farmed to cheap workers

User says: "Orchestrate this — add rate limiting to the API, use subagents."

Actions:
1. Read the brain, then `roles/parent-orchestrator.md` and `loop/session-loop.md`.
2. Clarify the goal and success criteria; break it into three gated phases
   (middleware, config surface, tests).
3. Follow `smart-worktree-flow` to create `feat-rate-limit` / `feat/rate-limit`.
4. Route per `routing/model-selection.md`: middleware to a mid-tier worker,
   config and tests to the cheapest worker, the shared-contract review stays
   on the parent.
5. Dispatch one brief per subtask using `contracts/brief-template.md`, each
   scoped to non-overlapping files inside that worktree.
6. Collect handoffs; reject any reply that is not the contract shape.
7. Verify diffs against the plan, then run `smart-commit` to close each gate.

Result: The feature lands in one worktree, one coherent design, frontier tokens
spent only on planning, integration, and verification.

### Example 2: Escalation mid-run

User says: "Keep orchestrating, but the auth worker came back confused."

Actions:
1. Read `contracts/handoff-contract.md` — the return says `needs-escalation`.
2. Apply `routing/model-selection.md`: the brief touches a shared auth contract,
   so it was mis-routed to the cheapest tier.
3. Re-scope the brief with the missing constraints, and either re-dispatch to a
   stronger worker or keep the subtask on the parent.
4. Append the mis-route to `decisions.md` so the next run routes auth higher.

Result: The gate reopens with a correct brief instead of a cheap worker guessing
at a shared contract.

## Troubleshooting

**A worker returns prose instead of the handoff contract**
Cause: the brief omitted `Return: handoff contract (required)`, or the worker
model is too small to hold the format.
Fix: re-dispatch with the literal contract block pasted into the brief. Treat the
free-form reply as a failed gate — do not hand-summarise it into the plan.

**Two parallel workers edited the same file**
Cause: overlapping file scopes inside one worktree.
Fix: give each parallel worker a disjoint file list, or a separate worktree. See
`references/wiki/orchestrate/isolation/worktree-first.md`.

**Workers committed or opened a PR**
Cause: the brief's `Do NOT` line was dropped.
Fix: parent alone owns commits and ship steps. Reset the worker's commits into the
working tree and re-run `smart-commit` from the parent.

**The parent is doing all the implementation itself**
Cause: no gating — the plan never got split into bounded briefs.
Fix: stop, write the phase list, and route each phase per
`references/wiki/orchestrate/routing/model-selection.md`.

## Success Signals

- Every worker ran inside a worktree, never the user's main checkout
- Every subtask had scope, brief, model tier, done-when, and a return contract
- Every worker reply matched the handoff contract, or was rejected and re-dispatched
- No parallel workers shared a file without a named merge owner
- Only the parent committed, opened PRs, or merged
- Frontier tokens went to planning, integration, and verification — not mechanical edits
- `bash scripts/lint.sh` exits 0 and `log.md` gained exactly one entry for the run
