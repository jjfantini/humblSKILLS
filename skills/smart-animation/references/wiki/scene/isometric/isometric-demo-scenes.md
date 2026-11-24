---
title: "Isometric Demo Scenes: SVG First, Three.js When Earned"
context: scene
category: isometric
concept: isometric-demo-scenes
description: "Animated isometric (2.5D) demo scenes of a warehouse, pharmacy, factory or any site: hand-built SVG plus vanilla JS where all state is a function of a sim clock, data-driven triggers, panels and a SCENE API. Includes when to switch to Three.js and how to carry the same clock, trigger and API pattern into it. By Carter Stonecipher, who built two demo scenes with it."
tags: isometric, 3d, threejs, svg, warehouse, demo, simulation, digital-twin, gantt
sources:
  - "references/raw/isometric-3d-models-skill.md"
last_ingested: 2026-11-23
---

## Isometric Demo Scenes: SVG First, Three.js When Earned

Use this concept for any animated scene of a place: a warehouse or distribution center, a pharmacy, a
hospital floor, a factory, a dock, a store, a data center. It also covers any request that says
"3D", "isometric" or "Three.js". The recipe is Carter Stonecipher's. He built a distribution center
and a pharmacy cold chain with it. Both were demo-ready, with a deterministic clock, triggers, status
tags and a Gantt panel. Read the raw source for the full kit. This file gives the decisions, the
order of work and the fixes that this skill adds on top.

### 1. Pick the renderer first

| Need | Use |
|---|---|
| A fixed isometric view, flat illustration, demo or exec storytelling, ≤ a few hundred objects | **SVG + vanilla JS** (default) |
| The viewer orbits or zooms the camera, real lighting or shadows, perspective, imported GLB/glTF models, thousands of objects, VR | **Three.js** |
| A static diagram with no motion | Not this skill |

SVG is the default for a reason. It is crisp at any zoom, has no dependency and no CDN or CSP risk,
and every face is a DOM node you can hover. A "3D warehouse" request usually means this look, not a
free camera. Ask one question when the request is unclear: "Fixed isometric view, or must the
viewer rotate the camera?"

### 2. Order of work

1. **Research the place**: zones, fixtures, flows and real thresholds (for example, CDC fridge range
   2-8 °C). Cite sources and say what you inferred.
2. **Brand**: try `/brand`, `/press`, `/media-kit`, then search for "<company> brand guidelines hex".
   Take colors and plain-text names only. Never redraw a logo or mascot. Keep status colors
   (ok #1D9E75, warn #E8961C, danger red) separate from the brand colors.
3. **Scene spec, written down**: zones, static props, actors (2-6 per type), the trigger storyline
   with sim times, what the panel shows, and the deliverable. Ask only for what you cannot infer.
4. **Build** one file in the architecture below.
5. **Verify** with Node (logic) and then a browser (pixels). See step 8.
6. **Deliver**: write the `.html` file and give its full path. In Claude Code, publish it as an
   Artifact when the user wants to share it. Summarize the storyline, controls and API in a few lines.

### 3. Architecture (both renderers)

```
KPI cards -> scene -> control bar (play/pause, clock, scrubber, rate, trigger buttons, reset)
-> panel (status cards with sparklines | Gantt | event log) -> one shared hover tooltip
```

- **All state is a function of sim time `T`** (minutes of day). Loop: `T += dt * rate` with `dt`
  clamped to 0.05 s, wrap at the end, then `render(T)`. Precompute per-minute series at `init()` so
  charts, min/max and minutes out of range are lookups.
- **No `Math.random()` in the model.** Seed noise once at init (for example, mulberry32 with a
  fixed seed) or use sinusoids. Otherwise scrubbing back to 13:03 shows a different scene each time.
- Seed data (`CFG0`, `ITEMS0`) is deep-cloned for Reset. Never JSON-clone objects that hold DOM refs.
- Start the demo at a moment where every interesting state is visible. Use a gentle rate (0.5-1
  sim-min/s) with faster options.

### 4. Isometric kit (SVG)

The kit uses true isometric (`C = .866 = cos 30°`, `SC = .5 = sin 30°`: each floor axis runs at 30° on
screen, not the 2:1 pixel-art slope), so `P(x,y,z) = [ox+(x-y)*C*s, oy+(x+y)*SC*s - z*s]`,
and 1 unit ≈ 1 m. The raw source only sketches `P`, `pts`, `poly`, `box`, `fy`, `fx` and `ftext`; write them
from `isometric-build-kit.md`, which has the complete code. The fit is
`s = min(22, 640/((W+D)*C))` and `ox = (680-(W+D)*C*s)/2 + D*C*s`. Set `oy` from the tallest height
plus the tallest floating tag.

- Materials are `[top, left, right]` triples, with the top lightest. Stroke = fill hides seams.
- **Painter's order by `x+y`**, far to near. Static order: slab -> floor markings -> things behind
  the back wall -> the two far walls -> wall fixtures -> mid islands -> near racks -> **actor
  layers** -> front counter -> foreground.
- **Cutaway**: draw only the back (y≈0) and left (x≈0) walls. Put docks, trucks and entrances behind
  the back wall, so they never cover the interior.
- **Actors among tall rows**: make one actor layer before each tall row plus a top layer. Each frame,
  an actor's layer is the number of rows it is fully in front of. Re-append nodes only when that
  order changes. Leave at least one object height of clear floor between rows.
- Build actor variants (heading × state) once in `<defs>` and move `<use>` by `transform` only.

**Incorrect (actors float above the floor):**

```js
use.setAttribute('transform', `translate(${(x-y)*C*s},${(x+y)*SC*s})`);
```

**Correct (include the origin offset):**

```js
use.setAttribute('transform', `translate(${ox+(x-y)*C*s},${oy+(x+y)*SC*s})`);
```

### 5. Behavior, triggers and labels

- **Itineraries**: cyclic stops `[[station, dwell], ...]` plus specials `{t, station, dwell, tag}`.
  Route on orthogonal corridors. Machines use `loopPos(path, tau, speed, dwellA, dwellB)` and dwell
  at handoff points.
- **Couple motion to data**: doors open only while a vehicle is docked, staff work only on shift,
  and a door left ajar raises the temperature. Signals are `base + small noise + drift + Σ event
  responses`, with a rise of `amp·(1-e^(-t/τ))` and a recovery of `e^(-t/τr)`. Tune them until noise
  never flickers across a threshold by accident.
- **Triggers are data** (`TRIGGERS0`): scheduled (`at`), conditional (`when(state)`), manual
  (buttons) and external (`SCENE.trigger`). A fired trigger inserts an event into the timeline and
  the model is rebuilt, so scrubbing stays consistent. Each trigger writes to the event log and
  emits `scene:trigger`. Each status change emits `scene:status` (prev -> next).
- **Status tags**: dark pill, status dot, ID and value, sized to the longest text. Add a leader line
  to the object. Stagger leader heights so neighboring tags are at least 2× the pill height apart,
  and raise the far-back tags more. Show status in more than one way: light, readout, tint, floor
  halo and panel chip.
- **Calm over busy**: few actors, no decorative roaming, purposeful motion only.

### 6. Integration API, made safe

```js
window.SCENE = { setTime, getTime, play, pause, setRate, select, state, setData, upsert, remove,
  updateItem, setReading, trigger, configure, timeline };
addEventListener('message', (e) => {
  if (e.source !== parent || !ALLOWED_ORIGINS.includes(e.origin)) return;  // the raw source has no check
  const { cmd, ...args } = e.data || {};
  if (Object.hasOwn(SCENE, cmd)) SCENE[cmd](...(args.args || []));
});
```

Post events to `parent` only when the scene is embedded (`parent !== window`). Insert all injected
text with `textContent`, never with `innerHTML`. Data from an API or postMessage is untrusted.

### 7. Performance and reduced motion (this skill's floor)

- Pause the rAF loop on `document.visibilitychange` and when an `IntersectionObserver` reports the
  scene offscreen. While paused, render only on `setTime`, a scrub or a trigger.
- Move only `transform` on actors. Change `fill` or `class` only when a status changes, not every
  frame.
- **`prefers-reduced-motion: reduce`**: start paused at the key moment, with no walk bob, no pulsing
  halos and no looping dots. Show status with color and text. The scrubber, triggers and API still
  work. The user moves time, not the page. See `references/wiki/motion/principles/accessibility.md`.
- **Accessibility**: give the SVG `role="img"` and an `aria-label` that names the current state.
  Make the controls real `<button>`s. Show the tooltip on keyboard focus as well as on hover.

### 8. Verify with logic first, then pixels

1. Extract the `<script>`, run `node --check`, then run it against a tiny DOM stub (the raw source
   lists what the stub needs). Step 2000-4000 frames, `setTime` to each storyline checkpoint, assert
   the expected statuses, call every API method, and scan all attributes and text for
   `NaN|undefined|Infinity`.
2. **Then look at it.** The raw source ended with "say visuals were not browser-checked". Here a
   browser is available, so check them. Read `?t=<minutes>&paused=1` at boot, then take one
   screenshot per checkpoint:

```bash
for t in 540 783 858; do
  "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" --headless=new --hide-scrollbars \
    --window-size=1360,1000 --screenshot="shots/t$t.png" "file://$PWD/scene.html?t=$t&paused=1"
done
```

   Open every PNG. Check that actors stand on the floor, tags do not overlap, no tall row hides an
   actor it should not, and the dark theme reads.

### 9. Three.js: the same scene with a real camera

Keep sections 2, 3, 5, 6 and 7 unchanged. Only the drawing changes.

- **Camera**: an `OrthographicCamera` at `(d, d, d)` looking at the floor center matches the SVG kit
  exactly. That is 45° azimuth and 35.26° elevation, which is true isometric. A 30° elevation gives the
  flatter 2:1 pixel-art look. Use `PerspectiveCamera` only when the user wants depth. Add `OrbitControls` only when
  the user must rotate the view, and clamp its polar angle so the floor never flips.
- **Render on demand**: call `renderer.render()` only when `T` changes, a control moves or the window
  resizes. Do not render every frame while paused. In React, use `@react-three/fiber` with
  `frameloop="demand"` and `invalidate()`.
- `renderer.setPixelRatio(Math.min(devicePixelRatio, 2))`. Use one `InstancedMesh` per repeated prop
  (racks, totes, pallets). Share geometries and materials. Use `MeshLambertMaterial` or a flat
  `MeshToonMaterial` with one directional light plus a hemisphere light for the flat illustration look.
- Picking for the tooltip: use one `Raycaster` on `pointermove`, throttled to one call per frame.
- **Load three from a pinned version**, with an import map to an exact `three@x.y.z` or a bundle.
  Never use an unversioned CDN URL. If the host blocks CDNs (CSP), that is the raw source's reason to
  stay on SVG.
- Teardown: cancel rAF, call `dispose()` on geometries, materials, textures and the renderer, and
  remove listeners. React effects must do this in their cleanup.
- Verify the same way. Run the logic through the clock functions without WebGL, then take headless
  Chrome screenshots at the checkpoints. Headless Chrome on macOS drew WebGL with no extra flag
  (checked 2026-11-23). If a CI machine renders black, add `--use-angle=swiftshader`.

## Sources

- `references/raw/isometric-3d-models-skill.md` - Carter Stonecipher's skill: workflow, brand
  styling, architecture, isometric kit and fit math, draw order and actor layers, behavior from
  time, triggers, labels, panels, API, Node verification, and the two reference scenes. This skill
  adds the renderer choice, seeded noise, the postMessage origin check, pausing and reduced motion,
  browser screenshots and the Three.js mapping.
