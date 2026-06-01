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

### 2026-06-01 | Wiki taxonomy + text-only raw / assumed-flow handling
- Context: migrating the flat refs (delivery_framework, core_concepts, examples)
  into the CCCCC wiki, and the raw Hello Interview pages are text-only scrapes
  whose original HLD/architecture diagrams did not survive the markdown export.
- Options: (A) one flat context (e.g. `ml`) with many categories; (B) mirror the
  source site sections; (C) four intent-based contexts -
  `delivery` (the 6-step framework), `concepts` (core knowledge),
  `intro` (interview meta), `breakdown` (worked problems). For diagrams: (A)
  skip diagram content silently, (B) fabricate a precise diagram, (C) add an
  explicit `> Assumed flow (...)` note inferred from surrounding prose.
- Chose: (C) four intent-based contexts; (C) explicit assumed-flow notes.
- Why: contexts map to how a candidate actually reaches for material mid-
  interview (framework step vs core concept vs worked example), keeping concepts
  atomic and routable from SKILL.md. Assumed-flow notes keep distilled claims
  honest and auditable against text-only raw without inventing source content.
- Result: 14 concepts, all citing raw pages; the 3 breakdowns + delivery
  overview carry assumed-flow notes for the omitted diagrams. Also removed the
  extraneous `raw/learn/system-design/` tree the staging cp pulled in (sibling
  skill territory; expected result was ml-system-design only). Lint clean.
