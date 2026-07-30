---
title: "Whiteboard: ChatGPT (Simplicity Bar)"
context: examples
category: whiteboards
concept: chatgpt
description: "The simplicity bar for an LLM chat: one stateless service over one DB, with a queue + worker added only to handle the slow AI call and stream tokens via SSE."
tags: chatgpt, llm, sse, streaming, simplicity, whiteboard
sources:
  - "references/raw/whiteboards/chatgpt.png"
last_ingested: 2026-06-01
---

## Whiteboard: ChatGPT (Simplicity Bar)

A reference design holding the simplicity bar for an LLM chat product.

**Incorrect (over-built):**

```text
A model-serving mesh, vector DB, and multi-stage RAG pipeline before the
prompt asks for any of it. The infra question is just slow-call + streaming.
```

**Correct (the simple shape):**

```text
Client -> API Gateway (auth, rate limiting, LB) -> stateless Chat Service -> DB
(Users, Chat, Messages). Chat Service publishes to Kafka; a worker calls the
third-party AI service and streams tokens back via SSE.
```

- **Functional:** send a prompt and receive a response; see past chats and resume with prior context.
- **Non-functional:** selective availability over consistency; 100M DAU; low-latency token streaming (HTTP + SSE).
- **Why simple:** one stateless service + one DB; the queue + worker exist purely to handle the slow, long-running AI call and stream results.

Lesson: the only added complexity (queue + worker + SSE) maps to the one hard requirement, slow streamed responses.

## Sources

- `references/raw/whiteboards/chatgpt.png` - the reference whiteboard for the ChatGPT design.
