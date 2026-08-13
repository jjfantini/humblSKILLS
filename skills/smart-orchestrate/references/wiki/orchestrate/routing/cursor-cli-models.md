---
title: "Which Cursor CLI Model IDs Are Actually Reachable"
context: orchestrate
category: routing
concept: cursor-cli-models
description: "Reliability is per-model-ID, never auto, and --list-models advertises IDs the API rejects"
tags: cursor, cursor-agent, model-selection, reliability, worker, backend, cli
sources:
  - "references/raw/cursor-cli-model-sweep-2026-08-13.md"
last_ingested: 2026-08-13
---

## This Concept Is About Backend Reachability, Not Task Difficulty

`references/wiki/orchestrate/routing/model-selection.md` decides *how strong* a
worker a subtask deserves. This concept decides *which concrete `--model` value*
can be dispatched to at all. Both apply: pick the tier from blast radius, then
pick an ID from this table.

Measured 2026-08-13 (`cursor-agent` 2026.08.11-e8db854, macOS arm64), 3 trials
per arm, with `gpt-5.3-codex-low-fast` repeated as a control throughout — every
control instance returned 3/3, so no vendor incident is confounding the numbers.

## Never Dispatch on `auto`

`auto` scored **0/12** across two sweeps. It never succeeded once.

`auto` lets Cursor route to whatever provider it likes, including one that is
mid-incident, and the routing failure surfaces as a *client-side* stream
teardown rather than an honest upstream error. That misdirection cost a full
investigation into the transport, the sandbox, stdin, workspace trust, and
chat-session state before the model turned out to be the variable.

Always pass an explicit `--model`.

## `--list-models` Over-Reports

It advertises IDs the API rejects. `gpt-5.4-nano-medium` and `gpt-5.4-nano-low`
are both listed and both return `AI Model Not Found`. Appearing in
`--list-models` is not evidence an ID works — a 3-trial ping is.

## Reliability Is Per-Model-ID, Not Per-Family and Not Per-Tier

Three plausible shortcuts were tried and all three are wrong:

| Shortcut | Killed by |
|---|---|
| "Failure is per-family" | Grok 4.5: `low-fast` 9/9 while `medium` 0/3 and `high` 0/3, same family |
| "Slow-to-first-token arms hit a stream idle timeout" | `gpt-5.4-nano-medium` fails in ~4s |
| "The `-fast` suffix (priority capacity) is what works" | Grok 4.6: `low-fast` 0/3 and `xhigh-fast` 0/3 |

Measure the specific ID you intend to route to. Nothing generalises — not the
family, not the effort tier, not the `-fast` suffix.

Corollary: **do not trust a 3/3 that sits inside a mostly-broken family.**
`cursor-grok-4.5-high-fast` scored 3/3 in one sweep, then 4/6 on re-test — 7/9
overall, not reliable. Its sibling `low-fast` held at 9/9. A 3-trial pass is
enough to shortlist an ID and not enough to adopt one.

## Verified Good

All 3/3. Prefer these; they are ordered by measured latency.

| Model ID | avg | Note |
|---|---|---|
| `gpt-5.3-codex-low-fast` | 4-7s | Control arm, 15/15 across all sweeps. Cheapest reliable worker slot. |
| `glm-5.2-high` | 4s | |
| `gemini-3-flash` | 4s | |
| `gpt-5.6-luna-high` | 5s | Fastest GPT-5.6 tier. Tool use confirmed on a real brief. |
| `gpt-5.4-medium-fast` | 5s | |
| `gpt-5.4-mini-medium` | 5s | |
| `gpt-5-mini` | 5s | |
| `gpt-5.6-terra-medium-fast` | 5s | |
| `claude-opus-4-8-medium-fast` | 5s | |
| `claude-4.5-opus-high` | 5s | |
| `gpt-5.6-sol-xhigh-fast` | 6s | |
| `gpt-5.5-high-fast` | 6s | |
| `gpt-5.1` | 6s | |
| `gemini-3.6-flash-medium` | 6s | |
| `claude-4.6-sonnet-medium` | 6s | |
| `gpt-5.6-sol-high-fast` | 7s | |
| `gpt-5.2` | 7s | |
| `claude-4.6-opus-high` | 7s | |
| `claude-4.5-sonnet` | 7s | |
| `claude-4-sonnet` | 7s | |
| `claude-sonnet-5-medium` | 8s | Anthropic — see caveat below |
| `claude-fable-5-medium` | 8s | Anthropic — see caveat below |
| `claude-opus-5-thinking-high` | 9s | Strongest verified worker. Use for hard briefs. |
| `gemini-3.1-pro` | 9s | |
| `cursor-grok-4.5-low-fast` | 9s | 9/9. The ONLY reliable Grok ID of 14 tested. |
| `gemini-3.5-flash` | 10s | |
| `kimi-k3-low` | 11s | |
| `kimi-k2.7-code` | 13s | |
| `kimi-k3-high` | 18s | Slowest verified. |

## Verified Bad

| Model ID | Result | Failure class |
|---|---|---|
| `auto` | 0/12 | never succeeded — do not use |
| `gpt-5.4-nano-medium` | 0/3 | invalid ID |
| `gpt-5.4-nano-low` | 0/1 | invalid ID |
| `claude-opus-4-7-medium-fast` | 0/3 | provider error |
| `composer-2.5` | 1/4 | stream teardown |
| `composer-2.5-fast` | 1/3 | stream teardown |
| `cursor-grok-4.5-high-fast` | 7/9 | stream teardown — passed 3/3 once, then 4/6. Not reliable. |
| `cursor-grok-4.5-low` | 1/3 | stream teardown |
| `cursor-grok-4.5-medium` | 0/3 | stream teardown |
| `cursor-grok-4.5-medium-fast` | 1/6 | stream teardown |
| `cursor-grok-4.5-high` | 0/3 | stream teardown |
| `cursor-grok-4.6-low` | 0/3 | stream teardown |
| `cursor-grok-4.6-low-fast` | 0/4 | stream teardown |
| `cursor-grok-4.6-medium` | 1/3 | stream teardown |
| `cursor-grok-4.6-medium-fast` | 1/3 | stream teardown |
| `cursor-grok-4.6-high` | 0/3 | stream teardown |
| `cursor-grok-4.6-high-fast` | 1/3 | stream teardown |
| `cursor-grok-4.6-xhigh` | 1/3 | stream teardown |
| `cursor-grok-4.6-xhigh-fast` | 0/3 | stream teardown |

**Avoid Grok on the Cursor CLI.** All 14 tiers of 4.5 and 4.6 were measured:
Grok 4.6 came in around 3/24 with every single tier failing, and Grok 4.5 around
11/27 with all of its successes concentrated in `low-fast`. Only
`cursor-grok-4.5-low-fast` (9/9) is safe, and it is one reliable ID inside an
otherwise broken family — there is no cheaper Grok fallback if it regresses, so
prefer a GPT ID for anything load-bearing.

## Four Failure Classes, and the Retry Policy for Each

This mapping is the actionable part; the score tables are supporting evidence.

| Error | Speed | Retry? |
|---|---|---|
| `ActionRequiredError: AI Model Not Found` | ~3-4s | **No.** The ID is invalid. Retrying burns every attempt. Fix the ID. |
| `RetriableError: WritableIterable is closed` | ~26-48s | **Yes** — this is the only retryable class. |
| `NonRetriableError: Provider Error` | ~3s | **No.** Upstream is down. Vendor labels it non-retriable; switch model or wait. |
| `auto` producing any of the above | ~31s | **No.** Pass an explicit `--model` instead. |

`scripts/dispatch-cursor-worker.sh` implements exactly this: it greps for
`WritableIterable is closed` and retries only that, surfacing everything else to
the parent immediately.

## A Ping Proves the Transport, Not Tool Use

`"reply with only: OK"` exercises the stream and nothing else. Before adopting a
new ID as a worker backend, dispatch one real file-writing brief and verify the
output yourself. `gpt-5.6-luna-high` was confirmed this way on a bracket matcher
(correct on the interleaved `([)]` case); `gpt-5.3-codex-low-fast` on a roman
numeral converter (correct on every subtractive form).

Note that even a verified-good ID is not perfectly immune:
`gpt-5.6-luna-high` printed one `Connection lost` and recovered on
`cursor-agent`'s own internal retry. Keep the wrapper's retry loop.

## Anthropic Rows Are Dated

status.cursor.com carried an **open** incident throughout these measurements
(2026-08-13 14:46 UTC), naming Claude Mythos 5, Claude Fable 5, and Claude
Sonnet 5. The `claude-sonnet-5-*`, `claude-fable-5-*`, and Opus rows are
therefore today's weather, not the climate — they are lower-confidence than the
GPT and Gemini rows. Re-measure after any Cursor status-page incident rather
than trusting this table.

## Sources

- `references/raw/cursor-cli-model-sweep-2026-08-13.md`
