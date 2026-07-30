---
title: "Next.js App Router Setup: Client Boundary and Lenis"
context: nextjs
category: integration
concept: app-router-setup
description: "Where ScrollFrameCanvas lives in an App Router project. Covers the 'use client' directive requirement, server/client boundary placement, and the optional Lenis smooth-scroll wrapper in layout.tsx. Hydration-safe by construction."
tags: nextjs, app-router, use-client, hydration, lenis
sources: []
last_ingested: 2026-04-20
---

## App Router Integration

The generated component is a **client component** (it holds refs, uses
`window.matchMedia`, and subscribes to scroll). It cannot be rendered on
the server.

**Where it goes:**

```
app/
├── layout.tsx           # root layout (server component by default)
├── page.tsx             # landing page
└── _components/
    └── ScrollFrameCanvas.tsx   # the generated file — 'use client' at top
```

**Importing from a server page:**

```tsx
// app/page.tsx — this stays a server component
import { ScrollFrameCanvas } from './_components/ScrollFrameCanvas';

export default function Page() {
  return (
    <main>
      <ScrollFrameCanvas />
      {/* other server-rendered sections here */}
    </main>
  );
}
```

Server components can import and render client components. The reverse
isn't true (client components can't import server components as
children), but we don't need the reverse direction here.

**Incorrect — putting `'use client'` on the page itself:**

```tsx
'use client';
// BAD: converts the entire page to a client component
// for no reason, losing streaming SSR + RSC benefits.
import { ScrollFrameCanvas } from './_components/ScrollFrameCanvas';
export default function Page() { return <ScrollFrameCanvas />; }
```

**Correct — keep the page server-rendered, just import the client component:**

```tsx
// No directive. Server component.
import { ScrollFrameCanvas } from './_components/ScrollFrameCanvas';
export default function Page() {
  return <main><ScrollFrameCanvas /></main>;
}
```

The `'use client'` directive already lives at the top of
`ScrollFrameCanvas.tsx` — that's the boundary. Everything upstream stays
on the server.

## Optional: Lenis smooth scroll

Lenis makes native scroll feel like `css-snap` + momentum. Works
alongside Framer Motion's `useScroll` (Lenis syncs with the browser's
scroll, which is what `useScroll` observes).

**Wire it into the root layout:**

```tsx
// app/layout.tsx
'use client';

import { ReactLenis } from 'lenis/react';

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <ReactLenis root options={{ lerp: 0.1, smoothWheel: true }}>
          {children}
        </ReactLenis>
      </body>
    </html>
  );
}
```

**Default:** don't include Lenis. Add it only if the user explicitly
asks for smooth scroll. One fewer dependency, one fewer scroll-handler
layer, one fewer thing to break on iOS.

**Install:** `npm i lenis`

## Hydration safety

- `useMatchMedia` and `useScroll` are wrapped in `useEffect` in the
  generated component. They never run during SSR.
- Initial state values (`prefersReducedMotion: false`, `ready: false`)
  are deterministic and match between server + client render.
- Canvas is inert until the first `useEffect` runs post-hydration.

You should see zero hydration warnings in the DevTools console after
integration. If you see one, check that no consumer is wrapping
`ScrollFrameCanvas` in a `dynamic(() => ..., { ssr: false })` — that's
unnecessary and can cause its own hydration issues.

## Route segment config

The page hosting this component should stay **static**:

```tsx
// app/page.tsx
export const dynamic = 'force-static'; // optional — default is already static
```

Frames live in `/public/frames/` → CDN-cacheable forever. No SSR needed.
No ISR needed. Edge cache does the work.

## Sources

Synthesis concept. The App Router client/server boundary rules are
standard Next.js 14+ behavior documented in the Next.js docs.
