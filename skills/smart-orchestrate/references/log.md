# Log

Append-only session log. Every session MUST append at least one entry.
Never edit old entries - they are the historical record. Most recent
entries appear at the bottom.

Entry shape:

```
[INGEST|QUERY|LINT <YYYY-MM-DD>] <one-line summary>
  <optional indented detail line(s)>
```

---

[INGEST 2026-08-12] Scaffolded smart-orchestrate via scripts/scaffold.sh.
  - Directory layout created: references/{wiki,raw}/, brain meta files, templates
  - Awaiting first raw material and wiki concepts

[INGEST 2026-08-12] Migrated the flat `orchestrate` skill into smart-orchestrate.
  - Raw: references/raw/orchestrate-SKILL.md (163-line flat SKILL.md, verbatim)
  - New context `orchestrate/` with 7 categories and 9 concepts:
    - roles/{parent-orchestrator, worker-agent}
    - isolation/worktree-first
    - routing/model-selection
    - loop/session-loop
    - contracts/{brief-template, handoff-contract}
    - closeout/commit-and-ship
    - anti-patterns/avoid
  - All 9 concepts cite the raw file in `sources:`
  - SKILL.md rewritten as a router: Brain Protocol, CCCCC table, When/How to Use,
    2 Examples, Troubleshooting, Success Signals
  - Preserved upstream `disable-model-invocation: true`; see decisions.md
  - humblSKILLS extension fields nested under `metadata:` per repo convention

[LINT 2026-08-12] 9 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[QUERY 2026-08-12] Pointed the model tiers at Grok 4.6 (was 4.5).
  - Updated: wiki/orchestrate/roles/parent-orchestrator.md (Cursor parent tier)
  - Updated: wiki/orchestrate/roles/worker-agent.md (cheap worker tier)
  - references/raw/orchestrate-SKILL.md deliberately NOT touched - raw is
    immutable and is the provenance baseline; see decisions.md 2026-08-12
  - metadata.version 1.0.0 -> 1.0.1

[LINT 2026-08-12] 9 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[LINT 2026-08-12] 9 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[QUERY 2026-08-13] Smoke-tested Cursor worker dispatch end-to-end (user request: "does a Cursor sub agent respond and do work").
  - No worktree - trivial single-file test, legitimate skip per isolation/worktree-first.md
  - Dispatched via `cursor-agent -p --force --workspace <scratch-dir> "<brief>"`, brief followed contracts/brief-template.md verbatim
  - Worker (cursor-agent 2026.08.11-e8db854) wrote fizzbuzz.py, ran verify command, returned the handoff contract shape exactly - no free-form prose
  - Independently verified: file correct, no scope creep, nothing committed
  - Confirms the CLI plumbing (`cursor-agent` on PATH at ~/.local/bin) works for this skill's Cursor-worker path

[LINT 2026-08-13] 9 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[LINT 2026-08-13] 9 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[QUERY 2026-08-13] Diagnosed the cursor-agent transport failure and shipped a retry wrapper.
  - Reproduced verbatim: "Connection lost, reconnecting to https://agentn.global.api5.cursor.sh"
    x3 then "RetriableError: WritableIterable is closed", rc=1, ~30s. ~40% success/attempt.
  - Ruled OUT: auth (status logged in), DNS/network (curl to same host = 200),
    harness sandbox (fails equally with sandbox disabled), stdin lifecycle
    (< /dev/null and held-open pipe both fail identically), chat/session state
    (create-chat + --resume <id> fails the same way), workspace trust
    (-p mode never prompts; --trust changes nothing).
  - Matches a Cursor-side known bug; vendor calls it transient client stream teardown.
  - CORRECTED the earlier patterns.md entry from this session, which claimed a working
    backend on the basis of one green dispatch. See patterns.md 2026-08-13.
  - Added scripts/dispatch-cursor-worker.sh: retries ONLY the WritableIterable failure,
    refuses to dispatch into a dirty tree, and aborts if the stream dies after the
    worker wrote files (so a retry cannot double-apply a brief).
  - Verified end to end: wrapper dispatched a primes.py brief, worker returned the
    handoff contract, output independently checked correct.
  - Still open: whether the same failure rate occurs outside the Claude Code Bash tool
    (needs a run from a real terminal to confirm it is purely vendor-side).

[QUERY 2026-08-13] Corrected the diagnosis: `--model auto` was the cause, not the transport.
  - Failure rate degraded to 0/6 on the default model; status.cursor.com showed an
    active incident (2026-08-13 14:46 UTC, model-provider degradation).
  - Head-to-head measurement: auto 0/9, composer-2.5 1/4, cursor-grok-4.6-low-fast 0/1,
    gpt-5.3-codex-low-fast 6/6, claude-sonnet-5-thinking-high 3/3.
  - `auto` routes to any provider including one mid-incident, and that routing failure
    surfaces as a client-side "WritableIterable is closed" stream teardown rather than
    an honest upstream error - which is why it read as a transport bug.
  - scripts/dispatch-cursor-worker.sh now defaults to gpt-5.3-codex-low-fast
    (override with $CURSOR_WORKER_MODEL or arg 3) and documents the measured table.
    The retry loop stays as secondary insurance.
  - Verified: roman-numeral brief returned the handoff contract in 20s, first attempt,
    output independently checked correct on all subtractive-form edge cases.
  - patterns.md: prior entry marked SUPERSEDED rather than deleted, so the dead-end
    hypotheses (sandbox, stdin, trust, chat-session) are not re-litigated later.

[LINT 2026-08-13] 9 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[QUERY 2026-08-13] Swept 13 Cursor model arms x 3 runs to pick safe worker backends.
  - Control (gpt-5.3-codex-low-fast) run at both ends: 3/3 then 3/3 - no incident drift.
  - GOOD 3/3: gpt-5.6-luna-high 5s, gpt-5.6-sol-xhigh-fast 6s, gpt-5.5-high-fast 6s,
    gpt-5.6-sol-high-fast 7s, gpt-5.2 7s, gpt-5.3-codex-low-fast 7s,
    claude-opus-5-thinking-high 9s, kimi-k3-high 18s.
  - BAD: all three cursor-grok-4.6 tiers (2/9 combined), composer-2.5 1/4, auto 0/12.
  - Confirmed gpt-5.6-luna-high on a REAL brief (bracket matcher), not just a ping -
    output independently verified correct including the interleaved `([)]` case.
  - scripts/dispatch-cursor-worker.sh header now carries the full measured table.

[LINT 2026-08-13] 9 wiki, 1 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[LINT 2026-08-13] 10 wiki, 2 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[LINT 2026-08-13] 10 wiki, 2 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[INGEST 2026-08-13] Added routing/cursor-cli-models concept + its raw measurement log.
  - New wiki concept: references/wiki/orchestrate/routing/cursor-cli-models.md
    (30 verified-good IDs, 19 verified-bad, 4 failure classes with retry policy)
  - New raw: references/raw/cursor-cli-model-sweep-2026-08-13.md (5 sweeps, verbatim)
  - Cross-linked from routing/model-selection.md and SKILL.md How to Use.
    Deliberately did NOT put model IDs into model-selection.md's tier table -
    that concept is generic on purpose and would rot; see decisions.md.
  - Grok answer for the record: Grok 4.6 unusable (~3/24, all 8 tiers).
    Grok 4.5 unusable EXCEPT cursor-grok-4.5-low-fast (9/9).
  - decisions.md: recorded authoring a raw file as measurement provenance, and
    the asymmetry vs the 2026-08-12 "raw is immutable" entry.
  - scripts/dispatch-cursor-worker.sh header synced to final numbers.

[LINT 2026-08-13] 10 wiki, 2 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[LINT 2026-08-13] 10 wiki, 2 raw. Hard: 0, Soft: 0. Regenerated _index.md.
