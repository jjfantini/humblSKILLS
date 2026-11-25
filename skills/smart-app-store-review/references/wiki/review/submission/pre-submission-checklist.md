---
title: "Pre-Submission Checklist (Code and App Store Connect)"
context: review
category: submission
concept: pre-submission-checklist
description: "Final gate before Submit for Review: the items code scanning cannot prove, from metadata to demo accounts to reviewer devices"
tags: checklist, app-store-connect, metadata, submission, review-notes
sources:
  - "references/raw/upstream/SKILL.md"
  - "references/raw/upstream/2-performance.md"
last_ingested: 2026-11-25
---

## Pre-Submission Checklist

Run after cleanup. Items marked **ASC** live in App Store Connect; ask the
user to confirm them, because code cannot.

### Privacy (5.1)
- [ ] **ASC** Privacy policy URL set; the policy is also reachable inside the app
- [ ] **ASC** App Privacy details match what the app and its SDKs collect
- [ ] Purpose strings specific and accurate; only needed permissions requested
- [ ] ATT before Apple-defined tracking or IDFA access
- [ ] Account creation implies in-app deletion initiation
- [ ] Pre-permission screens: one neutral button, no Cancel/Close

### Payments (3.1)
- [ ] Digital goods use StoreKit unless a storefront entitlement or 3.1.3 exception applies
- [ ] Restore purchases; subscription terms shown before purchase; loot box odds disclosed
- [ ] Credits/currencies do not expire

### Safety (1.x)
- [ ] UGC/AI generation: filter, report, block, contact; terms (zero tolerance) agreed before sign-up via unchecked box or explicit button; reports acted on within 24 h
- [ ] NCII safeguards for AI image features; Kids Category parental gates

### Performance (2.x)
- [ ] **ASC** Age rating questionnaire, including the social media questions (mandatory from September 2026)
- [ ] Tested on iPhone **and** iPad. App Review currently uses iPad Air 11-inch (M3) and iPhone 17 Pro Max (as of August 2026)
- [ ] No crashes, no placeholder content, every feature reachable
- [ ] **ASC** Notes for Review describe new features and how to reach them (2.3.1); **demo account** with working credentials if login is required
- [ ] No hidden, dormant, or post-approval-enabled features, including remote flags (2.5.19)
- [ ] **ASC** No "Free" or other pricing words in name, subtitle, keywords, screenshots (including captured app UI) (2.3.7); creative assets comply with 2.3.14
- [ ] IPv6-only network tested; public APIs only; background modes justified

### Design and legal (4.x, 5.x)
- [ ] Real native functionality beyond a website; original name and branding
- [ ] Social login comes with a 4.8-compliant option (Sign in with Apple is simplest)
- [ ] Push / Live Activities not used for marketing spam without opt-in
- [ ] Licensed third-party content; correct Apple trademark use; system review prompt only

## Sources

- `references/raw/upstream/SKILL.md` - "Pre-Submission Checklist" (condensed here; full text there).
- `references/raw/upstream/2-performance.md` - 2.3.1 Notes for Review, 2.5.19 hidden features, reviewer devices.
