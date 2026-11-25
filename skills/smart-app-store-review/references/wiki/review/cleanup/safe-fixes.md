---
title: "Clean Up: Safe Fixes vs Report-Only"
context: review
category: cleanup
concept: safe-fixes
description: "Which findings the agent fixes after one confirmation and which it only reports, so cleanup never changes business logic silently"
tags: cleanup, fixes, remediation, react-native, swift
sources:
  - "references/raw/upstream/SKILL.md"
  - "references/raw/upstream/2-performance.md"
  - "references/raw/upstream/4-design.md"
  - "references/raw/upstream/5-legal.md"
last_ingested: 2026-11-25
---

## Clean Up

Cleanup means removing what would get the build rejected without changing
what the app does. Show the full list of planned edits, get **one**
confirmation, apply, then re-run `scripts/scan.sh`.

### Safe: apply after confirmation

| Finding | Fix |
|---|---|
| secret-in-log | Delete the log line, or redact the value (`token.slice(0,4)+'...'`); gate debug logs behind `__DEV__` / `#if DEBUG` |
| android-reference | Remove or platform-gate the copy (`Platform.OS === 'android'`) |
| placeholder-content | Remove the screen/feature from the build, or ask the user for the real copy. Never invent product claims |
| precheck-terms | Initial state `false`; disable sign-up until checked |
| pre-permission-wording | One neutral button ("Continue"/"Next") that opens the system prompt; remove Cancel/Close on that screen |
| custom-review-prompt | Replace with `SKStoreReviewController` / `AppStore.requestReview` (Swift) or `expo-store-review` `StoreReview.requestReview()` |
| pricing-word | Remove "Free" from in-app copy that will appear in screenshots |
| deprecated-uiwebview | Migrate to `WKWebView` or `react-native-webview` |

### Report-only: the user decides

| Finding | Why not auto-fix |
|---|---|
| hardcoded-secret | Rotating the key is the real fix; moving it server-side is architecture. Removing it from the bundle alone leaves it in git history |
| external-checkout / IAP / restore | Storefront, entitlement and revenue decisions |
| account-deletion-missing | Needs a backend deletion path |
| att-missing | Depends on what the SDK collects; may need an SDK config change instead of a prompt |
| login-alternative-missing | Adds an auth provider and an App ID capability |
| vague-purpose-string | Draft a specific string ("Scan receipts so the app can fill in the amount") and let the user approve it: it must be true |
| ota-update, private-api, dynamic-code | Requires knowing intent and alternatives |
| privacy-manifest-missing | Declared reasons must match real API use |

**Incorrect:** silently swapping a Stripe link for StoreKit code.
**Correct:** report 3.1.1 with the storefront question, and propose
`react-native-purchases` / StoreKit 2 as the remediation for the user to approve.

## Sources

- `references/raw/upstream/SKILL.md` - remediation hints in the Quick Reference.
- `references/raw/upstream/2-performance.md` - OTA and dynamic code (2.5.2), placeholders (2.1).
- `references/raw/upstream/4-design.md` - 4.8 login services.
- `references/raw/upstream/5-legal.md` - purpose strings and pre-permission screens (5.1.1(ii), (iv)), 5.6.1 review prompts.
