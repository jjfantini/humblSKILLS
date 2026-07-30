---
title: "Design WhatsApp (Messaging)"
context: breakdown
category: core
concept: whatsapp
description: "Realtime group chat with offline delivery and media; WebSockets over an L4 LB, per-recipient inbox for offline, and pub/sub routing across many chat servers."
tags: whatsapp, chat, websocket, inbox, pubsub
sources:
  - "references/raw/learn/system-design/problem-breakdowns/whatsapp.md"
last_ingested: 2026-06-01
---

## Design WhatsApp (Messaging)

**Functional:** group chats (<=100 participants); send/receive messages; receive messages sent while offline (up to 30 days); send/receive media.

**Non-functional:** delivery <500ms; guaranteed deliverability; billions of users; messages stored only as long as needed; resilient to component failure.

**Core entities:** User, Chat, Message, Client (a user may have many devices).

**Key API (WebSocket commands):**

```text
-> createChat, sendMessage, modifyChatParticipants, getAttachmentTarget
<- chatUpdate, newMessage   (clients send ack on receipt)
```

**High-level design:**

> Assumed flow (diagram in source omitted; inferred): Clients <-WebSocket-> L4 LB -> Chat Server (in-memory userId->connection map) -> DynamoDB (Chat, ChatParticipant w/ GSI, Messages, Inbox); media via presigned S3.

Write to a per-recipient **Inbox** table; deliver immediately if online, else on reconnect; client **ack** deletes from inbox. TTL cleans old messages.

**Key deep dives:**
- **Scale to billions:** hundreds of chat servers (~1-2M connections each). Senders/recipients land on different servers -> route via **pub/sub** (or a consistent-hashing registry tracking which server holds which connection).
- **Media:** clients upload directly to blob storage via **presigned URLs** (chat server never proxies bytes); 30-day TTL.
- **Delivery guarantees:** acks + inbox retries until confirmed.

## Sources

- `references/raw/learn/system-design/problem-breakdowns/whatsapp.md` - inbox/offline delivery, multi-server routing, media handling.
