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

### 2026-07-23 | "beams" library resolved to border-beam
- Context: the requested URL `beams.jakubantalik.com` does not resolve; the only
  matching Antalik library is `border-beam` (site `beam.jakubantalik.com`), a
  traveling/breathing border glow — not free-floating light beams.
- Options: (A) document border-beam as the "beams" package, (B) find a different
  free-floating light-beam library, (C) drop beams entirely.
- Chose: (A) border-beam, with a naming note and a pointer to the third-party
  `border-beam-vanilla` port for non-React stacks.
- Why: user confirmed border-beam is the intended library.
- Result: `references/wiki/lib/beams/usage.md` documents border-beam.

### 2026-07-23 | Standalone skill, no cross-links to existing animation skills
- Context: the repo already has `create-scroll-animation`,
  `create-video-transition`, and `smart-frontend-design`.
- Options: (A) fully standalone, (B) standalone + "see also" cross-links,
  (C) absorb the others.
- Chose: (A) fully standalone.
- Why: user chose it; keeps scope tight and avoids coupling/duplication.
- Result: no references to sibling skills anywhere in this skill.

### 2026-07-23 | Native-first per language; libraries documented as published
- Context: three of the four libraries (thinking-orbs, border-beam, metal-fx) are
  React-first; only transitions.dev is framework-agnostic CSS. The skill must
  still offer per-language HTML/CSS/JS/TS files.
- Options: (A) native-first per language + libs documented as published (React,
  with vanilla-port notes), (B) hand-write vanilla adapters for every React lib.
- Chose: (A) native-first.
- Why: user chose it; forcing vanilla wrappers (esp. metal-fx WebGL) is high
  maintenance for little gain, and native CSS/WAAPI covers non-React motion well.
- Result: `lang/*` files teach native techniques; `lib/*` files document each
  library as shipped and note vanilla ports where they exist.

### 2026-07-24 | Added Canvas UI as a fifth library
- Context: user surfaced canvasui.dev (David Haz) — 27 creative canvas/WebGL
  effects over real HTML, shadcn copy-paste, React/Vue/Svelte/vanilla.
- Options: (A) add it as `lib/canvas-ui/usage`, (B) skip (out of scope),
  (C) expand language routing to Vue/Svelte to match its framework matrix.
- Chose: (A) add the lib concept only; did NOT expand language routing.
- Why: strong thematic fit for the `lib` context; Vue/Svelte are outside this
  skill's HTML/CSS/JS/TS/React routing, so adding them is scope creep for little
  gain (noted their availability in the lib file instead).
- Why-caveats: documented two non-obvious risks prominently — (1) flagship
  html-in-canvas mode is an experimental Chrome origin-trial feature (prefer its
  WebGL components for guaranteed support), (2) license is MIT + Commons Clause,
  not pure MIT like the Antalik libs.
- Result: `references/wiki/lib/canvas-ui/usage.md` + `references/raw/canvas-ui.md`;
  skill version bumped 0.1.0 -> 0.2.0.
