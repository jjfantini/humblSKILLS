---
title: "The Parent-to-Worker Brief Template"
context: orchestrate
category: contracts
concept: brief-template
description: "Nine fixed lines are what make a cheap worker's output predictable"
tags: brief, dispatch, template, scope, contract
sources:
  - "references/raw/orchestrate-SKILL.md"
last_ingested: 2026-08-12
---

## The Template

```text
Goal: …
Worktree path: …
Scope (files / surfaces): …
Out of scope: …
Constraints: …
Done when: …
Verify with: …
Do NOT: commit, open PRs, touch main checkout, expand scope
Return: handoff contract (required)
```

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
Scope (files / surfaces): internal/middleware/ratelimit.go, internal/middleware/chain.go
Out of scope: router setup, config loading, auth middleware, any test file
Constraints: use the existing golang.org/x/time/rate dependency; no new deps
Done when: chain.go wires the limiter and `go build ./...` passes
Verify with: go build ./... && go vet ./internal/middleware/...
Do NOT: commit, open PRs, touch main checkout, expand scope
Return: handoff contract (required)
```

## Do Not Paraphrase the Last Two Lines

`Do NOT` and `Return` are the enforcement lines. Softening them into prose ("try
not to commit", "let me know how it went") reliably produces committed work and a
free-form reply — see `references/wiki/orchestrate/contracts/handoff-contract.md`.

## Sources

- `references/raw/orchestrate-SKILL.md`
