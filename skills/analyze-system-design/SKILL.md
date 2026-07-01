---
name: analyze-system-design
description: Drive a system design interview from a blank prompt to a simple, complete, defensible design using a 6-step framework (requirements, core entities, API, optional data flow, high-level design, deep dives). Use it for system design interviews, "design X system" prompts, high-level design and architecture sketches, deep dives, scaling and data-modeling questions, and choosing technologies (Redis, Kafka, Cassandra, DynamoDB, Postgres). For ML-centric problems (recommendations, fraud or bot detection, ranking, content moderation) use analyze-ml-system-design instead.
license: MIT
metadata:
  author: Jennings Fantini
  version: "2.0.3"
  category: development
  tags: [system-design, interviews, architecture]
  platforms: [claude-code, cursor, codex]
  preserve:
    - references/raw/
    - references/wiki/
    - references/decisions.md
    - references/log.md
    - references/patterns.md
---

# Analyze System Design

Drive a system from a blank prompt to a simple, complete, defensible design. The biggest failure mode is not delivering a working system, usually from over-engineering early. Build the simplest thing that satisfies the functional requirements, then harden it against non-functional requirements in deep dives.

## Brain Protocol (read BEFORE answering)

1. `references/_index.md`       - what this skill knows (map)
2. `references/patterns.md`     - what worked, with numbers
3. `references/decisions.md`    - past reasoning, don't repeat mistakes
4. `references/log.md`          - last 5 session entries
5. Relevant `references/wiki/<context>/<category>/` concepts per task

After completing work, UPDATE the brain:
- New source material lives in `references/raw/` (never renamed or edited)
- Distilled insights -> new/updated `references/wiki/<context>/<category>/<concept>.md` (cite raw in `sources:`)
- Performance data (if reported) -> `patterns.md`
- Non-obvious decisions -> `decisions.md`
- Session summary (always) -> append to `log.md`
- Run `scripts/lint.sh` to regenerate `_index.md` and verify structure

_Full spec: `references/_brain.md`._

## Is this the right skill?

Use **analyze-system-design** (this skill) when the problem is infrastructure and data movement: URL shortener, chat, ride-sharing, file sync, rate limiter, news feed, ticketing, payments, metrics.

If the heart of the problem is choosing data, features, a model, and an evaluation strategy (recommendations, fraud/bot detection, content moderation, ranking, search relevance, personalization), use the **analyze-ml-system-design** skill instead. ML designs still need an infra backbone; pull those components from this skill's `tech/` concepts when you reach serving and scaling.

## CCCCC Architecture

| Layer        | Role                            | Location                                          |
|--------------|---------------------------------|---------------------------------------------------|
| **Core**     | Root structure of the skill     | `SKILL.md`, `references/`, `scripts/`             |
| **Context**  | Top-level taxonomy grouping     | First segment under `references/wiki/`            |
| **Category** | Topic within a context          | Second segment under `references/wiki/`           |
| **Concept**  | One atomic idea per file        | Filename stem AND required frontmatter field      |
| **Command**  | Deterministic executable script | `scripts/<command>.sh` linked from wiki           |

## When to Use

- A "design X" prompt or system design interview, mock, or prep.
- Gathering functional and non-functional requirements and scoping a system.
- Sketching a high-level design (boxes and arrows) that satisfies an API.
- Driving deep dives: scaling reads/writes, contention, realtime, large blobs.
- Picking a database, cache, queue, or search engine and justifying it.
- Studying a specific problem breakdown (29 worked designs under `breakdown/core/`).

## How to Use

Walk the framework in order; load only the wiki concept you need for the current step.

- **Step 1 Requirements** -> `wiki/delivery/framework/` (`requirements`, `nonfunctional-checklist`, `estimation`, `overview`) plus `wiki/concepts/theory/` (`cap-theorem`, `numbers-to-know`).
- **Step 2 Core entities** -> `wiki/concepts/data/data-modeling`.
- **Step 3 API** -> `wiki/concepts/network/` (`api-design`, `networking-essentials`).
- **Step 4 Data flow** -> `wiki/patterns/core/` (`multi-step-processes`, `long-running-tasks`).
- **Step 5 High-level design** -> `wiki/tech/core/` and `wiki/tech/advanced/` to pick components; `wiki/examples/whiteboards/` for the simplicity bar (url-shortener, uber, whatsapp, dropbox, chatgpt) and full final designs (payment-system, metrics-monitoring, job-scheduler, youtube, youtube-top-k, online-auction, web-crawler, fb-post-search).
- **Step 6 Deep dives** -> `wiki/patterns/core/` (realtime, contention, scaling reads/writes, large blobs) + `wiki/concepts/scaling/` (`sharding`, `consistent-hashing`) + `wiki/concepts/data/` (`caching`, `db-indexing`).
- **Studying a specific problem** -> `wiki/breakdown/core/<problem>.md` (bitly, ticketmaster, uber, whatsapp, instagram, chatgpt, and 23 more).

For the live, lint-generated map of every concept and its raw sources, read `references/_index.md`.

## Keep it simple (the bar to hit)

The whiteboard examples set the target: one client, one API gateway, one or two services, one database. Add a cache, queue, CDN, replica, or shard only when a specific non-functional requirement forces it, and say which one. If the design looks busy before step 6, cut back.

**Simplicity bar:** `url-shortener`, `uber`, `whatsapp`, `dropbox`, `chatgpt`.

**Full final designs (core + deep dives):** `payment-system`, `metrics-monitoring`, `job-scheduler`, `youtube`, `youtube-top-k`, `online-auction`, `web-crawler`, `fb-post-search`. See `wiki/examples/whiteboards/overview.md`.
