---
title: "Which Cursor CLI Model IDs Are Actually Reachable"
context: orchestrate
category: routing
concept: cursor-cli-models
description: "Reliability is per-model-ID, never auto, and --list-models advertises both invalid IDs and IDs your account is blocked from"
tags: cursor, cursor-agent, model-selection, reliability, worker, backend, cli
sources:
  - "references/raw/cursor-cli-model-sweep-2026-08-13.md"
  - "references/raw/orchestrator-model-sweep-2026-09-10.md"
last_ingested: 2026-09-10
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

## `--list-models` Over-Reports, in Two Different Ways

It advertises IDs the API rejects. `gpt-5.4-nano-medium` and `gpt-5.4-nano-low`
are both listed and both return `AI Model Not Found`. Appearing in
`--list-models` is not evidence an ID works — a 3-trial ping is.

It also advertises IDs that are perfectly valid but that *this account* is not
entitled to use. Measured 2026-09-10: all ten `claude-fable-5-1-*` IDs are
listed, and all three tested returned `ActionRequiredError: Model Blocked`
(0/9). This is the more dangerous of the two, because the ID is real, the
spelling is right, and the failure is invisible until dispatch.

## The Newest Frontier Models, Measured 2026-09-10

`cursor-agent` 2026.08.25-3e8eec8, same protocol as the 2026-08-13 sweep, with
`gpt-5.3-codex-low-fast` as the control at **both** ends — 6/6, matching its
15/15 from August, so nothing below is incident drift.

| Model ID | Result | Note |
|---|---|---|
| `gpt-5.3-codex-low-fast` (control) | 6/6, 5-14s | now 21/21 lifetime |
| `claude-fable-5-1-high` | 0/3, 3-4s | `Model Blocked` — account entitlement |
| `claude-fable-5-1-thinking-high` | 0/3, 3-13s | `Model Blocked` |
| `claude-fable-5-1-max` | 0/3, 5-6s | `Model Blocked` |
| `claude-fable-5-high` (probe, n=1) | 0/1, 4s | `Model Blocked` — same message, so the block covers the whole Fable family, not just 5.1 |

**GPT-6 Astra is not on the Cursor CLI at all.** `--list-models` returns 227 IDs
and none of them matches `gpt-6` or `astra`. It is reachable through the Codex
CLI — `codex -m gpt-6-astra`, measured 4/4 at 6-7s on `codex-cli` 0.154.0 — so a
brief that needs Astra routes to Codex, not Cursor.

**Every listed Fable ID carries `(NO ZDR)`.** That is a routing constraint in its
own right and it survives the entitlement being enabled: Fable 5.1 carries 30-day
retention and isn't offered under zero data retention without express Anthropic
authorization. Proprietary source plus a ZDR policy rules the ID out on grounds
that have nothing to do with its score.

## Listed, Not Yet Measured

Do not promote anything here into **Verified Good** without pinging it. Left
unmeasured on 2026-09-10 because the family-wide entitlement block made a ping
uninformative: `claude-fable-5-1-{low,medium,xhigh}` and
`claude-fable-5-1-thinking-{low,medium,xhigh,max}`.

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
| `gpt-5.3-codex-low-fast` | 4-7s | Control arm, 21/21 across all sweeps. Cheapest reliable worker slot. |
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
| `claude-fable-5-medium` | 8s | 3/3 on 2026-08-13, but the whole Fable family was **entitlement-blocked** on the same account on 2026-09-10. Do not route here without re-pinging. |
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

## Five Failure Classes, and the Retry Policy for Each

This mapping is the actionable part; the score tables are supporting evidence.

| Error | Speed | Retry? |
|---|---|---|
| `ActionRequiredError: AI Model Not Found` | ~3-4s | **No.** The ID is invalid. Retrying burns every attempt. Fix the ID. |
| `ActionRequiredError: Model Blocked` | ~3-13s | **No.** The ID is valid; this *account* isn't entitled to it. Only an admin enabling the model fixes it. Route to a different model. |
| `RetriableError: WritableIterable is closed` | ~26-48s | **Yes** — this is the only retryable class. |
| `NonRetriableError: Provider Error` | ~3s | **No.** Upstream is down. Vendor labels it non-retriable; switch model or wait. |
| `auto` producing any of the above | ~31s | **No.** Pass an explicit `--model` instead. |

`Model Blocked` is the only class that is **environment-specific**: the same ID
may be 3/3 on another Cursor account whose admin has enabled it. So a `Model
Blocked` row in this table is a fact about the measured account, not about the
model — unlike an invalid ID, which is wrong everywhere. Re-check it yourself
before concluding a model is unavailable to you.

`scripts/dispatch-cursor-worker.sh` implements exactly this: it greps for
`WritableIterable is closed` and retries only that, surfacing everything else to
the parent immediately — so the new `Model Blocked` class already fails loudly
on the first attempt without a script change.

## A Ping Proves the Transport, Not Tool Use

`"reply with only: OK"` exercises the stream and nothing else. Before adopting a
new ID as a worker backend, dispatch one real file-writing brief and verify the
output yourself. `gpt-5.6-luna-high` was confirmed this way on a bracket matcher
(correct on the interleaved `([)]` case); `gpt-5.3-codex-low-fast` on a roman
numeral converter (correct on every subtractive form).

Note that even a verified-good ID is not perfectly immune:
`gpt-5.6-luna-high` printed one `Connection lost` and recovered on
`cursor-agent`'s own internal retry. Keep the wrapper's retry loop.

## Anthropic Rows Are Dated — and Entitlements Drift

status.cursor.com carried an **open** incident throughout the August
measurements (2026-08-13 14:46 UTC), naming Claude Mythos 5, Claude Fable 5, and
Claude Sonnet 5. The `claude-sonnet-5-*`, `claude-fable-5-*`, and Opus rows are
therefore that day's weather, not the climate — they are lower-confidence than
the GPT and Gemini rows. Re-measure after any Cursor status-page incident rather
than trusting this table.

The Fable rows demonstrate a second, slower kind of drift. `claude-fable-5-medium`
was 3/3 on 2026-08-13 and the entire Fable family returned `Model Blocked` on
2026-09-10 — nothing about the model changed, the account's entitlements did. An
availability table therefore has **two** expiry mechanisms: the vendor's uptime
and your own organisation's policy. A row that was green a month ago is a
hypothesis, not a fact.

## Sources

- `references/raw/cursor-cli-model-sweep-2026-08-13.md` — the 30-good/19-bad
  table and the four original failure classes.
- `references/raw/orchestrator-model-sweep-2026-09-10.md` — the Fable 5.1
  entitlement block, the fifth failure class, and GPT-6 Astra's absence from the
  227-ID list.
