---
title: "The Parent-to-Worker Brief Template"
context: orchestrate
category: contracts
concept: brief-template
description: "Eleven fixed lines plus a vendor appendix are what make a cheap worker's output predictable"
tags: brief, dispatch, template, scope, contract
sources:
  - "references/raw/orchestrate-SKILL.md"
  - "references/raw/anthropic-prompting-claude-fable-5-1-2026-09-10.md"
  - "references/raw/openai-gpt-6-astra-latest-model-2026-09-10.md"
last_ingested: 2026-09-10
---

## The Template

```text
Goal: …
Worktree path: …
Model: …            # concrete id, never `auto`
Effort: …           # low | medium | high | xhigh | max
Scope (files / surfaces): …
Out of scope: …
Constraints: …
Done when: …
Verify with: …
Do NOT: commit, open PRs, touch main checkout, expand scope
Return: handoff contract (required)
```

`Model` and `Effort` are one routing decision written as two lines. On the
current frontier models the effort level carries most of the cost and latency,
so naming the model alone leaves the dominant variable at the harness default —
usually `high`, which is how a mechanical rename ends up paying frontier
thinking. See `references/wiki/orchestrate/routing/model-selection.md`.

## Every Line Earns Its Place

**Incorrect (a brief missing three lines):**

```text
Goal: add rate limiting to the API
Worktree path: ../feat-rate-limit
```

No `Out of scope`, so the worker refactors the router. No `Done when`, so it stops
wherever it feels finished. No `Do NOT`, so it commits. No `Return`, so the parent
gets a transcript. Each omission maps directly to one failure the parent then pays
to undo.

**Correct (complete brief):**

```text
Goal: Add a per-API-key rate limit to the /v1 request path.
Worktree path: /Users/dev/proj-worktrees/feat-rate-limit
Model: gpt-5.3-codex-low-fast
Effort: n/a (baked into the Cursor ID)
Scope (files / surfaces): internal/middleware/ratelimit.go, internal/middleware/chain.go
Out of scope: router setup, config loading, auth middleware, any test file
Constraints: use the existing golang.org/x/time/rate dependency; no new deps
Done when: chain.go wires the limiter and `go build ./...` passes
Verify with: go build ./... && go vet ./internal/middleware/...
Do NOT: commit, open PRs, touch main checkout, expand scope
Return: handoff contract (required)
```

## Add a Vendor Appendix, Not a Second Brief

The eleven lines are vendor-neutral. What is not neutral is the way a given
backend under-delivers by default: Claude Fable 5.1 ends a turn describing what
it would do next, GPT-6 Astra stops to ask a question nobody is there to answer,
and the two need **opposite** subagent-delegation nudges. Append the short block
that pre-empts the failure you have actually seen from that backend — the
verbatim blocks and the picker are in
`references/wiki/orchestrate/prompting/brief-prompting-by-vendor.md`.

Where a vendor block and the brief disagree, the brief wins; it is the more
specific instruction. The clearest case: OpenAI's own autonomy block sanctions
"creating isolated worktrees / checkouts if needed … creating draft PRs," which
the `Do NOT` line forbids. Keep both — the vendor block sets the autonomy floor
so the worker doesn't stall, the `Do NOT` line sets the ceiling so it doesn't
ship.

## Do Not Paraphrase the Last Two Lines

`Do NOT` and `Return` are the enforcement lines. Softening them into prose ("try
not to commit", "let me know how it went") reliably produces committed work and a
free-form reply — see `references/wiki/orchestrate/contracts/handoff-contract.md`.

## Sources

- `references/raw/orchestrate-SKILL.md` — the original nine lines.
- `references/raw/anthropic-prompting-claude-fable-5-1-2026-09-10.md` and
  `references/raw/openai-gpt-6-astra-latest-model-2026-09-10.md` — the effort
  parameter and the vendor appendix blocks.
