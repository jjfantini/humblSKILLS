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

### 2026-08-12 | Keep `disable-model-invocation: true` on a registry-distributed skill
- Context: the upstream flat `orchestrate` skill set `disable-model-invocation: true`, so the model can never auto-trigger it. No other skill in the humblSKILLS registry uses that field, and the smart-skill validation checklist emphasises trigger phrases in `description:`.
- Options: (A) drop the field so the skill auto-triggers like every other registry skill, (B) keep it and rely on explicit `/smart-orchestrate` invocation, (C) keep it but omit trigger phrases from the description since they can never fire automatically.
- Chose: B - keep the field, and still write a full trigger-phrase description.
- Why: a full orchestration run spends frontier tokens, creates git worktrees, and dispatches subagents across CLIs. Auto-triggering that on a vague "orchestrate this" is a real cost and blast-radius event the user did not opt into, which is exactly why the upstream author set the flag. Trigger phrases stay in the description because they are what the skill picker and `humblskills search` match on - they are not wasted just because auto-invocation is off. The CLI's frontmatter parser accepts unknown top-level keys, so the field passes `build-registry` untouched.
- Result: TBD - watch whether registry users report "installed it and nothing happened". If so, the fix is a README note, not dropping the flag.

### 2026-08-12 | Rename the raw source to `orchestrate-SKILL.md`
- Context: the brain spec says the LLM never renames files in `references/raw/`. The migration source was literally named `SKILL.md`, which would land at `references/raw/SKILL.md`.
- Options: (A) keep the name verbatim per the raw-territory rule, (B) rename to `orchestrate-SKILL.md`.
- Chose: B.
- Why: a second file named `SKILL.md` inside a skill directory is a genuine hazard for humans grepping the tree, and would be a hazard for tooling too if `build-registry` ever walked recursively (today it only joins `skills/<name>/SKILL.md`, so it is safe by one implementation detail). The no-rename rule exists to protect human-dropped provenance; here the drop was performed by the migration itself, and the prefix adds provenance rather than erasing it.
- Result: raw file is `references/raw/orchestrate-SKILL.md`, cited by all 9 concepts, byte-identical to the upstream flat skill.

### 2026-08-12 | Refresh model names in the wiki, leave `references/raw/` at 4.5
- Context: Grok 4.5 was superseded by Grok 4.6. Two wiki concepts named 4.5 as a model tier (`roles/parent-orchestrator` for the Cursor parent slot, `roles/worker-agent` for the cheap worker slot), and so does the preserved upstream skill at `references/raw/orchestrate-SKILL.md`.
- Options: (A) update the wiki only, (B) update the wiki and the raw file so nothing in the skill reads "4.5", (C) update the wiki and annotate the raw file's line as superseded.
- Chose: A - wiki only. The raw file keeps saying Grok 4.5.
- Why: raw is human territory and immutable by the brain spec ("Never edit raw to satisfy a wiki claim"). It is a point-in-time snapshot of what was handed over, and it doubles as the provenance baseline every concept's `sources:` array points at - editing it to match a later fact destroys what makes the citation meaningful, and silently rewrites history for anyone diffing the distillation against its origin. A wiki that has moved ahead of its raw source is the normal state of a living skill, not drift to be reconciled.
- Result: `roles/parent-orchestrator.md` and `roles/worker-agent.md` say Grok 4.6; the raw file still says 4.5. Read the difference as "wiki is current, snapshot is historical" - that is exactly what `sources:` is for. Skill version 1.0.0 -> 1.0.1.
