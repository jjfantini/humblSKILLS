---
name: smart-animation
description: >
  Performance-first frontend animation, routed by codebase (HTML, CSS, JS,
  TS, React) and by ready-made libraries (thinking-orbs, border-beam,
  metal-fx, liquid-gooey, transitions.dev, Canvas UI, Motion, React Bits,
  Beautiful UI), plus a code-built video loop (HyperFrames, earshot sound,
  self-critique) and animated isometric or Three.js scenes. Use when adding
  animations, page or route transitions, scroll reveals, hover
  micro-interactions, loading/thinking indicators, or agent UI states
  (streaming text, tool calls, approval gates); when
  the user says "animate this", "add motion", "make it feel premium", or asks
  which animation library to use; for "make a video", "launch video",
  "render to MP4" or sound on a video; or for "isometric", "3D scene",
  "Three.js", "warehouse visualization", "digital twin demo". Enforces
  compositor-only performance and a reduced-motion floor. For whole-interface
  design use smart-frontend-design. Do NOT use for non-frontend work or
  static visual design with no motion.
license: MIT
compatibility: Requires bash, POSIX utilities (awk, sed, find, grep), python3, writable filesystem at the skill target path. The video loop also needs Node 22+ (npx hyperframes), ffmpeg/ffprobe, Chrome, and for sound a Python with earshot installed (CPU torch, ~1.1 GB models on first listen); narration recommends Chatterbox TTS in its own Python 3.11 venv (ask the user before installing, ~2 GB), with macOS say as the fallback.
metadata:
  author: jjfantini
  version: "0.7.0"
  previous_names: ["use-smart-animation"]
  category: design
  tags: [animation, transitions, frontend, motion, css, react, webgl, performance, accessibility, agent-ui, video, hyperframes, earshot, isometric, threejs, humblskill]
  platforms: [claude-code, cursor, codex]
  preserve:
    - references/raw/
    - references/wiki/
    - references/decisions.md
    - references/log.md
    - references/patterns.md
---

# Smart Animation

Give any frontend the right motion for its stack: read the shared principles,
then load only the file for the language you're in and the library you're
reaching for. Motion is deliberate here — one orchestrated moment beats
scattered effects, everything animates on compositor-only properties, and every
snippet honors `prefers-reduced-motion`.

## Brain Protocol (read BEFORE animating anything)

1. `references/_index.md`    - what this skill knows (map)
2. `references/patterns.md`  - what worked, with numbers
3. `references/decisions.md` - past reasoning, don't repeat mistakes
4. `references/log.md`       - last 5 session entries
5. The relevant `references/wiki/<context>/<category>/` concepts (see routing below)

After completing work, UPDATE the brain:
- New user-provided material -> `references/raw/` (never renamed or edited)
- Distilled insights -> new/updated `references/wiki/<ctx>/<cat>/<concept>.md`,
  citing every raw file used in `sources:`
- Quantified results (fps, bundle size, LCP) -> `patterns.md`
- Non-obvious decisions -> `decisions.md`
- Session summary (always) -> append to `log.md`
- Run `scripts/lint.sh` to regenerate `_index.md` and verify structure

_Full spec: `references/_brain.md`._

## CCCCC Architecture

| Layer        | Role                                | Location                                       |
|--------------|-------------------------------------|------------------------------------------------|
| **Core**     | Root structure of the skill         | `SKILL.md`, `references/`, `scripts/`          |
| **Context**  | Top-level grouping                   | 1st segment under `references/wiki/` (`motion`, `lang`, `lib`, `video`, `scene`) |
| **Category** | Specific topic within a context      | 2nd segment (`principles`; `css`/`react`/…; `orbs`/`beams`/…; `loop`; `isometric`) |
| **Concept**  | One atomic idea per file             | Filename stem + required frontmatter field     |
| **Command**  | Deterministic executable script      | `scripts/lint.sh`, `scripts/video-*.{sh,py}`, `scripts/dead-frames.sh` |

## When to Use

- Adding or reviewing animations/transitions in a web frontend
- Choosing a motion approach for a specific language (HTML/CSS/JS/TS/React)
- Wiring one of the nine bundled libraries (orbs, border-beam, metal-fx, liquid-gooey, transitions.dev, Canvas UI, Motion, React Bits, Beautiful UI)
- Building the UI around an agent: thinking states, streaming answers, tool calls, human-in-the-loop approval
- Page/route transitions, scroll reveals, hover micro-interactions, loading states
- Fixing janky animation (dropped frames) or missing reduced-motion handling
- Making a short animation video in code (intro, launch film, explainer, logo sting) and rendering it to MP4
- An animated isometric or 3D scene of a place for a demo or visualization (warehouse, distribution
  center, pharmacy, factory, dock), or any Three.js work

## How to Use

**Always read first — the shared principles apply to every task:**
Read `references/wiki/motion/principles/design.md` (where/if to animate),
`references/wiki/motion/principles/performance.md` (compositor-only, 60fps),
and `references/wiki/motion/principles/accessibility.md` (reduced-motion floor + tokens).

**Then route by the codebase language:**

- **CSS-only / static site:** Read `references/wiki/lang/css/css-animation.md`
- **Plain HTML:** Read `references/wiki/lang/html/html-animation.md`
- **Vanilla JavaScript:** Read `references/wiki/lang/js/js-animation.md`
- **TypeScript (non-React):** Read `references/wiki/lang/ts/ts-animation.md`
- **React / Next.js:** Read `references/wiki/lang/react/react-animation.md`

**And route by the effect / library you need:**

- **Thinking / loading indicator for an AI or agent UI:** Read `references/wiki/lib/orbs/orbs-usage.md` (thinking-orbs)
- **Animated glow around a card, button, or input border:** Read `references/wiki/lib/beams/beams-usage.md` (border-beam)
- **Metallic / liquid-metal accent on one premium element:** Read `references/wiki/lib/metal/metal-usage.md` (metal-fx)
- **Elements that merge like drops of liquid (expanding FAB cluster, morphing card, slider thumb with a rubbery trail):** Read `references/wiki/lib/gooey/gooey-usage.md` (liquid-gooey)
- **Standard UI transitions (modals, toasts, tabs) in any stack:** Read `references/wiki/lib/transitions/transitions-usage.md` (transitions.dev)
- **Creative canvas/WebGL effects over content (particle reveal, liquid, glass, shatter) for a hero or signature moment:** Read `references/wiki/lib/canvas-ui/canvas-ui-usage.md` (Canvas UI)
- **Exit animations, layout/FLIP animation, spring physics, or interruptible orchestration in React:** Read `references/wiki/lib/motion-react/motion-react-usage.md` (Motion, formerly Framer Motion) — the four things CSS cannot do
- **A specific finished effect (split-text reveal, aurora background, magnet button) from a large catalog:** Read `references/wiki/lib/react-bits/react-bits-usage.md` (React Bits) — check the per-component dependency cost first
- **AI / agent interface states — thinking traces, streaming text, tool chips, approval gates, confidence, context cards, diffs:** Read `references/wiki/lib/beautiful-ui/beautiful-ui-usage.md` (Beautiful UI)

**And route scene work (an animated place, live in the browser):**

- **Isometric or 3D scene, warehouse or site visualization, digital-twin demo, any Three.js work:**
  Read `references/wiki/scene/isometric/isometric-demo-scenes.md` (SVG first, Three.js when the camera
  must move; sim clock, data triggers, SCENE API, headless screenshots at checkpoints).
- **Writing the scene code** (projection, box/fy/fx/ftext, sprites, itineraries, actor layering,
  dirty-checked frames, chart and Gantt helpers) and the quality bar a finished scene must meet:
  Read `references/wiki/scene/isometric/isometric-build-kit.md`.
- **The scene inside a Next.js / React app** (imperative engine, client wrapper, a board that owns the
  sim clock, linked chart, Gantt, stat row, tooltips): Read
  `references/wiki/scene/isometric/isometric-nextjs-components.md`.

**And route video work (a rendered MP4, not live UI motion):**

- **Make / fix an animation video (plan, render loop, scoring, sound):** Read the
  `references/wiki/video/loop/` concepts in order: `plan.md` (interview, context, research, beat sheet,
  approval - before any scene), `reference-style.md` (when the user gives a "make it like this" video:
  measure it, borrow its grammar, copy nothing brand-specific), `render-engine.md` (HyperFrames, one command, delivery and review),
  `reference-assets.md` (real assets, closed-form springs, banned defaults), `sound.md` (earshot kits
  on picture cues; effects-only default, narrated voice mode with a ducked bed, music mode) and `self-critique.md` (reasoned rubric, adjacent
  pairs, critic subagent, mechanical rules, exit codes). Install HyperFrames' own skills
  (`npx hyperframes skills update`) for the composition contract.

**Live enumeration of everything this skill knows:**
Read `references/_index.md` (auto-regenerated by `scripts/lint.sh`).

**Brain protocol, naming, linking contract, lint checks, `patterns.md` shape:**
Read `references/_brain.md`. **Wiki concept file shape:** Read `references/_template.md`.

### Scripts

- `scripts/lint.sh` — brain health check: regenerates `_index.md`, validates
  every wiki file's path/frontmatter triple and that each cites a real
  `references/raw/` source. Run after editing any wiki concept.
- `scripts/video-init.sh <dir>` — `hyperframes init` plus two starter scenes (springs, `data-cues`),
  `scenes.json`, `PLAN.md`, `assets/SOURCES.md`, `critique/`.
- `scripts/video-loop.sh [--deliver] [dir]` — one loop pass: plan check -> timeline -> cues -> sound ->
  check -> draft render -> dead frames, earshot listening, frames for changed scenes and cuts -> gate.
  Exit 0 pass, 1 plan not ready or tool error, 2 waiting for scores, 3 failing, 4 iteration cap.
  `--deliver` then renders the final, re-checks it, and writes a poster and review frames. Set `EARSHOT_PYTHON`.
- `scripts/tests/test_video_gate.py` — hermetic gate test (26 checks, no HyperFrames or earshot).
- `scripts/dead-frames.sh <mp4>` — freeze/black stretches from the encoded pixels (TSV).
- `scripts/video-reference.sh <ref.mp4> <dir>` — measure a reference video (cuts, palette and small
  accents, loudness, earshot sound summary, motion strips) into a STYLE.md to fill.
- `scripts/video-sound.py`, `scripts/video-gate.py` — the sound build/listen and the gate, called by
  `video-loop.sh`.

## Examples

### Example 1: "Add a page-load animation to my React landing page"

1. Read the three `motion/principles/*` concepts (deliberate, performant, accessible).
2. Read `lang/react/react-animation.md` — decision rule says use CSS for the entrance,
   or Motion only if you need staggered orchestration/exit.
3. Implement ONE orchestrated entrance (staggered children), compositor-only
   (`transform`/`opacity`), with a `prefers-reduced-motion` fallback.
4. If they want a premium accent on the CTA, read `lib/metal/metal-usage.md` and gate
   the shader with `paused` under reduced motion.
5. Append a `log.md` entry; if you measured LCP/fps, add a `patterns.md` entry.

### Example 2: "Make my modals and toasts feel polished in a plain HTML/CSS site"

1. Read the `motion/principles/*` concepts.
2. Read `lib/transitions/transitions-usage.md` — copy the `t-*` pattern (open + closing
   states, motion tokens) or install the upstream transitions.dev skill.
3. Read `lang/html/html-animation.md` — prefer `<dialog>`/Popover so the motion hangs
   on accessible native elements.
4. Verify each transition keeps its `@media (prefers-reduced-motion: reduce)` block.

### Example 3: "Make a 15-second launch video for my CLI, with sound"

1. Read the five `video/loop/*` concepts; `bash scripts/video-init.sh launch`.
2. Work through `launch/PLAN.md` with the user one question at a time; write `approved:` only after
   they confirm the beat sheet.
3. Put real assets in `launch/assets/` and list them in `SOURCES.md`; write one scene per beat of
   the story as a sub-composition, its sound events in `data-cues`, motion on `springEase`.
4. `EARSHOT_PYTHON=... bash scripts/video-loop.sh launch` until it exits 2; hand each stub to a critic
   subagent (prompt in `self-critique.md`), apply the fix notes, rerun. Exit 3 means fixes to apply,
   unless the report prints a `stall` line: then rewrite that unit's PLAN row with the user first.
5. At exit 0, run `video-loop.sh --deliver launch`, look through `critique/final/frames`, then ask the
   user to watch `renders/final.mp4` with sound. At exit 4, stop and show the user what still fails.

### Example 4: "Make a video like this one" (a narrated product film, reference attached)

1. `bash scripts/video-init.sh film`, then `bash scripts/video-reference.sh ref.mp4 film/reference`
   and the smart-watch skill on `ref.mp4` for the transcript; fill the style sheet in your own words.
2. Plan with the user: the message, the story spine, one narration line per scene. Measure each line
   (`say -o` + ffprobe) and size scenes as line + 0.5-1 s.
3. `scenes.json`: `"audio": "voice"`, `"cut_on": "free"`, a `voice` engine (ask the user to install
   Chatterbox for an expressive voice; `say` is the flat fallback), a `bed`, `max_iterations` 8. Each scene's
   `data-cues` holds its line (`kind: vo`) and its UI ticks; words resolve with the voice.
4. Loop, critics, `--deliver` as in Example 3. Keep the reference's grammar, not its brand: its
   grays may fail contrast, its master may clip - the gate stays at AA and -1 dBTP.

## Troubleshooting

**`video-loop.sh` exits 1 "earshot not importable":** install earshot into a venv (CPU torch) and
point `EARSHOT_PYTHON` at its python, or set `"audio": "none"` in `scenes.json`.

**A scene fails "freeze ... (unplanned)" on an intentional rest:** declare it in that scene's
`holds` in `scenes.json` (scene-local seconds, `"to": "end"` allowed, at most 40% of the scene).
Do not raise the threshold.

**`stall <unit>: 7 -> 8 -> 7`:** the unit failed again at a score it already failed at. Fresh critics
are pulling it in different directions. Stop editing it, show the user its fix notes, rewrite its PLAN
row ("A, then B"), and rescore against the new row. Do not spend more renders on it.

**"score ...: not counted":** every rubric line needs `<0-2> - <reason>` and the lines must sum to the score.

**Render shows static frames after a long wait:** a host `data-composition-id` does not match the
sub-composition's inner id and timeline key. They must be identical.

**Animation is janky / drops frames:** you're likely animating a layout
property. Read `motion/principles/performance.md` — move to `transform`/`opacity`
or FLIP.

**`scripts/lint.sh` exits 1 with "path/frontmatter mismatch":** a wiki file's
`context/category/concept` doesn't match its path. Rename the file or fix the
frontmatter, then re-run.

**`scripts/lint.sh` exits 1 with "broken source path":** a wiki concept's
`sources:` points at a missing `references/raw/` file. Fix the path or the raw file.

**Skill never triggers:** the `description:` lacks the phrases a user would type.
Ask "When would you use smart-animation?" and adjust until it matches.

## Success Signals

- `scripts/lint.sh` exits 0 after every wiki change (all triples valid, all sources resolve).
- Every animation uses `transform`/`opacity` (no layout-property tweens).
- Every motion-producing snippet has a `prefers-reduced-motion` fallback.
- Motion is concentrated on one signature moment, not scattered across the view.
- The right file loads for the stack: React work reads `lang/react`, a modal reads `lib/transitions`, etc.
- A video ships only after `video-loop.sh` exits 0 and the user approved the preview.
