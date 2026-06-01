---
title: "Design Ticketmaster (Event Ticketing)"
context: breakdown
category: core
concept: ticketmaster
description: "View/search events and book tickets without double-booking; reserve seats with a Redis TTL lock, scale reads with caching, and handle viral events with a waiting queue."
tags: ticketmaster, booking, contention, redis-lock, search
sources:
  - "references/raw/learn/system-design/problem-breakdowns/ticketmaster.md"
last_ingested: 2026-06-01
---

## Design Ticketmaster (Event Ticketing)

**Functional:** view events, search events, book tickets.

**Non-functional:** availability for view/search, consistency for booking (no double-booking); handle bursty popular events (10M users, one event); search <500ms; read-heavy (100:1).

**Core entities:** Event, User, Performer, Venue, Ticket, Booking.

**Key API:**

```text
GET  /events/{id} -> Event & Venue & Performer & Ticket[]
GET  /events/search?keyword=&start=&end=&page=
POST /bookings/{eventId} { ticketIds[], paymentDetails }
```

**High-level design:**

> Assumed flow (diagram in source omitted; inferred): Client -> API Gateway -> Event/Search/Booking services -> shared Postgres (events, tickets, bookings) + Stripe.

A shared DB is fine here: data is tightly coupled and booking needs ACID.

**Key deep dives:**
- **Reserve without double-booking:** avoid long-running DB locks; use **implicit status + expiration** in short transactions, or a **Redis distributed lock with TTL** (`SET NX EX`) that auto-expires abandoned reservations. OCC on the DB is the safety net.
- **Scale view path:** aggressive caching of event/venue details (read-through, TTL), load balancing, stateless horizontal scaling.
- **Viral events:** **virtual waiting queue** (Redis sorted set + SSE) gating the booking page.
- **Search:** Postgres FTS, then **Elasticsearch** (fed by CDC) for fuzzy/typeahead.

## Sources

- `references/raw/learn/system-design/problem-breakdowns/ticketmaster.md` - reservation locking, read scaling, waiting queue, search.
