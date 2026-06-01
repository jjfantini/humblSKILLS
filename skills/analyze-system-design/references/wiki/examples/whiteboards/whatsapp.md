---
title: "Whiteboard: WhatsApp (Simplicity Bar)"
context: examples
category: whiteboards
concept: whatsapp
description: "The simplicity bar for chat: one chat-server tier over WebSockets behind an L4 LB, one DB, and S3 for media; realtime justifies the only added complexity."
tags: whatsapp, chat, websocket, simplicity, whiteboard
sources:
  - "references/raw/whiteboards/whatsapp_chat.png"
last_ingested: 2026-06-01
---

## Whiteboard: WhatsApp (Simplicity Bar)

A reference design holding the simplicity bar for a realtime chat system.

**Incorrect (over-built):**

```text
Per-message microservices, a queue per chat, and a graph DB for the social
graph before the realtime requirement even forces WebSockets.
```

**Correct (the simple shape):**

```text
Clients hold WebSocket connections through an L4 load balancer to a Chat Server.
Messages persist to a DB (Chat, ChatParticipant, Messages, Inbox tables).
Media -> S3 (messages carry s3 links). Offline delivery via a per-participant Inbox.
```

- **Functional:** group chats (limit 100); send/receive; receive messages sent while offline (up to 30 days); media.
- **Non-functional:** realtime sub-1s (ideally <500ms); 1B DAU; availability over consistency; always delivered, even offline; graceful failure on server death.
- **Why simple:** one chat-server tier + one DB + S3 for blobs. The realtime requirement justifies WebSockets and the L4 LB; nothing else added.

Lesson: add one specialized component per hard non-functional requirement, no more.

## Sources

- `references/raw/whiteboards/whatsapp_chat.png` - the reference whiteboard for the WhatsApp chat design.
