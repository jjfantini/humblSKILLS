---
title: "Which Cursor CLI Model IDs Are Actually Reachable"
context: orchestrate
category: routing
concept: cursor-cli-models
description: "Reliability is per-model-ID, never auto, flat IDs run the 300K variant, a ping is not a brief - and every gpt-* row expires at OpenAI's proposed 2026-11-12 Cursor cutoff"
tags: cursor, cursor-agent, model-selection, reliability, worker, backend, cli
sources:
  - "references/raw/cursor-cli-model-sweep-2026-08-13.md"
  - "references/raw/orchestrator-model-sweep-2026-09-10.md"
  - "references/raw/orchestrator-model-sweep-2026-10-08.md"
  - "references/raw/cursor-cli-model-sweep-2026-10-08.md"
  - "references/raw/anthropic-prompting-claude-haiku-5-5-2026-10-08.md"
  - "references/raw/cursor-forum-openai-models-after-nov-12-2026-10-08.md"
  - "references/raw/cursor-cli-changelog-2026-10-08.md"
  - "references/raw/cursor-forum-cli-bracket-overrides-172338-2026-10-08.md"
  - "references/raw/cursor-model-claude-opus-5-5-2026-10-08.md"
  - "references/raw/cursor-model-claude-fable-5-1-2026-10-08.md"
  - "references/raw/cursor-enterprise-privacy-data-governance-2026-10-08.md"
last_ingested: 2026-10-08
---

## This Concept Is About Backend Reachability, Not Task Difficulty

`references/wiki/orchestrate/routing/model-selection.md` decides *how strong* a
worker a subtask deserves. This concept decides *which concrete `--model` value*
can be dispatched to at all. Both apply: pick the tier from blast radius, then
pick an ID from this table.

Measured 2026-08-13 (`cursor-agent` 2026.08.11-e8db854, macOS arm64), 3 trials
per arm, with `gpt-5.3-codex-low-fast` repeated as a control throughout — every
control instance returned 3/3, so no vendor incident is confounding the numbers.

## Read This First: Measured 2026-10-08 on CLI 2026.10.01

Measured after `cursor-agent update` (2026.08.25 -> 2026.10.01-e373342) with
every dispatch's init-event `model` recorded - see
`references/raw/cursor-cli-model-sweep-2026-10-08.md`. Controls
`gpt-5.3-codex-low-fast` 6/6 and `gemini-3-flash` 13/13 at both ends.

| Model ID | Pings | Real brief | What ran (init event) | Verdict |
|---|---|---|---|---|
| `claude-haiku-5-5-thinking-medium` | 6/6, 5-6s | 9/9, 15s | Claude Haiku 5.5 300K Medium | **Cheapest reliable worker** |
| `claude-sonnet-5-5-medium` | 6/6, 5-8s | 9/9, 18s | Claude Sonnet 5.5 300K Medium | **Mid-tier worker** |
| `claude-opus-5-thinking-high` | 3/3, 20-28s | 9/9, 24s | Claude Opus 5 300K High | **Hard briefs** (still) |
| `claude-opus-5-5-high` | 7/7, 7-23s | 0/2 plain; with capacity retry 1/2, 272s | Claude Opus 5.5 300K High | Throttled on real briefs - re-test |
| `claude-opus-5-5-medium-fast` | 4/6 | - | Claude Opus 5.5 300K Medium Fast | `resource_exhausted` |
| `claude-opus-5-5-medium` | 1/6 | - | Claude Opus 5.5 300K Medium | `resource_exhausted` |
| `claude-haiku-5-5-medium` | 3/3 | - | ... Medium **No Thinking** | Thinking off - prefer the `-thinking-` ID |
| `claude-haiku-5-5-xhigh` | 2/2 | - | ... Extra High **No Thinking** | Do not use: Anthropic returns 400 for thinking-off at `xhigh`, so what Cursor sends is unknown |
| `claude-fable-5-1-high` | 0/3 | - | Fable 5.1 300K High No Thinking | `Model Blocked` (entitlement, unchanged) |
| `grok-4.7-high` | 1/3 | - | Grok 4.7 256K High | stream teardown |
| `grok-4.7-low-fast` | 0/3 | - | Grok 4.7 256K Low Fast | stream teardown |
| `composer-2.5` | 0/3 | - | Composer 2.5 | stream teardown |

Five things this sweep established:

**1. Flat IDs run the 300K variant, whatever the label says.** Every Claude
flat ID ran a "300K" model, although `--list-models` labels the Opus and Fable
IDs "1M". Thinking shows in the init name too: flat Haiku and Fable IDs ran
"No Thinking"; use the `-thinking-` IDs where thinking matters.

**2. A ping is not a brief on Opus 5.5.** Opus 5.5 `high` passed every ping and
failed real briefs with `RetriableError: [resource_exhausted]` until the
retry-with-backoff added to `scripts/dispatch-cursor-worker.sh` got one through
in 4.5 minutes. Opus 5 thinking-high did the same brief in 24s. Route hard
Cursor briefs to `claude-opus-5-thinking-high` until a re-test shows Opus 5.5
taking briefs without retries.

**3. Every `gpt-*` row has an expiry date.** OpenAI announced it will end
Cursor's model access after SpaceX acquired Cursor; "The proposed cutoff date is
**November 12, 2026**, although it is not yet final." No GPT-6 ID exists on
Cursor (256 IDs, none match), and a bring-your-own OpenAI key "will not cover
features such as Tab, Auto, Cloud Agents, Automations, or Cursor CLI." Route
GPT-6 briefs to the Codex CLI. **The sweep control is now `gemini-3-flash`**
(16/16 lifetime); keep `gpt-5.3-codex-low-fast` alongside it only until the
cutoff.

**4. Grok and Composer are still broken on the CLI.** Grok 4.7 - now Cursor's
flagship and without the `cursor-` prefix - scored 1/6 across two tiers, and
`composer-2.5` 0/3, all stream teardowns. The pattern from Grok 4.5 / 4.6
carried straight over.

**5. Rows from builds before v2026.09.28 prove the transport, not the model.**
The v2026.09.28 changelog: "The CLI runs the exact model you pick. `--model`
and `/model` now run the model you choose even when its ID starts with another
model's ID. Previously the CLI could silently run the shorter base model
instead." The older tables below ran on 2026.08.11 / 2026.08.25. Re-measure a
row on 2026.09.28+ with the init event before relying on it; the four re-checked
today (Codex 5.3 Low Fast, Gemini 3 Flash, Opus 5 thinking-high, Fable 5.1)
ran the model their ID names.

The `--model` variant-string syntax is also documented by Cursor staff -
"`--model` only accepts the full variant string (base id plus every parameter,
in order)," e.g.
`'claude-opus-4-8[thinking=true,context=1m,effort=high,fast=false]'` - and is
the only way to get the 1M context from the CLI. It was not measured here.

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
| `gpt-5.3-codex-low-fast` | 4-7s | 27/27 across all sweeps. Expires with the proposed 2026-11-12 OpenAI cutoff; `claude-haiku-5-5-thinking-medium` replaces it as the cheap slot. |
| `glm-5.2-high` | 4s | |
| `gemini-3-flash` | 4s | The sweep control from 2026-10-08 (16/16). |
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
| `claude-opus-5-thinking-high` | 9s | Strongest verified worker. Use for hard briefs. Re-confirmed 2026-10-08 on 2026.10.01 (20-28s, real brief 24s). |
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
| `composer-2.5` | 1/4, then 0/3 on 2026-10-08 | stream teardown |
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

## Six Failure Classes, and the Retry Policy for Each

This mapping is the actionable part; the score tables are supporting evidence.

| Error | Speed | Retry? |
|---|---|---|
| `ActionRequiredError: AI Model Not Found` | ~3-4s | **No.** The ID is invalid. Retrying burns every attempt. Fix the ID. |
| `ActionRequiredError: Model Blocked` | ~3-13s | **No.** The ID is valid; this *account* isn't entitled to it. For Fable, "requests to these models fail until the model's data retention policy is approved" - an admin approves it under the dashboard's `restricted_models` page. Otherwise route to a different model. |
| `RetriableError: WritableIterable is closed` | ~26-48s | **Yes** — stream teardown before any work. |
| `RetriableError: [resource_exhausted] Error` | ~22-28s | **Yes, with backoff** — provider capacity (or a team quota; the client cannot tell). Seen only on Opus 5.5, 2026-10-08. The script backs off 15s x attempt. |
| `NonRetriableError: Provider Error` | ~3s | **No.** Upstream is down. Vendor labels it non-retriable; switch model or wait. |
| `auto` producing any of the above | ~31s | **No.** Pass an explicit `--model` instead. |

A passing Fable ping on Cursor is weaker evidence than it looks: "When a
request trips one of those guardrails, Cursor routes it to Claude Opus
automatically so your work continues without an error." Check the init event's
`model` field, not the exit code.

`Model Blocked` is the only class that is **environment-specific**: the same ID
may be 3/3 on another Cursor account whose admin has enabled it. So a `Model
Blocked` row in this table is a fact about the measured account, not about the
model — unlike an invalid ID, which is wrong everywhere. Re-check it yourself
before concluding a model is unavailable to you.

`scripts/dispatch-cursor-worker.sh` implements exactly this: it retries
`WritableIterable is closed` immediately and `resource_exhausted` with backoff,
and surfaces everything else to the parent on the first attempt - so `Model
Blocked` and invalid IDs still fail loudly. The dirty-tree guard applies to both
retryable classes.

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
- `references/raw/cursor-cli-model-sweep-2026-10-08.md` — the 2026-10-08
  measurement on CLI 2026.10.01: init-event models, the `resource_exhausted`
  class, real briefs, the `gemini-3-flash` control.
- `references/raw/anthropic-prompting-claude-haiku-5-5-2026-10-08.md` — why
  `claude-haiku-5-5-xhigh` "No Thinking" cannot be what it says.
- `references/raw/orchestrator-model-sweep-2026-10-08.md` — the morning run,
  when Cursor auth had expired.
- `references/raw/cursor-forum-openai-models-after-nov-12-2026-10-08.md` — the
  proposed OpenAI cutoff. OpenAI's own announcement returned 403 to `curl`, so
  this forum summary of it is the saved source.
- `references/raw/cursor-cli-changelog-2026-10-08.md` — the prefix-ID fix in
  v2026.09.28 and the team-restriction exit.
- `references/raw/cursor-forum-cli-bracket-overrides-172338-2026-10-08.md` —
  staff-confirmed `--model` variant syntax.
- `references/raw/cursor-model-claude-opus-5-5-2026-10-08.md` and
  `references/raw/cursor-model-claude-fable-5-1-2026-10-08.md` — Opus 5.5 ZDR and
  coordinator notes; Fable approval and silent Opus fallback.
- `references/raw/cursor-enterprise-privacy-data-governance-2026-10-08.md` —
  which models fall outside Cursor's ZDR agreements.
