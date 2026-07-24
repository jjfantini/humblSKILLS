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

[INGEST 2026-07-24] Added Canvas UI (canvasui.dev) as a fifth library.
  - New raw: references/raw/canvas-ui.md; new wiki: lib/canvas-ui/usage.
  - SKILL.md router gained a Canvas UI routing line; description/tags updated (added webgl); version 0.1.0 -> 0.2.0.
  - Decision recorded: lib concept only (no Vue/Svelte language routing); flagged experimental html-in-canvas API + MIT+Commons-Clause license.

[LINT 2026-07-23] 13 wiki, 7 raw. Hard: 0, Soft: 8. Regenerated _index.md.

[INGEST 2026-07-24] Added Safari/SVG-workaround note to Canvas UI docs.
  - Documented that html-in-canvas is Chromium-only (Safari/WebKit + Firefox unsupported), matching observed reduced clarity in Safari.
  - Noted the SVG <foreignObject> -> drawImage cross-browser workaround and its Safari caveats; steer to WebGL components for Safari/production.
  - Updated raw/canvas-ui.md + lib/canvas-ui/usage.md (no version change; 0.2.0 still in PR #195).

[LINT 2026-07-23] 13 wiki, 7 raw. Hard: 0, Soft: 8. Regenerated _index.md.
