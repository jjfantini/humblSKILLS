# Notes: porting standalone isometric HTML scenes into a Next.js demo app

Observed 2026-11-23 in a customer-demo Next.js repo (App Router, React, Tailwind, `motion/react`) that
ported three standalone single-file isometric scenes (a cold-chain pharmacy, a dock/DC with trucks, a
crew-staffing DC) into components. Customer-specific names are left out on purpose.

## Layout that repeated across all three ports
```
components/<demo>/scene/
  engine.ts          imperative SVG renderer: createEngine(svg, scenario) -> { setTime, setSelected, setHover, info(id, t), destroy }
  people.ts          itineraries, door model, poses as pure functions of time
  index.tsx          "use client" wrapper: <Scene scenario simMin selectedId onSelect />
  *-tooltip.tsx      React tooltips rendered from engine.info(id, simMin)
components/<demo>/temp-chart.tsx + temp-chart-lib.ts   pure geometry/series helpers + the SVG chart
components/<demo>/*-gantt.tsx + *-gantt-lib.ts          Gantt + pure helpers
components/<demo>/<demo>-board.tsx                      owns the sim clock; scene over chart/Gantt
lib/<demo>/temperature.ts, scenario.ts, types.ts        deterministic model, series(), statusOf()
app/demo/<demo>-scene-preview/page.tsx                  throwaway preview: ?t= &sel= &dark= &play=
```

## Facts worth keeping
- The engine is a port of the HTML file and is controlled by a sim minute. There is no internal clock and no
  requestAnimationFrame inside it. `setTime(simMin)` writes only attributes that changed.
- The static world is built once per scenario in a ref callback (`svgRef`), cleaned up by returning a destroy
  function; `useLayoutEffect` pushes `simMin`, selection, and optional extras (a technician) into the engine.
- The board owns the clock (rAF, `dt` clamped to 0.1 s, state painted at ~30 fps, a `timeRef` for the true value)
  so only the scene + chart subtree re-renders each frame, not the whole stage.
- Clock policy per story phase: a rate per phase (calm / warning / escalation), a cap so the clock never runs
  past the next scripted moment, snap-forward to the latest beat's `sim_min`, and an eased time-lapse
  (`easeInOut`, ~7-8 s, 600 ms under reduced motion) to the outcome once the story is told.
- Hover tooltips are positioned by a ref (`transform: translate`) from the pointer, flip near the edge, and are
  recomputed from the current minute so they tick while hovered. Mouse only (`pointerType === "mouse"`).
- Selection: click or Enter/Space on `[data-unit]` / `[data-door]` / `[data-person]` elements, focus-visible dashed outline.
- Pulses (halo, dot, ring, tint) are CSS keyframes on opacity/transform with a `prefers-reduced-motion` fallback
  that shows a static emphasized state.
- Tokens: status colors green #1D9E75, amber #E8961C, red for danger; chips and chart lines use Tailwind design
  tokens (`bg-success/15 text-success`, `stroke-open`), never raw hex outside the scene SVG.
- Chart helpers: per-unit `series()` precomputed over the day, with running min / max / out-of-range-minute prefix
  arrays so the cards read lookups; markers (`text`, `call`, `dispatch`, `repair`, `excursion`, `cleared`) laid out
  on the same x axis, a now-line at the sim minute, safe band, dashed limits.
- Verification used: a real-time Playwright run with `channel: "chrome"`; virtual-time headless screenshots misled
  because library staggers never completed.
