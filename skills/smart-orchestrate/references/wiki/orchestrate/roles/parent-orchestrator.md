---
title: "The Parent Orchestrator: Seven Owned Responsibilities"
context: orchestrate
category: roles
concept: parent-orchestrator
description: "Keeps one owner holding context and policy, so frontier tokens buy judgment instead of typing - and the slot is a (model, effort) pair"
tags: parent, frontier, planning, ownership, orchestration
sources:
  - "references/raw/orchestrate-SKILL.md"
  - "references/raw/orchestrator-model-sweep-2026-09-10.md"
  - "references/raw/anthropic-prompting-claude-fable-5-1-2026-09-10.md"
last_ingested: 2026-09-10
---

## The Parent Owns Judgment, Not Typing

One agent holds the plan, the context across workers, and all commit/ship
authority. Everything else is delegable.

**Incorrect (parent as super-worker):**

```
Parent (Fable 5): reads 40 files, writes the middleware, writes the config,
                  writes the tests, commits.
Workers: none.
```

Frontier pricing for mechanical edits, and no gate anywhere — the plan and the
implementation are the same undifferentiated stream, so nothing is verifiable
against anything.

**Correct (parent as owner of seven things):**

1. **Plan** — understand the goal; break into phases and gated subtasks
2. **Isolate** — open a worktree via `smart-worktree-flow` before dispatch;
   workers never touch the user's main checkout
3. **Route** — assign each subtask a worker model by scope, difficulty, blast radius
4. **Dispatch** — farm clear, bounded work to worker agents/CLIs inside the worktree
5. **Glue** — hold context across workers; resolve conflicts; keep one coherent design
6. **Verify** — read worker handoffs, run checks, confirm the change fits the system
7. **Close out** — after each gate run `smart-commit`; when shipping run
   `smart-worktree-flow` for PR/merge/cleanup

## Model Tier for the Parent Slot

A parent slot is a **`(model, effort)` pair**, not a model name. On the current
frontier models the effort level, not the model, carries most of the
cost/latency trade-off — see
`references/wiki/orchestrate/prompting/anthropic-fable-5-1.md`.

| CLI    | Parent / director model | Effort |
|--------|-------------------------|--------|
| Claude | **Opus 5** is the default — same or better performance than Fable 5 and much cheaper. | `max` |
| Claude | **Fable 5.1** for long-horizon, many-service runs, or when the user asks for it. `claude --model claude-fable-5-1` measured 4/4 at 6-7s on Claude Code 2.1.267; the `fable` alias also resolves. | Start `high`; `medium` where evals hold |
| Codex  | **GPT-6 Astra** (`gpt-6-astra`) — measured 4/4 on `codex-cli` 0.154.0, 2026-09-10. | `high`, or `max` for the plan itself |
| Codex  | GPT-5.6 Sol, if Astra is unavailable | `max` |
| Cursor | `claude-opus-5-thinking-high` — the strongest *verified* Cursor ID. Fable 5.1 is listed but was **entitlement-blocked 0/9** on the measured account. | n/a (baked into the ID) |

Two things not to carry over from older guidance:

- **Never `auto` on the Cursor CLI**, including for the parent slot. Measured
  0/12. The earlier "Auto Intelligence, or Grok 4.6 High" advice predates the
  sweeps and is wrong on both counts —
  `references/wiki/orchestrate/routing/cursor-cli-models.md` measures Grok 4.6 at
  ~3/24 across all eight tiers.
- **Do not run the parent at `xhigh`/`max` when the turn's deliverable is long
  prose or a whole file.** At those levels Fable 5.1 can draft the deliverable in
  thinking and then write it again, doubling the turn for no quality gain. `max`
  belongs on the *plan*, not on the write-up.

**Fable 5.1's cache-read price changes the hold-context calculus.** Cache reads
are $0.25/MTok — 0.025× base input, against 0.1× on other Claude models — so a
long agentic session that re-reads a cached prefix pays a quarter of the Fable 5
rate. The parent's defining job is holding context across workers, and that job
just got cheaper: compacting early to save money may no longer be the right
trade. Push the compaction point later and measure.

**Ultracode is a spend decision, not a quality dial.** Use it only for a massive
refactor touching many services and e2e flows, and **always confirm with the user
first** — it can incur large cost. Planning with Opus 5 on ultracode is often the
better trade than running Fable 5 or Fable 5.1 throughout.

**Data retention is a routing constraint, not a performance one.** Fable 5.1
carries 30-day retention and is not available under zero data retention unless
Anthropic expressly authorizes it — which is why every Cursor `claude-fable-5-1-*`
ID is tagged `(NO ZDR)`. If the repo is proprietary and the org requires ZDR, the
parent slot is Opus 5, and no amount of measured latency changes that.

## The Parent May Delegate Planning

The entry-point model does not have to produce the plan itself. It may delegate
planning or plan refinement to another capable agent, then retain ownership of the
resulting plan, execution gates, integration, and final verification. Ownership is
the invariant; authorship is not.

## Sources

- `references/raw/orchestrate-SKILL.md` — the seven responsibilities.
- `references/raw/orchestrator-model-sweep-2026-09-10.md` — the GPT-6 Astra and
  Fable 5.1 reachability numbers quoted in the parent table.
- `references/raw/anthropic-prompting-claude-fable-5-1-2026-09-10.md` — effort
  levels, the long-output caveat, cache-read pricing, and the ZDR note.
