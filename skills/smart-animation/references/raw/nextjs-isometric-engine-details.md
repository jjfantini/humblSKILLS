# Notes: implementation details of the isometric engines (Next.js ports)

Observed 2026-11-23 in the same Next.js demo repo as nextjs-isometric-port-notes.md. Generic facts only;
customer names, layouts, and copy left out.

- Projection: `P(x,y,z) = [ox + (x-y)*0.866*s, oy + (x+y)*0.5*s - z*s]`, viewBox 680 wide,
  `s = min(22, 640/((W+D)*0.866))`, `ox = (680-(W+D)*0.866*s)/2 + D*0.866*s`, `oy = tallest*s + ~40`,
  `vbH = oy + (W+D)*0.5*s + ~26`. Coordinates printed with `toFixed(1)`.
- `box()` draws the +x face (darkest), then the +y face (mid), then the top (lightest); stroke = fill,
  width 0.6, `stroke-linejoin: round`. `fy`/`fx` draw quads on +y / +x planes; `ftext` skews text with
  `matrix(0.866, 0.5, 0, 1, X, Y)`, weight 500.
- Sprites: part lists `[x,y,z,w,d,h,palette]`, sorted by `x+y+w/2+d/2` then `z`, drawn once into `<defs>`
  with the origin offset switched off; placed with `<use href>` inside a group carrying the floor shadow
  ellipse (`rx = r*s*1.2247`, `ry = r*s*0.7071`, `rgba(0,0,0,.12)`) and a hidden selection ellipse.
- Deterministic clutter: an LCG (`seed = (seed*1664525 + 1013904223) % 2**32`) fills shelves with small
  product quads at ~75-85% occupancy in 3-4 colors, so every build looks identical and rich.
- People: itineraries are segments `{t0, t1, walk, path, at, st, tag}`; `poseAt` binary-searches the segment
  and interpolates along an orthogonal path (`route(a,b)` through a corridor x); `along(path, d)` walks the
  polyline by distance.
- Doors swing about a hinge: free edge at `(hx + w*cos a, y + w*sin a)`; angle rounded to 0.1 deg and only
  rewritten when it changes.
- Status tags: dark pill (#1F1E1D) 104x22 with dot and value, thin leader from the object top; leader lift
  `1.2 + (n-1-index)*1.0` units so back units step higher; pill offset +16 px right of its leader.
- Redundant status: top light bar, on-device readout, glass tint (rgba danger .16 / warn .12), floor halo
  (danger .34 pulsing / warn .2 / ok .08), pill color, aria-label "<name>, <status>".
- Per-frame `apply(t)`: caches last status, text, and angle per object and writes only on change; then
  places people and re-orders layers.
- Reduced motion: engine listens to `matchMedia('(prefers-reduced-motion: reduce)')` change and re-applies.
- Chart helpers: `geometry()` maps minutes and values to px with clamping; `bands()` gives in-range, near-limit
  and limit lines; `linePath()` samples every 2 minutes and the component reveals it to the cursor with a
  clip; `placeMarkers()` fans markers within 15 px of each other toward the roomier side of the plot.
