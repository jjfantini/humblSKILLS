---
title: "Guideline Section Map: Where Each Rule Lives"
context: review
category: guidelines
concept: section-map
description: "Routes any guideline number or topic to the exact upstream rule file, so the agent reads only the section it needs"
tags: guidelines, safety, performance, business, design, legal, routing
sources:
  - "references/raw/upstream/1-safety.md"
  - "references/raw/upstream/2-performance.md"
  - "references/raw/upstream/3-business.md"
  - "references/raw/upstream/4-design.md"
  - "references/raw/upstream/5-legal.md"
  - "references/raw/upstream/SKILL.md"
  - "references/raw/upstream/LICENSE"
last_ingested: 2026-11-25
---

## Section Map

The full rule text, checklists, and Swift / React Native / Expo examples
live in `references/raw/upstream/`: five files, about 150 KB together. Never
load all five. Find the section here, then read only that heading
(`grep -n '^#' references/raw/upstream/<file>` lists them).

| Section | File | Read it when the app has |
|---|---|---|
| 1 Safety | `raw/upstream/1-safety.md` | UGC, AI generation (1.2, incl. NCII safeguards), kids (1.3), medical/physical harm (1.4), developer contact info (1.5), data security (1.6) |
| 2 Performance | `raw/upstream/2-performance.md` | Incomplete features (2.1), TestFlight (2.2), metadata/screenshots/age rating/creative assets (2.3, 2.3.14), device support and Mac rules (2.4), public APIs, OTA, background modes, IPv6 (2.5), hidden features (2.5.19) |
| 3 Business | `raw/upstream/3-business.md` | IAP, subscriptions, purchase links (3.1.1(a)), reader/multiplatform exceptions (3.1.3), crypto (3.1.5), other models (3.2) |
| 4 Design | `raw/upstream/4-design.md` | Copycat risk (4.1), thin or web-wrapper app (4.2), spam/clones (4.3), extensions (4.4), Apple services and push (4.5), mini apps/chatbots/emulators (4.7), login (4.8), Apple Pay (4.9) |
| 5 Legal | `raw/upstream/5-legal.md` | Privacy policy, permissions and purpose strings (full Info.plist reference), account deletion, tracking/ATT (5.1), IP (5.2), gambling (5.3), VPN (5.4), MDM (5.5), review prompts and code of conduct (5.6) |

Each rule file ends with a React Native packages reference and an RN
checklist. Use those for RN/Expo apps.

## Currency

The upstream rules are current through Apple's **October 16, 2026** App Review
Guidelines update. Before a real submission, if the current date is more than
about 3 months past that, check https://developer.apple.com/news/ (filter:
App Store) and the guidelines page for later changes, and say in the report
which date the review was checked against.

## Provenance

Content is adapted from safaiyeh/app-store-review-skill v1.5.0 (MIT, see
`raw/upstream/LICENSE`), upstream commit 7197cff. raw/ is never edited;
re-ingest by replacing the files from upstream and updating this page.

## Sources

- `references/raw/upstream/1-safety.md` ... `references/raw/upstream/5-legal.md` - section headings and scope.
- `references/raw/upstream/SKILL.md` - "Guideline Sections" table and currency date.
- `references/raw/upstream/LICENSE` - MIT terms for the adapted content.
