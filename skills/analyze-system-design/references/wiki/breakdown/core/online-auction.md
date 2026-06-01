---
title: "Design an Online Auction"
context: breakdown
category: core
concept: online-auction
description: "Post auctions, place bids accepted only above the current max, and view the live highest bid; enforce strong consistency with OCC and broadcast updates in real-time."
tags: auction, bidding, contention, occ, realtime
sources:
  - "references/raw/whiteboards/online_auction.png"
  - "references/raw/learn/system-design/problem-breakdowns/online-auction.md"
last_ingested: 2026-06-01
---

## Design an Online Auction

**Functional:** post an item for auction (starting price, end date); bid (accepted only if higher than current max); view an auction including the current highest bid.

**Non-functional:** strong consistency for bids (everyone sees the same max); fault-tolerant and durable (no dropped bids); real-time highest-bid display; 10M concurrent auctions.

**Core entities:** Auction, Item, Bid, User.

**Key API:**

```text
POST /auctions { item, startDate, endDate, startingPrice }
POST /auctions/{id}/bids { Bid }
GET  /auctions/{id} -> Auction & Item
```

**High-level design (from whiteboard):**

```text
CRUD:  Client -> API Gateway -> Auction Service -> DB (Postgres, sharded by auctionId)
Bids:  createBid() -> Producer -> Message Queue (partitioned by auctionId)
                            -> Bid Service -> DB ; SSE push max bid to Client
```

Keep a full **bids table** (accepted | rejected status), not just a `maxBidPrice` field.

**Key deep dives (visible on whiteboard):**
- **Async bid pipeline:** queue decouples burst bid traffic from processing (~100x auction CRUD volume).
- **Partition by auctionId:** queue + DB shard co-locate one auction's ordered bid stream.
- **SSE realtime:** Bid Service pushes max bid updates; no stale polling.
- **Strong consistency:** conditional update / OCC on auction max so stale reads cannot accept lower bids.

See also `wiki/examples/whiteboards/online-auction.md`.

## Sources

- `references/raw/whiteboards/online_auction.png` - async bid queue, SSE, sharding by auctionId.
- `references/raw/learn/system-design/problem-breakdowns/online-auction.md` - OCC, audit trail, realtime prose.
