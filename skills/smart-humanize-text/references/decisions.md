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

### 2026-04-17 | Atomic per-tell concepts grouped by mechanism (humanize migration)
- Context: migrating the flat humanize-text SKILL.md (232 lines, 24 numbered tells + vocabulary + voice + process sections) into the Smart Skill structure. Needed to decide concept granularity.
- Options: (A) minimal ~5 concepts (one per top-level section), (B) grouped ~11 concepts (tells grouped by shared mechanism), (C) atomic - one concept per numbered tell plus one per vocabulary list, voice subsection, and process step.
- Chose: C - atomic concepts organized into 11 categories by shared mechanism (analysis-reflexes, list-patterns, word-variation, punctuation, attribution, wrap-ups, padding, formatting, voice, vocabulary, process).
- Why: each numbered tell in the source has its own incorrect/correct example pair - that IS the atomic unit. Runtime lookups ("am I hitting the rule-of-three tell?") can load a single small file instead of pulling in a bigger bundle. Categories group files by the mechanism they share (e.g. "wrap-ups" covers section-summaries + generic-positive-conclusions + sycophantic-tone, all "how AI ends things"). User explicitly requested atomic granularity with context/category/concept split.
- Result: 27 atomic wiki concepts, 11 categories, 1 context (`humanize`). Thin SKILL.md routes to specific concept files. Every concept cites the single raw PDF.

