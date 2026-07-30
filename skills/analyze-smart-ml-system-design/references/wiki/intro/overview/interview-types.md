---
title: "The Four ML Interview Types and Assessment Dimensions"
context: intro
category: overview
concept: interview-types
description: "The four broad ML interview types and which one this skill targets (Applied ML System Design), plus the five dimensions interviewers assess."
tags: interview-types, applied-ml, assessment, rubric
sources:
  - "references/raw/learn/ml-system-design/in-a-hurry/introduction.md"
last_ingested: 2026-06-01
---

## The Four ML Interview Types and Assessment Dimensions

The "ML Engineer" role is poorly standardized, so interviews vary. Spend time
with your recruiter to learn what the role is, then work backwards. There are
four broad types:

1. **Applied ML System Design** - designing practical ML solutions in
   production, assuming garden-variety ML infra (serving, pipelines) exists.
   The majority of ML engineering interviews and the focus of this skill.
   Examples: design a recommendation system, build fraud detection, content
   moderation.
2. **ML Infra Design** - the infrastructure itself: model serving and scaling,
   training systems, feature stores, pipeline orchestration. This skill helps
   partly but emphasizes modeling/data more than infra interviews do.
3. **AI/ML Research** - theoretical foundations, latest papers, novel
   algorithms. This skill is not useful here.
4. **AI/ML Research Engineering** - implementing/optimizing research papers,
   frameworks, hardware acceleration. This skill is not useful here.

**Do not pre-qualify yourself.** Prefixing the interview with "I have only
worked on recommendation systems" undersells you before the interviewer can
assess your skills. If you got the interview, they think you might be the right
candidate. Do not poison the well.

**Assessment dimensions** (rubrics overlap heavily across companies):

- **Problem navigation** - frame a vague business goal as a measurable ML
  problem; decide if ML is even appropriate; justify classification vs ranking.
- **Input data, features, labels** - recruit the right data, design labels,
  avoid leakage and feedback loops, discuss representation.
- **Model design** - the heart of the interview; select a model, detail
  architecture, reason about trade-offs and multi-component systems.
- **Integration and evaluation** - deploy, monitor, iterate; bridge the gulf
  between a notebook idea and production; a separate evaluation discussion.
- **Communication** - implicit in every interview; can you collaborate and
  explain clearly?

Entry-level ML roles rarely include system design; it becomes common at
mid-level and is the norm at senior. More senior candidates get more ambiguous
problems and are expected to find optimal formulations themselves.

## Sources

- `references/raw/learn/ml-system-design/in-a-hurry/introduction.md` - the four
  interview types and the five assessment dimensions.
