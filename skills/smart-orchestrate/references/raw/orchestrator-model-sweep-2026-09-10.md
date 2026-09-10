# Orchestrator / director model sweep — 2026-09-10

Purpose: establish which of the two newest frontier models (Claude Fable 5.1,
GPT-6 Astra) are actually reachable as an orchestrator or worker backend from
the CLIs this skill dispatches through, before naming either one in a routing
table.

Method follows the 2026-08-13 sweep so the numbers stay comparable: explicit
`--model` on every arm, 3 trials per arm, `gpt-5.3-codex-low-fast` as the
control at **both** ends of the run so a live vendor incident cannot be
mistaken for a bad model.

## Environment

```
date                2026-09-10
host                macOS arm64 (Darwin 25.5.0)
cursor-agent        2026.08.25-3e8eec8
codex-cli           0.154.0
claude (Claude Code) 2.1.267
ping                "reply with only: OK"
cursor invocation   cursor-agent -p --force --trust --model <ID> --workspace <tmp> "<ping>"
codex invocation    echo "" | codex exec -m <ID> --sandbox read-only --skip-git-repo-check -C <tmp> "<ping>"
pass criterion      rc=0 AND stdout contains "OK"
```

## Inventory checks (before any ping)

```
$ cursor-agent --list-models | wc -l
227

$ cursor-agent --list-models | grep -Ei 'gpt-6|astra'
(no output)

$ cursor-agent --list-models | grep -E 'fable-5-1'
claude-fable-5-1-low - Claude Fable 5.1 1M Low (NO ZDR)
claude-fable-5-1-medium - Claude Fable 5.1 1M Medium (NO ZDR)
claude-fable-5-1-high - Claude Fable 5.1 1M (NO ZDR)
claude-fable-5-1-xhigh - Claude Fable 5.1 1M Extra High (NO ZDR)
claude-fable-5-1-max - Claude Fable 5.1 1M Max (NO ZDR)
claude-fable-5-1-thinking-low - Claude Fable 5.1 1M Low Thinking (NO ZDR)
claude-fable-5-1-thinking-medium - Claude Fable 5.1 1M Medium Thinking (NO ZDR)
claude-fable-5-1-thinking-high - Claude Fable 5.1 1M Thinking (NO ZDR)
claude-fable-5-1-thinking-xhigh - Claude Fable 5.1 1M Extra High Thinking (NO ZDR)
claude-fable-5-1-thinking-max - Claude Fable 5.1 1M Max Thinking (NO ZDR)

$ claude --help | grep -A4 -- '--model'
  --model <model>   Model for the current session. Provide an alias for the
                    latest model (e.g. 'fable', 'opus', or 'sonnet') or a
                    model's full name (e.g. 'claude-fable-5').
```

So GPT-6 Astra is absent from the Cursor CLI's 227-ID list entirely, and every
Fable 5.1 ID that IS listed carries a `(NO ZDR)` marker — no zero-data-retention
guarantee.

## Raw results

Tab-separated, in execution order. `verdict / backend / model / trial /
wall-clock / rc / first-or-last 160 chars of combined output`.

```
PASS	cursor	gpt-5.3-codex-low-fast	trial=1	14s	rc=0	OK
PASS	cursor	gpt-5.3-codex-low-fast	trial=2	5s	rc=0	OK
PASS	cursor	gpt-5.3-codex-low-fast	trial=3	6s	rc=0	OK
FAIL	cursor	claude-fable-5-1-high	trial=1	4s	rc=1	ActionRequiredError: Model Blocked Please ask your admin to enable access to Claude Fable 5.
FAIL	cursor	claude-fable-5-1-high	trial=2	4s	rc=1	ActionRequiredError: Model Blocked Please ask your admin to enable access to Claude Fable 5.
FAIL	cursor	claude-fable-5-1-high	trial=3	3s	rc=1	ActionRequiredError: Model Blocked Please ask your admin to enable access to Claude Fable 5.
FAIL	cursor	claude-fable-5-1-thinking-high	trial=1	13s	rc=1	ActionRequiredError: Model Blocked Please ask your admin to enable access to Claude Fable 5.
FAIL	cursor	claude-fable-5-1-thinking-high	trial=2	3s	rc=1	ActionRequiredError: Model Blocked Please ask your admin to enable access to Claude Fable 5.
FAIL	cursor	claude-fable-5-1-thinking-high	trial=3	5s	rc=1	ActionRequiredError: Model Blocked Please ask your admin to enable access to Claude Fable 5.
FAIL	cursor	claude-fable-5-1-max	trial=1	6s	rc=1	ActionRequiredError: Model Blocked Please ask your admin to enable access to Claude Fable 5.
FAIL	cursor	claude-fable-5-1-max	trial=2	6s	rc=1	ActionRequiredError: Model Blocked Please ask your admin to enable access to Claude Fable 5.
FAIL	cursor	claude-fable-5-1-max	trial=3	5s	rc=1	ActionRequiredError: Model Blocked Please ask your admin to enable access to Claude Fable 5.
PASS	cursor	gpt-5.3-codex-low-fast	trial=4	7s	rc=0	OK
PASS	cursor	gpt-5.3-codex-low-fast	trial=5	5s	rc=0	OK
PASS	cursor	gpt-5.3-codex-low-fast	trial=6	5s	rc=0	OK
PASS	codex	gpt-6-astra	trial=1	7s	rc=0	codex OK / tokens used 9,278
PASS	codex	gpt-6-astra	trial=2	7s	rc=0	codex OK
PASS	codex	gpt-6-astra	trial=3	6s	rc=0	codex OK
```

Plus one earlier ad-hoc probe of `codex -m gpt-6-astra` that also returned `OK`,
making that arm 4/4 in total.

## Follow-up: Fable 5.1 through the Claude Code CLI

Run after the table above, because the Cursor block says nothing about a Claude
Code session, which authenticates against the caller's own subscription rather
than a Cursor entitlement.

```
$ claude -p --model claude-fable-5-1 "reply with only: OK"     # x4
OK   7s
OK   6s
OK   6s
OK   (initial ad-hoc probe)

$ claude -p --model fable "reply with only: OK"
OK
```

4/4 at 6-7s on Claude Code 2.1.267. Both the full ID and the `fable` alias
resolve, so the Cursor entitlement block is specific to that account's Cursor
organisation and does not describe Fable 5.1's availability generally.

## Follow-up probe: is the block specific to 5.1?

```
$ cursor-agent -p --force --trust --model claude-fable-5-high --workspace <tmp> "reply with only: OK"
ActionRequiredError: Model Blocked Please ask your admin to enable access to Claude Fable 5.
```

No. Claude Fable 5 is blocked with the identical message, so this is an
account/organisation entitlement covering the whole Fable family on this Cursor
account — consistent with the `(NO ZDR)` marker and an org policy that permits
only zero-data-retention models. It is **not** a Fable 5.1 regression and not a
transport failure.

## Summary

| Backend | Model ID | Score | Wall clock | Class |
|---|---|---|---|---|
| cursor-agent | `gpt-5.3-codex-low-fast` (control) | 6/6 | 5-14s | good — no incident drift, matches its 15/15 on 2026-08-13 |
| cursor-agent | `claude-fable-5-1-high` | 0/3 | 3-4s | account/entitlement block |
| cursor-agent | `claude-fable-5-1-thinking-high` | 0/3 | 3-13s | account/entitlement block |
| cursor-agent | `claude-fable-5-1-max` | 0/3 | 5-6s | account/entitlement block |
| cursor-agent | `claude-fable-5-high` (probe, n=1) | 0/1 | 4s | account/entitlement block |
| codex | `gpt-6-astra` | 4/4 | 6-7s | good |
| claude (Claude Code) | `claude-fable-5-1` | 4/4 | 6-7s | good — subscription auth, unaffected by the Cursor block |
| claude (Claude Code) | `fable` (alias) | 1/1 | ~6s | good |

## Findings

1. **A fifth failure class exists.** The 2026-08-13 table names four
   (`AI Model Not Found`, `WritableIterable is closed`,
   `NonRetriableError: Provider Error`, `auto`). This is a fifth:
   `ActionRequiredError: Model Blocked`, ~3-13s, **never retryable**, and unlike
   the other four it is *environment-specific* — the same ID may work fine on
   another Cursor account whose admin has enabled the model.
2. **`--list-models` over-reports in a second, different way.** Previously it
   advertised IDs the API rejects as invalid (`gpt-5.4-nano-*`). Here it
   advertises IDs that are perfectly valid but that this *account* is not
   entitled to use. Listing remains no evidence of reachability.
3. **GPT-6 Astra is reachable via the Codex CLI and not via Cursor.** 4/4 at
   6-7s on `codex-cli` 0.154.0; absent from all 227 Cursor IDs.
4. **The control design earned its keep again.** Six control passes bracketing
   nine consecutive failures is what makes "the Fable arms are blocked" a
   defensible conclusion rather than "something was down at 11:50."
5. **"Is model X available?" has no single answer — it is per-CLI.** Fable 5.1
   is 0/9 on cursor-agent and 4/4 on Claude Code on the same machine at the same
   hour, because the two authenticate against different accounts. So a routing
   table must be indexed by backend, not by model alone, and "blocked" is never
   a property of the model.
6. **`(NO ZDR)` is a routing constraint in its own right**, independent of
   reachability. Even with the entitlement enabled, dispatching proprietary
   source to a no-zero-data-retention worker is a policy decision, not a
   performance one.

## Not measured

- `claude-fable-5-1-{low,medium,xhigh}` and the remaining `-thinking-*` tiers:
  pointless to ping while the family-wide entitlement block stands.
- Which concrete model the `fable` alias resolves to. It answers, but the alias
  is documented as "the latest model" of that family, so it is not a stable
  routing target — name the full ID in a brief.
- `gpt-6-astra` at each of `low` / `high` / `max` effort: only the CLI default
  was pinged.
- Real file-writing briefs. Per the 2026-08-13 rule, a ping proves the transport,
  not tool use.
