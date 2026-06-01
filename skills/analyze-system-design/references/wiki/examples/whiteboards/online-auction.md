---
title: "Whiteboard: Online Auction (Real-Time Bidding)"
context: examples
category: whiteboards
concept: online-auction
description: "Reference design with async bid pipeline: API Gateway producer, queue partitioned by auctionId, Bid Service with SSE push, Postgres sharded by auctionId."
tags: auction, sse, kafka, sharding, contention, whiteboard
sources:
  - "references/raw/whiteboards/online_auction.png"
  - "references/raw/learn/system-design/problem-breakdowns/online-auction.md"
last_ingested: 2026-06-01
---

## Whiteboard: Online Auction (Real-Time Bidding)

Core CRUD is Auction Service + DB. Deep dives are the async bid queue, partition strategy, and SSE realtime updates.

**Auction management (lower volume):**

```text
Client -> API Gateway (routing, auth)
       -> Auction Service -> Database (Postgres, sharded by auctionId)
       getAuction() / createAuction()
```

**Bidding pipeline (deep dive - ~100x volume):**

```text
createBid() -> Producer -> Message Queue (partitioned by auctionId)
                        -> Bid Service -> Database
Bid Service -> Pub/Sub + SSE connection -> Client (max bid updates)
```

**Entities (from diagram):**

- **Auction:** id, itemId, startTime, endTime, creatorId, startingPrice, maxBidPrice
- **Item:** id, name, description, imageLinks
- **Bid:** id, auctionId, price, userId, createdAt, status (accepted | rejected)

**Key design choices on the board:**

- **Queue partitioned by auctionId:** bids for one auction stay ordered on one partition.
- **DB sharded by auctionId:** co-locate auction state with its bid stream.
- **SSE not polling:** push max bid updates in real time after Bid Service accepts.
- **Append bids table:** status accepted/rejected preserves audit trail, not just maxBidPrice field.

## Sources

- `references/raw/whiteboards/online_auction.png` - async bid queue, SSE, sharding by auctionId.
- `references/raw/learn/system-design/problem-breakdowns/online-auction.md` - OCC, durability, realtime prose.
