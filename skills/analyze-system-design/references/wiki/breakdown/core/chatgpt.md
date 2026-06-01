---
title: "Design ChatGPT (LLM Chat Serving)"
context: breakdown
category: core
concept: chatgpt
description: "Stream LLM responses with low time-to-first-token, resume past chats, and schedule scarce GPUs fairly; SSE + gRPC streaming with Redis Streams for gapless reconnects."
tags: chatgpt, llm, sse, gpu-scheduling, streaming
sources:
  - "references/raw/learn/system-design/problem-breakdowns/chatgpt.md"
  - "references/raw/whiteboards/chatgpt.png"
last_ingested: 2026-06-01
---

## Design ChatGPT (LLM Chat Serving)

The LLM is a black box; the design is the serving system around it (streaming, GPU scheduling, cost). Text-only.

**Functional:** send a prompt and receive a streamed response; view past chats and resume with prior context carried into the prompt.

**Non-functional:** time-to-first-token <~500ms then smooth streaming (full response up to ~30s); availability > consistency for chat state; scale under GPU constraints with fair tiered allocation (200M DAU, ~20k prompts/s, ~120k concurrent streams).

**Core entities:** User (tier), Chat, Message; later a Generation.

**Key API:**

```text
POST /chats -> { chatId }
POST /chats/{chatId}/messages { content } -> Message (streamed via SSE, returns runId)
GET  /chats?cursor= ;  GET /chats/{chatId}/messages?cursor=
```

**High-level design:**

> Assumed flow (diagram in source omitted; inferred): Client -> API Gateway -> stateless Chat Service (Postgres for chats/messages) -> Inference Service (GPU workers); see chatgpt whiteboard.

**Key deep dives:**
- **Fast streaming:** **SSE** to the browser (one-way; lighter than WebSockets) + **gRPC server-streaming** worker -> Chat Service; TTFT bounded by model, not poll interval.
- **Gapless reconnects:** decouple worker from connection via a **Redis Stream keyed by runId** (replay from last-seen ID), not fire-and-forget pub/sub; publish token deltas, persist final message to Postgres.
- **GPU scheduling:** queue + admission control; fair allocation across free/paid tiers.
- **Context/cost:** summarize long history + prefix caching instead of resending full context each turn.

## Sources

- `references/raw/learn/system-design/problem-breakdowns/chatgpt.md` - SSE+gRPC streaming, Redis Streams replay, GPU scheduling, context cost.
- `references/raw/whiteboards/chatgpt.png` - the simple Chat/Inference service split.
