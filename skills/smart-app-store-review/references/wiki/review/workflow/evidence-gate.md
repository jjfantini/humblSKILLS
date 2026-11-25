---
title: "Evidence Gate: Candidate to Finding"
context: review
category: workflow
concept: evidence-gate
description: "Prevents false rejections: an import, API call, or URL alone never becomes a finding without the evidence named here"
tags: evidence, false-positive, verification, att, account-deletion, payments
sources:
  - "references/raw/upstream/SKILL.md"
  - "references/raw/upstream/3-business.md"
  - "references/raw/upstream/5-legal.md"
last_ingested: 2026-11-25
---

## Evidence Gate

Upstream's core rule: do not declare a rejection from an API call, SDK import,
or URL alone. Inspect the configuration and the actual behavior; if that
evidence is unavailable, report **what needs verification** instead.

**Incorrect:**

```text
CRITICAL 5.1.2  App imports AppsFlyer and never calls ATT. Will be rejected.
```

**Correct:**

```text
NEEDS-VERIFICATION 5.1.2  src/analytics.ts:12 initializes AppsFlyer before any
ATT prompt. Rejected only if it links user data with third-party data for ads
or shares it with a data broker. Check the SDK config (IDFA collection,
ad-network sharing) and the App Privacy answers.
```

## What each check needs before it is CONFIRMED

| check-id | Evidence required |
|---|---|
| att-missing | The SDK actually performs Apple-defined tracking (IDFA, cross-company ad linking, data brokers). First-party analytics alone does not need ATT |
| account-deletion-missing | The app has account creation and no in-app path that **initiates** deletion. A button linking straight to a web page that completes deletion is allowed; a link to generic help/instructions is not |
| external-checkout | The purchase is a digital good/feature consumed in the app **and** no storefront entitlement or 3.1.3 exception (reader, multiplatform, person-to-person, physical goods, enterprise...) applies. US storefront purchase links need no entitlement (3.1.1(a)) |
| login-alternative-missing | Third-party/social login is the account's primary login and no option meeting 4.8's criteria exists. Own-account-only or enterprise/education logins are exempt |
| secret-in-log / hardcoded-secret | The value is real and reaches a production build (not a test key, not stripped by a build flag). Check redaction |
| ota-update | The update path delivers new features or changes the app's primary purpose, not just bug fixes |
| dynamic-code | Downloaded or remote code is executed; an eval on a bundled literal in a vendored lib is a lower-risk note |
| private-api | The selector is not public API (check Apple docs); dlopen of a bundled framework is normal |
| restore-purchases-missing | The products are restorable (non-consumables, subscriptions). Consumables-only apps need no restore |
| vague-purpose-string | Read it: does it say what is accessed **and** why, with an example? Length alone is a hint |
| privacy-manifest-missing | The app or an SDK uses a required-reason API or tracking domains. Beyond the upstream reference: cite the App Store Connect warning (ITMS-91053), not a guideline number |
| webview-wrapper | The app adds nothing a website would not (4.2 minimum functionality) |

Pattern hits on comments, tests, dead code, and vendored dependencies are
DROPPED with one line of reasoning, so the user can audit the drop.

## Sources

- `references/raw/upstream/SKILL.md` - the "Quick Reference" verification preamble (ATT, deletion, logging, payments).
- `references/raw/upstream/3-business.md` - 3.1.1(a) and 3.1.3 exceptions.
- `references/raw/upstream/5-legal.md` - 5.1.1(v) account sign-in/deletion, 5.1.2 tracking.
