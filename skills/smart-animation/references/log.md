# Log

Append-only session log. Every session MUST append at least one entry.
Never edit old entries - they are the historical record. Most recent
entries appear at the bottom.

Entry shape:

```
[INGEST|QUERY|LINT <YYYY-MM-DD>] <one-line summary>
  <optional indented detail line(s)>
```

---

[INGEST 2026-07-23] Authored smart-animation skill.
  - Router SKILL.md dispatches by codebase language (html/css/js/ts/react) and by library (orbs/beams/metal/transitions).
  - 12 wiki concepts: motion/principles/{design,performance,accessibility}; lang/{css,html,js,ts,react}/animation; lib/{orbs,beams,metal,transitions}/usage.
  - 6 raw sources seeded from research: thinking-orbs, border-beam, metal-fx, transitions-dev, frontend-design-motion-principles, web-animation-standards.
  - 3 decisions recorded: beams=border-beam, fully standalone, native-first per language.

[LINT 2026-07-23] 12 wiki, 6 raw. Hard: 0, Soft: 7. Regenerated _index.md.

[INGEST 2026-07-24] Added Canvas UI (canvasui.dev) as a fifth library.
  - New raw: references/raw/canvas-ui.md; new wiki: lib/canvas-ui/usage.
  - SKILL.md router gained a Canvas UI routing line; description/tags updated (added webgl); version 0.1.0 -> 0.2.0.
  - Decision recorded: lib concept only (no Vue/Svelte language routing); flagged experimental html-in-canvas API + MIT+Commons-Clause license.

[LINT 2026-07-23] 13 wiki, 7 raw. Hard: 0, Soft: 8. Regenerated _index.md.

[INGEST 2026-07-24] Added Safari/SVG-workaround note to Canvas UI docs.
  - Documented that html-in-canvas is Chromium-only (Safari/WebKit + Firefox unsupported), matching observed reduced clarity in Safari.
  - Noted the SVG <foreignObject> -> drawImage cross-browser workaround and its Safari caveats; steer to WebGL components for Safari/production.
  - Updated raw/canvas-ui.md + lib/canvas-ui/usage.md (no version change; 0.2.0 still in PR #195).

[LINT 2026-07-23] 13 wiki, 7 raw. Hard: 0, Soft: 8. Regenerated _index.md.

[INGEST 2026-07-30] Added three libraries: Motion, React Bits, Beautiful UI (5 -> 8).
  - New raw: motion-react.md, react-bits.md, beautiful-ui.md.
  - New wiki: lib/motion-react/usage, lib/react-bits/usage, lib/beautiful-ui/usage.
  - SKILL.md: 3 router lines, description gained the AI/agent-state trigger, "four
    bundled libraries" -> eight (was already stale, omitted Canvas UI), tag agent-ui,
    version 0.2.0 -> 0.3.0.
  - Motion: documented as "only for the four things CSS can't do" (exit, layout,
    spring, interruption) with the verbatim bundle table (34kb vs 4.6kb LazyMotion).
  - Beautiful UI: page payload inspected directly; extracted all 17 component sources
    to audit imports (react + liveline only). Documented the three things a copied
    file leaves behind (tokens, 7 keyframes, global reduced-motion block) — the
    silent-failure mode where the component renders unstyled and still, with no error.
  - 3 decisions recorded: motion-react slug, Beautiful UI despite no license,
    React Bits dependency contract as headline.

[LINT 2026-07-30] 16 wiki, 10 raw. Hard: 0, Soft: 11. Regenerated _index.md.

[BOUNDARY 2026-07-30] Declared the smart-frontend-design boundary; version 0.3.0 -> 0.4.0.
  - description gained: this skill owns motion implementation; whole-interface
    design direction routes to smart-frontend-design.
  - raw/frontend-design-motion-principles.md gained a provenance header naming the
    shared upstream (claude-plugins-official/frontend-design) and stating that
    smart-frontend-design mirrors the same sentence shallowly by design.
  - 2 decisions recorded: reversal of the 2026-07-23 "fully standalone" call, and
    holding scope at motion-only with Beautiful UI as the canary for a future split.

[LINT 2026-07-30] 16 wiki, 10 raw. Hard: 0, Soft: 11. Regenerated _index.md.

[LINT 2026-07-30] 16 wiki, 10 raw. Hard: 0, Soft: 11. Regenerated _index.md.

[LINT 2026-07-30] 16 wiki, 10 raw. Hard: 0, Soft: 11. Regenerated _index.md.

[LINT 2026-07-30] 16 wiki, 10 raw. Hard: 0, Soft: 11. Regenerated _index.md.

[LINT 2026-08-12] 17 wiki, 11 raw. Hard: 0, Soft: 12. Regenerated _index.md.

[INGEST 2026-08-12] Added liquid-gooey (gooey.jakubantalik.com); version 0.4.0 -> 0.5.0.
  - raw/liquid-gooey.md: full README distillation from the npm package README in
    Jakubantalik/Libraries (packages/liquid-gooey), plus registry metadata
    (v0.1.0, MIT, zero deps, react/react-dom >=18 peers).
  - wiki/lib/gooey/usage.md: ninth bundled library. Headline is the two-layer
    architecture (SVG silhouette below carries goo + shadow, real DOM above stays
    unfiltered) — that is what separates it from the CSS `filter: blur+contrast`
    gooey hack, which blurs text and can't draw one shadow on a merged shape.
  - The knob users get wrong first: bridging is a ratio of `blur` to gap, not a
    mode. 8px apart barely bridges at blur 5, merges cleanly at blur 12.
  - Reduced motion is honored for component-driven transitions (collapse to
    instant snaps) — unlike metal-fx, which needs manual `paused` wiring.
  - SKILL.md description, When-to-Use count (eight -> nine) and library routing
    list updated so the skill can actually be selected for gooey work.

[LINT 2026-08-12] 17 wiki, 11 raw. Hard: 0, Soft: 12. Regenerated _index.md.

[LINT 2026-10-20] 21 wiki, 16 raw. Hard: 0, Soft: 12. Regenerated _index.md.

[INGEST 2026-10-20] Added the code-built video loop; version 0.5.0 -> 0.6.0.
  - raw: video-loop-brief.md (the three-stage loop + decisions taken with the user), hyperframes.md,
    earshot.md (incl. FFmpeg measurements of the earshot launch film), claude-remotion-skill.md,
    hyperframes-motion-reel-skill.md.
  - wiki/video/loop/: render-engine, reference-assets, sound, self-critique (new context `video`).
  - scripts: video-init.sh, video-loop.sh, dead-frames.sh (bash 3.2), video-gate.py (stdlib),
    video-sound.py (earshot), templates/video/ (2 spring scenes with data-cues, scenes.json).
  - Verified end to end in a scratch project: exit 2 (unscored) -> 3 (6/10) -> 0 (pass) ->
    2 after a scene edit (stale score cleared) -> 4 at the iteration cap; music mode caught a cut
    5.9 frames off the beat.
  - Sound rules: stacked lockup verified (a pop 50 ms after the hit failed the scene); the dirty-take
    rule is untested because every calm-preset take is clean.
  - 5 decisions recorded; 1 patterns entry with timings and thresholds.

[LINT 2026-10-20] 21 wiki, 16 raw. Hard: 0, Soft: 12. Regenerated _index.md.

[LINT 2026-10-20] 21 wiki, 16 raw. Hard: 0, Soft: 12. Regenerated _index.md.

[LINT 2026-10-20] 22 wiki, 17 raw. Hard: 0, Soft: 12. Regenerated _index.md.

[UPDATE 2026-10-20] Loop hardening after the gap review (same PR, still 0.6.0).
  - New: wiki/video/loop/plan.md, raw/hyperframes-planning.md, templates/video/PLAN.md,
    scripts/tests/test_video_gate.py (21 hermetic checks).
  - video-gate.py: reasoned rubric (0-2 + reason per line, sum = score), shared fingerprint
    (index.html + scenes.json "shared"), hold cap (max_hold_share 0.4), `final`, `plan`, `scenes`.
  - video-loop.sh: plan gate, earshot version warning, per-scene sound.png, `--deliver`.
  - Starter lockup shortened to 1.8 s (its hold covered 56%). Self-critique gains a critic-subagent prompt.
  - Real run verified: plan exit 1 -> honest 7/10 exit 3 -> pass -> deliver exit 0.

[LINT 2026-10-20] 22 wiki, 17 raw. Hard: 0, Soft: 12. Regenerated _index.md.

[LINT 2026-10-20] 23 wiki, 18 raw. Hard: 0, Soft: 12. Regenerated _index.md.

[TEST 2026-10-20] Gold-reference test: a narrated product film in a reference's style.
  - New: scripts/video-reference.sh, wiki/video/loop/reference-style.md, raw/reference-product-film-
    2026-10-20.md (anonymized), voice mode (say / hyperframes / files) + ducked bed in video-sound.py,
    cut_on (beat | free), check --strict by default, GSAP vendored by video-init.sh, 23-check test.
  - Run: reference measured in 20 s; 28 s six-scene film; contrast gate caught copied grays; three
    critic rounds took scenes 1/6 -> 5/6 and cuts 1/5 -> 3/5; exit 4 at the 5-render cap with
    s05-fix, s03->s04 and s04->s05 at 7/10. Escalated to the user as the skill prescribes.

[LINT 2026-10-20] 23 wiki, 18 raw. Hard: 0, Soft: 12. Regenerated _index.md.

[LINT 2026-10-20] 23 wiki, 19 raw. Hard: 0, Soft: 13. Regenerated _index.md.

[LINT 2026-10-20] 23 wiki, 19 raw. Hard: 0, Soft: 13. Regenerated _index.md.

[TEST 2026-10-20] Same film, cap 8, Chatterbox voice, motion.js. Units passing: 2 -> 4 -> 8 -> 8 -> 8 -> 9
  of 11. Exit 4 at 8/8 (s05-fix 7, s05->s06 7). Added: Chatterbox engine, templates/video/motion.js,
  plan questions for voice engine and iteration budget, cut/sound/critic lessons in the wiki.

[LINT 2026-10-20] 23 wiki, 19 raw. Hard: 0, Soft: 13. Regenerated _index.md.

[LINT 2026-11-23] 23 wiki, 19 raw. Hard: 0, Soft: 13. Regenerated _index.md.

[LINT 2026-11-23] 23 wiki, 19 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[UPDATE 2026-11-23] Lint clean-up, gate stall rule, isometric scenes (v0.7.0).
  - Renamed 14 wiki concepts to unique stems (`<category>-animation`, `<category>-usage`); sound.md
    cites raw/chatterbox.md. Lint 0 hard, 0 soft.
  - video-gate.py: per-iteration score history in .loop/scores.json and a `stall` line for a unit
    failing again at a score it already failed at (the exit-4 pattern of the cap-8 film); a PLAN.md edit resets it. 26-check test.
  - New raw/isometric-3d-models-skill.md (author Carter Stonecipher) and wiki
    scene/isometric/isometric-demo-scenes.md. SKILL.md routes isometric, 3D, Three.js and warehouse
    visualization requests to it.

[LINT 2026-11-23] 24 wiki, 20 raw. Hard: 0, Soft: 0. Regenerated _index.md.

[INGEST 2026-11-24] Merged develop into feat/smart-animation-isometric-scenes. Dropped this branch's duplicate raw copy and
  motion/isometric/demo-scenes.md (superseded by scene/isometric/isometric-demo-scenes.md). Kept and moved the two
  additions to scene/isometric with unique stems: isometric-build-kit.md (complete helper, sprite, layering, frame and
  chart/Gantt code plus a quality bar, from raw/nextjs-isometric-engine-details.md) and isometric-nextjs-components.md
  (engine / wrapper / board split, clock policy, linked chart and Gantt aids, from raw/nextjs-isometric-port-notes.md).

[LINT 2026-11-24] 26 wiki, 22 raw. Hard: 0, Soft: 0. Regenerated _index.md.
