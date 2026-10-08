---
title: "Which Vendor Prompt Lines Belong in a Worker Brief"
context: orchestrate
category: prompting
concept: brief-prompting-by-vendor
description: "The brief template is vendor-neutral; the fix for a given model's default failure is not - Fable over-delegates, GPT-6 under-delegates, and the cheap Claude workers stop early"
tags: brief, prompting, vendor, anthropic, openai, delegation, dispatch
sources:
  - "references/raw/anthropic-prompting-claude-fable-5-1-2026-09-10.md"
  - "references/raw/openai-gpt-6-astra-latest-model-2026-09-10.md"
  - "references/raw/orchestrate-SKILL.md"
  - "references/raw/anthropic-prompting-claude-opus-5-5-2026-10-08.md"
  - "references/raw/anthropic-prompting-claude-sonnet-5-5-2026-10-08.md"
  - "references/raw/anthropic-prompting-claude-haiku-5-5-2026-10-08.md"
  - "references/raw/openai-latest-model-2026-10-08.md"
last_ingested: 2026-10-08
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

| | Claude Fable 5.1 | Claude Opus 5.5 | Claude Sonnet 5.5 / Haiku 5.5 | GPT-6 family (Astra, 6.1 Sol, Luna) |
|---|---|---|---|---|
| Subagent delegation | Over-delegates; **damp** it | Not documented; the Opus 5 damping block is the stated starting point | Sonnet at `xhigh`/`max` launches its own reviewer subagents; **stop** it | Under-delegates; **encourage** it - and never run a worker at `ultra` |
| Asking vs assuming | Assumes and continues; may end a turn describing next steps | Ends turns on a text-only progress update; the harness must nudge | Checks in early at `low`/`medium` on long tasks | Asks a clarifying question and stops |
| Scope | Rewrites whole files more than needed | - | Sonnet adds unrequested tests, docs and files at every effort | - |
| Testing | Commits more tests than asked | - | Skips the check at `low` (Haiku also at `medium`) | Runs broader test suites than the change warrants |
| Formatting | Uses *less* bold/lists than earlier models | - | - | Uses *more* lists/tables/Markdown |
| Instruction sensitivity | - | - | - | Highly sensitive to `AGENTS.md` / skill files; can stall silently |
| "Show your reasoning" in a brief | - | May be declined (`reasoning_extraction`) | Sonnet: may be declined | - |

A dash means the vendor's current guide says nothing about that behaviour for
that model - not that the behaviour is absent. Fill a cell from a vendor page or
a measurement, never by analogy with a sibling model.

The delegation row is the trap. A prompt library that carries one
"don't spawn subagents unnecessarily" line and pastes it into every brief brakes
the model that was already under-delegating, while the formatting row means an
anti-formatting rule written for an older Claude will over-correct Fable 5.1
into wall-of-text handoffs. The two cheap Claude workers fail in the opposite
direction to Fable: they stop *early*, so their appendix is the keep-working and
verify blocks, not a damping line.

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

For a Haiku 5.5 or Sonnet 5.5 worker the appendix is different in kind: the
model's keep-working block plus the shared verification paragraph, both in
`anthropic-claude-5-5.md`, and `Effort: medium` rather than `low`.

That GPT-6 appendix is *abridged for the example* — the lines are trimmed and
re-indented to fit a brief. In a real dispatch, paste the unabridged blocks.

Full verbatim blocks live in
`references/wiki/orchestrate/prompting/anthropic-fable-5-1.md`,
`references/wiki/orchestrate/prompting/anthropic-claude-5-5.md` and
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
