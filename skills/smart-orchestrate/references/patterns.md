# Patterns

Performance memory. Each entry records a concrete attempt, its numeric
outcome, and the lesson. Read before every session; append after every
session where quantified results appear.

Entry shape (see `wiki/brain/patterns/how-to-log-results.md` for the full
worked example):

```
### <YYYY-MM-DD> | <short title>
- Context: <what was attempted, in one line>
- Approach: <the method used>
- Result: <metrics, numbers, outcomes>
- Worked: <what helped>
- Didn't: <what hurt>
- Lesson: <the rule to apply next time>
```

---

### 2026-08-13 | Cursor worker dispatch: the fix is an explicit `--model`, NOT a retry loop
- Context: same investigation as the superseded entry below, carried further. That entry blamed the transport and concluded the model was irrelevant; it was WRONG, confounded by `--model auto`. This is the entry to act on.
- Approach: after the failure rate degraded to 0/6 on the default model, checked status.cursor.com - active incident 2026-08-13 14:46 UTC, model-provider degradation - then measured explicit models head to head.
- Result: `auto` 0/9. `composer-2.5` 1/4. `cursor-grok-4.6-low-fast` 0/1. `gpt-5.3-codex-low-fast` 6/6 at 5-10s each. `claude-sonnet-5-thinking-high` 3/3. With an explicit model the entire failure class disappears. End-to-end proof: a roman-numeral brief returned the handoff contract in 20s on the first attempt, output independently verified correct on every subtractive-form edge case.
- Worked: passing an explicit `--model`. `gpt-5.3-codex-low-fast` is the cheapest reliable worker slot measured; `claude-sonnet-5-thinking-high` for harder briefs.
- Didn't: every hypothesis in the entry below - transport, sandbox, stdin, trust, chat-session - was a dead end. The "failure precedes the model" reasoning was the specific trap: the reconnect line fires at attempt 1 regardless, so the timing looks transport-shaped even when the cause is routing.
- Lesson: NEVER dispatch a worker on `auto`. It lets Cursor route to any provider including one mid-incident, and that routing failure surfaces as a client-side stream teardown instead of an honest upstream error. Generalises past Cursor: a vendor "auto" model selector is an unpinned dependency, and a stream-teardown message is not evidence the stream is at fault. Check the vendor status page before building a workaround.

### 2026-08-13 | SUPERSEDED (see entry above) - Cursor worker misread as a transport failure
- Context: verify a Cursor sub-agent responds and does real work as a worker for this skill. An earlier n=1 success in this same session was written up here as "0 retries, working backend" - that entry was WRONG and is replaced by this one. Do not trust a single green dispatch as evidence of a working backend.
- Approach: `cursor-agent -p --force --trust --workspace <dir> "<brief>"` with the brief built verbatim from contracts/brief-template.md. Then repeated bare-ping dispatches to measure the per-attempt success rate.
- Result: first dispatch wrote the file, ran its own verify, and returned the handoff-contract shape exactly - the contract path works. But across ~8 subsequent dispatches only ~40% succeeded. Failures are bimodal and easy to fingerprint: `Connection lost, reconnecting to https://agentn.global.api5.cursor.sh (attempt 1..3)` then `RetriableError: WritableIterable is closed`, rc=1, ~29-30s. Successes take ~11s and print no reconnect line at all.
- Worked: `-p --force --trust` is the right non-interactive flag set. Plain `curl` to the same host returns 200, so it is not DNS, network, or auth - only the long-lived duplex stream dies. Disabling the harness sandbox changed nothing (1/3 success unsandboxed), so the sandbox is not the cause.
- Didn't: retrying blindly is only safe because every observed failure happened BEFORE the worker wrote anything. Cursor's own forum reports a separate mid-turn variant (`http/2 stream closed CANCEL` after 15-20 min) which would leave a brief partially applied - a blind rerun there double-applies the work.
- Lesson: treat `cursor-agent` as a flaky-transport backend, not a reliable one. Wrap dispatch in a retry loop keyed to a chat id (`cursor-agent create-chat` then `--resume <id>` on rc=1, which is the vendor's stated workaround) rather than re-dispatching from scratch, and keep briefs short so pre-work failure stays the only realistic mode. Codex CLI (`codex:codex-rescue`) is an already-installed non-Claude worker backend with no such transport problem - prefer it when the task does not specifically require Cursor.

### 2026-08-13 | Cursor model sweep: GPT-5.x and Opus 5 good, Grok 4.6 and `auto` bad
- Context: follow-up to the entry above - which Cursor models are actually safe worker backends, including the newer GPT-5.6 and Grok 4.6 tiers.
- Approach: 13 arms x 3 dispatches, `gpt-5.3-codex-low-fast` as a control at BOTH ends of the sweep so a worsening vendor incident could not be mistaken for a bad model. Control was 3/3 at the start and 3/3 at the end, so the numbers are comparable across the run.
- Result: GOOD, all 3/3 - gpt-5.6-luna-high (5s), gpt-5.6-sol-xhigh-fast (6s), gpt-5.5-high-fast (6s), gpt-5.6-sol-high-fast (7s), gpt-5.2 (7s), gpt-5.3-codex-low-fast (7s), claude-opus-5-thinking-high (9s), kimi-k3-high (18s). BAD - cursor-grok-4.6-medium-fast 1/3, cursor-grok-4.6-high-fast 1/3, cursor-grok-4.6-xhigh-fast 0/3 (2/9 across all Grok tiers), composer-2.5 1/4, `auto` 0/12.
- Worked: the two-control sandwich. Without it, `auto` 0/3 landing mid-sweep is indistinguishable from "the incident got worse", and the whole sweep would be uninterpretable. Also: a real file-writing brief on gpt-5.6-luna-high produced correct code (bracket matcher, right on the interleaved `([)]` case) - a bare ping proves the transport but NOT tool use, so always confirm a candidate with a real brief.
- Didn't: nothing new failed. Note gpt-5.6-luna-high printed one "Connection lost" and recovered via cursor-agent's own internal retry, so a good model is not perfectly immune - keep the wrapper's retry loop.
- Lesson: failure is per-model-family, not global. GPT-5.x is uniformly reliable here and Grok 4.6 is uniformly not, across every effort tier - so tier is irrelevant and family is what to pick on. Default to `gpt-5.3-codex-low-fast` for cheap worker slots, `gpt-5.6-luna-high` when speed matters, `claude-opus-5-thinking-high` for hard briefs. Re-run the sweep rather than trusting this table after any Cursor status-page incident.

### 2026-08-13 | Full Cursor CLI sweep: reliability is per-model-ID, nothing generalises
- Context: user asked about Grok 4.5 specifically and for the routing wiki to carry up-to-date Cursor CLI model guidance. Supersedes the "GPT good / Grok bad by family" framing in the entry above, which was too coarse.
- Approach: `--list-models` reports 201 IDs; testing all at n=3 is ~8h wall clock, so one representative per family (~26 arms) plus EVERY Grok tier (14 arms), plus n=6 re-tests of suspicious outliers. `gpt-5.3-codex-low-fast` interspersed as a control 6 times - all 3/3, so no incident drift.
- Result: 30 IDs verified good, 19 verified bad. Grok 4.6 ~3/24 across all 8 tiers. Grok 4.5 ~11/27 with successes concentrated in `low-fast` (9/9). `auto` 0/12. `composer-2.5` and `-fast` both ~1/4. `gpt-5.4-nano-*` invalid IDs. Everything GPT-5.x, Gemini, GLM, Kimi and most Claude tiers 3/3 at 4-18s.
- Worked: interspersed controls (6 of them, not just endpoints) - they are what make a 30-arm sweep interpretable when a vendor incident is live. Re-testing outliers at n=6 - `cursor-grok-4.5-high-fast` looked 3/3, fell to 4/6, final 7/9 and got demoted. Classifying failures by stderr rather than by exit code, which is what revealed there are FOUR failure classes, not one.
- Didn't: three successive generalisations all died. "Per-family" died to Grok 4.5 low-fast 9/9 vs medium 0/3. "Slow-to-first-token hits an idle timeout" died to gpt-5.4-nano-medium failing in 4s. "The -fast suffix is what works" died to Grok 4.6 low-fast 0/3. Each felt well-supported when proposed.
- Lesson: reliability is per-model-ID and does not follow family, effort tier, or the `-fast` suffix - measure the exact ID, at n>=6 before adopting it, and never trust a 3/3 sitting inside a mostly-broken family. `--list-models` over-reports, so listing is not evidence an ID resolves. Four failure classes with different retry policies: invalid ID and `NonRetriableError: Provider Error` must fail loudly; only `WritableIterable is closed` is retryable. Full table: wiki/orchestrate/routing/cursor-cli-models.md.
