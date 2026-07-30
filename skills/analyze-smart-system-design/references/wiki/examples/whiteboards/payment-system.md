---
title: "Whiteboard: Payment System (Stripe-like)"
context: examples
category: whiteboards
concept: payment-system
description: "Full reference design for card payments: iframe tokenization, PaymentIntent/Transaction services, CDC to Kafka, reconciliation worker, and merchant webhooks."
tags: payment, stripe, kafka, webhooks, reconciliation, whiteboard
sources:
  - "references/raw/whiteboards/payment_system.png"
  - "references/raw/learn/system-design/problem-breakdowns/payment-system.md"
last_ingested: 2026-06-01
---

## Whiteboard: Payment System (Stripe-like)

A complete reference design showing the core payment path plus deep-dive infrastructure (CDC, reconciliation, webhooks). Use as the target shape once non-functional requirements force async durability and merchant notification.

**Core flow (functional path):**

```text
Customer -> Merchant (iframe collects card; encrypted payload never hits merchant server)
        -> API Gateway -> Payment Service + Transaction Service -> External Payment Network
Payment Service writes PaymentIntent; Transaction Service writes Transaction; both persist to DB.
```

**Entities (from diagram):**

- **PaymentIntent:** id, merchantId, customerId, status (e.g. created), amount, product, metadata
- **Transaction:** id, paymentId, amount, type, status, network, metadata
- **Merchant:** id, name, api key, callbackUrl, subscribedEvents, metadata

**Deep-dive components (added after core works):**

- **Kafka event stream (CDC):** DB writes publish change events so downstream workers react without polling.
- **Reconciliation Worker:** consumes Kafka, queries External Payment Network (batch files or one-off), stores reconciliation artifacts in S3, writes corrections back to DB.
- **Webhook Consumer:** consumes Kafka, reads Merchant subscribedEvents + callbackUrl from DB, POSTs status updates to merchant server.

**Why this shape:** the interview core is two services + one DB + external network. Kafka, reconciliation, and webhooks each answer a named non-functional: auditability/correctness, network drift, and real-time merchant updates.

## Sources

- `references/raw/whiteboards/payment_system.png` - full diagram including CDC, reconciliation, and webhook paths.
- `references/raw/learn/system-design/problem-breakdowns/payment-system.md` - PaymentIntent vs Transaction, idempotency, security prose.
