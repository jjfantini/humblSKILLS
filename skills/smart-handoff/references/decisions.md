# Decisions

Reasoning memory. Each entry records a non-obvious choice: the context,
the options considered, what was chosen, why, and the observed result.
Never delete entries - if a decision is reversed, add a new entry that
references the old one.

Entry shape:

```
### <YYYY-MM-DD> | <short title>
- Context: <the situation that required a choice>
- Options: (A) <opt>, (B) <opt>, (C) <opt>
- Chose: <letter and name>
- Why: <the rationale, ideally citing evidence>
- Result: <what happened after, or "TBD">
```

---

### 2026-04-17 | Add `compatibility` frontmatter, skip `allowed-tools`
- Context: agentskills.io spec defines 4 optional SKILL.md frontmatter fields (`license`, `compatibility`, `metadata`, `allowed-tools`). Needed to decide which to propagate through this skill and its scaffold.
- Options: (A) add both `compatibility` and `allowed-tools`, (B) add `compatibility` only, (C) add neither and only document in a wiki concept.
- Chose: B - add `compatibility` to this skill's SKILL.md, emit a loud TODO placeholder in scaffold.sh, and propagate through workflow/validation docs. Skip `allowed-tools` entirely.
- Why: `compatibility` carries real signal (this skill has bash scripts, consumers need to know). Max 500 chars, cheap, per-spec appropriate. `allowed-tools` is flagged experimental by the spec itself, agent support varies, and this is a meta-skill with open-ended tool use - pre-approving a list would either be decorative (too permissive) or restrictive (breaks the skill). Scaffold emits a TODO-or-delete line (not silent omit) so new-skill authors consciously decide instead of forgetting the field exists.
- Result: SKILL.md now declares bash + POSIX requirements. Scaffold forces explicit decision. Future skill authors see the field in the new `wiki/smart/spec/skill-frontmatter.md` concept.

### 2026-04-19 | New `anthropic/` wiki context (alongside existing `smart/`)
- Context: Anthropic published "The Complete Guide to Building Skills for Claude" PDF. Needed to ingest best practices so every scaffolded humblSKILL inherits them. Existing `smart/spec/skill-frontmatter.md` already distills the agentskills.io spec.
- Options: (A) merge new material into `smart/` context as new categories, (B) create a dedicated `anthropic/` context with categories `frontmatter`, `description`, `structure`, `testing`, `patterns`, `troubleshooting`, (C) overwrite existing `smart/spec/skill-frontmatter.md` with distilled PDF content.
- Chose: B - dedicated `anthropic/` context.
- Why: (1) attribution is first-class - the `context:` folder name carries provenance back to Anthropic's guide vs agentskills.io. (2) When Anthropic ships a new version of the PDF we update every concept under `anthropic/*` and know exactly which pages to re-read. (3) `smart/spec/skill-frontmatter.md` (agentskills source) and `anthropic/frontmatter/requirements.md` (Anthropic source) can legitimately coexist as complementary views on the same underlying schema. Lint's contradiction heuristic will flag duplicate `concept:` values across contexts for human audit - acceptable trade-off.
- Result: 8 new concepts under `anthropic/` (frontmatter/requirements, frontmatter/security, description/trigger-design, structure/progressive-disclosure, structure/file-layout, testing/three-layer-approach, patterns/five-patterns, troubleshooting/common-failures). All cite `references/raw/anthropic-skill-building-guide.pdf`. SKILL.md `How to Use` now routes Anthropic-sourced questions into `anthropic/<category>/` concepts.

### 2026-04-19 | Reverse 2026-04-17 decision: include `allowed-tools` after all
- Context: Anthropic's official guide (Reference B) and Chapter 5 "Instructions not followed" both surface `allowed-tools` as a real tool that improves security posture and reduces ambiguity. The 2026-04-17 rationale was "experimental, agent support varies". The PDF (Oct 2025 vintage) lists it as a first-class optional field with a clear syntax. Reality moved faster than the 2026-04-17 decision.
- Options: (A) keep skipping `allowed-tools` to honor prior decision, (B) include `allowed-tools` with an explicit value in this skill + TODO-or-delete in scaffold template, (C) include but leave as comment-only placeholder.
- Chose: B - declare `allowed-tools: "Bash(bash:*) Bash(sh:*) Read Write Edit Glob Grep"` in this skill's SKILL.md, emit TODO-or-delete in scaffold template so new-skill authors decide explicitly.
- Why: field is now canonical in Anthropic's guide, not experimental. Meta-skill can declare a tight set because its own tool usage is predictable (bash scripts + file ops). Scaffold template stays permissive (TODO-or-delete) because per-skill tool needs vary.
- Result: supersedes the "skip `allowed-tools`" part of the 2026-04-17 decision. `compatibility` decision from 2026-04-17 remains in force.

### 2026-04-19 | Move humblSKILLS extension fields from top-level to `metadata:`
- Context: Anthropic's spec (Reference B) treats top-level frontmatter as `name`/`description`/`license`/`compatibility`/`allowed-tools` + a free-form `metadata:` map for everything else. humblSKILLS had been putting `version`, `tags`, `platforms`, `requires`, `preserve` at the top level, which would be rejected as non-standard by strict validators.
- Options: (A) keep top-level (remain non-compliant but historical), (B) move all humblSKILLS fields under `metadata:` with a hard break, (C) move to `metadata:` with a soft-transition fallback (parser reads metadata first, falls back to top-level, warns).
- Chose: C - soft transition.
- Why: keeps the in-repo migration risk-free (existing skills keep loading during the cutover), surfaces deprecation warnings on every registry build so external consumers learn the new shape, and preserves a rollback path if a downstream consumer breaks. Remove the fallback in a later release once no warnings appear.
- Result: `cli/internal/frontmatter/Frontmatter` now exposes `Version()`, `Requires()`, `Platforms()`, `Tags()`, `Preserve()` accessor methods that fall back to legacy top-level fields. `DeprecationWarnings()` surfaces remaining top-level usage. All 5 in-repo skills migrated to the new shape; `build-registry` output is warning-free.

### 2026-08-03 | Live-context capture as default; transcript parsing opt-in
- Context: a handoff doc needs both the *why* (only in the conversation) and the *where* (only in git). Claude Code stores transcripts as JSONL under `~/.claude/projects/<slug>/`, which is tempting to parse for an exhaustive capture.
- Options: (A) always parse the session transcript, (B) always write from live context + git, (C) live context + git by default, transcript only on explicit user request or after a compaction.
- Chose: C.
- Why: transcript paths are Claude-Code-only, so (A) makes the skill's core path unusable in Codex and Cursor — the exact harnesses it exists to hand off to. Parsing also costs a large read for information the agent already holds. (B) alone loses nothing in a normal session but degrades badly after a context compaction, which is precisely when a handoff gets written. C keeps the cross-harness path deterministic and names the escape hatch.
- Result: `scripts/preflight.sh` covers the git half with no harness assumptions; `wiki/handoff/capture/session-context.md` documents the opt-in transcript lookup as Claude-Code-only.

### 2026-08-03 | Mask matches with bash substitution, not sed
- Context: `scan-secrets.sh` prints a finding with the match replaced by `<REDACTED:name>`. First implementation used `sed -E "s|$re|...|g"`.
- Options: (A) `sed` with a `|` delimiter, (B) `sed` with an exotic delimiter (`\001`), (C) extract matches with `grep -oE` and do literal bash `${var//match/repl}` substitution.
- Chose: C.
- Why: the detection patterns contain `|` (alternations), `/` (URLs, bearer charset) and `#`, so every printable sed delimiter collides with at least one pattern — the `|` version failed with "RE error: parentheses not balanced" on the `inline-credential` pattern. Literal substitution has no delimiter at all. Added a re-scan of the masked string that suppresses the line outright if masking did not take, so a masking bug degrades to silence rather than to a printed credential.
- Result: 22/22 tests pass under `/bin/bash` 3.2, including an explicit assertion that the raw fixture secret never appears in the report.

### 2026-08-03 | `EXAMPLE` stays on the safe-list even though it hid a test fixture
- Context: the safe-list suppresses lines containing `EXAMPLE`, `op://`, `REDACTED`, `example.com` etc. The AWS test fixture used AWS's own documented `AKIAIOSFODNN7EXAMPLE`, which the safe-list correctly swallowed — so the AWS assertion failed.
- Options: (A) drop `EXAMPLE` from the safe-list so the fixture is detected, (B) keep the safe-list and change the fixture.
- Chose: B.
- Why: `EXAMPLE` is the single most common placeholder marker in real docs and no leaked credential contains it. Removing it would trade a lower false-negative rate on a value that cannot be live for a higher false-positive rate on every doc that says `AKIA…EXAMPLE`. False positives are what make a scanner get disabled.
- Result: fixture became `AKIA` + `ZZTESTKEY000000Q`, still assembled at runtime so no literal credential shape is committed.
