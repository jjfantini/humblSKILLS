---
title: "Audit, Clean Up, Verify: The Review Loop"
context: review
category: workflow
concept: audit-loop
description: "Five-phase loop that turns a repo into a verified, cleaned-up submission with every finding cited to a guideline"
tags: workflow, audit, cleanup, verify, app-review
sources:
  - "references/raw/upstream/SKILL.md"
last_ingested: 2026-11-25
command: scripts/scan.sh
---

## The Review Loop

One pass, five phases. Skipping verification (phase 3) is the most common
way an audit goes wrong: it either buries the real blocker under noise or
asks the user to fix something Apple accepts.

### 1. Scope

Establish before reading code. Every later judgment depends on it.

| Question | Why it matters |
|---|---|
| Stack: Swift / Objective-C / React Native / Expo (managed or bare)? | Fixes differ per stack; upstream rules carry both |
| Platforms: iOS, iPadOS, macOS, tvOS, watchOS, visionOS? | 2.4.x and Mac sandbox rules |
| Storefronts: US, EU, Brazil, Korea, worldwide? | 3.1.1(a) purchase links and alternative payments are storefront-specific |
| Accounts? UGC or AI generation? IAP or subscriptions? Ads or attribution SDKs? Kids Category? | Each switches on whole rule blocks |
| First submission, update, or a response to a rejection? | A rejection message names the guideline; start there |

### 2. Scan (candidates)

```bash
bash scripts/scan.sh <repo> | sort
```

The output is a list of **candidates**, one per line:
`SEVERITY  GUIDELINE  file:line  check-id  text`. A grep hit is not a rejection.
Then walk the five sections in `review/guidelines/section-map.md` for
what grep cannot see: UX flows, metadata, business model, IPv6, iPad layout.

### 3. Verify (evidence gate)

Apply `review/workflow/evidence-gate.md` to every candidate. Each becomes
CONFIRMED, NEEDS-VERIFICATION (with the exact missing evidence), or DROPPED.

### 4. Clean up

Apply `review/cleanup/safe-fixes.md`: mechanical fixes after the user
confirms the list; business, legal, and flow decisions are reported, never
silently edited.

### 5. Re-verify and report

Re-run `scripts/scan.sh`; confirm fixed candidates are gone and nothing new
appeared. Run the project's own build/tests. Then write the report per
`review/report/findings-format.md`, ending with the App Store Connect items
from `review/submission/pre-submission-checklist.md`, which code cannot prove.

## Sources

- `references/raw/upstream/SKILL.md` - "When to Apply", risk levels, and the instruction to
  establish storefront and platform before flagging payments.
