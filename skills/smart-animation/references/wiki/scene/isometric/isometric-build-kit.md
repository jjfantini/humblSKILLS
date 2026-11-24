---
title: "Isometric Build Kit: Write the Engine, Sprites, and Chart Helpers"
context: scene
category: isometric
concept: isometric-build-kit
description: "How to write the code an isometric demo scene needs, from projection and box helpers to actor layering, dirty-checked setTime, status tags, and chart geometry, plus the quality bar a finished scene must meet."
tags: isometric, svg, engine, sprites, layering, chart, gantt, quality-bar, typescript
sources:
  - "references/raw/nextjs-isometric-engine-details.md"
  - "references/raw/nextjs-isometric-port-notes.md"
  - "references/raw/isometric-3d-models-skill.md"
last_ingested: 2026-11-23
---

## Isometric Build Kit

No template ships with this skill: write these pieces fresh for each scene, in
this order. The snippets are complete and framework-free (TypeScript; strip
types for a standalone HTML file). A scene that skips the quality bar at the end
looks like a diagram, not a demo.

**Incorrect (stub helpers, random clutter, rebuild every frame):**

```ts
function box(...) { /* faces */ }                 // never written: draw order breaks
shelf.forEach(() => Math.random() < .8 && item()); // different scene every reload
function render(t) { svg.innerHTML = ""; drawWorld(); drawPeople(t); }
```

**Correct (real helpers, seeded clutter, world built once, frames write deltas):**

```ts
const world = createEngine(svg, scenario);  // builds static world + sprites once
world.setTime(t);                           // pure: same t, same frame; writes only changes
```

### 1. Projection and layout

```ts
const C = 0.866, SC = 0.5, VBW = 680;
export function computeLayout(W: number, D: number, tallest: number) {
  const s = Math.min(22, 640 / ((W + D) * C));
  const ox = (VBW - (W + D) * C * s) / 2 + D * C * s;
  const oy = tallest * s + 40;                 // headroom for the highest status tag
  return { s, ox, oy, vbW: VBW, vbH: Math.round(oy + (W + D) * SC * s + 26) };
}
// inside createEngine:
let zero = false;                              // true while drawing sprites into <defs>
const P = (x: number, y: number, z = 0): [number, number] =>
  zero ? [(x - y) * C * s, (x + y) * SC * s - z * s]
       : [ox + (x - y) * C * s, oy + (x + y) * SC * s - z * s];
const pts = (a: number[][]) => a.map(p => P(p[0], p[1], p[2] ?? 0).map(v => v.toFixed(1)).join(",")).join(" ");
```

1 unit = 1 m. `W` runs along x, `D` along y; the viewer looks from +x,+y.

### 2. Drawing primitives

```ts
type Pal = readonly [top: string, left: string, right: string];
function el<K extends keyof SVGElementTagNameMap>(tag: K, a: Record<string, string | number>, parent?: Element) {
  const e = document.createElementNS("http://www.w3.org/2000/svg", tag);
  for (const k in a) e.setAttribute(k, String(a[k]));
  parent?.appendChild(e);
  return e;
}
const poly = (a: number[][], fill: string, g: Element, stroke = fill, sw = 0.6) =>
  el("polygon", { points: pts(a), fill, stroke, "stroke-width": sw, "stroke-linejoin": "round" }, g);
function box(x: number, y: number, z: number, w: number, d: number, h: number, c: Pal, g: Element) {
  poly([[x+w,y,z],[x+w,y+d,z],[x+w,y+d,z+h],[x+w,y,z+h]], c[2], g);   // +x face (right, darkest)
  poly([[x,y+d,z],[x+w,y+d,z],[x+w,y+d,z+h],[x,y+d,z+h]], c[1], g);   // +y face (left, mid)
  poly([[x,y,z+h],[x+w,y,z+h],[x+w,y+d,z+h],[x,y+d,z+h]], c[0], g);   // top (lightest)
}
const fy = (x1: number, x2: number, y: number, z1: number, z2: number, f: string, g: Element, st?: string) =>
  poly([[x1,y,z1],[x2,y,z1],[x2,y,z2],[x1,y,z2]], f, g, st);         // fronts, screens, bands, doors
const fx = (x: number, y1: number, y2: number, z1: number, z2: number, f: string, g: Element, st?: string) =>
  poly([[x,y1,z1],[x,y2,z1],[x,y2,z2],[x,y1,z2]], f, g, st);         // side walls, windows
function ftext(x: number, y: number, z: number, str: string, size: number, fill: string, g: Element) {
  const [X, Y] = P(x, y, z);
  const t = el("text", { transform: `matrix(${C},${SC},0,1,${X.toFixed(1)},${Y.toFixed(1)})`,
    "font-size": size, "font-weight": 500, fill, "text-anchor": "middle" }, g);
  t.textContent = str;
  return t;
}
let seed = 11;                                                         // deterministic clutter
const rnd = () => ((seed = (seed * 1664525 + 1013904223) % 4294967296) / 4294967296);
```

Offset a quad 0.001-0.003 off its parent plane (`y + 0.001`) when it sits on
a face, so it never z-fights.

### 3. Static world, far to near

Floor slab (white, plus a 0.7-deep brand-color edge on the two visible sides),
floor zones with tile lines, things behind the back wall, the two cutaway walls,
wall fixtures sorted by x, mid islands, then tall rows **with an actor layer
created before each row**, then a top actor layer, then front counters and the
foreground. Fill shelves and bins with seeded clutter: small `fy`/`fx` product
quads at 75-85% occupancy in 3-4 colors (neutrals plus one brand accent). This
clutter is most of what makes a scene read as a real place.

### 4. Sprites and actors

```ts
type Part = readonly [number, number, number, number, number, number, Pal];
function drawParts(parts: Part[], g: Element) {
  parts.slice().sort((a, b) => a[0]+a[1]+a[3]/2+a[4]/2 - (b[0]+b[1]+b[3]/2+b[4]/2) || a[2]-b[2])
       .forEach(q => box(q[0], q[1], q[2], q[3], q[4], q[5], q[6], g));
}
const person = (skin: Pal, uniform: Pal, accent: Pal, hair: Pal): Part[] => [
  [-.16,-.12,0,   .32,.24,.62, DARK],    // legs
  [-.25,-.18,.55, .5, .36,.95, uniform], // torso
  [-.05, .17,1.05,.1, .03,.35, accent],  // badge / lanyard
  [-.13,-.13,1.5, .26,.26,.28, skin],    // head
  [-.14,-.14,1.72,.28,.28,.09, hair],
];
zero = true; drawParts(person(...), el("g", { id: `${uid}-p0` }, defs)); zero = false;

function makeActor(id: string, sprite: string, layer: Element, r = .5) {
  const g = el("g", { "data-person": id }, layer);
  el("ellipse", { rx: r*s*1.2247, ry: r*s*.7071, fill: "rgba(0,0,0,.12)" }, g);          // floor shadow
  const ring = el("ellipse", { rx: .62*s*1.2247, ry: .62*s*.7071, fill: "none", stroke: SELECT, "stroke-width": 1.5, display: "none" }, g);
  const body = el("g", {}, g);
  el("use", { href: `#${uid}-${sprite}` }, body);
  return { g, ring, body, tf: "" };
}
function place(a: Actor, x: number, y: number, walking: boolean, stride: number) {
  const tf = `translate(${(ox+(x-y)*C*s).toFixed(1)} ${(oy+(x+y)*SC*s).toFixed(1)})`;   // ALWAYS include ox, oy
  if (tf !== a.tf) { a.g.setAttribute("transform", tf); a.tf = tf; }
  a.body.setAttribute("transform", walking ? `translate(0 ${(-Math.abs(Math.sin(stride*6))*1.5).toFixed(1)})` : "");
}
```

Use a unique `uid` per engine instance in `<defs>` ids, so two scenes on one page never collide.

### 5. Movement as a pure function of time

```ts
const route = (a: Pt, b: Pt): Pt[] =>                        // orthogonal, through a known corridor x
  Math.abs(a[1]-b[1]) < .6 ? [a, [b[0], a[1]], b] : [a, [CX, a[1]], [CX, b[1]], b];
function along(p: Pt[], d: number): Pt {
  for (let i = 1; i < p.length; i++) {
    const L = Math.hypot(p[i][0]-p[i-1][0], p[i][1]-p[i-1][1]);
    if (d <= L || i === p.length-1) { const u = L ? Math.min(1, d/L) : 1;
      return [p[i-1][0]+(p[i][0]-p[i-1][0])*u, p[i-1][1]+(p[i][1]-p[i-1][1])*u]; }
    d -= L;
  }
  return p[p.length-1];
}
// Itinerary = segments {t0, t1, walk, path, len, at, task}; build once from [[station, dwell], ...] + specials.
function poseAt(segs: Seg[], t: number) {
  let lo = 0, hi = segs.length - 1;
  while (lo < hi) { const m = (lo + hi + 1) >> 1; segs[m].t0 <= t ? (lo = m) : (hi = m - 1); }
  const g = segs[lo];
  if (g.walk && t < g.t1) { const d = (t - g.t0) / (g.t1 - g.t0) * g.len; const [x, y] = along(g.path, d);
    return { x, y, walk: true, stride: d, task: `Walking to ${g.toName}` }; }
  return { x: g.at[0], y: g.at[1], walk: false, stride: 0, task: g.task };
}
```

Doors and lids swing on a hinge: free edge at `(hx + w*cos a, y + w*sin a)`;
round the angle to 0.1 degrees and skip the write when unchanged. Couple motion
to data (a door opens only while someone is at it; a truck backs in only while
its appointment is active).

### 6. Layering actors among tall rows

```ts
// rows: tall rows sorted far -> near; layers[i] was appended just BEFORE rows[i]; layers[rows.length] is on top.
const layerFor = (x: number, y: number) => { let i = 0; while (i < rows.length && x + y > rows[i].depth) i++; return i; };
function reorder() {
  const sig = actors.map(a => `${a.layer}:${a.depth.toFixed(1)}`).join("|");
  if (sig === lastSig) return; lastSig = sig;
  actors.slice().sort((a, b) => a.depth - b.depth).forEach(a => layers[a.layer].appendChild(a.g));
}
```

`depth` is the front-most `x+y` of the row or actor. Leave at least one object
height of clear floor between rows so the layer is never ambiguous.

### 7. Status in the scene

- Tag: dark pill `#1F1E1D`, about 104x22, status dot, value right-aligned, thin
  leader from the object top. Lift the leader `1.2 + (n-1-index) * 1.0` units so
  rows further back step higher, and offset the pill about 16 px right of its leader.
- Show each status at least three ways: light bar, on-device readout (`ftext`),
  surface tint (danger `rgba(...,.16)`, warn `.12`), floor halo (danger `.34`
  pulsing, warn `.2`, ok `.08`), pill color, and `aria-label="<name>, <status>"`.
- Pulses are CSS classes toggled on status change (opacity/transform keyframes)
  with a static `prefers-reduced-motion` state. The engine listens to the media
  query's `change` and re-applies.

### 8. The frame function

```ts
function apply(t: number) {
  for (const o of objects) {
    const v = model.valueAt(o.id, t), st = model.statusOf(o, v);
    if (st !== o.st) applyStatus(o, st);                     // cached: write only on change
    const txt = format(v); if (txt !== o.txt) { o.txt = txt; o.readout.textContent = txt; o.pillText.textContent = txt; }
    setDoor(o, model.doorAngle(o.id, t));
  }
  actors.forEach((a, i) => { const p = poseAt(itins[i], t); place(a, p.x, p.y, p.walk, p.stride); a.layer = layerFor(p.x, p.y); a.depth = p.x + p.y; });
  reorder();
}
return { setTime: apply, setSelected, setHover, info: (id, t) => ({ ...}), destroy: () => { mq?.removeEventListener("change", onMq); svg.replaceChildren(); } };
```

No clock, no rAF, no `innerHTML` resets inside the engine. `info(id, t)` returns
everything a tooltip shows (status, value, range, min/max so far, minutes out of
range, task, meta) computed from the model.

### 9. Chart and Gantt helpers (pure, unit-testable)

```ts
// Precompute once per scenario: value per minute + running stats, so cards and tooltips are lookups.
function buildSeries(model, id, start, end, lo, hi) {
  const vals = [], pMin = [], pMax = [], pOut = []; let mn = Infinity, mx = -Infinity, out = 0;
  for (let t = start; t <= end; t++) { const v = model.valueAt(id, t); vals.push(v);
    mn = Math.min(mn, v); mx = Math.max(mx, v); if (v < lo || v > hi) out++; pMin.push(mn); pMax.push(mx); pOut.push(out); }
  return { vals, pMin, pMax, pOut };
}
const geometry = (start, end, [lo, hi], w, h, top = 6, bottom = h - 16) => ({
  x: (t) => (Math.min(end, Math.max(start, t)) - start) / (end - start || 1) * w,
  y: (v) => bottom - (Math.min(hi, Math.max(lo, v)) - lo) / (hi - lo) * (bottom - top), top, bottom });
const linePath = (s, g, start, step = 2) =>
  s.vals.filter((_, i) => i % step === 0).map((v, k) => `${k ? "L" : "M"}${g.x(start + k*step).toFixed(1)} ${g.y(v).toFixed(1)}`).join("");
function placeMarkers(markers, g, valueAt, gap = 15) {                    // fan close markers apart
  const out = [];
  for (const m of [...markers].sort((a, b) => a.at - b.at)) {
    const cx = g.x(m.at), ly = g.y(valueAt(m.at)), near = out.filter(p => Math.abs(p.cx - cx) < gap).length;
    const dir = ly < (g.top + g.bottom) / 2 ? 1 : -1;
    out.push({ m, cx, ly, cy: Math.max(g.top + 7, Math.min(g.bottom - 7, ly + dir * gap * Math.min(near, 4))) });
  }
  return out;
}
// Gantt: x(t) as above, rowY(i) = header + i*rowH; snap(t) = Math.round(t / 5) * 5;
const conflicts = (bars, b) => bars.some(o => o.id !== b.id && o.row === b.row && o.start < b.end && b.start < o.end);
```

Draw the chart in SVG: in-range band, near-limit bands, dashed limit lines, the
full-day line revealed up to the cursor with a `clipPath` whose width is
`g.x(simMin)`, a now-line, hour ticks every 2 hours, markers with a short leader
to the line. Bars in the Gantt: scheduled (outline), active (fill with progress),
complete (muted); invalid drags turn red and are not applied.

### 10. Quality bar (check before calling it done)

| Check | Pass when |
|---|---|
| Fit | viewBox 680 wide, scene fills it, tallest tag fully visible |
| Depth | painter's order holds at every storyline minute; no actor floats or shows through a wall |
| Richness | shelves, bins, and counters carry seeded clutter; floor has zones and tile lines; slab edge in brand color |
| Calm | 2-8 people, purposeful motion only, readable at the default rate |
| Status | each status shown 3+ ways, tags never overlap, colors semantic |
| Determinism | reload and scrub to the same minute: identical frame |
| Performance | world built once; a frame writes only changed attributes; about 30 fps repaint |
| Accessibility | objects focusable with Enter/Space, aria-labels carry status, reduced motion shows static emphasis; the SVG is `role="group"` when objects inside are focusable (`role="img"` hides them from assistive tech), `role="img"` only for a non-interactive scene |
| Proof | screenshot at each storyline minute in a real browser and compare with the spec |

## Sources

- `references/raw/nextjs-isometric-engine-details.md` - projection constants, box face order, sprite and shadow sizes, seeded clutter, itinerary and door math, tag stagger, redundant status, chart helpers.
- `references/raw/nextjs-isometric-port-notes.md` - engine contract, dirty-checked frames, reduced-motion handling.
- `references/raw/isometric-3d-models-skill.md` - the recipe whose stubbed helpers this concept completes.
