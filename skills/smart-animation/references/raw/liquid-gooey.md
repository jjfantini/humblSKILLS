# liquid-gooey (the "gooey" library)

Source: https://gooey.jakubantalik.com/ and
https://github.com/Jakubantalik/Libraries (monorepo, `packages/liquid-gooey`).
npm: `liquid-gooey` (v0.1.0 at time of research). Author: Jakub Antalik.
License: MIT. Researched 2026-08-12.

Same author as `border-beam` and `metal-fx`. The monorepo root README lists two
published libraries: `border-beam` and `liquid-gooey`.

## What it is
"Liquid effects for React UI. Two effects and one modifier cover the whole
family."

- **Morph** (default) — touching pieces merge gooily and change shape like jelly
  (metaball / gooey-filter behavior).
- **Move** — the surface trails a moving element as liquid rubber with a droplet
  tail.
- **dissolve** — an orthogonal modifier: the item's imagery melts into whatever
  it touches at the contact point (a liquid warp with two-liquid mixing, not a
  blur). Text is unaffected; only imagery warps.

## The architectural point (why it isn't just a CSS gooey filter)
Two rendering layers, filters never touch content:

- **Silhouette layer (SVG, below):** one blob per item, carrying the goo filter
  and the shadow chain.
- **Content layer (DOM, above):** the real, unfiltered, interactive elements.

Consequences the README calls out:
- Real shadows parsed from `box-shadow` syntax, rendered on the **merged**
  silhouette — one consistent shadow hugs the liquid through every merge/split.
  `inset` layers paint inside the liquid edge.
- Text, icons and images stay sharp (content never receives a filter).
- Works in Safari, because it filters via SVG rather than CSS `url()`.
- Focus, ARIA and handlers stay live on real DOM.
- GPU-composited animation, zero idle overhead.

## Install
```bash
npm install liquid-gooey
```

## Framework / requirements
React >= 18 (`react` and `react-dom` peer deps `>=18`). Zero runtime
dependencies. Ships ESM + CJS + types, `sideEffects: false`.

## Minimal usage
```tsx
import { Liquid } from 'liquid-gooey'

<Liquid blur={6} contrast={18} fill="#fff" shadow="0 2px 6px rgba(0,0,0,.08)">
  <Liquid.Item x={open ? -54 : 0} y={open ? -34 : 0} transition="bouncy">
    <button className="round-btn">…</button>
  </Liquid.Item>
  <Liquid.Item x={0} y={open ? -64 : 0} transition="bouncy" delay={40}>
    <button className="round-btn">…</button>
  </Liquid.Item>
</Liquid>
```

## Morph
Default. Items that touch bridge and merge automatically. Opt into jelly
shape-change physics:

```tsx
<Liquid.Item morph={{ shape: true }}>
  <div className={open ? 'panel-open' : 'panel-closed'}>…</div>
</Liquid.Item>
```

| Knob | Default | Meaning |
| --- | --- | --- |
| `shape` | `false` | Liquid shape-change physics (springs + droplet travel + corner timeline). |
| `speed` | `1` | Tempo multiplier for the shape physics (2 = twice as fast, same character). |
| `bounce` | `0.5` | `0..1` overshoot/wobble. `0` = calm, critically damped. |
| `contentBlur` | `7` | Max px **your content** cross-blurs by while the liquid moves, sharpening as it settles. `0` disables. |

## Move
```tsx
<Liquid.Item effect="move" move={{ springiness: 0.5, trail: 0.575 }}>
  <div className="thumb" style={{ transform: `translateX(${x}px)` }} />
</Liquid.Item>
```
You move the element; the liquid trails behind it.

| Knob | Default | Meaning |
| --- | --- | --- |
| `springiness` | `0.5` | `0..1` how tightly the liquid chases. 0 = heavy syrup, 1 = near-instant. |
| `wobble` | `0.5` | `0..1` overshoot on arrival. |
| `stretch` | `0.36` | `0..1` velocity stretch of the drop. |
| `trail` | `0.575` | `0..1` trailing droplet size. 0 disables the tail. |

## dissolve (modifier, not an effect)
```tsx
<Liquid.Item dissolve />                                   // the tuned look
<Liquid.Item dissolve={0.5} />                             // scaled intensity
<Liquid.Item dissolve={{ mix: 0.7, active: dragging }} />  // full DissolveOptions
```
Applies to **morph items only**; under `effect="move"` it emits a developer
warning.

## `<Liquid>` group props
Renders a `position: relative` container with the liquid silhouette SVG behind
its children. Accepts all standard `div` props. Size the group to contain the
full travel of its items.

| Prop | Default | Description |
| --- | --- | --- |
| `blur` | `6` | Goo blur sigma (px). **Bridging is a ratio of blur to gap** — neighbours merge once `blur` ≳ the gap. Items 8px apart barely bridge at `blur={5}`, merge cleanly at `blur={12}`. If pieces that should merge look separate, raise `blur` (or close the gap) first. |
| `contrast` | `18` | Alpha-contrast slope. Larger = sharper liquid edge. |
| `fill` | `'#fff'` | Liquid surface color. Any CSS color; `var(--surface)` works for theming. |
| `shadow` | — | `box-shadow` syntax, multi-layer, rendered on the merged silhouette. |
| `filterPadding` | `24` | Extra filter-region slack (px) if blobs travel outside the group's box. |

## `<Liquid.Item>` props
| Prop | Description |
| --- | --- |
| `effect` | `'morph'` (default) or `'move'`. |
| `morph` / `move` | The knobs above, plus `advanced`. |
| `dissolve` | `true`, `0..1`, or raw `DissolveOptions`. Orthogonal to `effect`. |
| `x` `y` `scale` `transition` `delay` | Component-driven position: the library animates element + liquid in one commit (pixel-perfect sync). `transition` is `'snappy' \| 'smooth' \| 'bouncy'`, `{ stiffness, damping, mass }` or `{ duration, ease }`. |
| `observe` | For plain-merge items animated by *your* code: the liquid follows the child's rendered rect. Implied by `morph.shape`, `dissolve` and `effect="move"`. |
| `radius` | Override the measured border-radius (`number` or `[tl, tr, br, bl]`). |

Border-radius is measured from computed style; percentages, circles and pills
work automatically. Item backgrounds should be transparent — the blob *is* the
visual surface.

## Advanced (raw engine values)
```tsx
<Liquid.Item
  morph={{
    shape: true,
    advanced: {
      evolve: { massStiffness: 320, cornerDuration: 820, roundness: 1 /* EvolveOptions */ },
      blobInset: 2,   // shrink the blob so an opaque photo covers its own liquid
      bridgeGrow: 8,  // liquid coat that swells out and necks into a neighbour
    },
  }}
  dissolve={{ warp: 26, mix: 0.7, gravity: 60, active: dragging /* DissolveOptions */ }}
/>

<Liquid.Item effect="move" move={{ advanced: { stiffness: 380, damping: 18 /* MoveOptions */ } }} />
```

- **EvolveOptions:** `massStiffness/Damping`, `sizeStiffness/Damping`,
  `radiusStiffness/Damping`, `contentBlur`, `roundness`, `anticipation`/`travel`,
  corner timeline (`cornerDuration`/`cornerDelay`/`cornerEase`).
- **DissolveOptions:** `blur`, `warp`, `zone`, `range`, `mix`, `gravity`/`taper`,
  `warpFreq`, `flowSpeed`, `warpStyle`, `detail`, `active`/`releaseMs`.
- **MoveOptions:** `stiffness`, `damping`, `stretch`, `tail`.

## Performance
- Component-driven motion compiles to CSS `linear()` easings (GPU-composited).
- One rAF loop per group; it sleeps entirely after stillness — zero idle cost.
- Dissolve overlay renders only near contact areas.

Caveats:
- Dissolve warps a snapshot of each image's `src` (rebuilt on resize).
- Dissolve writes `mask-image` onto `<img>` elements during contact.
- `morph.shape` applies `filter: blur()` mid-morph.
- Rotation is not mirrored in v1.

## Accessibility
- "The content layer is real DOM — focus rings, hit targets and ARIA are never
  filtered."
- **Reduced motion is respected**: component-driven transitions collapse to
  instant snaps. (Unlike metal-fx, this does not need manual wiring — but
  motion you drive yourself, e.g. the `effect="move"` transform, is still
  yours to gate.)

## Theming
Pass `fill="var(--surface)"` for light/dark switching.

## Playground
```bash
npm install && npm run dev
```
Append `?pro` to the URL for full raw-physics control panels.

## Credits (from the README)
"Filter architecture distilled from the PRO7 'Gooey plus menu' prototype; API
philosophy inspired by Sileo."
