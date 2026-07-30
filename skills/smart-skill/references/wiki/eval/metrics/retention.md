---
title: "Retention Check - Brain Doesn't Forget"
context: eval
category: metrics
concept: retention
description: "A retention_check on session N+k asserts that the agent still follows a lesson taught in session N, without being retold. Catches brain-protocol regressions."
tags: eval, metrics, retention, humblskills
sources: []
last_ingested: 2026-04-19
---

## What it is

A `retention_check` is a free-text note on a `scenarios.json` session that
documents what earlier-session lesson the current assertions are testing.
The harness treats it as metadata - it's surfaced in the report and used
by the `retention_at_k` derived metric.

## The shape

```json
{
  "n": 3,
  "prompt": "Ingest this second transcript.",
  "retention_check": "Wiki concept filenames from session 2 must remain kebab-case; session 3 must not regress.",
  "assertions": [
    {
      "text": "every wiki filename is kebab-case",
      "check": "exec:find ... | awk ... | grep -Ev '^[a-z][a-z0-9-]*\\.md$' | head -1 | ..."
    }
  ]
}
```

## Why it matters

Without retention checks you cannot distinguish a smart skill that happens
to perform well on session N from one that actually remembers what it
learned in session N-1. A retention-free scenario could be aced by a
stateless agent.

## Incorrect pattern

```
session 1: teach kebab-case
session 2: re-explain kebab-case, assert it
```

This tests whether the current-session prompt works, not whether the
brain is carrying knowledge forward.

## Correct pattern

```
session 1: teach kebab-case (via a scenarios.json prompt that names the rule)
session 2: DON'T mention kebab-case; assertions test the filenames
```

If the brain is working, the agent read patterns.md / log.md at session 2
start, saw the kebab-case rule recorded by session 1, and followed it
unprompted.

## The claim to highlight

The clearest-signal table in the final report is:

```
metric                    smart_skill    flat_skill
retention_at_k[k=1]       0.90           0.50
retention_at_k[k=2]       0.85           0.45
```

At k=1 and k=2 sessions back, smart retains. Flat doesn't - because it
has no brain state to carry forward. This is the humblskills-specific
claim that eval dashboards need to reify with numbers, not prose.

## Sources

Implementation of the retention hint surface lives in
[cli/internal/eval/scenarios/scenarios.go](../../../../../../cli/internal/eval/scenarios/scenarios.go) via the `RetentionCheck` field.
