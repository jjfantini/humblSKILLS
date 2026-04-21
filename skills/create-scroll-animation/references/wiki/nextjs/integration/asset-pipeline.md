---
title: "Frame Assets: public/frames/ Convention and URL-Pattern Prop"
context: nextjs
category: integration
concept: asset-pipeline
description: "Default asset location is public/frames/frame_NNNN.webp — served as static assets by Next.js, cacheable at the edge. Component accepts a frameUrlPattern prop so later migrations to Vercel Blob / Supabase Storage / UploadThing require zero code changes."
tags: nextjs, public, assets, frames, CDN, hosting, url-pattern
sources: []
last_ingested: 2026-04-20
---

## Asset Pipeline

The v0.1.0 default is the simplest thing that works: frames in
`public/frames/`, served as static assets by Next.js, cached at the edge
by Vercel's CDN. No API route, no database, no upload flow.

**Default layout:**

```
public/
└── frames/
    ├── frame_0001.webp
    ├── frame_0002.webp
    ├── ...
    └── frame_0100.webp
```

**Default URL pattern (baked into the generated component):**

```
/frames/frame_{index}.webp
```

Where `{index}` is a zero-padded integer (default 4 digits: `0001`,
`0002`, ..., `0100`).

## URL-Pattern Prop Contract

The component accepts a `frameUrlPattern` prop that overrides the
default. Any URL with a `{index}` placeholder works. This is the forward
path to external hosting:

```tsx
// Default — public/frames/:
<ScrollFrameCanvas />

// Later — Vercel Blob:
<ScrollFrameCanvas
  frameUrlPattern="https://xyz.public.blob.vercel-storage.com/frames/frame_{index}.webp"
/>

// Later — Supabase Storage:
<ScrollFrameCanvas
  frameUrlPattern="https://abc.supabase.co/storage/v1/object/public/frames/frame_{index}.webp"
/>

// Later — UploadThing:
<ScrollFrameCanvas
  frameUrlPattern="https://utfs.io/f/<file-id>-frame_{index}.webp"
/>
```

A future skill can manage the upload + return the correct pattern
string. For v0.1.0 of `create-scroll-animation`, the user lives with
`public/frames/` and gets a working hero.

## Why not `next/image` for frames

The component fetches frames into `new Image()` to paint onto a canvas.
`next/image` is designed for `<img>` rendering with layout, srcset,
placeholder, priority, etc. None of that applies to canvas image sources.
Using `next/image` here would:

- Add DOM `<img>` elements we'd then have to read back from to draw
  (extra indirection)
- Pay for Next.js's on-the-fly image optimization when we've already
  pre-optimized to WebP at build time
- Complicate the preload contract (we want direct control of decode
  order, priority hints)

**Rule:** canvas-consumed images bypass `next/image`. Rendered `<img>`
elements use `next/image`. The reduced-motion fallback uses a plain
`<img>` (because it's rendered, not canvas-drawn) — and even there,
`next/image` is overkill since the WebP is already edge-cached and
appropriately sized.

## `next.config.js` for remote URLs (forward-looking)

Not needed for the default (`/frames/` is same-origin). If you later
point the component at a remote host, the component doesn't load images
through `next/image`, so `next.config.js remotePatterns` config isn't
required either. It's only required for `<Image />` components.

## Cache headers

Vercel serves `/public/*` with `Cache-Control: public, max-age=0,
must-revalidate` by default. For hero frames that should never change,
add a custom cache header via `next.config.js`:

```js
// next.config.js
module.exports = {
  async headers() {
    return [
      {
        source: '/frames/:path*',
        headers: [
          { key: 'Cache-Control', value: 'public, max-age=31536000, immutable' },
        ],
      },
    ];
  },
};
```

Frames are content-hashed by URL (filename includes a version in
practice, or a commit hash) — `immutable` is safe as long as you bust
the cache by changing the path when frames change.

**If you don't hash frame URLs:** drop `immutable`, keep `max-age=31536000`.
Return visits still hit cache; forced re-deploys invalidate via Vercel's
deployment-scoped cache layer.

## Sources

Synthesis concept. Next.js static asset behavior, cache headers, and
image-optimization semantics are standard documented behaviors.
