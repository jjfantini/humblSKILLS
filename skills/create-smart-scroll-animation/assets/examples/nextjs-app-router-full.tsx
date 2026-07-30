// =============================================================================
// nextjs-app-router-full.tsx — full Next.js App Router integration
//
// Reference example. Not a template — no placeholders.
// Shows:
//   - Lenis smooth scroll wrapped at the root layout
//   - ScrollFrameCanvas with brand colors wired via CSS custom properties
//   - Edge cache headers for /frames/* (configured in next.config.js)
//
// Three files in one reference example — split them into the paths below
// when integrating.
// =============================================================================

// -----------------------------------------------------------------------------
// app/layout.tsx
// -----------------------------------------------------------------------------
'use client';

import { ReactLenis } from 'lenis/react';
import type { ReactNode } from 'react';
import './globals.css';

export default function RootLayout({ children }: { children: ReactNode }) {
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

// -----------------------------------------------------------------------------
// app/globals.css (excerpt — add to your existing globals.css)
// -----------------------------------------------------------------------------
/*
:root {
  --brand-bg: #0A1A2F;
  --brand-accent: #D4AF37;
}

body {
  background-color: var(--brand-bg);
  color: #fff;
  margin: 0;
}
*/

// -----------------------------------------------------------------------------
// app/page.tsx
// -----------------------------------------------------------------------------
// Note: remove `'use client'` from this file — the page stays a server
// component. ScrollFrameCanvas is already marked 'use client'.
/*
import { ScrollFrameCanvas } from './_components/ScrollFrameCanvas';

export default function Page() {
  return (
    <main>
      <ScrollFrameCanvas />
      <section style={{ padding: '8rem 1.5rem', textAlign: 'center' }}>
        <h2>About the product</h2>
        <p>Copy that lives under the scrolling hero.</p>
      </section>
    </main>
  );
}
*/

// -----------------------------------------------------------------------------
// next.config.js (excerpt — frame cache headers)
// -----------------------------------------------------------------------------
/*
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
*/

// -----------------------------------------------------------------------------
// package.json diff
// -----------------------------------------------------------------------------
/*
"dependencies": {
  "framer-motion": "^11.3.0",
  "lenis": "^1.1.13",          // only if using Lenis
  ...
}
*/
