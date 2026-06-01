---
name: analyze-ml-system-design
description: Drive an ML system design from an ambiguous prompt to a production-minded solution with a 6-step framework (problem framing, high-level design, data and features, modeling, inference and evaluation, deep dives). Use for an ML system design interview or any "design X" ML prompt - recommendation systems, fraud or bot detection, content moderation, ranking, search, personalization - or whenever the core problem is choosing data, features, a model, and an evaluation strategy. For infra or non-ML system design (URL shortener, chat, rate limiter, file sync), use analyze-system-design instead.
license: MIT
metadata:
  author: Jennings Fantini
  version: "2.0.0"
  tags: [ml-system-design, interviews, machine-learning]
  platforms: [claude-code, cursor]
  preserve:
    - references/decisions.md
    - references/log.md
    - references/patterns.md
---

# Analyze ML System Design

Take an ambiguous business problem and show how ML drives real impact: not just
a model, but structured thinking and reasonable trade-offs under time pressure.

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

Use **analyze-ml-system-design** (this skill) when the heart of the problem is
data, features, model, and evaluation: design a recommendation feed, build fraud
or bot detection, content moderation, ranking and search relevance.

Use **analyze-system-design** when the problem is infrastructure and data
movement (URL shortener, chat, ride-sharing, file sync, rate limiter). Many ML
problems still need an infra backbone (serving, feature store, queues); pull
those components from that sibling skill when you reach serving and scaling.

## CCCCC Architecture

| Layer        | Role                            | Location                                          |
|--------------|---------------------------------|---------------------------------------------------|
| **Core**     | Root structure of the skill     | `SKILL.md`, `references/`, `scripts/`             |
| **Context**  | Top-level taxonomy grouping     | First segment under `references/wiki/`            |
| **Category** | Topic within a context          | Second segment under `references/wiki/`           |
| **Concept**  | One atomic idea per file        | Filename stem AND required frontmatter field      |
| **Command**  | Deterministic executable script | `scripts/<command>.sh` linked from wiki           |

## When to Use

- Preparing for or running an applied ML system design interview.
- Any "design a recommender / fraud detector / moderation / ranking" prompt.
- Choosing data sources, features, a model family, and how to evaluate it.
- Reasoning about cold start, drift, feedback loops, or serving at scale.

## How to Use

Read `references/_index.md` first for the full concept map, then load only the
concepts the task needs:

- Interview type and assessment rubric -> `references/wiki/intro/overview/interview-types.md`
- The 6-step framework and timing -> `references/wiki/delivery/framework/overview.md`
  - Step 1 framing -> `references/wiki/delivery/framework/problem-framing.md`
  - Step 3 data/features -> `references/wiki/delivery/framework/data-features.md`
  - Step 4 modeling -> `references/wiki/delivery/framework/modeling.md`
  - Step 5 inference/eval -> `references/wiki/delivery/framework/inference-eval.md`
  - Step 6 deep dives -> `references/wiki/delivery/framework/deep-dives.md`
- Core knowledge -> `references/wiki/concepts/core/` (`feature-engineering`,
  `embeddings`, `generalization`, `evaluation`)
- Worked problems -> `references/wiki/breakdown/core/` (`harmful-content`,
  `bot-detection`, `video-recommendations`)
- Serving/scaling infra -> the `analyze-system-design` skill.

## Green vs Red Flags (quick reminder)

- Clear business objective and concrete ML objective before modeling. (Red:
  assuming a naive "maximize accuracy" objective.)
- Enumerate signal sources, then a few signals each. (Red: a feature dump.)
- Ship a simple baseline first. (Red: jumping to a complex, expensive model.)
- Tie every metric to the business objective. (Red: metrics in a vacuum.)
- Consider inference constraints; give architecture detail that proves you have
  built one. (Red: hand-waving so the interviewer doubts you.)
