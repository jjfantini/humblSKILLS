---
title: "Prompting GPT-6 Astra as a Director or Worker"
context: orchestrate
category: prompting
concept: openai-gpt-6-astra
description: "Astra asks instead of assuming and delegates less than you want - both are prompt-fixable, and skill files can silently outrank the brief"
tags: openai, gpt-6-astra, codex, delegation, autonomy, prompting, instruction-following
sources:
  - "references/raw/openai-gpt-6-astra-latest-model-2026-09-10.md"
last_ingested: 2026-09-10
---

## The Model and Its Effort Levels

`gpt-6-astra`, reached through the Responses API or the Codex CLI (`codex -m
gpt-6-astra`). Reasoning effort is `low`, `high`, or `max` — **there is no
`none`**, and `minimal` is gone. `temperature`, `top_p`, `logprobs` and
`top_logprobs` are not accepted; sending them is a migration bug, not a tuning
knob. Verified reachable from `codex-cli` 0.154.0 on 2026-09-10; see
`references/wiki/orchestrate/routing/cursor-cli-models.md` for the Cursor CLI,
which does not carry it.

OpenAI's framing of the cost picture: Astra "achieves stronger results while
using substantially fewer output tokens — delivering a lower estimated API cost
per task than earlier models despite its higher per-token pricing." Per-token
price is therefore the wrong number to route on. Cost per completed task is.

## The Four Behaviours That Matter to an Orchestration Run

OpenAI names five behaviour patterns worth prompting for. Four of them are
directly orchestration-shaped, and each has an official prompt block. Paste
them verbatim.

### 1. It Asks Where Earlier Models Assumed

Astra "is designed to be a more effective collaborator and is thus more likely
to ask the user a question when additional input could materially change the
result. This can cause it to stop when the user may expect it to make reasonable
assumptions and persist." For a worker dispatched into a worktree with nobody
watching, a question is a dead gate.

```text
You should infer the user's intent and task scope from the instructions and prior conversation context. Your job is to bias towards action and carry the user's intended task to completion.

When the user expresses intent to perform new work or fix an existing issue, persist until the user's intended goal is complete. Progress autonomously towards the user's goal (e.g. creating isolated worktrees / checkouts if needed, resolving merge conflicts, read-only actions, creating draft PRs etc.) unless they are clearly destructive or irreversible.
```

Note the parenthetical: OpenAI's own example of legitimate autonomous progress
includes *creating isolated worktrees* and *draft PRs*. A worker brief in this
skill forbids both (see
`references/wiki/orchestrate/contracts/brief-template.md`). Keep the brief's
`Do NOT` line — the vendor block sets the autonomy floor, the brief sets the
ceiling, and the brief wins because it is the more specific instruction.

Second block, for when a brief's phrasing reads as a proposal rather than an
order:

```text
When the user's prompt indicates a request for action, such as "can you...", "I want to...", "help me..." and similar expressions, treat these as instructions to do the work and take action. Do not stop at acknowledging capability (e.g. "Yes…"), proposing a plan, or offering to continue. Do not settle for a partial or "helpful enough" solution that does not fully satisfy the user's task to save time, effort or tokens. If a task requires sustained work, complete all the necessary work until the intended outcome is fulfilled.
```

Third — and this one is a gate-design principle, not just a prompt. Approval
should come *after* the reviewable artifact exists:

```text
Before asking the user clarifying questions, you should complete the work that is already authorized from context and necessary to make the proposed action concrete and reviewable. The user should be approving a concrete, reviewable result. For example, before deploying a change, writing to an external application, merging a PR or publishing a site, do all the required work first so that user approval is the final step. You don't need user permission for reversible tasks, read-only actions, reviews or fixes, or anything for which authorization is provided earlier in the session or strongly implied from the task instruction.

Do not introduce unsolicited warnings, disclaimers, approval flows, or safety/compliance checklists due to hypothetical risk.
```

### 2. Skill Files Can Silently Outrank the Brief

Astra "can be more sensitive to instructions contained in skills and other
files, such as `AGENTS.md`," and OpenAI **strongly recommends** auditing every
skill and instruction file the model can reach. This is a live hazard for a
worker dispatched into a repo that carries its own `AGENTS.md`, `CLAUDE.md`, or
an installed skill set: unclear or conflicting guidance there "may cause the
model to pause and block work early," and the parent sees a stalled gate with no
stated cause.

Two blocks. The first sets precedence:

```text
The user's instructions take precedence over guidelines provided in a skill. If explicit user instructions conflict with a skill's instructions, prioritize the user's instructions.
```

The second makes a silent block diagnosable — worth adding to any brief whose
worker reads repo instruction files:

```text
If a skill causes you to ask for permission or confirmation, pause, leave requested work unfinished, or diverge from the user's intent, name and link to the exact SKILL.md file you read, quote the relevant instruction, and briefly explain how it applies. Distinguish explicit skill requirements from your interpretation of guidelines.
```

**Incorrect (unexplained stall):**

```text
Status: blocked
Summary: I stopped before applying the migration to check with you first.
```

The parent cannot tell whether the brief was underspecified, the change is
genuinely risky, or a repo file it never read told the worker to stop. All three
have different fixes; the handoff names none of them.

**Correct:**

```text
Status: blocked
Summary: AGENTS.md:14 ("never modify files under db/migrate without human
  review") stops this brief's second file. That is an explicit repo requirement,
  not my interpretation. The first file is done and verified.
Escalate?: yes - parent must decide whether the brief overrides AGENTS.md
```

### 3. It Delegates Less Than You Want

Astra "is trained to be able to divide and delegate work to subagents that work
in parallel," but "may delegate less often than desired for your workflow."
This is the exact mirror image of recent Claude models, which over-spawn and
need damping (see
`references/wiki/orchestrate/prompting/anthropic-fable-5-1.md`). Do not carry an
anti-subagent instruction across vendors — it lands as a brake on the model that
was already under-delegating.

```text
If at any point you can parallelize work by delegating tasks to another agent (no matter if you are the root or subagent), you should do so using collaboration tools if it could save time or improve quality.
```

Astra-to-Astra messages also need a legibility rule, because "messages between
agents may contain grammar or spacing errors" — which matters here since a
handoff contract is read by the parent and by a human:

```text
Messages that you send to other agents and your final answer may be read by a human, so ensure they are legible. Always put proper spaces between words and/or numbers.
```

### 4. It Over-Tests Small Changes

On coding work Astra "tends to be thorough in testing before considering a task
complete. For smaller tasks, this can result in broader tests than the task
requires." A cheap mechanical brief that runs the full suite twice has spent its
saving.

```text
Do not write tests for reversible, low-impact changes that mirror the implementation. If you do choose to verify your work with tests, make sure that the tests are meaningful and necessary to verify implementation.

Run tests appropriate to the change and complete required checks. Once those pass, broaden or repeat testing only when new changes, failures, or unresolved concerns justify it; otherwise, continue toward completing the task.
```

This composes with — and does not replace — the brief's `Verify with:` line. The
brief says which command proves the gate closed; this block stops the model
inventing a wider suite around it.

## Writing Style, for Handoffs a Human Reads

Astra "tends to use lists, tables and Markdown to make responses scannable."
That is fine inside a handoff contract, which *is* a fixed structure, and wrong
for the `Summary` field. If handoff summaries come back as nested bullets, the
prose block and the slop-word blocklist are in
`references/raw/openai-gpt-6-astra-latest-model-2026-09-10.md` under "Personality
and writing style" — including a named blocklist ("delve", "leverage", "it's
worth noting", "Bottom Line:", "X, not Y" contrastive framing).

## Sources

- `references/raw/openai-gpt-6-astra-latest-model-2026-09-10.md` — OpenAI's
  "Using GPT-6 Astra" guide (`developers.openai.com/api/docs/guides/latest-model`),
  fetched as markdown 2026-09-10. Every block quoted above is verbatim from its
  "Prompting best practices" section.
