---
title: "Which Vendor Prompt Lines Belong in a Worker Brief"
context: orchestrate
category: prompting
concept: brief-prompting-by-vendor
description: "The brief template is vendor-neutral; the fix for a given model's default failure is not - and two vendors need opposite delegation nudges"
tags: brief, prompting, vendor, anthropic, openai, delegation, dispatch
sources:
  - "references/raw/anthropic-prompting-claude-fable-5-1-2026-09-10.md"
  - "references/raw/openai-gpt-6-astra-latest-model-2026-09-10.md"
  - "references/raw/orchestrate-SKILL.md"
last_ingested: 2026-09-10
---

## The Brief Stays Vendor-Neutral; the Appendix Does Not

`references/wiki/orchestrate/contracts/brief-template.md` is the contract, and
it does not change per model. What changes is a short appendix that pre-empts
the specific way the chosen backend under-delivers by default.

Two lines are worth adding to the brief itself, independent of vendor, because
`(model, effort)` is now one routing decision rather than two:

```text
Model: <concrete id>
Effort: <low | medium | high | xhigh | max>
```

Naming the model without the effort leaves the dominant cost variable at the
harness default — usually `high` — which is how a mechanical rename ends up
paying frontier thinking.

## The Vendor Defaults Point in Opposite Directions

| | Claude Fable 5.1 | GPT-6 Astra |
|---|---|---|
| Subagent delegation | Over-delegates; **damp** it | Under-delegates; **encourage** it |
| Asking vs assuming | Assumes and continues; may end a turn describing next steps | Asks a clarifying question and stops |
| File edits | Rewrites whole files more than needed | — |
| Testing | Commits more tests than asked | Runs broader test suites than the change warrants |
| Formatting | Uses *less* bold/lists than earlier models | Uses *more* lists/tables/Markdown |
| Instruction sensitivity | — | Highly sensitive to `AGENTS.md` / skill files; can stall silently |

The delegation row is the trap. A prompt library that carries one
"don't spawn subagents unnecessarily" line and pastes it into every brief brakes
the model that was already under-delegating, while the formatting row means an
anti-formatting rule written for an older Claude will over-correct Fable 5.1
into wall-of-text handoffs.

**Incorrect (one appendix for every worker):**

```text
Do NOT: commit, open PRs, touch main checkout, expand scope
Return: handoff contract (required)

Appendix (all workers): Avoid spawning subagents. Do not use bullet points or
headers. Be thorough — run the full test suite before reporting done.
```

On a GPT-6 Astra worker that is three own-goals: it suppresses the parallelism
the model needs a nudge toward, and "be thorough / full test suite" amplifies
exactly the over-testing OpenAI documents.

**Correct (appendix chosen by backend):**

```text
Do NOT: commit, open PRs, touch main checkout, expand scope
Return: handoff contract (required)

Appendix — GPT-6 Astra worker:
  When the user's prompt indicates a request for action, such as "can you...",
  "I want to...", "help me..." and similar expressions, treat these as
  instructions to do the work and take action. Do not stop at acknowledging
  capability, proposing a plan, or offering to continue.

  The user's instructions take precedence over guidelines provided in a skill.
  If explicit user instructions conflict with a skill's instructions, prioritize
  the user's instructions. If a skill causes you to pause or leave requested work
  unfinished, name the exact SKILL.md file, quote the instruction, and explain
  how it applies.

  Run tests appropriate to the change and complete required checks. Once those
  pass, broaden or repeat testing only when new changes, failures, or unresolved
  concerns justify it.
```

That appendix is *abridged for the example* — the lines are trimmed and
re-indented to fit a brief. In a real dispatch, paste the unabridged blocks.

Full verbatim blocks for both vendors live in
`references/wiki/orchestrate/prompting/anthropic-fable-5-1.md` and
`references/wiki/orchestrate/prompting/openai-gpt-6-astra.md`. Pick from there
rather than paraphrasing — both vendors note that specific wordings carry the
effect.

## The Shared Floor: Three Lines Every Unattended Worker Needs

Whatever the backend, a worker dispatched into a worktree with nobody watching
needs the same three things stated, because both vendors' defaults are tuned for
an interactive user who can answer:

1. **Nobody is watching.** Anthropic's block opens with "You are operating
   autonomously. The user is not watching in real time and cannot answer
   questions mid-task" and the docs say that opening sentence carries much of
   the effect. OpenAI's equivalent is "bias towards action and carry the user's
   intended task to completion."
2. **Don't end a turn on a promise.** Both models will otherwise describe the
   next step instead of taking it. Anthropic states the check explicitly: if the
   last paragraph is a plan, a question, or an "I'll…", do that work now.
3. **Extras are a follow-up, not a change.** The brief's `Out of scope` line is
   the parent's version of this; both vendors have a block that enforces it in
   the model's own idiom.

Where the vendor block and the brief disagree, the brief wins — it is the more
specific instruction. The clearest case: OpenAI's autonomy block explicitly
sanctions "creating isolated worktrees / checkouts if needed, resolving merge
conflicts, read-only actions, creating draft PRs," and this skill's `Do NOT`
line forbids the PR and the checkout change. Keep both. The vendor block sets
the autonomy floor so the worker doesn't stall; the `Do NOT` line sets the
ceiling so it doesn't ship.

## An Appendix Is Not a Licence to Grow the Brief

Every line added is a line the worker has to reconcile against the other lines,
and OpenAI's own warning is that "unclear or conflicting guidance in a skill
file may cause the model to pause and block work early." Add the block that
addresses a failure you have actually observed from that backend, not the whole
vendor guide. If the brief's appendix is longer than the brief, the routing
decision was probably wrong: a subtask needing that much behavioural correction
belongs on a stronger worker or on the parent.

## Sources

- `references/raw/anthropic-prompting-claude-fable-5-1-2026-09-10.md`
- `references/raw/openai-gpt-6-astra-latest-model-2026-09-10.md`
- `references/raw/orchestrate-SKILL.md` — the brief this appendix attaches to.
