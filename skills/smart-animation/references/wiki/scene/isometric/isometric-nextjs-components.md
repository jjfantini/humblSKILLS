---
title: "Isometric Scenes as Next.js Components with Charts and Visual Aids"
context: scene
category: isometric
concept: isometric-nextjs-components
description: "Port a standalone isometric HTML scene into React/Next.js: an imperative SVG engine, a client wrapper, a board that owns the sim clock, plus a linked chart, Gantt, stats, and legend."
tags: isometric, nextjs, react, svg, sim-clock, chart, gantt, tooltip, reduced-motion
sources:
  - "references/raw/nextjs-isometric-port-notes.md"
  - "references/raw/isometric-3d-models-skill.md"
last_ingested: 2026-11-23
---

## Isometric Scenes as Next.js Components

The standalone HTML recipe (`isometric-demo-scenes.md`) is the prototype. In a Next.js app
the same scene becomes three layers: a **pure model** (`series`, `statusOf`), an
**imperative SVG engine** driven by a sim minute, and a **board** that owns the
clock and renders the scene next to a chart, Gantt, and stat row. Keep the
engine out of React state: React owns props, the engine owns the SVG nodes.

**Incorrect (scene re-rendered through React on every frame, clock in the page):**

```tsx
// 60 setState calls a second re-render the whole stage; the SVG is rebuilt each time.
const [t, setT] = useState(0);
useEffect(() => { const id = setInterval(() => setT(v => v + 1), 16); return () => clearInterval(id); }, []);
return <Stage>{rooms.map(r => <Room key={r.id} temp={temperatureAt(r, t)} />)}</Stage>;
```

**Correct (engine built once, clock in a leaf board, painted at about 30 fps):**

```tsx
"use client";
export function Scene({ scenario, simMin, selectedId, onSelect }: SceneProps) {
  const [engine, setEngine] = useState<Engine | null>(null);
  const svgRef = useCallback((svg: SVGSVGElement | null) => {
    if (!svg) return;
    const eng = createEngine(svg, scenario);          // static world, built once per scenario
    setEngine(eng);
    return () => { eng.destroy(); setEngine(null); };
  }, [scenario]);
  useLayoutEffect(() => { engine?.setTime(simMin); }, [engine, simMin]);      // writes only changed attributes
  useLayoutEffect(() => { engine?.setSelected(selectedId); }, [engine, selectedId]);
  return <svg ref={svgRef} viewBox={`0 0 ${vbW} ${vbH}`} role="group" aria-label="..." />;
}
```

The engine internals (projection, primitives, sprites, layering, frame function,
chart and Gantt helpers) and the quality bar are in `isometric-build-kit.md`.

### Structure

```
components/<demo>/scene/{engine.ts, people.ts, index.tsx, *-tooltip.tsx}
components/<demo>/{chart.tsx, chart-lib.ts, gantt.tsx, gantt-lib.ts, board.tsx}
lib/<demo>/{scenario.ts, model.ts, types.ts}      # deterministic, importable by tests
app/demo/<demo>-scene-preview/page.tsx            # ?t= &sel= &dark= &play=
```

### The board owns the clock

- rAF loop, `dt` clamped to 0.1 s, a `timeRef` for the true time, `setTime` at
  ~33 ms so only the scene and chart re-render.
- Per phase: a rate, a cap (never run past the next scripted moment), snap
  forward to the latest beat's sim minute, and an eased time-lapse to the
  outcome (about 7-8 s; 600 ms with `useReducedMotion`).
- Expose the clock only as a number; the engine, chart cursor, and Gantt now-line
  all read the same `simMin`.

### Charts and visual aids that pay for themselves

Add each one only if it answers a question the scene alone cannot.

| Aid | What it adds | Notes |
|---|---|---|
| Time-series chart | The history behind the status colors | Precompute `series()` once; prefix arrays for running min/max/out-of-range minutes so cards are lookups; safe band, dashed limits, now-line at `simMin` |
| Event markers | Why the line moved | Kinds like text, call, dispatch, repair, excursion, cleared on the same x axis; lay out to avoid overlap |
| Gantt | Capacity and conflicts | Rows per resource, now chip, drag with snap, invalid moves red and not applied; selecting a bar selects the scene object |
| Stat row | The headline number | Count in range, minutes out of range, unstaffed doors, at-risk deliveries; derived from the same model |
| Status chip + legend | Meaning of the colors | Tailwind tokens (`bg-success/15 text-success`), same words as the scene tags |
| Floor halo / ring | Where to look | CSS keyframes on opacity/transform only, with a static reduced-motion state |
| Hover tooltip | Detail on demand | Ref-positioned `translate`, flips at the edge, recomputed from the current minute so it ticks while hovered, mouse only |
| Selection sync | One focus across views | Click or Enter/Space on `[data-*]` elements; scene, chart, and Gantt share `selectedId` |

### Rules

- Status colors stay semantic (green, amber, red); chrome uses design tokens, raw hex only inside the scene SVG.
- Every pulse has a `prefers-reduced-motion` fallback that shows a static emphasized state.
- Keep a throwaway preview route with URL params to jump to a storyline minute and screenshot it.
- Verify in a real-time browser run (Playwright `channel: "chrome"`); virtual-time screenshots can mislead when library staggers never complete.

## Sources

- `references/raw/nextjs-isometric-port-notes.md` - what repeated across three real ports (layout, clock policy, tooltips, chart helpers, verification).
- `references/raw/isometric-3d-models-skill.md` - the standalone recipe these components port.
