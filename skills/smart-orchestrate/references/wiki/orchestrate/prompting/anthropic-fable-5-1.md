---
title: "Prompting Claude Fable 5.1 as an Orchestrator or Worker"
context: orchestrate
category: prompting
concept: anthropic-fable-5-1
description: "Effort is the cost dial, and four verbatim system-prompt blocks fix the four ways Fable 5.1 under-delivers on a long orchestration run"
tags: anthropic, fable-5-1, effort, autonomy, subagents, prompting, parent, worker
sources:
  - "references/raw/anthropic-prompting-claude-fable-5-1-2026-09-10.md"
  - "references/raw/anthropic-claude-prompting-best-practices-2026-09-10.md"
last_ingested: 2026-09-10
---

## Effort Is the Routing Dial, Not the Model Name

On Claude Fable 5.1 (`claude-fable-5-1`) the model name no longer carries the
cost/quality trade-off on its own — `effort` does. Anthropic's guidance is
explicit: *"Effort is the primary control for trading off intelligence, latency,
and cost on Claude Fable 5.1."* The levels are `low`, `medium`, `high` (default),
`xhigh`, `max`.

Three consequences that change how this skill routes:

| Documented behaviour | Routing consequence |
|---|---|
| At `medium`, "results roughly match Claude Fable 5 at lower cost" | A mid-tier subtask does not need a mid-tier *model* — it needs Fable 5.1 at `medium` |
| At `low`, "often competitive with Claude Opus and Claude Sonnet models on cost per task while scoring higher" | Fable 5.1 `low` belongs in the *cheap worker* comparison, not only the parent slot |
| "effort level names don't correspond to the same amount of thinking across models" | A sweep run on Fable 5 does not transfer. Re-measure per model |

So a routing decision is now a **pair**: `(model, effort)`. Naming only the
model leaves the dominant cost variable unset.

**Incorrect (model named, effort left at default):**

```text
Scope: rename the config keys across 6 leaf files
Model: claude-fable-5-1
```

Mechanical, leaf-file, zero blast radius — and it runs at `high` because that is
the default, paying frontier thinking for a rename.

**Correct:**

```text
Scope: rename the config keys across 6 leaf files
Model: claude-fable-5-1 at effort=low
```

Do not carry `xhigh`/`max` into long-deliverable subtasks. At those levels the
model "can think for longer before it starts writing its reply" and may draft the
deliverable in thinking and then write it again — a double-length turn for no
quality gain. Anthropic's own recommendation is to run those at `high` and move
up only where a measured quality gain exists.

## The Parent Nudge: Let the Lead Keep Working While Subagents Run

This is the single most orchestration-specific line in Anthropic's guidance:

> "If your coding agent lets Claude Fable 5.1 delegate work to subagents, don't
> force the lead agent to stop and wait for each one. On coding tasks, letting
> the lead continue while subagents run lowers average time to completion at
> similar quality, token usage, and cost."

The harness shape it asks for: the dispatch tool returns immediately, each
result comes back to the lead in a later `user` message, and the lead gets a
*separate* tool for when it does want to block.

**This does not weaken the gate in `references/wiki/orchestrate/loop/session-loop.md`.**
A gate is a **commit boundary, not a blocking boundary**. The lead may dispatch
worker B and keep reading code while worker A runs; it may not *open gate 2*
before gate 1 is verified and committed. Non-blocking dispatch buys wall clock
inside a gate. Committing between gates buys a verified baseline across them.
Conflating the two is what produces three phases in flight with nothing
committed.

## Four Verbatim Blocks Worth Pasting

Each of these fixes a documented Fable 5.1 tendency that costs a real round trip
on a long run. Paste them; do not paraphrase — Anthropic notes the specific
opening sentences carry most of the effect.

**1. Autonomy / finish the whole task** — for any worker or parent running
unattended. Without it the model "sometimes describes what it would do next
instead of doing it" or stops to ask permission for a step the original request
already covered.

Anthropic pairs this with a second "Delivering work" block and says: *"Two
system prompt additions together mitigate this. Apply both. If you need to limit
prompt length, use only the first, which keeps most of the effect."* The first
is reproduced in full below; the second is under "Finish the whole task" in
`references/raw/anthropic-prompting-claude-fable-5-1-2026-09-10.md` — paste it
too unless prompt length forces the trade.

```text
You are operating autonomously. The user is not watching in real time and cannot answer questions mid-task, so asking 'Want me to…?' or 'Shall I…?' will block the work. For reversible actions that follow from the original request, proceed without asking. Stop only for destructive actions or genuine scope changes the user must decide. Offering follow-ups after the task is done is fine; asking permission before doing the work is not.

Exception: when the user is describing a problem, asking a question, or thinking out loud rather than requesting a change, the deliverable is your assessment. Report your findings and stop. Don't apply a fix until they ask for one.

Before ending your turn, check your last paragraph. If it is a plan, an analysis, a question, a list of next steps, or a promise about work you have not done ('I'll…', 'let me know when…'), do that work now with tool calls. That includes retrying after errors and gathering missing information yourself. Do not stop because the context or session is long. End your turn only when the task is complete or you are blocked on input only the user can provide.

Before running a command that changes system state (such as restarts, deletes, or config edits), check that the evidence actually supports that specific action. A signal that pattern-matches to a known failure may have a different cause.
```

Keep the opening sentence about the user not watching exactly as written —
Anthropic notes it "carries much of the effect." If the product needs the model
to stop for specific confirmations, add a sentence listing them *after* it
rather than editing it. One trade-off to check on your own tasks: this block can
also make the model less likely to ask about genuinely ambiguous requests.

**2. Scope containment** — the brief's `Out of scope` line, enforced in the
model's own idiom. Anthropic measured that "unrequested additions and committed
test code drop substantially with no measurable change in task success."

```text
If, while working or testing, you find a pre-existing bug, a performance concern, or behavior the task doesn't mention, don't fix, optimize or extend it in this change unless the requested behavior cannot work without it; report it as a follow-up in your summary. Where the task is ambiguous, implement the reading its wording and the surrounding code most directly support, state that assumption in your summary, and don't build for the other readings as well. Verify your work however you like; scratch scripts and quick checks need not be kept. Commit tests only where the task asks for them or this repository already keeps tests for this kind of change, sized like the neighboring test files — roughly one focused test per stated behavior — and don't turn scratch checks into additional permanent test files. This is about extras only: implement every behavior the task asks for, completely.
```

**3. Targeted edits** — Fable 5.1 rewrites whole files more readily than Fable 5
did. Same resulting file, more output tokens and wall clock.

```text
The number of tokens used to edit files is best minimized, all else being equal. Therefore, when it will not affect the end result, try to surgically edit a file rather than rewrite the entire thing.
```

**4. Tool batching in an agent loop** — in bash-and-editor harnesses Fable 5.1
may issue implied independent calls one per turn. Each extra turn costs tokens,
a round trip, and wall clock.

```text
First privately list what you need next; then request every item that doesn't depend on another's result in this one response.
```

Keep the word *privately*. Anthropic's note: without it "the model sometimes
answers the reminder instead of the user." Send it as a turn-scoped system
message after each set of tool results, never by rewriting an earlier turn.

## Do Not Rewrite History Mid-Run

For accounts created on or after 2026-08-31, Fable 5.1's thinking blocks are
valid only in the exact conversation that produced them: replaying one after its
prefix changed returns a 400. The edits that trip this are the same ones that
restart the prompt cache — injecting and removing per-turn reminders,
summarizing older turns in place, changing the system prompt mid-session.

For a parent orchestrator this maps to one rule: **append, never edit.** Per-turn
reminders go in turn-scoped system messages; instruction or tool changes go in a
mid-conversation system message; trimming is server-side compaction's job. If the
harness compacts on the client, replace the whole history with one summary
message plus the new user turn and replay no thinking blocks at all.

When compacting a long orchestration session by hand, state what the summary must
keep. The parent's own value is the context it holds across workers, and an
unguided summary drops exactly the parts a worker brief is built from — what was
ruled out, and why. Anthropic's client-side summarization instruction is in
`references/raw/anthropic-prompting-claude-fable-5-1-2026-09-10.md` under "Tell
the model what to preserve in compaction summaries"; the six items it enumerates
are worth honouring verbatim.

## Damp Subagent Spawning, Don't Command It

Anthropic's cross-model guidance is that recent Claude models "orchestrate
subagents natively" and "will delegate appropriately without explicit
instruction" — the failure mode to prompt against is *over*-use, not under-use:
Opus 5 and Opus 4.6 "may spawn them in situations where a simpler, direct
approach would suffice," such as spawning an agent for exploration when a direct
grep is faster.

That is the opposite of the GPT-6 Astra default (see
`references/wiki/orchestrate/prompting/openai-gpt-6-astra.md`, which needs a
nudge *toward* delegation). Damping prompt:

```text
Use subagents when tasks can run in parallel, require isolated context, or involve
independent workstreams that don't need to share state. For simple tasks, sequential
operations, single-file edits, or tasks where you need to maintain context across steps,
work directly rather than delegating.
```

## Sources

- `references/raw/anthropic-prompting-claude-fable-5-1-2026-09-10.md` — the
  Fable 5.1 / Mythos 5.1 model-specific guide; every quoted block above is
  copied from it verbatim.
- `references/raw/anthropic-claude-prompting-best-practices-2026-09-10.md` — the
  cross-model guide; source for the subagent-orchestration and overeagerness
  sections.
