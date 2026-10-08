---
title: "Prompting the GPT-6 Family (Astra, 6.1 Sol, Luna) as Director or Worker"
context: orchestrate
category: prompting
concept: openai-gpt-6-astra
description: "One prompt set covers the whole GPT-6 family; Astra asks instead of assuming and delegates less than you want, skill files can silently outrank the brief, and ultra is parent-only"
tags: openai, gpt-6-astra, gpt-6-1-sol, gpt-6-luna, codex, delegation, autonomy, prompting, instruction-following
sources:
  - "references/raw/openai-gpt-6-astra-latest-model-2026-09-10.md"
  - "references/raw/openai-latest-model-2026-10-08.md"
  - "references/raw/openai-models-2026-10-08.md"
  - "references/raw/openai-model-gpt-6-astra-2026-10-08.md"
  - "references/raw/openai-codex-models-2026-10-08.md"
  - "references/raw/openai-codex-subagents-2026-10-08.md"
  - "references/raw/openai-codex-non-interactive-mode-2026-10-08.md"
  - "references/raw/openai-blog-rethinking-skills-and-prompts-for-gpt-6-astra-2026-10-08.md"
  - "references/raw/orchestrator-model-sweep-2026-10-08.md"
last_ingested: 2026-10-08
---

## The Family, and Which Member Goes Where

"GPT-6" is a family name, not a model ID: `codex -m gpt-6` returns a 400. The
callable members, measured 6/6 each from `codex-cli` 0.160.0 on 2026-10-08:

| ID | OpenAI's positioning | API $/MTok in / out | Efforts | Slot in this skill |
|---|---|---|---|---|
| `gpt-6-astra` | "Frontier intelligence for the most demanding work." | $10 / $50 | `low` `medium` `high` `xhigh` `max`, plus `ultra` in Codex | Director for the hardest, most ambiguous plans |
| `gpt-6.1-sol` | "Near-Astra performance for complex work at a lower cost." | $2 / $10 | `low` `medium` (default) `high` `xhigh` `max`, plus `ultra` in Codex | Default Codex parent; demanding workers |
| `gpt-6-sol` | "Previous generation workhorse." | $2 / $10 | adds `none` | Fallback only |
| `gpt-6-luna` | "Fast and affordable model for easier tasks." | $0.10 / $0.50 | adds `none`; **no `ultra`** | Cheap worker |

GPT-6.1 is Sol only - `gpt-6.1-astra` and `gpt-6.1-luna` do not exist. A higher
version number is not a higher tier: OpenAI ranks 6.1 Sol *below* Astra on
capability and makes it the Codex default anyway - "For complex coding and
agentic workflows, use GPT-6.1 Sol when available to your account and client.
Use Luna for focused, repeatable tasks." Astra is kept for the work that "needs
the strongest capability across steps and tools."

**Effort correction.** An earlier version of this concept said Astra's effort
was "`low`, `high`, or `max`". That was wrong: "`reasoning.effort` supports
`low`, `medium`, `high`, `xhigh`, and `max`." Still true from that version:
Astra and 6.1 Sol accept no `none`, no GPT-6 model accepts `minimal`, and with
reasoning on, `temperature` / `top_p` are not accepted. Codex's own starting
points for explicit settings: "start with `high` for GPT-6 Luna or `low` for
GPT-6 Astra."

OpenAI's framing of the cost picture: "Astra achieved stronger results using
substantially fewer output tokens. Its estimated API cost per task was lower
than earlier models despite its higher per-token pricing." Per-token price is
therefore the wrong number to route on. Cost per completed task is.

## `ultra` Is a Parent Setting, Never a Worker Setting

In Codex, `ultra` "uses subagents to handle separate parts of a complex task in
parallel. Choose it when you can divide the work into meaningful parts. Most
tasks do not need Max or Ultra." It is not an API effort value, and "GPT-6 Luna
supports reasoning efforts up to **Max**, but not **Ultra**" - the CLI accepted
`ultra` on Luna in the 2026-10-08 sweep, which shows only that the value was
passed through, not that anything was delegated.

A worker at `ultra` spawns its own subagents. That breaks two rules this skill
depends on: the parent owns every dispatch, and parallel workers hold disjoint
file scopes. Never set it in a worker brief. On a Codex parent it is a choice to
let Codex own the fan-out instead of this skill's briefs - and then the next
section applies.

## Codex Subagents Inherit the Parent's Model

"If you don't configure a subagent model or `model_reasoning_effort`, the
subagent inherits the parent agent's model and reasoning effort." An Astra
parent that fans out natively therefore fans out on Astra pricing. Set
`agents.default_subagent_model` (and `default_subagent_reasoning_effort`) or a
custom agent file, so native subagents land on `gpt-6-luna` or `gpt-6.1-sol` -
the tiers OpenAI names for workers: `gpt-6.1-sol` "Start here for demanding
agents"; `gpt-6-luna` "Use for fast, narrowly scoped agents handling clear,
repeatable, or high-volume work." Local Codex releases spawn agents only "after
a direct request or applicable project or skill instruction," so a brief that
says nothing about delegation gets none.

## Dispatching a Codex Worker: Sandbox and Stdin

"By default, `codex exec` runs in a read-only sandbox." A worker that must edit
files needs `--sandbox workspace-write` (`--full-auto` is deprecated). Send the
brief on stdin:

```text
printf '%s' "$BRIEF" | codex exec -m gpt-6-luna -c model_reasoning_effort=high \
  --sandbox workspace-write --skip-git-repo-check -C "$WORKTREE" -
```

On the measuring machine, a brief over roughly 1,000 characters passed as the
prompt *argument* was killed instantly (rc=137, no output) four times; the same
brief on stdin ran every time. The cause was not isolated - see
`references/raw/orchestrator-model-sweep-2026-10-08.md` - but stdin is safe
everywhere and a real brief is always longer than 1 KB.

## One Prompt Set for the Whole Family

OpenAI publishes no separate 6.1 Sol or Luna guide: "Use the following prompts
as a starting point across the GPT-6 model family. They address behavior
observed with GPT-6 Astra; evaluate them with your chosen model and workload."
The reverse also holds - "Guidance that helps Sol or Luna may overconstrain
GPT-6 Astra" - and "Previous models needed encouragement to run tests and check
their work. GPT-6 Astra does that on its own." A testing nudge written for Luna
becomes over-testing on Astra.

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

Astra-to-Astra messages also need a legibility rule - OpenAI: "Messages between
agents may contain grammar or spacing errors." That matters here because a
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
`references/raw/openai-latest-model-2026-10-08.md` under "Personality and
writing style" — including a named blocklist ("delve", "leverage", "it's
worth noting", "Bottom Line:", "X, not Y" contrastive framing).

## Sources

- `references/raw/openai-latest-model-2026-10-08.md` - OpenAI's "Using GPT-6"
  family guide (`developers.openai.com/api/docs/guides/latest-model`), fetched as
  markdown 2026-10-08. Every prompt block quoted above is verbatim from its
  "Prompting best practices" section; all of them are unchanged from the
  2026-09-10 snapshot, which is kept for diffing.
- `references/raw/openai-gpt-6-astra-latest-model-2026-09-10.md` - the earlier
  "Using GPT-6 Astra" snapshot.
- `references/raw/openai-models-2026-10-08.md` and
  `references/raw/openai-model-gpt-6-astra-2026-10-08.md` - positioning, prices,
  and Astra's five effort levels.
- `references/raw/openai-codex-models-2026-10-08.md` - Codex default model,
  `ultra`, and Luna's lack of it.
- `references/raw/openai-codex-subagents-2026-10-08.md` - subagent model
  inheritance, `default_subagent_model`, worker tiers and starting efforts.
- `references/raw/openai-codex-non-interactive-mode-2026-10-08.md` - the
  read-only default sandbox.
- `references/raw/openai-blog-rethinking-skills-and-prompts-for-gpt-6-astra-2026-10-08.md`
  - why Sol/Luna guidance can overconstrain Astra.
- `references/raw/orchestrator-model-sweep-2026-10-08.md` - reachability, the
  400 on `gpt-6`, and the argv kill.
