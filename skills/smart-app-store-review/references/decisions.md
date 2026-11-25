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

### 2026-11-25 | Adopt upstream rules verbatim in raw/, route via one section map
- Context: safaiyeh/app-store-review-skill v1.5.0 (MIT) already holds about 150 KB of per-guideline rules with Swift and RN/Expo examples, kept current with Apple's updates.
- Options: (A) rewrite the rules into wiki concepts, (B) copy upstream into raw/upstream/ unedited and add a thin wiki layer (workflow, evidence gate, cleanup, report, section map), (C) link to upstream at runtime.
- Chose: B - raw/upstream/ plus the section map.
- Why: rewriting would fork the content and go stale fast. Runtime links need network access and may change under us. Raw files stay diffable against upstream, so re-ingesting a new version is a file replace plus a section-map date bump.
- Result: TBD - re-ingest when upstream ships a guidelines update.

### 2026-11-25 | scan.sh lists candidates, never verdicts
- Context: upstream insists that ATT, account deletion, logging and payment issues need config and flow evidence, not an import or URL.
- Options: (A) scanner exits non-zero on CRITICAL hits like a CI gate, (B) scanner lists candidates and the agent applies the evidence gate.
- Chose: B.
- Why: storefront exceptions (3.1.1(a), 3.1.3), first-party analytics and web deletion links make grep verdicts wrong often enough to train users to ignore them.
- Result: TBD.
