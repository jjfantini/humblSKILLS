---
title: "Connection-Aware Frame Count and Preload Tuning"
context: mobile
category: network
concept: connection-aware
description: "Use navigator.connection.effectiveType to pick sensible defaults: 4g gets the full frame set, 3g gets a reduced set, slow-2g/2g falls back to a static image. Avoids blowing mobile data budgets and shipping 8MB of frames to a phone on a subway."
tags: mobile, network, connection, effectiveType, adaptive, bandwidth
sources: []
last_ingested: 2026-04-20
---

## The Navigator Connection API

`navigator.connection` is implemented in Chrome, Edge, Samsung Internet,
and Chrome-on-Android. **It is not implemented in Safari or Firefox.**

So it's useful on Android (covers most mobile web users worldwide) but
not a universal signal.

**Fields we care about:**

| Field            | Values                                                   |
|------------------|----------------------------------------------------------|
| `effectiveType`  | `'slow-2g'`, `'2g'`, `'3g'`, `'4g'`                       |
| `saveData`       | `true` if the user has Data Saver / Lite Mode enabled    |
| `downlink`       | Mbps estimate (approximate)                              |

## Adaptive Strategy

```tsx
function pickFrameBudget(): { count: number; staticFallback: boolean } {
  const conn = (navigator as any).connection;
  const eff = conn?.effectiveType ?? '4g';  // default to full experience if unknown
  const saveData = conn?.saveData === true;

  if (saveData) return { count: 0, staticFallback: true };

  switch (eff) {
    case 'slow-2g':
    case '2g':
      return { count: 0, staticFallback: true };
    case '3g':
      return { count: 60, staticFallback: false };
    case '4g':
    default:
      return { count: 100, staticFallback: false };
  }
}
```

**Rules:**

1. **`saveData: true` respects the user.** If they're telling the
   browser to conserve data, serve the static image. Full stop.
2. **Unknown effectiveType → full experience.** Safari iOS won't
   expose this; we can't penalize those users. Fall back to the
   memory-based mobile mitigations from `mobile/memory/ios-safari-cap.md`
   instead.
3. **2g gets the static fallback**, not a tiny frame set. A
   scroll-scrubbed video on 2g is broken no matter how few frames.
4. **3g gets 60 frames** (down from 100). Noticeably choppier but the
   effect still reads.
5. **4g / unknown gets the full 100.**

## Picking the right asset set

The frame count isn't just a number — you need multiple asset sets
extracted up front. Extract three variants during the pipeline:

| Variant       | Count | Resolution | URL prefix              |
|---------------|-------|------------|-------------------------|
| desktop/4g    | 100   | 1920×1080  | `/frames/`              |
| mobile/3g     | 60    | 1280×720   | `/frames-mobile/`       |
| static        | 1     | any        | `/frames-static/frame_0001.webp` |

Wire the component:

```tsx
const { count, staticFallback } = pickFrameBudget();
const isMobileViewport = window.matchMedia('(max-width: 768px)').matches;

if (staticFallback) {
  return <img src="/frames-static/frame_0001.webp" alt="" aria-hidden="true" />;
}

const urlPattern = isMobileViewport
  ? '/frames-mobile/frame_{index}.webp'
  : '/frames/frame_{index}.webp';

return <ScrollFrameCanvas frameCount={count} frameUrlPattern={urlPattern} />;
```

## Default for v0.1.0

**Ship one desktop set only.** Multiple asset pipelines are
operationally expensive and most users aren't blocked by network budget.
The mobile resolution downshift (separate concept) is more important.

Add connection-aware switching as a **v0.2.0 follow-up** when you have
real analytics showing users on slow networks. Over-engineering this up
front spends complexity budget for marginal gains.

## Privacy note

`navigator.connection.effectiveType` does not reveal PII, but it's a
fingerprinting signal. In regulated contexts (EU analytics) you may
need to disclose its use in your privacy policy. Safari's refusal to
implement is partly a privacy-stance call.

## Sources

Synthesis concept. The effectiveType semantics come from the Network
Information API spec; the 100/60/0 budget split is our own opinion based
on user-experience targets.
