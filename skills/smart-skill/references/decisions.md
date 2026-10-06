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

### 2026-08-20 | Validate decisions.md/patterns.md entry schema; gate on structure, not expressiveness
- Context: `lint.sh` only ever counted `### ` headings for these two files (via `countHeadings` in the eval harness) - an entry missing `Why`/`Result`/etc, or a malformed heading, passed silently. The data model (required field labels per entry) existed only as prose instruction, never enforced.
- Options: (A) leave it prose-only, trust the LLM to follow the template, (B) hard-fail on any missing required field label, (C) hard-fail on missing/malformed structure only, soft-warn on empty fields or a non-numeric pattern `Result`, never gate on wording or length.
- Chose: C - structural hard-fail, quality soft-warn, no expressiveness gate.
- Why: the point of validation is to keep the schema (the labels) intact, not to police how much an agent writes under each label - a terse but complete entry is valid, a well-written entry missing a label is not. (B) would punish brevity; (A) already proved insufficient (this exact gap). Distinguishing missing-field (hard) from empty-field (soft) also means a field the agent intentionally left blank fails loud enough to notice but doesn't block a session that has real content elsewhere.
- Result: added to `scripts/lint.sh` (canonical here, propagated to smart-orchestrate's copy), `_brain.md`, `wiki/brain/lint/checks.md`, `SKILL.md`'s Brain Operations table, and `docs/smart_skills.md`. Verified against this skill's own 4 decisions (clean) and against a synthetic broken entry (5 hard fails: 1 malformed heading + 4 missing fields). Version 1.1.3 -> 1.2.0.

### 2026-10-02 | Promotion ships inside smart-skill as promote.sh, not as a CLI command
- Context: FDEs create smart skills locally and need a reviewed path into humblSKILLS or happySKILLS (`raw/user-request-promote.md`). The maintainer asked for a `promote.sh` in this skill, a target question after create, and enforced category/role/tags.
- Options: (A) a `humblskills promote` Go subcommand, (B) `scripts/promote.sh` plus a stdlib-only Python engine (`scripts/lib/promote.py`) shipped in this skill, (C) a prose workflow the agent runs by hand with git and gh.
- Chose: B - promote.sh + engine, with the CLI's validator kept authoritative.
- Why: the agent that just created the skill already has this skill loaded, so the question and the tool sit where the work happens, and they update with the skill instead of waiting for a CLI release. (C) leaves the quality bar to memory. To keep (B) honest, promote.sh runs the real `build-registry` when Go 1.23+ is present, and a Go test pins the closed sets the engine duplicates (categories, roles, platforms).
- Result: 135-case hermetic suite passes under bash 5.2 and bash 3.2.57 (macOS default); a real `--dry-run` against jjfantini/humblSKILLS develop validated a new skill with build-registry in 9s.

### 2026-10-02 | Promotion PRs never carry registry.json
- Context: registry.json changes with every skill edit and every Registry-workflow run, and a conflicted PR gets no CI at all.
- Options: (A) regenerate and commit registry.json in the promotion PR, as hand-made skill PRs do, (B) leave it out, validate locally by building into a temp file, and let the Registry workflow regenerate it after merge.
- Chose: B.
- Why: a PR that only touches `skills/<name>/` cannot conflict on the repo's hottest file. The Registry workflow already repins `source.sha` on develop after the merge, and fork PRs could not run it anyway (it needs the release environment).
- Result: every promotion diff is exactly `skills/<name>/`; the test suite asserts registry.json is untouched on the branch.

### 2026-10-02 | Public promotions fail on HappyRobot mentions
- Context: a branch pushed to a public repo is public before anyone reviews it, so review cannot catch an internal leak in time.
- Options: (A) warn and rely on review, (B) fail, with `--allow-internal` for deliberate public-safe mentions, (C) fail with no override.
- Chose: B.
- Why: (A) warns after the damage is possible; (C) blocks legitimate mentions such as this skill naming happySKILLS as a target. An explicit flag records intent in the run without a code change.
- Result: calibration flagged smart-orchestrate (a raw sweep cites an @happyrobot.ai account) and this skill (by design); no other published skill trips it.
