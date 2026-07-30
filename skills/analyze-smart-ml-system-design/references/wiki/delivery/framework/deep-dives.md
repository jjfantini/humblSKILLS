---
title: "Step 6: Deep Dives"
context: delivery
category: framework
concept: deep-dives
description: "The remaining time: go deep on edge cases, scaling, and monitoring/maintenance, led by the interviewer or by the hard parts you flagged earlier."
tags: deep-dives, cold-start, scaling, monitoring, retraining
sources:
  - "references/raw/learn/ml-system-design/in-a-hurry/delivery.md"
last_ingested: 2026-06-01
---

## Step 6: Deep Dives

The final stretch. Go where the interviewer drives or where the hard parts you
earmarked lead. Three common categories:

**Edge cases.** Cold start for new users and items (content-based filtering for
new items, epsilon-greedy or exploration for new users), data sparsity,
seasonality, and bias mitigation for non-representative training data.

**Scaling.** Distributed training, efficient serving architectures, caching
strategies as user base and data volume grow.

**Monitoring and maintenance.** Track drift, product metrics (CTR,
recommendation diversity), and set alert thresholds. Define when you retrain:
automated retraining triggers when performance drops below a threshold.

This is your chance to show depth in your strongest areas. Keep a running
shortlist of topics you flagged earlier ("cold start, explore/exploit, feedback
loops") and offer the interviewer a choice. That signals you noticed and did not
forget, without spending time on every one.

**Correct (offer a menu, follow their lead):**

```text
"A few minutes left. I would like to cover cold start, explore/exploit, or
feedback loops. Any preference, or should I start with cold start since it gates
new-creator growth?"
```

If the interviewer is driving, follow them. You will not get to your pet
reinforcement-learning topic if they want the ranking model fleshed out. The
deep dive is collaborative, not a monologue.

## Sources

- `references/raw/learn/ml-system-design/in-a-hurry/delivery.md` - the three
  deep-dive categories and guidance on letting the interviewer steer.
