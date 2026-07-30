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

### 2026-06-12 | Keep one question before frontend design
- Context: The skill needs enough design intent to avoid generic output without slowing every frontend task with a long intake form.
- Options: (A) ask a full brand questionnaire, (B) ask exactly one essence/style question, (C) skip questions and infer everything from the repo.
- Chose: B - ask exactly one essence/style question unless the user already supplied style, audience, brand, screenshot, or emotional direction.
- Why: One question creates a strong design constraint while preserving momentum. Existing codebase discovery then supplies implementation constraints.
- Result: `design/intake/one-question.md` defines the contract and `SKILL.md` routes every task through it.

### 2026-06-12 | Store original flat brief as raw source
- Context: The user required all supplied frontend-design text to be included, but also asked for a smart skill rather than one flat file.
- Options: (A) paste the full brief into `SKILL.md`, (B) split everything into wiki concepts and discard the original, (C) store the full brief in `references/raw/` and cite it from distilled concepts.
- Chose: C - immutable raw source plus cited wiki concepts.
- Why: This preserves every word while keeping progressive disclosure. Agents can read focused concepts at runtime and still audit the original brief.
- Result: `references/raw/user-frontend-design-brief.md` contains the full supplied text and all seven wiki concepts cite it.

### 2026-07-30 | Re-synced against current upstream; the mirror was a generation stale
- Context: this skill mirrors the Claude `frontend-design` plugin skill. The
  preserved copy under `references/raw/` (misnamed `user-frontend-design-brief.md`,
  45 lines) shared *zero* section headings with the current upstream (55 lines) —
  upstream had been rewritten, not amended, and nothing detected it.
- Impact found: `anti-patterns/generic-ai-slop` was teaching the previous
  generation's tells (Inter, purple gradients, Space Grotesk) while current
  upstream names three different clusters (cream+serif+terracotta;
  near-black+acid accent; broadsheet). The skill was steering toward what
  upstream now flags as default.
- Options: (A) leave as-is, (B) re-sync content only, (C) re-sync and add a
  provenance mechanism so the next drift is detectable.
- Chose: (C).
- Why: (A) ships knowingly stale calibration. (B) fixes today and guarantees a
  repeat, because the failure was never content — it was the absence of an
  upstream pointer, a sync date, and a declared delta list.
- Result: preserved copy replaced and renamed to `frontend-design-SKILL.md`;
  `generic-ai-slop` rewritten to the three clusters; 3 concepts added
  (`aesthetics/hero-thesis`, `aesthetics/structural-honesty`,
  `process/two-pass-plan`) for upstream material with no prior home; 4 concepts
  updated. 7 wiki concepts -> 10. Version 1.0.3 -> 1.1.0.

### 2026-07-30 | Provenance lives in an `upstream:` block; ATTRIBUTION.md is derived
- Context: two inheritance mechanisms existed in the repo — `ATTRIBUTION.md`
  prose in the 7 `better-*` skills, and nothing at all here. Neither is
  machine-readable, so no tool can answer "is this mirror current?".
- Options: (A) write an ATTRIBUTION.md here too, (B) frontmatter `upstream:`
  block as the single source of truth with ATTRIBUTION.md derived from it,
  (C) drop ATTRIBUTION.md for the block alone.
- Chose: (B).
- Why: (A) leaves the data unparseable and duplicated across 8 skills. (C) loses
  the human-readable license notice that MIT attribution conventions expect and
  that a browsing user looks for. (B) gives one editable source, a rendered file
  for humans and licence compliance, and a structure the CLI can display.
- Why-caveats: the block carries `preserved:`, `synced:`, and `deltas:` — the
  three fields whose absence caused the staleness above. `deltas:` is what stops
  a future sync from silently re-litigating intentional differences.
- Result: `upstream:` block added here and to all 7 `better-*` skills; all 8
  ATTRIBUTION.md files regenerated to one shape with a mirror-status table and a
  declared-deltas section, each marked as derived.
- Deferred: teaching `scripts/lint.sh` to generate ATTRIBUTION.md and warn when
  `preserved:` differs from the installed upstream. Out of scope here because
  `lint.sh` exists as 20 per-skill copies, 2 of which have already diverged
  substantially — that is a repo-wide refactor. Tracked with the "update
  mirrors" command idea (auto-resync every mirrored skill from its upstream).
