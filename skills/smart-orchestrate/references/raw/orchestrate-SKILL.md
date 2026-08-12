---
name: orchestrate
description: >
  Run a frontier-parent / cost-optimized-worker orchestration pattern across
  Claude, Codex, and Cursor CLIs. Use when the user asks to orchestrate a
  feature, goal, or loop; farm work to subagents; plan then gate execution;
  or split planning on a frontier model from implementation on cheaper models.
disable-model-invocation: true
---

# Orchestrate

Teach the current session agent to act as (or under) a **parent orchestrator**:
plan on a frontier model, execute on cheaper models, verify the whole.

Prefer **Claude** as the entry point. **Codex** or **Cursor** can also be the
parent, and any parent may farm work to agents running through Claude, the
Cursor CLI, or the Codex plugin.

## Roles

### Parent orchestrator (frontier)

Use a strong model appropriate to the orchestration work, e.g.:

- Claude: Fable 5 (high - max, ultracode if a massive refactor or implementation touching many services and e2e test flows. always confirm with the user as it can incure massive cost. planning with Opus 5 on ultracode might be better. opus 5 on max is the same if not better performance than Fable 5 and much cheaper, so opt for this unless Fable 5 is requestd. Typically Fable 5 should only be used as the parent orchestrator)
- Codex: GPT-5.6 Sol Max
- Cursor: Auto Intelligence, or Grok 4.5 High

Owns:

1. **Plan** — understand the goal; break into phases and gated subtasks
2. **Isolate** — open a worktree via `/smart-worktree-flow` before dispatch; workers never use the user's main checkout
3. **Route** — assign each subtask a worker model by scope, difficulty, blast radius
4. **Dispatch** — farm clear, bounded work to worker agents/CLIs inside the worktree
5. **Glue** — hold context across workers; resolve conflicts; keep one coherent design
6. **Verify** — read worker handoffs, run checks, confirm the change fits the larger system
7. **Close out** — after each gate, parent runs `/smart-commit`; when shipping, parent runs `/smart-worktree-flow` for PR/merge/cleanup

The entry-point model does not have to produce the plan itself. It may delegate
planning or plan refinement to another capable agent, then retain ownership of
the resulting plan, execution gates, integration, and final verification.

The parent does **not** burn frontier tokens on mechanical implementation when
a smaller model can execute a crisp brief. The parent **does** own all commits
and ship steps — workers implement; they do not invent commit or PR policy.

### Worker agents (cost-optimized)

Prefer the smallest, fastest model that can finish the brief cleanly, e.g.:

- Grok 4.5, Composer 2.5
- GPT-5.6 Terra, Luna
- Claude Sonnet 5, Opus 5 (reserve Opus for harder worker slots)

Owns:

1. Execute one gated subtask from the parent's brief, inside the assigned worktree
2. Stay inside the given scope and files
3. Return the **handoff contract** below — summary only, not a full transcript

Workers do **not** re-plan the whole feature, expand scope, commit, open PRs,
or touch the main checkout without escalating.

## Isolation (`/smart-worktree-flow`)

Default: isolate before any worker runs.

1. Parent follows `/smart-worktree-flow` to create a paired worktree + branch
2. Every worker is pointed at that worktree path (or a separate worktree if the
   subtasks are truly independent)
3. Parallel workers get non-overlapping file scopes inside the same worktree,
   or separate worktrees when scopes would collide
4. Parent alone merges, opens PRs, and cleans up via `/smart-worktree-flow`

Skip isolation only when the user explicitly wants in-place work on the current
checkout, or the change is a trivial single-file edit with no parallel workers.

## Routing guidelines

Not hard rules — judgment calls. Bias cheap and narrow; escalate when unsure.

| Signal | Prefer |
| --- | --- |
| Clear brief, small blast radius, mechanical change | Fastest / cheapest worker |
| Multi-file but well-specified; some judgment | Mid-tier worker (e.g. Sonnet 5, Terra) |
| Ambiguous design, high blast radius, tricky correctness | Stronger worker (e.g. Opus 5) or keep on parent |
| Cross-cutting architecture, phase planning, final integration check | Parent frontier only |

Escalate to the parent (or a stronger worker) when:

- The brief is underspecified or contradicts the codebase
- The change touches shared contracts, auth, data, or many callers
- Verification fails and the fix isn't obvious
- A worker wants to expand scope

## Session loop

1. **Clarify goal** — feature, fix, goal, or loop; success criteria
2. **Plan or delegate planning** — phases, dependencies, risks, verification points
3. **Isolate** — `/smart-worktree-flow` worktree + branch (unless user opts out)
4. **Gate subtasks** — each task has: scope, brief, model tier, done-when, return contract
5. **Dispatch workers** — one subtask per worker; no overlapping file ownership when avoidable
6. **Collect handoffs** — reject incomplete returns; re-dispatch or escalate
7. **Integrate + verify** — parent reviews diffs against the plan and larger system
8. **Close the gate** — parent `/smart-commit` for that phase; only then open the next gate
9. **Ship or replan** — when done, parent `/smart-worktree-flow` for PR/merge/cleanup; if reality diverges, parent revises the plan before more dispatch

## Brief template (parent → worker)

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

## Handoff contract (worker → parent)

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

Parent treats this as the only worker context worth keeping — not the full
worker transcript.

## Close-out (parent only)

| When | Skill | Who |
| --- | --- | --- |
| After each verified gate | `/smart-commit` | Parent |
| Feature ready to ship | `/smart-worktree-flow` (PR → develop → main, cleanup) | Parent |
| Plan still in flux | neither — replan first | Parent |

Workers never run these.

## Anti-patterns

- Parent implementing everything itself on the frontier model
- Workers inventing new phases or rewriting the plan
- Workers committing, opening PRs, or editing the main checkout
- Parallel workers on the same files without a merge owner
- Farming across CLIs without a worktree (shared dirty tree)
- Accepting a free-form worker reply instead of the handoff contract
- Skipping parent verification or `/smart-commit` between phases
- Sending vague, high-blast-radius work to the cheapest model
