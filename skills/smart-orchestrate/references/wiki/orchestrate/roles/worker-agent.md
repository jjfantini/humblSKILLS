---
title: "The Worker Agent: One Gated Subtask, One Contract Back"
context: orchestrate
category: roles
concept: worker-agent
description: "Bounded worker authority is what makes cheap models safe to dispatch to - and the worker slot is a measured (model, effort) pair the harness must actually honour"
tags: worker, subagent, scope, cost, delegation
sources:
  - "references/raw/orchestrate-SKILL.md"
  - "references/raw/orchestrator-model-sweep-2026-09-10.md"
  - "references/raw/cursor-cli-model-sweep-2026-08-13.md"
  - "references/raw/orchestrator-model-sweep-2026-10-08.md"
  - "references/raw/anthropic-prompting-claude-haiku-5-5-2026-10-08.md"
  - "references/raw/openai-codex-subagents-2026-10-08.md"
  - "references/raw/anthropic-claude-code-sub-agents-2026-10-08.md"
  - "references/raw/anthropic-claude-code-model-config-2026-10-08.md"
  - "references/raw/cursor-forum-openai-models-after-nov-12-2026-10-08.md"
last_ingested: 2026-10-08
---

## A Worker Executes One Brief and Returns One Contract

Prefer the smallest, fastest **`(model, effort)` pair** that can finish the brief
cleanly. Every row below is a measured ID, not a plausible one:

| Slot | Pick | Effort | Evidence |
|---|---|---|---|
| Cheap, Claude Code | `claude-haiku-5-5` | `medium` | 6/6 ping, 9/9 real brief (2026-10-08). Anthropic: `low` "is more likely to skip a search, stop early, or skip a check" in long agent prompts |
| Cheap, Codex | `gpt-6-luna` | `high` (Codex's stated start) | 6/6 ping, 9/9 real brief via stdin. "Use for fast, narrowly scoped agents" |
| Mid-tier, Claude Code | `claude-sonnet-5-5` | `medium`; `high` for harder briefs | 6/6 ping, 9/9 real brief |
| Demanding, Codex | `gpt-6.1-sol` | `medium` | 6/6 ping, 9/9 real brief. "Start here for demanding agents" |
| Hard brief, high blast radius | `claude-opus-5-5` | `low` or `medium` | 6/6. Or keep the subtask on the parent |
| Cursor, cheapest reliable | `gpt-5.3-codex-low-fast` | n/a | 21/21 across all sweeps - **but every Cursor `gpt-*` ID is expected to disappear at OpenAI's proposed 2026-11-12 cutoff** |
| Cursor, hard brief | `claude-opus-5-thinking-high` | n/a | 3/3, 9s (2026-08-13 build) |

**Never put a worker at `ultra`.** On Codex it makes the worker spawn its own
subagents, which breaks parent-owned dispatch and disjoint file scopes - see
`references/wiki/orchestrate/prompting/openai-gpt-6-astra.md`.

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
a question nobody is there to answer; Haiku 5.5 and Sonnet 5.5 at low effort
stop early or report "done" without running a check. Pick the appendix from
`references/wiki/orchestrate/prompting/brief-prompting-by-vendor.md` — the two
vendors need *opposite* delegation nudges, so one shared appendix is wrong for
at least one of them.

## Routing a Worker Inside Claude Code

When the parent is Claude Code and the worker is one of its own subagents, the
model in the brief only takes effect if the harness honours it:

- **A fork ignores `model`.** A fork "sees the same system prompt, tools, model,
  and message history as the main session." Routing a worker to Haiku needs a
  non-fork subagent.
- **Resolution order** is: the per-invocation `model` parameter, then the
  subagent definition's `model` frontmatter, then `CLAUDE_CODE_SUBAGENT_MODEL`,
  then the main conversation's model. Before v2.1.251 the environment variable
  came first; forcing one model now needs `CLAUDE_CODE_SUBAGENT_MODEL_FORCE=1`.
- **Aliases depend on the provider.** On the Anthropic API `haiku` is Haiku 5.5;
  on Amazon Bedrock and Google Cloud it is Haiku 4.5, and `sonnet` is Sonnet 4.5.
  Write the full ID (`claude-haiku-5-5`) in a brief that may run on either.
- **Effort has to be set, not hoped for.** A subagent definition takes an
  `effort` field that "Overrides the session effort level, but not the
  `CLAUDE_CODE_EFFORT_LEVEL` environment variable." The Agent tool in Claude Code
  2.1.294 also accepts a per-call `effort`, but only when a skill or the user asks
  for one - this skill's brief `Effort:` line is that instruction. Without either,
  the worker runs at its model's default.

## Why the Limits Are the Point

Cheap models are safe to dispatch to precisely because their authority is small.
Every constraint removed from a worker is a constraint the parent must re-verify
by hand, which cancels the cost saving that motivated delegation.

## Sources

- `references/raw/orchestrate-SKILL.md` — worker authority and the three things
  a worker owns.
- `references/raw/orchestrator-model-sweep-2026-10-08.md` — Haiku 5.5, Sonnet
  5.5, Opus 5.5, GPT-6 Luna and GPT-6.1 Sol reachability and real-brief results.
- `references/raw/anthropic-prompting-claude-haiku-5-5-2026-10-08.md` — why
  Haiku workers run at `medium`.
- `references/raw/openai-codex-subagents-2026-10-08.md` — Codex's worker tiers
  and starting efforts.
- `references/raw/anthropic-claude-code-sub-agents-2026-10-08.md` and
  `references/raw/anthropic-claude-code-model-config-2026-10-08.md` — fork
  behaviour, model resolution order, provider-dependent aliases.
- `references/raw/cursor-forum-openai-models-after-nov-12-2026-10-08.md` — the
  proposed OpenAI-on-Cursor cutoff.
- `references/raw/cursor-cli-model-sweep-2026-08-13.md` — the Grok and Composer
  scores that retire the old recommendation.
- `references/raw/orchestrator-model-sweep-2026-09-10.md` — the earlier GPT-6
  Astra worker slot.
