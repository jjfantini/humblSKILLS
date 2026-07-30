# React Bits (reactbits.dev)

Source: https://reactbits.dev and https://github.com/DavidHDev/react-bits
Author: David Haz (DavidHDev). License: **MIT + Commons Clause**.
Researched 2026-07-30.

## What it is
"The largest & most creative library of animated React components" — **140+**
free, customizable animated components across text effects, animated
backgrounds, UI elements, and shapes. New components ship weekly. Official
ports exist for **Vue** and **Svelte**.

This is a *catalog of set-piece effects*, not an animation engine. You reach for
it when you want a specific finished effect (split text reveal, aurora
background, magnet button) rather than a primitive to build with.

## Install
Copy-paste, jsrepo, or the shadcn CLI. The shadcn form encodes the variant in
the component name:

```
npx shadcn@latest add @react-bits/BlurText-TS-TW
```

You own the resulting source file — there is no runtime npm dependency on
React Bits itself to keep updated.

## The four variants
Every component ships in all four combinations, selected at install time:

1. JavaScript + CSS
2. JavaScript + Tailwind
3. TypeScript + CSS
4. TypeScript + Tailwind

Suffix convention in the shadcn name: `-JS-CSS`, `-JS-TW`, `-TS-CSS`, `-TS-TW`.

## Categories
- **Text animations** — blur text, split text, shiny text, decrypted text,
  scroll-driven reveals, variable proximity
- **Animated backgrounds** — aurora, particles, waves, grids, orbs, shaders
- **Components / UI** — carousels, docks, magnet buttons, spotlight cards,
  stacks, tilted cards
- **Shapes & effects** — blobs, ribbons, distortion, cursor-following effects

Also ships free creative tools: Background Studio, Shape Magic, Texture Lab.

## Dependencies — the part that matters
The site advertises "minimal dependencies … lightweight and tree-shakeable,"
but that is per-component, not global. The repo's own `package.json` pulls:

| Package | Version |
| --- | --- |
| gsap | ^3.13.0 |
| motion | ^12.23.12 |
| three | ^0.180.0 |
| ogl | ^1.0.11 |
| matter-js | ^0.20.0 |
| @react-three/fiber | ^9.3.0 |
| @react-three/drei | ^10.7.4 |
| react / react-dom | ^19.0.0 |

So the true cost is **per component**: a text effect may pull `gsap` or
`motion`; a 3D background pulls `three` + `@react-three/fiber` + `drei`, which
is a very different bundle conversation. Check the imports at the top of the
copied file before committing to it — that file is the dependency contract.

Requires **React 19**.

## License caveat
`MIT + Commons Clause` — free for personal and commercial use in your own
product, but the Commons Clause forbids *selling the components themselves*
(i.e. you cannot repackage React Bits as a paid component library). Same
license posture as Canvas UI, already flagged elsewhere in this skill. Note
also that the repo `package.json` is marked private with no `license` field;
the license is stated on the site and in the repo LICENSE file.

## Accessibility
No documented `prefers-reduced-motion` handling in the components themselves.
Scroll-driven text reveals and particle/shader backgrounds are exactly the
category that triggers vestibular discomfort, so the reduced-motion gate is
**the consumer's job** after copying the source in. Since you own the file,
add the guard directly rather than wrapping it.
