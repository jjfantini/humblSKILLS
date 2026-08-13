# Cursor CLI model reliability sweep — 2026-08-13

Raw measurement output. Provenance for
`references/wiki/orchestrate/routing/cursor-cli-models.md`.

Environment:

- `cursor-agent` 2026.08.11-e8db854, macOS arm64, `~/.local/bin/cursor-agent`
- Account: jennings.fantini@happyrobot.ai (Cursor team tier)
- Invocation per trial: `cursor-agent -p --force --trust --model <M> --workspace <dir> "reply with only: OK"`
- Trial counted as success on exit code 0
- 3 trials per model arm
- `gpt-5.3-codex-low-fast` used as a control arm, repeated throughout each
  sweep so that a worsening vendor incident could not be mistaken for a bad
  model. Every control instance returned 3/3.

Context: status.cursor.com carried an OPEN incident during these runs, opened
2026-08-13 14:46 UTC — "due to a service outage for one of our model
providers, agent surfaces using Anthropic Models Claude Mythos 5, Claude
Fable 5, and Claude Sonnet 5 may run into issues." Anthropic-backed rows are
therefore the least durable numbers here.

`cursor-agent --list-models` reported 201 model IDs. Exhaustively testing all
201 at 3 trials each would have taken roughly 8 hours of wall clock, so arms
were chosen as one representative per model family, plus every Grok tier
(Grok turned out to vary by tier, see below).

## Sweep 1 — initial model comparison

    auto                             0/9
    composer-2.5                     1/4
    cursor-grok-4.6-low-fast         0/1
    gpt-5.3-codex-low-fast           6/6
    claude-sonnet-5-thinking-high    3/3

## Sweep 2 — newer tiers, controls at both ends

    gpt-5.3-codex-low-fast           3/3  avg=7s     <- control
    gpt-5.6-sol-high-fast            3/3  avg=7s
    gpt-5.6-sol-xhigh-fast           3/3  avg=6s
    gpt-5.6-luna-high                3/3  avg=5s
    gpt-5.5-high-fast                3/3  avg=6s
    cursor-grok-4.6-medium-fast      1/3  avg=33s
    cursor-grok-4.6-high-fast        1/3  avg=27s
    cursor-grok-4.6-xhigh-fast       0/3  avg=48s
    claude-opus-5-thinking-high      3/3  avg=9s
    kimi-k3-high                     3/3  avg=18s
    gpt-5.2                          3/3  avg=7s
    auto                             0/3  avg=31s
    gpt-5.3-codex-low-fast           3/3  avg=5s     <- control

## Sweep 3 — one representative per family, controls interspersed

    gpt-5.3-codex-low-fast           3/3  avg=6s     <- control
    cursor-grok-4.5-low-fast         3/3  avg=9s
    cursor-grok-4.5-medium-fast      0/3  avg=26s
    cursor-grok-4.5-high-fast        3/3  avg=10s
    gpt-5.6-terra-medium-fast        3/3  avg=5s
    gpt-5.4-medium-fast              3/3  avg=5s
    gpt-5.3-codex-low-fast           3/3  avg=5s     <- control
    gpt-5.4-mini-medium              3/3  avg=5s
    gpt-5.4-nano-medium              0/3  avg=4s
    gpt-5.1                          3/3  avg=6s
    gpt-5-mini                       3/3  avg=5s
    composer-2.5-fast                1/3  avg=29s
    gpt-5.3-codex-low-fast           3/3  avg=6s     <- control
    claude-opus-4-8-medium-fast      3/3  avg=5s
    claude-opus-4-7-medium-fast      0/3  avg=3s
    claude-4.6-opus-high             3/3  avg=7s
    claude-4.5-opus-high             3/3  avg=5s
    claude-sonnet-5-medium           3/3  avg=8s
    gpt-5.3-codex-low-fast           3/3  avg=4s     <- control
    claude-4.6-sonnet-medium         3/3  avg=6s
    claude-4.5-sonnet                3/3  avg=7s
    claude-4-sonnet                  3/3  avg=7s
    claude-fable-5-medium            3/3  avg=8s
    kimi-k3-low                      3/3  avg=11s
    gpt-5.3-codex-low-fast           3/3  avg=10s    <- control
    kimi-k2.7-code                   3/3  avg=13s
    gemini-3.6-flash-medium          3/3  avg=6s
    gemini-3.1-pro                   3/3  avg=9s
    gemini-3-flash                   3/3  avg=4s
    gemini-3.5-flash                 3/3  avg=10s
    glm-5.2-high                     3/3  avg=4s
    gpt-5.3-codex-low-fast           3/3  avg=4s     <- control

## Failure classification

Re-ran each failing arm capturing stderr:

    cursor-grok-4.5-medium-fast   succeeded on re-run -> flaky, not dead
    gpt-5.4-nano-medium           ActionRequiredError: AI Model Not Found
                                    Model name is not valid: "gpt-5.4-nano-medium"
    gpt-5.4-nano-low              ActionRequiredError: AI Model Not Found
                                    Model name is not valid: "gpt-5.4-nano-low"
    composer-2.5-fast             RetriableError: WritableIterable is closed
    claude-opus-4-7-medium-fast   NonRetriableError: Provider Error
                                    We're having trouble connecting to the model
                                    provider. This might be temporary - please
                                    try again in a moment.

Four distinct classes:

1. `ActionRequiredError: AI Model Not Found` — the ID is not valid. Fast
   (~3-4s). `--list-models` over-reports: it advertises IDs the API rejects.
   Both `gpt-5.4-nano-*` tiers tested were rejected.
2. `RetriableError: WritableIterable is closed` — client-side stream teardown
   after 3 internal reconnect attempts. Slow (~26-48s).
3. `NonRetriableError: Provider Error` — upstream provider unreachable. Fast
   (~3s). Matches the open incident.
4. `auto` — 0/12 across two sweeps, never succeeded once.

## Ruled out as causes (measured, not assumed)

- Auth — `cursor-agent status` reports logged in throughout.
- DNS / network — `curl https://agentn.global.api5.cursor.sh` returns 200, as
  does `https://api2.cursor.sh`.
- Claude Code Bash-tool sandbox — same failure rate with the sandbox disabled.
- stdin lifecycle — `< /dev/null` and a held-open pipe both fail identically.
- Workspace trust — `-p` mode never raises the trust prompt; adding `--trust`
  changes nothing. (An earlier report that a missing `--trust` caused a hang
  and exit 137 did not reproduce.)
- Chat/session state — `create-chat` returns a clean UUID over one-shot HTTP,
  and `--resume <id>` fails exactly like a fresh dispatch.
- Effort tier as such — Grok 4.5 `low-fast` 3/3 while `medium-fast` 0/3, so the
  failure is not monotonic in reasoning effort. (See sweeps 4 and 5 below: the
  `high-fast` figure quoted here as 3/3 later fell to 7/9 on re-test.)

## Tool-use confirmation

A bare "reply with only: OK" ping proves the transport but NOT tool use.
Confirmed separately with real file-writing briefs, each returning the
handoff contract and each output independently verified:

- `gpt-5.3-codex-low-fast` — roman numeral converter, correct on all
  subtractive forms (IV, IX, XL, XC, CD, CM, MCMXCIV, MMMCMXCIX). 20s.
- `gpt-5.6-luna-high` — bracket matcher, correct including the interleaved
  `([)]` case a naive counter gets wrong. Printed one "Connection lost" and
  recovered on cursor-agent's own internal retry.
- `auto` (before the model cause was identified) — fizzbuzz and a prime
  generator, both correct.

## Sweep 4 — full Grok tier matrix

    gpt-5.3-codex-low-fast           3/3  avg=4s     <- control
    cursor-grok-4.5-low              1/3  avg=41s
    cursor-grok-4.5-medium           0/3  avg=29s
    cursor-grok-4.5-high             0/3  avg=30s
    cursor-grok-4.5-medium-fast      1/3  avg=22s
    cursor-grok-4.6-low              0/3  avg=33s
    cursor-grok-4.6-low-fast         0/3  avg=29s
    cursor-grok-4.6-medium           1/3  avg=45s
    cursor-grok-4.6-high             0/3  avg=45s
    cursor-grok-4.6-xhigh            1/3  avg=36s
    gpt-5.3-codex-low-fast           3/3  avg=6s     <- control

## Sweep 5 — re-verify the two Grok 4.5 outliers at n=6

Sweep 3 had `cursor-grok-4.5-low-fast` 3/3 and `cursor-grok-4.5-high-fast` 3/3,
which is a suspicious result inside a family where 12 of 14 tiers fail. Re-ran
both at 6 trials:

    cursor-grok-4.5-low-fast         6/6  avg=9s    -> 9/9 cumulative, genuinely good
    cursor-grok-4.5-high-fast        4/6  avg=16s   -> 7/9 cumulative, NOT reliable

Conclusion: `cursor-grok-4.5-low-fast` is the only Grok ID that holds up.
`high-fast` was an n=3 artifact.

## Hypotheses raised and killed by the Grok data

- "Failure is per-family" — killed by Grok 4.5: `low-fast` 9/9 while `medium`
  0/3 and `high` 0/3 in the same family.
- "Failure is monotonic in reasoning effort / slow-to-first-token arms hit a
  stream idle timeout" — killed by `gpt-5.4-nano-medium`, which fails in ~4s,
  and by Grok 4.5 where `medium` fails but `high` and `low` differ from it in
  both directions.
- "The `-fast` suffix (priority capacity) is what makes an arm reliable" —
  killed by Grok 4.6: `low-fast` 0/3 and `xhigh-fast` 0/3.

Grok totals across all tiers measured: Grok 4.6 roughly 3/24. Grok 4.5 roughly
11/27, and all of its successes concentrate in `low-fast`.
