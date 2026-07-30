---
title: "Step 1: Problem Framing"
context: delivery
category: framework
concept: problem-framing
description: "The first 5-7 minutes: clarify the problem, establish a business objective, then translate it into a concrete ML objective that steers the whole interview."
tags: problem-framing, business-objective, ml-objective, clarifying-questions
sources:
  - "references/raw/learn/ml-system-design/in-a-hurry/delivery.md"
  - "references/raw/learn/ml-system-design/in-a-hurry/introduction.md"
last_ingested: 2026-06-01
---

## Step 1: Problem Framing

The opening 5-7 minutes. Three moves, in order. Get these right and the rest of
the interview has a spine; get them wrong and you backtrack painfully later.

**1. Clarify the problem.** Ask targeted questions: who are the users, what is
the pain point, is there an existing system, what scale (DAU, QPS), real-time
vs batch inference, latency and privacy constraints. Even when handed a problem
statement, do not jump in cold. Not heavily graded, but strong candidates
quickly find what makes the problem hard and start probing the things teams
work on for years.

**2. Establish a business objective.** The end goal the business cares about
(engagement, cost, risk, revenue), specific and directional. The real objective
is usually NOT the model loss. For harmful content, the goal is reducing
unwanted exposure (weighted by views), not "maximize accuracy". Call out where
business and naive ML objectives diverge. Be specific: "increase click-through
on recommendations" beats "improve user experience". No bonus for precise
numbers; articulate what success looks like directionally.

**3. Decide an ML objective.** Translate into a concrete task (classification,
regression, ranking, clustering) and the metric you will optimize (precision@k,
NDCG). This sets the stage for everything that follows. Do not get hung up on
loss hyperparameters (the exact k); those discussions give little signal.

**Incorrect (skips framing):**

```text
"It is a classifier, so I will maximize accuracy and start picking a model."
```

**Correct (frames first):**

```text
"Goal is minimizing views of harmful content subject to a 95% precision
guardrail. That makes this binary classification weighted by exposure, and it
answers 'can we wait to classify?' objectively: waiting accrues views."
```

- Green: detailed clarifying questions; a clear business objective to optimize;
  an ML objective that guides the rest of the work.
- Red: assuming a naive ML objective; questions that miss what is interesting;
  an ML objective too vague to steer the design.

## Sources

- `references/raw/learn/ml-system-design/in-a-hurry/delivery.md` - the three
  framing moves and their green/red flags.
- `references/raw/learn/ml-system-design/in-a-hurry/introduction.md` - problem
  navigation as a rubric dimension (vague goal -> measurable ML problem).
