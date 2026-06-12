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
