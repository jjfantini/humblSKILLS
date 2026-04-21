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

### 2026-04-20 | Single-file React component output (no hook extraction)
- Context: generated component has meaningful sub-concerns (preload state machine, fit math, reduced-motion branch, DPR canvas sizing). Could split into `useScrollFrame` + `useFramePreloader` + component files.
- Options: (A) single `.tsx` file, hooks inline; (B) component + 2 hook files + types file; (C) publish as `@humblskills/scroll-frame-canvas` npm package.
- Chose: A — single file, everything inline.
- Why: $100k production teams hate skills that vomit 6 files into their codebase. One file is inspectable, auditable, and trivially deletable if the user regrets it. Hook extraction is a future signal — if users ask for `useScrollFrame` as a standalone export across 5+ sessions, revisit. Logged as a `patterns.md` trigger.
- Result: `assets/templates/ScrollFrameCanvas.tsx.tmpl` is one file, ~280 lines with all conditional blocks inline. Framer Motion is the only new dep added to the user's project.

### 2026-04-20 | Two-phase interview (technical mandatory, aesthetic optional)
- Context: reference skill `scroll-stop-builder` conflates technical inputs (video path, FFmpeg) and aesthetic inputs (brand, colors, vibe, logo, copy) in one mandatory questionnaire. For a component-generator (not a full-site-builder) this is wrong.
- Options: (A) keep a single phase matching the reference; (B) split into Phase 1 technical (always) + Phase 2 aesthetic (conditional); (C) drop the aesthetic side entirely.
- Chose: B — split into two phases.
- Why: most users want the canvas primitive with neutral styling (`#000` / `#fff`). Asking for logo/copy/vibe when they've asked for "just the scroll-scrubbed hero" wastes turns and confuses scope. Phase 2 runs only when the user signals intent (mentions colors, asks for "styled" output, or project has design tokens visible). Two questions max in Phase 2 (bg + accent) — no logo, no copy.
- Result: `workflow/interview/phase-order.md` documents the split. Template's `BRAND_BG` and `BRAND_ACCENT` placeholders default to `#000`/`#fff` when Phase 2 is skipped. Scope is kept tight — no logo rendering, no copy generation.

### 2026-04-20 | Defer hosting adapters (local directory only in v0.1.0)
- Context: user plans to host frames on Vercel Blob, Supabase Storage, or UploadThing. Skill could ship adapters for all three now, or defer and accept a URL pattern prop.
- Options: (A) defer — local directory + URL-pattern prop; (B) ship all three hosting adapters now; (C) ship only Vercel + Supabase (user's default stack).
- Chose: A — defer.
- Why: user explicitly said "for now, lets just focus on saving to a directory." Keeping v0.1.0 scope tight. The generated component accepts a `frameUrlPattern` prop so the user can swap to any host later with zero code change. Upload flows belong in future dedicated skills (one per host) that handle their own auth, CORS, and API specifics.
- Result: `nextjs/integration/asset-pipeline.md` documents the URL-pattern contract with examples for all three future hosts. Default path is `public/frames/` — works on Vercel's built-in static asset serving with edge caching.

### 2026-04-20 | Drop white-first-frame requirement
- Context: reference skill hard-requires the video's first frame to be a predominantly white background. Originally planned as a gate via `scripts/check-first-frame.sh`.
- Options: (A) keep the requirement and the validator script; (B) make it advisory (warn but don't halt); (C) drop entirely.
- Chose: C — drop entirely.
- Why: user request during plan-mode review. The requirement exists in the reference skill for UI layering reasons specific to that skill's annotation cards and logo overlay — features we're not porting. Without those features, the first-frame constraint is aesthetic overhead with no functional purpose. Dropping removes a script, a concept, and a pipeline halt that would only frustrate users.
- Result: no `check-first-frame.sh`, no `video/extract/white-first-frame.md` concept, no pipeline halt. `references/raw/scroll-stop-builder-SKILL.md` notes the drop explicitly so future ingests don't re-introduce the requirement.

### 2026-04-20 | Framer Motion over GSAP ScrollTrigger
- Context: choose a scroll-progress library for the generated component. Both Framer Motion's `useScroll` and GSAP ScrollTrigger are production-grade.
- Options: (A) Framer Motion v11 (`useScroll` + `useMotionValueEvent`); (B) GSAP + `@gsap/react`; (C) no library, raw `scroll` listener + rAF.
- Chose: A — Framer Motion.
- Why: lighter (~42KB vs GSAP's ~80KB), native React hooks (`useMotionValueEvent` is specifically designed to avoid re-renders on motion value changes — the critical perf property for canvas draw), simpler API for our narrow use case. GSAP wins for complex multi-timeline scrubs with pinning/snapping, which we don't need. Raw scroll listener is tempting but means we own forced-layout-avoidance and rAF scheduling — Framer gets those right by default.
- Result: generated component imports `framer-motion`. Only new dep. Lenis is an optional opt-in for smooth-scroll polish, not a default.
