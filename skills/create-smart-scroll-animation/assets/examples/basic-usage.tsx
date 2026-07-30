// =============================================================================
// basic-usage.tsx — minimal Next.js App Router integration
//
// Reference example. Not a template — no placeholders.
// Shows how to drop ScrollFrameCanvas into a page with zero brand chrome.
// =============================================================================

// app/page.tsx
import { ScrollFrameCanvas } from './_components/ScrollFrameCanvas';

export default function Page() {
  return (
    <main>
      <ScrollFrameCanvas />

      {/* Your other sections below */}
      <section style={{ padding: '8rem 1.5rem', textAlign: 'center' }}>
        <h2>About the product</h2>
        <p>Copy that lives under the scrolling hero.</p>
      </section>
    </main>
  );
}
