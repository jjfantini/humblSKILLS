---
title: "Eight Orchestration Anti-Patterns and What Each Costs"
context: orchestrate
category: anti-patterns
concept: avoid
description: "Each anti-pattern names the concrete failure it produces, so it is recognisable mid-run"
tags: anti-patterns, failure-modes, review, cost, scope
sources:
  - "references/raw/orchestrate-SKILL.md"
last_ingested: 2026-08-12
---

## The Eight

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

## The Pattern Behind the Pattern

Six of the eight are the same mistake: **a boundary that was supposed to be
explicit was left implicit**. Scope, file ownership, commit authority, and return
shape are all cheap to state at dispatch time and expensive to reconstruct
afterwards.

The other two — parent-does-everything and skipped verification — are the
degenerate cases where orchestration collapses back into a single undifferentiated
stream, at which point the frontier model is paying for typing.

## Sources

- `references/raw/orchestrate-SKILL.md`
