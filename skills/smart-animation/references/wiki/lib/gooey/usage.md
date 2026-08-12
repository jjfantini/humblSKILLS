---
title: "liquid-gooey: Metaball Merge and Liquid Trails"
context: lib
category: gooey
concept: usage
description: "liquid-gooey (React): Morph (touching pieces merge like jelly), Move (liquid-rubber trail), and the dissolve modifier. SVG silhouette layer keeps content crisp and interactive; blur-to-gap ratio controls bridging."
tags: liquid-gooey, react, metaball, gooey, svg-filter, fab, morph, spring
sources:
  - "references/raw/liquid-gooey.md"
last_ingested: 2026-08-12
---

## liquid-gooey: Metaball Merge and Liquid Trails

`liquid-gooey` makes React elements behave like drops of liquid: pieces that
touch merge, a moving element drags a rubbery tail, and imagery can melt into a
neighbour at the contact point. React >= 18, zero runtime dependencies.

The reason to reach for it over a hand-rolled CSS gooey filter: it renders **two
layers**. An SVG silhouette below carries the goo filter and the shadow; your
real DOM rides crisp on top and is never filtered. So text stays sharp, focus
rings and ARIA survive, and it works in Safari (SVG filtering, not CSS `url()`).

```
npm install liquid-gooey
```

**Incorrect (filter the content itself — the classic CSS gooey hack):**

```css
/* Blurs the text and icons along with the shapes, kills hit-testing
   precision, and the shadow is per-element so merged blobs show seams. */
.fab-group { filter: blur(6px) contrast(18); }
```

**Correct (silhouette below, real DOM above, one shadow on the merged shape):**

```tsx
import { Liquid } from 'liquid-gooey'

<Liquid blur={12} contrast={18} fill="var(--surface)" shadow="0 2px 6px rgba(0,0,0,.08)">
  <Liquid.Item x={open ? -54 : 0} y={open ? -34 : 0} transition="bouncy">
    <button className="round-btn" aria-label="Share">…</button>
  </Liquid.Item>
  <Liquid.Item x={0} y={open ? -64 : 0} transition="bouncy" delay={40}>
    <button className="round-btn" aria-label="Edit">…</button>
  </Liquid.Item>
</Liquid>
```

Two effects and one modifier cover everything:

- **Morph** (default) — touching items bridge and merge. `morph={{ shape: true }}`
  adds jelly shape-change physics (`speed` 1, `bounce` 0.5, `contentBlur` 7).
- **Move** — `effect="move"`; you move the child, the liquid trails it
  (`springiness` 0.5, `wobble` 0.5, `stretch` 0.36, `trail` 0.575).
- **dissolve** — orthogonal modifier (`true` | `0..1` | `DissolveOptions`); the
  item's imagery melts into what it touches. **Morph items only** — under
  `effect="move"` it just warns.

**Bridging is a ratio of blur to gap, not a mode.** Items merge once `blur` ≳ the
gap between them: 8px apart barely bridges at `blur={5}`, merges cleanly at
`blur={12}`. If pieces that should merge look separate, raise `blur` or close
the gap before suspecting anything else. Give items transparent backgrounds —
the blob *is* the surface — and size the group to contain the items' full
travel (or raise `filterPadding`, default 24).

Group props: `blur` 6, `contrast` 18, `fill` `'#fff'`, `shadow` (box-shadow
syntax, drawn on the *merged* silhouette so one shadow hugs it through every
merge and split; `inset` layers paint inside the liquid edge), `filterPadding`
24. Item props: `x`/`y`/`scale`/`transition`/`delay` (component-driven — element
and liquid animate in one commit, so they can't desync), `observe` (follow a
child your own code animates), `radius`.

Performance: component-driven motion compiles to CSS `linear()` easings, one rAF
loop per group that sleeps completely once still — zero idle cost. Watch the
caveats: dissolve snapshots each image `src` (rebuilt on resize) and writes
`mask-image` onto `<img>` during contact, `morph.shape` applies a real
`filter: blur()` mid-morph, and rotation is not mirrored in v1.

Reduced motion is honored — component-driven transitions collapse to instant
snaps. Motion you drive yourself (the transform you feed an `effect="move"`
child) is still yours to gate; see `motion/principles/accessibility`.

Fit: ONE liquid moment — an expanding FAB cluster, a toggle/slider thumb with a
trail, a card that morphs into a panel. It is a signature effect, not a
site-wide texture. For a border glow use `lib/beams`; for a metallic accent use
`lib/metal`.

## Sources

- `references/raw/liquid-gooey.md` — two-layer architecture, full prop and knob
  tables with defaults, blur-to-gap bridging rule, performance caveats,
  reduced-motion behavior, npm/peer-dep facts.
