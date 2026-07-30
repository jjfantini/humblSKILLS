---
title: "Design Robinhood (Brokerage)"
context: breakdown
category: core
concept: robinhood
description: "Show live stock prices and manage orders against an external exchange; fan out prices via SSE from a cache and dispatch orders through a low-latency gateway."
tags: robinhood, trading, sse, order-gateway, consistency
sources:
  - "references/raw/learn/system-design/problem-breakdowns/robinhood.md"
last_ingested: 2026-06-01
---

## Design Robinhood (Brokerage)

We are the brokerage, not the exchange. The exchange offers synchronous order placement/cancel and a push trade feed.

**Functional:** see live stock prices; manage orders (market/limit, create/cancel, list).

**Non-functional:** high consistency for order management; 20M DAU, ~5 trades/day, 1000s of symbols; price updates and order placement <200ms; minimize expensive exchange connections.

**Core entities:** User, Symbol, Order.

**Key API:**

```text
GET    /symbol/:name ;  GET /subscribe?symbols=  (SSE)
POST   /order { position, symbol, priceInCents, numShares }   (cents, not float)
DELETE /order/:id ;  GET /orders
```

**High-level design:**

> Assumed flow (diagram in source omitted; inferred): Symbol Price Processor tails the exchange feed -> price cache -> Symbol Service pushes via SSE; Order Service -> order gateway -> exchange; order DB partitioned by userId.

**Key deep dives:**
- **Live prices:** one **price processor** subscribes to the exchange and updates a cache; fan out to clients via **SSE** (not per-client polling of the exchange), with sticky sessions.
- **Order dispatch:** send orders to a low-latency **order gateway (NAT/elastic IP)** to keep exchange connections few, rather than a queue (which risks the 200ms SLA under load).
- **Order tracking:** ACID order DB with state (pending/submitted), `externalOrderId`; a trade processor tails the feed to update order status.

## Sources

- `references/raw/learn/system-design/problem-breakdowns/robinhood.md` - SSE price fan-out, order gateway, order state tracking.
