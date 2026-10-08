---
title: "Fourteen Orchestration Anti-Patterns and What Each Costs"
context: orchestrate
category: anti-patterns
concept: avoid
description: "Each anti-pattern names the concrete failure it produces, so it is recognisable mid-run"
tags: anti-patterns, failure-modes, review, cost, scope
sources:
  - "references/raw/orchestrate-SKILL.md"
  - "references/raw/anthropic-prompting-claude-fable-5-1-2026-09-10.md"
  - "references/raw/openai-gpt-6-astra-latest-model-2026-09-10.md"
  - "references/raw/anthropic-optimizing-for-cost-and-intelligence-2026-10-08.md"
  - "references/raw/openai-codex-models-2026-10-08.md"
  - "references/raw/cursor-cli-changelog-2026-10-08.md"
  - "references/raw/cursor-model-claude-fable-5-1-2026-10-08.md"
last_ingested: 2026-10-08
---

## The Fourteen

| Anti-pattern | What it actually costs |
|---|---|
| Parent implementing everything itself on the frontier model | Frontier pricing for mechanical edits, and no gate to verify against |
| Workers inventing new phases or rewriting the plan | The parent's plan and the code diverge with nothing flagging it |
| Workers committing, opening PRs, or editing the main checkout | Incoherent history, release tooling misreads it, user's tree in the blast radius |
| Parallel workers on the same files without a merge owner | Interleaved diffs nobody owns; last writer silently wins |
| Farming across CLIs without a worktree | One shared dirty tree; `git status` stops meaning anything |
| Accepting a free-form worker reply instead of the handoff contract | The parent re-derives files, verify status, and scope by hand |
| Skipping parent verification or `smart-commit` between phases | No verified baseline, so a late phase can invalidate an early one undetected |
| Sending vague, high-blast-radius work to the cheapest model | The worker guesses at a shared contract; the round trip costs more than routing right |
| Naming a model in the brief but not an effort level | The dominant cost variable stays at the harness default (usually `high`), so a mechanical rename pays frontier thinking |
| Reusing one prompt appendix across vendors | The two frontier families need **opposite** delegation nudges; a shared "don't spawn subagents" line brakes the model that was already under-delegating |
| Adopting a model ID because `--list-models` or a docs page lists it | Listing proves neither validity nor entitlement. Measured 2026-09-10: all ten `claude-fable-5-1-*` IDs listed, 0/9 on dispatch |
| Orchestrating work one model could do at lower effort | Anthropic measured it: "On work a single model could handle alone, the same model at lower effort was cheaper every time." The coordination overhead buys nothing |
| Running a worker at Codex `ultra` | The worker spawns its own subagents - on its own model unless configured - so parent-owned dispatch and disjoint file scopes both break |
| Treating a passing ping as proof of which model answered | Cursor builds before v2026.09.28 could silently run a shorter base model, and Cursor reroutes guardrail-tripping Fable requests to Opus. Read the `model` field of the init event |

## The Pattern Behind the Pattern

Seven of the fourteen are the same mistake: **a boundary that was supposed to be
explicit was left implicit**. Scope, file ownership, commit authority, return
shape, and now effort level are all cheap to state at dispatch time and
expensive to reconstruct afterwards.

Two — parent-does-everything and skipped verification — are the degenerate cases
where orchestration collapses back into a single undifferentiated stream, at
which point the frontier model is paying for typing. Their mirror image is
orchestrating work that never needed splitting: the cheapest correct answer is
often one model with its effort turned down.

Three more - the shared appendix, the listed-not-measured ID, and the
unverified ping - share a root: **treating a model as a name rather than a
measured, entitled, prompt-shaped backend.** A vendor's docs tell you how
a model behaves; only a ping tells you whether your account can reach it, and
only the vendor's own guide tells you which default it will under-deliver on.

## Sources

- `references/raw/orchestrate-SKILL.md` — the original eight.
- `references/raw/anthropic-prompting-claude-fable-5-1-2026-09-10.md` and
  `references/raw/openai-gpt-6-astra-latest-model-2026-09-10.md` — the opposite
  delegation defaults behind the shared-appendix row.
- `references/raw/anthropic-optimizing-for-cost-and-intelligence-2026-10-08.md`
  — lower effort beats orchestrating single-model work.
- `references/raw/openai-codex-models-2026-10-08.md` — what `ultra` does.
- `references/raw/cursor-cli-changelog-2026-10-08.md` and
  `references/raw/cursor-model-claude-fable-5-1-2026-10-08.md` — the prefix-ID
  bug and the silent Fable-to-Opus reroute.
