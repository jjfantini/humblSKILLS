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
- Context: the repo already has `create-smart-scroll-animation`,
  `create-smart-video-transition`, and `smart-frontend-design`.
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

### 2026-07-30 | lib slug `motion-react`, not `motion`
- Context: adding Motion (formerly Framer Motion) as a lib concept, while the
  wiki already has a top-level `motion` context holding the shared principles
  (`wiki/motion/principles/*`).
- Options: (A) `lib/motion/usage`, (B) `lib/motion-react/usage`,
  (C) `lib/framer/usage`.
- Chose: (B) `motion-react`.
- Why: (A) is filesystem-legal but makes `_index.md` read "lib -> motion" next to
  a separate "motion" context, which is exactly the kind of ambiguity a router
  should not have to disambiguate. (C) encodes the outdated `framer-motion`
  package name the library has moved off. `motion-react` matches the actual
  import path `motion/react`.
- Result: `references/wiki/lib/motion-react/usage.md`; no collision in the index.

### 2026-07-30 | Beautiful UI documented despite having no license and no package
- Context: user requested beautiful-ui-five.vercel.app. Inspection of the page
  payload confirmed: 17 components with full copyable TypeScript source, but no
  npm package, no shadcn registry, no GitHub repo, no docs, and no license
  statement anywhere.
- Options: (A) skip it (not a library), (B) document as a design/pattern
  reference only, (C) document as copy-paste with the full integration story.
- Chose: (C), with the license gap stated explicitly.
- Why: the source genuinely is on the page and copyable, so (B) would have
  understated it. Its dependency profile is near-perfect for this skill (only
  `react`, plus `liveline` in one component) — but the tokens, the seven
  `@keyframes`, and the single global `prefers-reduced-motion` block all live in
  the site's global CSS, so a naive copy renders unstyled and motionless with no
  error. That silent-failure mode is the most useful thing the skill can warn
  about, and it only exists if we document the copy path.
- Why-caveats: no license = all rights reserved by default; flagged to seek an
  explicit grant before shipping copied source commercially, and noted that the
  state taxonomy (the genuinely reusable part) is not copyrightable.
- Result: `references/wiki/lib/beautiful-ui/usage.md` + `references/raw/beautiful-ui.md`.

### 2026-07-30 | React Bits: dependency contract framed as the headline caveat
- Context: reactbits.dev advertises "minimal dependencies … tree-shakeable", but
  the repo pulls gsap, motion, three, ogl, matter-js, @react-three/fiber, drei.
- Options: (A) repeat the site's framing, (B) lead with the per-component
  dependency contract.
- Chose: (B).
- Why: both claims are true at different scopes — light per component, heavy in
  aggregate. A text effect pulls gsap; a 3D background pulls the whole three.js
  stack. The copied file's import block is the only honest signal, so the skill
  tells you to read it before installing.
- Why-caveats: components ship no `prefers-reduced-motion` handling at all, and
  scroll reveals / particle backgrounds are the highest-risk category for
  vestibular discomfort; since copy-paste means you own the file, the guidance is
  to add the gate in the file. License is MIT + Commons Clause (same as Canvas UI).
- Result: `references/wiki/lib/react-bits/usage.md` + `references/raw/react-bits.md`.

### 2026-07-30 | Reverses "fully standalone, no cross-links" (2026-07-23)
- Context: the 2026-07-23 entry "Standalone skill, no cross-links to existing
  animation skills" chose zero references to sibling skills, naming
  smart-frontend-design explicitly. That decision was already being violated in
  substance: `raw/frontend-design-motion-principles.md` is distilled from the
  same upstream plugin skill that smart-frontend-design mirrors, and
  `motion/principles/design.md` says so in its own body ("Rules distilled from
  the frontend-design skill").
- Options: (A) keep standalone and accept silent duplication, (B) declare the
  boundary in both descriptions, (C) merge the two skills.
- Chose: (B).
- Why: the two skills are not duplicates, they are different *depths* of one
  upstream sentence — smart-frontend-design mirrors "Leverage motion
  deliberately" shallowly (correct for a mirror), this skill expands it into
  three principle concepts. Undeclared, that reads as plagiarism in one
  direction and as redundancy in the other; declared, it is clean layering.
  (C) was rejected because narrow scope is the only thing keeping this skill's
  trigger from colliding with smart-frontend-design's on prompts like "make
  this look better".
- Why-caveats: the boundary is declared in *this* skill's description and in
  smart-frontend-design's router. smart-frontend-design is an upstream mirror,
  so its side is limited to a routing pointer that does not alter mirrored
  content.
- Result: description gained the ownership sentence; raw file gained a
  provenance header naming the shared upstream. Supersedes the 2026-07-23
  standalone decision, which stands as history.

### 2026-07-30 | Scope stays motion-only; Beautiful UI is the canary
- Context: Beautiful UI was added under `lib/` in v0.3.0, but its own page
  argues the state taxonomy matters more than its animation. It is the first
  entry that is not primarily a motion library.
- Options: (A) widen this skill to general frontend craft, (B) keep motion-only
  and split a component-library skill when a second non-motion find arrives,
  (C) reject Beautiful UI.
- Chose: (B).
- Why: (A) re-creates the exact trigger collision with smart-frontend-design
  that the boundary decision above resolves — narrow scope is load-bearing.
  (C) discards a genuinely useful reference over taxonomy purity. Beautiful UI
  is defensible here today (streaming, shimmer, expand transitions, seven
  keyframes) and is the signal to watch.
- Result: no scope change. Revisit when a second non-motion component library
  is worth documenting.
