---
title: "TypeScript Animation: Typed Motion Primitives"
context: lang
category: ts
concept: animation
description: "Typing the Web Animations API, typed motion tokens, and a typed reduced-motion helper so animation code stays safe and self-documenting."
tags: typescript, waapi-types, motion-tokens, reduced-motion, type-safety
sources:
  - "references/raw/web-animation-standards.md"
  - "references/raw/transitions-dev.md"
last_ingested: 2026-07-23
---

## TypeScript Animation: Typed Motion Primitives

Everything in `lang/js/animation` applies; TypeScript adds type safety. The wins
are: correct WAAPI types (no `as any`), a `const`-typed token union so durations
and easings are checked at compile time, and a typed reduced-motion helper.

**Incorrect (untyped, stringly-typed, easy to typo):**

```ts
function reveal(el: any) {                        // loses all safety
  el.animate([{ opacity: 0 }, { opacity: 1 }], { duration: "300" }); // wrong type
}
const D = { fats: 250 };                           // typo passes silently
```

**Correct (typed keyframes, typed tokens, typed guard):**

```ts
// Motion tokens as a const object → keys become a checked union.
export const DURATION = { micro: 80, quick: 150, fast: 250, medium: 350 } as const;
export const EASE = { smoothOut: 'cubic-bezier(0.22,1,0.36,1)', linear: 'linear' } as const;
type Duration = keyof typeof DURATION;

export const prefersReducedMotion = (): boolean =>
  typeof matchMedia === 'function' &&
  matchMedia('(prefers-reduced-motion: reduce)').matches;

export function reveal(el: HTMLElement, d: Duration = 'medium'): Animation {
  const frames: Keyframe[] = [
    { opacity: 0, transform: 'translateY(24px)' },
    { opacity: 1, transform: 'translateY(0)' },
  ];
  const opts: KeyframeAnimationOptions = {
    duration: prefersReducedMotion() ? 0 : DURATION[d],
    easing: EASE.smoothOut,
    fill: 'both',
  };
  return el.animate(frames, opts);
}
```

Notes:
- WAAPI types ship with `lib.dom.d.ts`: `Keyframe[]` /
  `PropertyIndexedKeyframes` for frames, `KeyframeAnimationOptions` for options,
  `Animation` for the return. `matchMedia` returns `MediaQueryList`.
- Type motion tokens `as const` and derive unions from them — a mistyped
  duration key becomes a compile error, not a silent no-op.
- For libraries: Motion (`motion`) and GSAP both ship first-class `.d.ts`; import
  their prop/config types (e.g. Motion's `Variants`, `Transition`) rather than
  re-declaring shapes.

## Sources

- `references/raw/web-animation-standards.md` — WAAPI type names and the
  reduced-motion query.
- `references/raw/transitions-dev.md` — the motion-token values encoded as a
  typed const.
