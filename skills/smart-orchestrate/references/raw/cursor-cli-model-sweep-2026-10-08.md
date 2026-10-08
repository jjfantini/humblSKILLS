# Cursor CLI model sweep — 2026-10-08 (afternoon)

Purpose: measure the Cursor CLI after re-authentication and a CLI update, so the
Claude 5.5 rows in `routing/cursor-cli-models.md` stop being
documented-not-measured, and establish a non-OpenAI control ahead of OpenAI's
proposed 2026-11-12 Cursor cutoff.

Method follows the 2026-08-13 and 2026-09-10 sweeps: explicit `--model`, 3
trials per arm, topped up to 6 for every ID entering a table, controls at both
ends of each run. New this time: every dispatch records the `model` field of
the first `system`/`init` event in `--output-format stream-json`, which names
the variant that actually ran.

## Environment

```
date                2026-10-08 (~18:30-19:30 UTC)
host                macOS arm64 (Darwin 25.5.0)
cursor-agent        2026.10.01-e373342 (updated from 2026.08.25-3e8eec8 via `cursor-agent update`)
auth                login, team HappyRobot
status page         "All Systems Operational" at start (the morning's Opus 5.5 429 incident had cleared)
ping                "reply with only: OK"
invocation          cursor-agent -p --force --trust --model <ID> --output-format stream-json --workspace <tmp> "<ping>"
pass criterion      rc=0 AND result event is_error=false AND result contains "OK"
```

## Inventory

`cursor-agent --list-models`: 256 IDs (227 on 2026-09-10), of which 83 start
with `gpt-` and 123 with `claude-`. No `gpt-6*` or `astra` ID. New families:

```
claude-opus-5-5-{low,medium,high,xhigh,max}[-fast]        labelled "Claude Opus 5.5 1M ..."
claude-sonnet-5-5-{low,medium,high,xhigh,max}
claude-haiku-5-5-{low,medium,high,xhigh,max}              labelled "... No Thinking"
claude-haiku-5-5-thinking-{low,medium,high,xhigh,max}
claude-fable-5-1-{low,medium,high,xhigh,max}              (NO ZDR)
claude-fable-5-1-thinking-{low,medium,high,xhigh,max}     (NO ZDR)
grok-4.7-{low,medium,high,xhigh}[-fast]                   (no cursor- prefix)
composer-2.5, composer-2.5-fast, auto
```

Still listed: `gpt-5.3-codex-low-fast`, `gpt-5.6-luna-high`,
`claude-opus-5-thinking-high`, `claude-4.6-sonnet-medium`, `gemini-3-flash`,
`cursor-grok-4.5-low-fast`.

## Results

| ID | Result | Latency | Init-event model | Failure |
|---|---|---|---|---|
| `gpt-5.3-codex-low-fast` (control) | 6/6 | 7-18s | Codex 5.3 Low Fast | |
| `gemini-3-flash` (new control) | 12/12 (+1 smoke) | 6-34s | Gemini 3 Flash | |
| `claude-opus-5-5-high` | 6/6 | 7-23s | Claude Opus 5.5 300K High | |
| `claude-opus-5-5-medium-fast` | 4/6 | 8-46s | Claude Opus 5.5 300K Medium Fast | 2x `RetriableError: [resource_exhausted] Error` |
| `claude-opus-5-5-medium` | 1/6 | 22-48s | Claude Opus 5.5 300K Medium | 5x `RetriableError: [resource_exhausted] Error` |
| `claude-sonnet-5-5-medium` | 6/6 | 5-8s | Claude Sonnet 5.5 300K Medium | |
| `claude-haiku-5-5-thinking-medium` | 6/6 | 5-6s | Claude Haiku 5.5 300K Medium | |
| `claude-haiku-5-5-medium` | 3/3 | 6-7s | Claude Haiku 5.5 300K Medium No Thinking | |
| `claude-haiku-5-5-xhigh` (n=2) | 2/2 | 5-6s | Claude Haiku 5.5 300K Extra High No Thinking | see note |
| `claude-fable-5-1-high` | 0/3 | 4-7s | Claude Fable 5.1 300K High No Thinking | `ActionRequiredError: Model Blocked Please ask your admin to enable access to Claude Fable 5.` |
| `grok-4.7-high` | 1/3 | 17-43s | Grok 4.7 256K High | 2x `RetriableError: WritableIterable is closed` |
| `grok-4.7-low-fast` | 0/3 | 32-46s | Grok 4.7 256K Low Fast | 3x `WritableIterable is closed` |
| `composer-2.5` | 0/3 | 41-77s | Composer 2.5 | 3x `WritableIterable is closed` |
| `claude-opus-5-thinking-high` (re-check) | 3/3 | 20-28s | Claude Opus 5 300K High | |

Latency rose for every arm in the second half of the first run (the controls
went from ~7s to ~17-34s), then fell back in the top-up run; no control failed.

Notes:

- **Flat IDs resolve to the 300K variant.** Every Claude 5.5 flat ID ran a
  "300K" model although `--list-models` labels the Opus and Fable IDs "1M" -
  the same mislabel Cursor staff confirmed for `claude-sonnet-5-high`.
- **Thinking is visible in the init name.** Flat `claude-fable-5-1-high` ran
  "High No Thinking"; flat `claude-haiku-5-5-medium` ran "Medium No Thinking";
  the `-thinking-` IDs drop the suffix. Opus 5.5 names never say "No Thinking"
  (its thinking cannot be turned off on the API).
- **`claude-haiku-5-5-xhigh` answers while claiming "No Thinking".**
  Anthropic's Haiku 5.5 guide says turning thinking off "works at `low`,
  `medium`, and `high` only. At `xhigh` and `max`, the request returns a 400
  error." What Cursor actually sends for this ID is therefore unknown.
- **A sixth failure class appeared:** `RetriableError: [resource_exhausted]
  Error`, ~22-28s, before any output. It hit only Opus 5.5 `medium` and
  `medium-fast` in pings, and Opus 5.5 `high` on real briefs, while
  Sonnet/Haiku 5.5 never saw it - consistent with provider capacity on Opus 5.5
  rather than a client fault.

## Real file-writing briefs (through `scripts/dispatch-cursor-worker.sh`)

Same bracket-matcher brief and independent 9-case grader as the morning's
Claude Code / Codex sweep (`orchestrator-model-sweep-2026-10-08.md`).

| Worker | Wall clock | Code | Contract fields | Retries |
|---|---|---|---|---|
| `claude-haiku-5-5-thinking-medium` | 15s | 9/9 correct | 7/7 | 0 |
| `claude-sonnet-5-5-medium` | 18s | 9/9 correct | 7/7 | 0 |
| `claude-opus-5-5-high` (script without capacity retry) | 20s, 2 runs | no file | 0/7 | n/a - exited on `resource_exhausted` both times |

The script then gained a retry with linear backoff for `resource_exhausted`
(sleep 15s x attempt, up to 6 attempts). Re-runs:

| Worker | Wall clock | Code | Contract fields | Retries |
|---|---|---|---|---|
| `claude-opus-5-5-high` run 1 | 453s | no file | 0/7 | 6, all `resource_exhausted` - gave up |
| `claude-opus-5-5-high` run 2 | 272s | 9/9 correct | 7/7 | 4 |
| `claude-opus-5-thinking-high` | 24s | 9/9 correct | 7/7 | 0 |

A ping on `claude-opus-5-5-high` right afterwards passed in 10s (7/7 lifetime).
So on this account, this afternoon, Opus 5.5 answered short pings reliably but
was capacity-throttled on a real (~15K-token) brief, while Opus 5 thinking-high
was not. Whether `resource_exhausted` is provider capacity or a team quota is
not distinguishable from the client.
