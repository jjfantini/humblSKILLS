---
title: "Probabilistic Data Structures for Big Data"
context: tech
category: advanced
concept: big-data-structures
description: "Memory-efficient structures (bloom filter, count-min sketch, HyperLogLog) that trade exactness for space; only use under volume + memory limit + approximation tolerance."
tags: bloom-filter, count-min-sketch, hyperloglog, probabilistic
sources:
  - "references/raw/learn/system-design/deep-dives/data-structures-for-big-data.md"
last_ingested: 2026-06-01
---

## Probabilistic Data Structures for Big Data

Memory-efficient probabilistic structures that trade exactness for space. Use only when massive data volume **and** a hard memory constraint **and** tolerance for approximate answers all hold.

**Incorrect (probabilistic where exact fits):**

```text
Use a bloom filter to dedupe a few thousand items -> a plain hash set fits
in memory and gives exact answers. Probabilistic here is a red flag.
```

**Correct (pick the structure to the question):**

```text
Bloom filter: set membership, no false negatives. "Have I probably seen
  this URL?" (web-crawler visited set). No deletions. ~1GB for 1B elements at 1% FP.
Count-Min Sketch: upper-bound frequency counts in a stream. Top-K / heavy
  hitters / per-item view counts at huge volume.
HyperLogLog: cardinality (unique count) in tiny near-constant memory.
  Unique-visitor / distinct-count analytics over billions of items.
```

All three need non-trivial memory and break when another component depends on exact counts. Reach for them only when the exact structure genuinely will not fit.

## Sources

- `references/raw/learn/system-design/deep-dives/data-structures-for-big-data.md` - bloom filter, count-min sketch, HyperLogLog, the three-condition rule.
