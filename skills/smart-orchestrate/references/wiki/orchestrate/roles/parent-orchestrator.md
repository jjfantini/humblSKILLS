---
title: "The Parent Orchestrator: Seven Owned Responsibilities"
context: orchestrate
category: roles
concept: parent-orchestrator
description: "Keeps one owner holding context and policy, so frontier tokens buy judgment instead of typing - the slot is a (model, effort) pair, and the first decision is whether to orchestrate at all"
tags: parent, frontier, planning, ownership, orchestration
sources:
  - "references/raw/orchestrate-SKILL.md"
  - "references/raw/orchestrator-model-sweep-2026-09-10.md"
  - "references/raw/anthropic-prompting-claude-fable-5-1-2026-09-10.md"
  - "references/raw/orchestrator-model-sweep-2026-10-08.md"
  - "references/raw/anthropic-models-overview-2026-10-08.md"
  - "references/raw/anthropic-prompting-claude-opus-5-5-2026-10-08.md"
  - "references/raw/anthropic-optimizing-for-cost-and-intelligence-2026-10-08.md"
  - "references/raw/anthropic-api-and-data-retention-2026-10-08.md"
  - "references/raw/openai-codex-models-2026-10-08.md"
  - "references/raw/cursor-model-claude-opus-5-5-2026-10-08.md"
last_ingested: 2026-10-08
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

## First Gate: Should This Be Orchestrated at All?

Anthropic measured the pattern this skill implements and published where it
loses: "If the work is one chain, fits in one context without a long cost tail,
or a single model at lower effort already meets your bar, don't build an
orchestrator." And: "On work a single model could handle alone, the same model
at lower effort was cheaper every time."

Where it wins, it wins on cost and wall clock, not score. On a 21.6M-token
corpus too large for any context window, a Fable 5.1 lead over 25 Sonnet 5
workers "cost about half as much as those settings (47% to 55% less) and scored
10 to 12 points below them, in about 2.3 hours per episode against 15 to 20."
And "Delegation paid on the routine, normally solvable share of the work, the
opposite of the intuition that workers are for hard problems."

So before step 1, ask: does the work fan out across genuinely independent files
or cases, or exceed one context window? If not, run one model and turn its
effort down instead. That is not a failure of this skill - it is the first
routing decision.

## Model Tier for the Parent Slot

A parent slot is a **`(model, effort)` pair**, not a model name. On the current
frontier models the effort level, not the model, carries most of the
cost/latency trade-off - see
`references/wiki/orchestrate/prompting/anthropic-claude-5-5.md`.

| CLI    | Parent / director model | Effort |
|--------|-------------------------|--------|
| Claude | **Opus 5.5** (`claude-opus-5-5`) is the default - Anthropic: "start with Claude Opus 5.5 for most workloads." $4 / $20 per MTok; not a Covered Model, so Fable's retention veto does not apply (Cursor states it is "Zero Data Retention compatible"). 6/6 on Claude Code 2.1.294, 2026-10-08. | `medium` (its default); `high` for the plan when evals justify it |
| Claude | **Fable 5.1** (`claude-fable-5-1`) for "demanding reasoning and long-horizon agentic work, or when your evals on Claude Opus 5.5 at higher effort still fall short." 2.5x Opus 5.5's per-token price; 30-day retention, no ZDR. | Start `high`; `medium` where evals hold |
| Codex  | **GPT-6.1 Sol** (`gpt-6.1-sol`) is the default - OpenAI's Codex recommendation "for complex coding and agentic workflows." 6/6 on `codex-cli` 0.160.0. | `medium` (its default); `high` for the plan |
| Codex  | **GPT-6 Astra** (`gpt-6-astra`) for the hardest, most ambiguous plans - "the strongest capability across steps and tools," at 5x 6.1 Sol's per-token price. 9/9 as the sweep control. | Codex's stated start is `low`; raise for the plan |
| Cursor | `claude-opus-5-thinking-high` remains the strongest *measured* Cursor ID. Opus 5.5 is documented on Cursor (ZDR-compatible, recommended as "a coordinator for subagents") but was **not measured** - auth expired on 2026-10-08. | n/a (baked into the ID) |

Three things not to carry over from older guidance:

- **"Opus 5 is the default parent" and "parent at `max`" are retired.** Opus 5 is
  now a legacy model. For Opus 5.5 Anthropic's rule is "Reserve `xhigh` and `max`
  for work where you've measured a quality gain," and at a given level it thinks
  more per turn than Opus 5 did. The reasoning is recorded in `decisions.md`.
- **Never `auto` on the Cursor CLI**, including for the parent slot. Measured
  0/12. The earlier "Auto Intelligence, or Grok 4.6 High" advice predates the
  sweeps and is wrong on both counts -
  `references/wiki/orchestrate/routing/cursor-cli-models.md` measures Grok 4.6 at
  ~3/24 across all eight tiers.
- **Do not run the parent at `xhigh`/`max` when the turn's deliverable is long
  prose or a whole file.** At those levels Fable 5.1 can draft the deliverable in
  thinking and then write it again, doubling the turn for no quality gain.

**An unattended Opus 5.5 parent can end a turn on a text-only progress update.**
A harness that reads that as "done" stops the run with workers in flight. Keep
the gate checklist outside the model, nudge on open items, and cap automatic
continuations at two or three - see `anthropic-claude-5-5.md`.

**Do not hand `ultra` to anything but a Codex parent.** It makes the model spawn
its own subagents, which inherit the parent's model unless
`agents.default_subagent_model` says otherwise - see `openai-gpt-6-astra.md`.

**Fable 5.1's cache-read price changes the hold-context calculus.** Cache reads
are $0.25/MTok — 0.025× base input, against 0.1× on other Claude models — so a
long agentic session that re-reads a cached prefix pays a quarter of the Fable 5
rate. The parent's defining job is holding context across workers, and that job
just got cheaper: compacting early to save money may no longer be the right
trade. Push the compaction point later and measure.

**Ultracode is a spend decision, not a quality dial.** Use it only for a massive
refactor touching many services and e2e flows, and **always confirm with the user
first** — it can incur large cost. Planning with Opus 5.5 on ultracode is often
the better trade than running Fable 5.1 throughout.

**Data retention is a routing constraint, not a performance one.** Fable 5.1
carries 30-day retention and is not available under zero data retention unless
Anthropic expressly authorizes it — which is why every Cursor `claude-fable-5-1-*`
ID is tagged `(NO ZDR)`. If the repo is proprietary and the org requires ZDR, the
parent slot is Opus 5.5, and no amount of measured latency changes that.

## The Parent May Delegate Planning

The entry-point model does not have to produce the plan itself. It may delegate
planning or plan refinement to another capable agent, then retain ownership of the
resulting plan, execution gates, integration, and final verification. Ownership is
the invariant; authorship is not.

## Sources

- `references/raw/orchestrate-SKILL.md` — the seven responsibilities.
- `references/raw/orchestrator-model-sweep-2026-10-08.md` — Opus 5.5, GPT-6.1
  Sol and GPT-6 Astra reachability; Cursor not re-measured.
- `references/raw/orchestrator-model-sweep-2026-09-10.md` — the earlier Astra
  and Fable 5.1 numbers.
- `references/raw/anthropic-models-overview-2026-10-08.md` — "start with Claude
  Opus 5.5", prices, default efforts.
- `references/raw/anthropic-prompting-claude-opus-5-5-2026-10-08.md` — effort
  guidance and text-only end-of-turn behaviour.
- `references/raw/anthropic-optimizing-for-cost-and-intelligence-2026-10-08.md`
  — when not to orchestrate, and the measured orchestrator trade-off.
- `references/raw/anthropic-prompting-claude-fable-5-1-2026-09-10.md` — the
  long-output caveat and cache-read pricing.
- `references/raw/anthropic-api-and-data-retention-2026-10-08.md` — the Covered
  Models ZDR rule.
- `references/raw/openai-codex-models-2026-10-08.md` — Codex's default model and
  when to choose Astra.
- `references/raw/cursor-model-claude-opus-5-5-2026-10-08.md` — Opus 5.5 on
  Cursor, ZDR-compatible.
