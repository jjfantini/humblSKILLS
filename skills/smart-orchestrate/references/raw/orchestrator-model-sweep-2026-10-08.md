# Orchestrator / worker model sweep — 2026-10-08

Purpose: establish which of the models released since the 2026-09-10 sweep
(Claude Opus 5.5, Sonnet 5.5, Haiku 5.5; OpenAI GPT-6.1 Sol, GPT-6 Sol, GPT-6
Luna) are reachable as a parent or worker backend from the CLIs this skill
dispatches through, before naming any of them in a routing table.

Method follows the 2026-08-13 and 2026-09-10 sweeps so the numbers stay
comparable: explicit model on every arm, 3 trials per arm (topped up to 6 for
every ID that enters a parent or worker table), a known-good control at
**both** ends of each run.

## Environment

```
date                 2026-10-08 (15:05-15:27 UTC)
host                 macOS arm64 (Darwin 25.5.0)
claude (Claude Code) 2.1.294
codex-cli            0.160.0, ChatGPT-account auth
cursor-agent         2026.08.25-3e8eec8 (NOT measured, see below)
ping                 "reply with only: OK"
claude invocation    claude -p --model <ID> --output-format json "<ping>"
codex invocation     echo "" | codex exec -m <ID> -c model_reasoning_effort=low \
                       --sandbox read-only --skip-git-repo-check -C <tmp> "<ping>"
pass criterion       rc=0 AND reply contains "OK" (claude: is_error=false too)
resolved model       claude: the key of `modelUsage` in the JSON result
```

Codex pings ran at `model_reasoning_effort=low` because the local Codex config
defaults to `xhigh`; latency numbers are therefore for `low` and are not
comparable to the 2026-09-10 Astra figure if that run used a higher default.

## Inventory checks (before any ping)

Codex model cache (`~/.codex/models_cache.json`, fetched 2026-10-08), listed
models and their efforts:

```
gpt-6.1-sol   "Latest workhorse model for coding and everyday work."   default medium   low..max + ultra
gpt-6-astra   "Frontier intelligence for the most demanding work."     default low      low..max + ultra
gpt-6-sol     "Previous generation workhorse model."                   default medium   low..max + ultra
gpt-6-luna    "Fast and affordable model for easier tasks."            default medium   low..max (no ultra)
gpt-5.6-sol / gpt-5.6-terra / gpt-5.6-luna  (older)
hidden: gpt-reserve, gpt-5.5, codex-auto-review
```

`ultra` is described in the cache as "Maximum reasoning with automatic task
delegation".

Claude Code: `claude --help` documents `--model` aliases (`fable`, `opus`,
`sonnet`) and `--effort <level>` (low, medium, high, xhigh, max).

## Claude Code results

| Arm | Result | Latency | Resolved to |
|---|---|---|---|
| `claude-fable-5-1` (control, start) | 3/3 | 6s | `claude-fable-5-1` |
| `claude-opus-5-5` | 6/6 | 5-11s | `claude-opus-5-5` |
| `claude-sonnet-5-5` | 6/6 | 4-5s | `claude-sonnet-5-5` |
| `claude-haiku-5-5` | 6/6 | 3-5s | `claude-haiku-5-5` |
| alias `opus` (n=1) | 1/1 | 6s | `claude-opus-5-5` |
| alias `sonnet` (n=1) | 1/1 | 4s | `claude-sonnet-5-5` |
| alias `haiku` (n=1) | 1/1 | 3s | `claude-haiku-5-5` |
| alias `fable` (n=1) | 1/1 | 6s | `claude-fable-5-1` |
| `claude-fable-5-1` (control, end of both runs) | 5/5 | 5-7s | `claude-fable-5-1` |

Control lifetime on Claude Code: 4/4 (2026-09-10) + 8/8 today.

Alias resolution is a property of this account's provider (Anthropic API). The
Claude Code model-config page documents that on Bedrock / Google Cloud `haiku`
resolves to Haiku 4.5 and `sonnet` to Sonnet 4.5.

## Codex results

| Arm | Result | Latency | Note |
|---|---|---|---|
| `gpt-6-astra` low (control, start) | 3/3 | 6-7s | |
| `gpt-6.1-sol` low | 6/6 | 6-8s | |
| `gpt-6-sol` low | 6/6 | 5-9s | |
| `gpt-6-luna` low | 6/6 | 5-6s | |
| `gpt-6` (n=1) | 0/1 | 4s | 400 `The 'gpt-6' model is not supported when using Codex with a ChatGPT account.` |
| `gpt-6.1-astra` (n=1) | 0/1 | 4s | same 400 message |
| `gpt-6.1-luna` (n=1) | 0/1 | 3s | same 400 message |
| `gpt-6-astra` medium (n=1) | 1/1 | 6s | `medium` accepted |
| `gpt-6.1-sol` ultra (n=1) | 1/1 | 7s | `ultra` accepted |
| `gpt-6-luna` ultra (n=1) | 1/1 | 6s | accepted by the CLI although docs and cache say Luna has no `ultra`; a ping cannot show whether anything was delegated |
| `gpt-6-astra` low (control, end of both runs) | 5/5 | 6-8s | |

Before each 400 the CLI prints `warning: Model metadata for '<id>' not found.
Defaulting to fallback metadata`. The 400 text is identical for an ID that does
not exist (`gpt-6.1-astra`) and for a family name (`gpt-6`), so on a ChatGPT
account this message does not distinguish "invalid ID" from "not entitled".

## Real file-writing briefs (tool use, not just transport)

Brief: create `brackets.py` with `is_balanced(s)`, written in the
contracts/brief-template.md shape with the handoff contract block and the
Anthropic autonomy opening sentence. Graded independently by importing the file
and running 9 cases, including the interleaved `([)]` and `)(`.

| Worker | Effort | Wall clock | Code | Contract fields | Extra files |
|---|---|---|---|---|---|
| `claude-haiku-5-5` (Claude Code `-p`) | low | 11-13s | 9/9 correct (2 runs) | 7/7 | none |
| `claude-sonnet-5-5` (Claude Code `-p`) | low | 11s | 9/9 correct | 7/7 | none |
| `gpt-6-luna` (codex exec, stdin) | low | 11-16s | 9/9 correct (2 runs) | 7/7 | none |
| `gpt-6.1-sol` (codex exec, stdin) | low | 28s | 9/9 correct | 7/7 | none |

Haiku 5.5 kept its self-check asserts inside the one file it was scoped to and
said so under Risks; nothing was committed.

The brief is short (~1 KB). It does not exercise the long-agent-prompt
conditions under which Anthropic documents Haiku 5.5 stopping early or skipping
checks at `low`.

## Codex: argv prompts over ~1 KB were SIGKILLed on this machine

Observed while dispatching the briefs above. The same brief passed as the
positional prompt argument exited **rc=137 in 0s with no output at all**, four
times across `gpt-6-luna` and `gpt-6.1-sol`; passed on stdin (`codex exec ...
-`) it ran normally every time.

Bisection with synthetic single-line prompts on `gpt-6-luna`:

```
len=915   rc=0   5s
len=960   rc=0  11s
len=1010  rc=137 0s
len=1023  rc=137 0s
len=1024  rc=137 0s
len=1115  rc=137 0s
len=1315  rc=137 1s
```

The threshold sits between 960 and 1010 characters. Cause not isolated: the
local Codex install runs several user hooks, and the kill happens before any
hook line is printed. Treat the number as this machine's, and the rule (send
briefs on stdin) as safe everywhere.

## Cursor CLI: not re-measured

`cursor-agent --list-models` and `cursor-agent update` both returned
`Error: Authentication required` / `[unauthenticated]` on 2026-10-08; auth had
expired and the interactive login was not performed during this session. The
local build is still 2026.08.25-3e8eec8, while the vendor installer pins
2026.10.01-e373342. No Cursor row was added or changed on the strength of a
measurement today.

Status page at the time (status.cursor.com summary, 15:13 UTC): open incident
"Opus 5.5 is Experiencing 429s", started 13:54 UTC, CLI component degraded.
