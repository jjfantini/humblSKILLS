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
