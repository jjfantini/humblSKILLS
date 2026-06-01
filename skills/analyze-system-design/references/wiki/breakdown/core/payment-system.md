---
title: "Design a Payment System"
context: breakdown
category: core
concept: payment-system
description: "Process card charges through external networks with durability and auditability; model PaymentIntent vs Transaction, enforce idempotency, and reconcile async state."
tags: payment, idempotency, multi-step, webhooks, ledger
sources:
  - "references/raw/whiteboards/payment_system.png"
  - "references/raw/learn/system-design/problem-breakdowns/payment-system.md"
last_ingested: 2026-06-01
---

## Design a Payment System

**Functional:** merchants initiate payment requests (charge an amount); users pay with credit/debit cards; merchants view payment status.

**Non-functional:** highly secure; durable and auditable (never lose transaction data); transaction safety despite async external networks; 10k+ TPS, bursty.

**Core entities:** Merchant, PaymentIntent (state machine: created -> authorized -> captured/canceled/refunded; owns idempotency), Transaction (polymorphic money-movement, 1:many under an intent).

**Key API:**

```text
POST /payment-intents { amountInCents, currency }
POST /payment-intents/{id}/transactions { type, card }   (tokenized, never raw)
GET  /payment-intents/{id} ;  webhook POST to merchant on status change
```

**High-level design (from whiteboard):**

```text
Customer -> Merchant iframe (card encrypted client-side, never hits merchant server)
        -> API Gateway -> Payment Service + Transaction Service -> External Payment Network
Both services persist to DB (PaymentIntent + Transaction entities).
```

**Key deep dives (visible on whiteboard):**
- **Multi-step / async:** PaymentIntent state machine (created -> authorized -> captured); Transaction Service talks to external network asynchronously.
- **Security / PCI:** iframe keeps raw card off merchant server; tokenized card in API calls.
- **CDC + Kafka:** DB changes stream to Kafka for downstream workers.
- **Reconciliation Worker:** consumes Kafka, pulls batch/one-off data from External Payment Network, stores artifacts in S3, corrects DB.
- **Webhook Consumer:** consumes Kafka, reads Merchant callbackUrl + subscribedEvents, POSTs status to merchant.
- **Idempotency:** reuse PaymentIntent + idempotency key on retries so a card is never double-charged.

See also `wiki/examples/whiteboards/payment-system.md`.

## Sources

- `references/raw/whiteboards/payment_system.png` - full diagram with iframe, CDC, reconciliation, webhooks.
- `references/raw/learn/system-design/problem-breakdowns/payment-system.md` - PaymentIntent vs Transaction, state machine, idempotency, security prose.
