---
title: "Prompting Claude Opus 5.5, Sonnet 5.5 and Haiku 5.5 for Orchestration"
context: orchestrate
category: prompting
concept: anthropic-claude-5-5
description: "Opus 5.5 is the default Claude parent at medium effort, Haiku 5.5 the cheap worker at medium not low - and each needs one verbatim block to stop it ending work early"
tags: anthropic, opus-5-5, sonnet-5-5, haiku-5-5, effort, autonomy, subagents, prompting, parent, worker
sources:
  - "references/raw/anthropic-prompting-claude-opus-5-5-2026-10-08.md"
  - "references/raw/anthropic-prompting-claude-sonnet-5-5-2026-10-08.md"
  - "references/raw/anthropic-prompting-claude-haiku-5-5-2026-10-08.md"
  - "references/raw/anthropic-prompting-claude-opus-5-2026-10-08.md"
  - "references/raw/anthropic-models-overview-2026-10-08.md"
  - "references/raw/anthropic-choosing-a-model-2026-10-08.md"
  - "references/raw/anthropic-api-and-data-retention-2026-10-08.md"
  - "references/raw/orchestrator-model-sweep-2026-10-08.md"
  - "references/raw/anthropic-prompting-claude-fable-5-1-2026-10-08.md"
last_ingested: 2026-10-08
---

## Where Each 5.5 Model Sits in a Run

Anthropic's own default is now explicit: *"If you're unsure which model to use,
start with Claude Opus 5.5 for most workloads. Use Claude Fable 5.1 for
demanding reasoning and long-horizon agentic work, or when your evals on Claude
Opus 5.5 at higher effort still fall short."* Opus 5 is listed as a legacy model.

| Model | ID | $/MTok in / out | Default effort | Slot in this skill |
|---|---|---|---|---|
| Claude Opus 5.5 | `claude-opus-5-5` | $4 / $20 | `medium` | Default Claude parent; strong worker for high-blast-radius briefs |
| Claude Sonnet 5.5 | `claude-sonnet-5-5` | $2 / $10 | `high` on the API, `medium` in Claude Code | Mid-tier worker for multi-file, well-specified briefs |
| Claude Haiku 5.5 | `claude-haiku-5-5` | from $0.10 / from $0.50 | `medium` | Cheap worker; Anthropic lists "sub-agent tasks" among its uses |
| Claude Fable 5.1 | `claude-fable-5-1` | $10 / $50 | `high` | Escalation parent - see `anthropic-fable-5-1.md` |

All four have a 1M context window and 128K max output. Haiku 5.5's price
rises above 100k input tokens, so a worker brief that drags a large file set
into Haiku's context loses most of the saving. The Opus 5.5, Sonnet 5.5 and
Haiku 5.5 rows are not Covered Models, so the 30-day-retention / no-ZDR veto
that applies to Fable 5.1 does not apply to them.

Measured reachable from Claude Code 2.1.294 on 2026-10-08: each ID 6/6, and the
`opus` / `sonnet` / `haiku` aliases resolved to the 5.5 IDs on this account.
Haiku 5.5 and Sonnet 5.5 each returned correct code and a complete handoff
contract on a real file-writing brief
(`references/raw/orchestrator-model-sweep-2026-10-08.md`).

## Effort: Start at the Default, Not at `max`

Effort is the main dial on all three, as on Fable 5.1. The Opus 5.5 guide is
the clearest statement of it: *"Start at `medium`, the default on Claude Opus
5.5 (Claude Opus 5 defaults to `high`), set it explicitly, and test several
levels against your own evals rather than carrying over the setting you used on
Claude Opus 5."* And: *"Reserve `xhigh` and `max` for work where you've measured
a quality gain."*

That retires this skill's older "parent at `max`" habit for the Claude side.
Opus 5.5 at `medium` "matches or exceeds Claude Opus 5 at `high` on coding and
knowledge-work evaluations," and at a given level it "tends to think more per
turn than Claude Opus 5, especially at `xhigh` and `max`." Carrying `max` over
from Opus 5 buys longer turns, not a better plan, unless your own evals say
otherwise.

**Haiku 5.5 workers run at `medium`, not `low`.** Anthropic: *"In long agent
prompts, the model is more likely to skip a search, stop early, or skip a check
at this level,"* and *"At `low` and `medium` effort, Claude Haiku 5.5 sometimes
reports a code change as done without running a check."* Moving from `low` to
`medium` roughly halved early stopping and more than doubled output tokens. The
sweep's 9/9 at `low` was a 1 KB brief; it does not test the long-prompt case.

**Sonnet 5.5 workers run at `medium` for well-specified agentic briefs** and
`high` for harder ones, per its guide. At `low` it "can skip verifying a
change"; at `low` and `medium` on long tasks it is "more likely to stop and check
in with the user before it finishes."

## The Opus 5.5 Parent Ends Turns on Text

This is the orchestration-specific failure in the Opus 5.5 guide. *"On long
tasks with several parts, Claude Opus 5.5 keeps the user updated as it works,
and some of those updates end the turn with text rather than a tool call."* An
unattended loop that treats that turn as "done" stops mid-run, with workers
still in flight.

**Incorrect (end of turn treated as end of task):**

```text
Parent (Opus 5.5): "Gate 1 verified. Next I'll dispatch the config worker and
                   then the tests worker."   <- stop_reason: end_turn
Harness:           run marked complete; gate 2 never opens.
```

**Correct (harness keeps a checklist and nudges, with a cap):**

Treat a text-only end of turn as a report. If items remain open and no blocker
is stated, send a short user message naming them - Anthropic's example:

```text
Your task list still has open items: migrate the remaining two endpoints and update their tests. Continue with them. If one is blocked, say what is blocking it.
```

and "stop after two or three automatic continuations on the same task rather
than repeating them indefinitely." If a worker the parent started is still
running, wait for it and return its output as the next user message rather
than ending the run.

For fully unattended parents, Anthropic also gives a system-prompt addition.
Add it at the end of the system prompt **from the first request** - adding it
mid-session invalidates earlier thinking blocks - and leave it out of
human-in-the-loop sessions:

```text
A standing instruction from the user, the person you are working for. It is about how your turns end. A message with no tool call in it ends your turn, and the work stops there until you are asked to continue. The user has seen you end turns in four ways while work they asked for was still owed, and does not want any of them. One: a long summary of what was done that closes by announcing the next step and has no tool call, so the next thing never starts. Two: an offer to carry on with something unless the user would prefer otherwise, which stops to wait for an answer the user was not going to give. Three: a list of decisions for the user when, by your own account, none of them blocks the rest of the work. Four: deciding that this is a good place to report, because the turn has been long or a milestone is done. Status notes are welcome, and so are your recommendations on open decisions, but put them in the same message as your next tool call and carry on with whatever does not depend on the user's answer. If you notice yourself inviting the user to redirect you or offering to wait, delete it and do the next thing. The stops the user does want are the ones where nothing can move without them, or where the thing blocking you is deliberately protected from you. This does not override the need for confirmation on risky or destructive actions.
```

It does not override the parent's own gates: commit authority, the `Do NOT`
line, and confirmation before destructive steps all stay.

**Give a multi-agent parent a time signal.** Opus 5.5 "pays close attention to
information about elapsed time," and in a lead-plus-subagents setup that buys
parallelism. Append `elapsed 340s / 1200s` to each harness message, or, with no
sensible budget, add:

```text
Time matters here: do not spend time that can be avoided, and the earlier a correct result is obtained, the better.
```

The budget is advisory; keep your own timeout.

## Delegation: What Is and Isn't Documented

No 5.5 prompting page states whether the model over- or under-delegates. Do not
fill that gap by inference. What is documented:

- **Opus 5.5** - its guide says the Opus 5 patterns "remain a reasonable
  starting point," and the Opus 5 guide says Opus 5 "delegates to subagents more
  readily than prior models." Its damping block is the one to carry:

```text
Delegate to a subagent only for large tasks that are genuinely independent and parallelizable, such as a wide multi-file investigation. Do not delegate work you can finish yourself in a handful of tool calls, and do not use subagents to verify or double-check your own work. If one subagent can complete the task, use one rather than several, and keep spawn counts low.
```

- **Sonnet 5.5 at `xhigh` / `max`** - "After it finishes a task, it can start its
  own rounds of review and verification, sometimes with subagents if your harness
  provides them." In a worker that is scope expansion the parent pays for.
  Anthropic measured this block cutting session cost by about a third at `max`
  with no change in quality:

```text
When the work the user asked for is done and its checks pass, stop and report. Don't start extra rounds of review or hardening on your own, and don't launch reviewer sub-agents unless the user asked for a review. If you think a deeper review is worth doing, say so at the end.
```

Verification belongs to the parent in this skill, so a worker's self-started
reviewer subagents duplicate a job that is already owned.

## Worker Blocks: Keep Going, Stay in Scope, Run a Real Check

Sonnet 5.5 and Haiku 5.5 share a two-paragraph block with **different wording**
- copy the one for the backend you dispatch to.

Sonnet 5.5 (note "features, tests, files, docs or refactors"; blank line between
paragraphs). The page adds that Sonnet 5.5 adds unrequested tests, docs and
small files "at every effort level, and more at higher effort" - for limits on
additions alone, use only the second paragraph:

```text
Keep working until everything the user asked for is done, and only stop to ask when you can't go on without the user or before a risky step.

When the work the user asked for is done and checked, stop and report. Don't add features, tests, files, docs or refactors that weren't asked for. If you think one would help, mention it at the end instead of doing it.
```

Haiku 5.5 (note "new features, docs, or refactors"; single line break):

```text
Keep working until everything the user asked for is done, and only stop to ask when you can't go on without the user or before a risky step.
When the work the user asked for is done and checked, stop and report. Don't add new features, docs, or refactors that weren't asked for. If you think one would help, mention it at the end instead of doing it.
```

Both pages carry the same verification paragraph, for workers that report
"done" without test or build output:

```text
When you change code that can be run, built, or type-checked, run a real check that exercises the change before reporting it done: the project's tests, type-checker, or build, or the changed command itself. A syntax-only check, or a check command that failed to start, does not count; if all that is missing is the project's declared dependencies, install them with its own package manager and lockfile (e.g. npm install, pip install -r requirements.txt), never via sudo or the system package manager, unless told not to. Only if no real check can run here, say which one you did not run and why instead of reporting the change as done.
```

It composes with the brief's `Verify with:` line - the brief names the check,
this block stops the worker skipping it.

## Never Ask a Worker to Write Out Its Reasoning

On Opus 5.5, Sonnet 5.5 and Fable 5.1, requests "that push the model to
reproduce its internal reasoning in the response text may be declined with the
`reasoning_extraction` category." A brief that says "explain your reasoning
step by step" or asks for `<thinking>` tags in the handoff risks a refusal
instead of a contract. The handoff contract's `Summary` asks for what changed,
which is the safe form; keep it that way.

## Sources

- `references/raw/anthropic-prompting-claude-opus-5-5-2026-10-08.md` - effort
  defaults, unattended runs, the nudge and standing-instruction blocks, time
  signals, reasoning extraction.
- `references/raw/anthropic-prompting-claude-sonnet-5-5-2026-10-08.md` - effort,
  scope, reviewer-subagent and verification blocks.
- `references/raw/anthropic-prompting-claude-haiku-5-5-2026-10-08.md` - effort
  levels, early stopping, verification.
- `references/raw/anthropic-prompting-claude-opus-5-2026-10-08.md` - the
  delegation damping block the Opus 5.5 guide still points to.
- `references/raw/anthropic-models-overview-2026-10-08.md` and
  `references/raw/anthropic-choosing-a-model-2026-10-08.md` - IDs, prices,
  default efforts, "start with Claude Opus 5.5".
- `references/raw/anthropic-api-and-data-retention-2026-10-08.md` - the four
  Covered Models (Fable 5.1, Mythos 5.1, Fable 5, Mythos 5); the 5.5 models are
  not among them.
- `references/raw/orchestrator-model-sweep-2026-10-08.md` - reachability and
  real-brief results.
- `references/raw/anthropic-prompting-claude-fable-5-1-2026-10-08.md` - the
  same `reasoning_extraction` decline on Fable 5.1.
