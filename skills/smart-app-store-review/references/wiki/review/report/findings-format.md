---
title: "Findings Report Format"
context: review
category: report
concept: findings-format
description: "One row per finding with severity, guideline, location, evidence and fix, so every item is actionable and auditable"
tags: report, findings, severity, format
sources:
  - "references/raw/upstream/SKILL.md"
last_ingested: 2026-11-25
---

## Findings Report

Most-severe first. One row per verified item. Every row has a guideline
number (or `beyond-reference` with the App Store Connect error code), a
location, and a concrete fix.

**Incorrect:** "Privacy issues found. Consider reviewing your permissions."

**Correct:**

```markdown
## App Store review readiness: 2 blockers, 3 to verify, 4 fixed

| Status | Sev | Guideline | Where | Problem | Fix |
|---|---|---|---|---|---|
| CONFIRMED | CRITICAL | 3.1.1 | src/paywall.tsx:41 | Stripe checkout for premium filters (digital, US+EU storefronts) | StoreKit IAP via react-native-purchases; EU link-out needs the entitlement |
| NEEDS-VERIFICATION | HIGH | 5.1.2 | src/analytics.ts:12 | AppsFlyer starts before ATT | Confirm IDFA collection in SDK config; if on, gate init behind ATT |
| FIXED | HIGH | 1.6 | src/api.ts:88 | Access token logged | Line removed |

Dropped (3): ... one line each with the reason.

App Store Connect (cannot verify from code): privacy policy URL, App Privacy
answers, demo account, Notes for Review, age rating questionnaire, screenshots.
```

Severity follows upstream risk levels: CRITICAL = privacy/data and payments,
HIGH = safety and performance, MEDIUM = design and legal-other. Raise a
MEDIUM to CRITICAL only with a reason in the Problem column.

## Sources

- `references/raw/upstream/SKILL.md` - "Risk Levels by Category" table.
