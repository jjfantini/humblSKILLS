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

[INGEST 2026-07-23] Authored use-smart-animation skill.
  - Router SKILL.md dispatches by codebase language (html/css/js/ts/react) and by library (orbs/beams/metal/transitions).
  - 12 wiki concepts: motion/principles/{design,performance,accessibility}; lang/{css,html,js,ts,react}/animation; lib/{orbs,beams,metal,transitions}/usage.
  - 6 raw sources seeded from research: thinking-orbs, border-beam, metal-fx, transitions-dev, frontend-design-motion-principles, web-animation-standards.
  - 3 decisions recorded: beams=border-beam, fully standalone, native-first per language.

[LINT 2026-07-23] 12 wiki, 6 raw. Hard: 0, Soft: 7. Regenerated _index.md.
