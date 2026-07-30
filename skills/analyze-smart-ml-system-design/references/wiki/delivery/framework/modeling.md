---
title: "Step 4: Modeling"
context: delivery
category: framework
concept: modeling
description: "The ~10-minute modeling step: ship a baseline, survey model families with trade-offs and pick one, then detail architecture at the right altitude."
tags: modeling, baseline, model-selection, architecture, trade-offs
sources:
  - "references/raw/learn/ml-system-design/in-a-hurry/delivery.md"
last_ingested: 2026-06-01
---

## Step 4: Modeling

About 10 minutes. Three moves: baseline, selection, architecture.

**Baseline.** Always establish a simple, fast baseline (heuristic, simple
statistical model, or basic ML) to compare against. It moves the conversation
from theory to grounded trade-offs: what does each increment of complexity buy?
Baselines are also gap-fillers: if you need a candidate generator plus a ranker,
a trivial candidate generator gives you a complete end-to-end system so you can
spend time on the ranker.

**Model selection.** Survey appropriate model families and their trade-offs
(cost, complexity, latency, interpretability, predictive power). Usually a
"classical" model vs a deep model: discuss both. Deep is not always best, but it
usually must be considered. Favor ideas proven over the last ~2-3 years; do not
lean only on this year's NeurIPS paper (too unproven) or research 5-7 years
dated (looks stale). Do not neglect non-parametric methods like ANN. Then pick
one model to elaborate; there is not time to cover all.

**Architecture.** Thread the needle between "I will use deep learning" (too
vague) and "4 FC layers of 1024 neurons" (only known empirically). Example: a
two-tower model with user/item embeddings, FC layers + ReLU, sigmoid output,
dropout/L2. This is a credibility test for whether you have built one; expect
follow-ups. Admitting unfamiliarity then generalizing from what you know is fine
and expected, given the breadth of ML.

**Correct (right altitude):**

```text
"Two-tower: user tower eats demographics + recent watch history, item tower eats
content features. FC layers with ReLU, dot-product head, dropout + L2. I would
distill it for the light ranker."
```

- Green: a fast baseline; a few approaches with trade-offs and a justified pick;
  enough architecture detail to be credible.
- Red: jumping to a complex/expensive model with no baseline; hand-waved
  architecture so the interviewer doubts you have built one.

## Sources

- `references/raw/learn/ml-system-design/in-a-hurry/delivery.md` - baseline,
  model selection, architecture guidance and green/red flags.
