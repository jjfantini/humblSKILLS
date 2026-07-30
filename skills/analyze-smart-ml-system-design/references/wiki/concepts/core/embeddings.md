---
title: "Embeddings: Why One-Hot Fails and the Two-Tower Answer"
context: concepts
category: core
concept: embeddings
description: "Learned dense vectors that cluster similar entities, fixing one-hot's waste and poor generalization; trained via matrix factorization, two-tower contrastive, graph, or pre-trained + fine-tune."
tags: embeddings, two-tower, contrastive, one-hot, retrieval
sources:
  - "references/raw/learn/ml-system-design/core-concepts/embeddings.md"
last_ingested: 2026-06-01
---

## Embeddings: Why One-Hot Fails and the Two-Tower Answer

An embedding is a list of numbers representing an entity (user, product, word,
image, query) so that **similar entities get similar vectors**. They show up
everywhere in interviews: recommendations, search, RAG, classification, fraud,
ad ranking.

**Why one-hot fails.** With 10M users, each is a 10M-dim vector with a single 1.
Two problems: (1) it is wasteful, the model learns each entity from scratch with
no shared structure (no notion that two cooking-video watchers are similar); (2)
it does not generalize, a brand-new user needs a new dimension and a retrain.
Embeddings fix both: a compact space where similar entities cluster, distances
are meaningful, you can average them, and you can embed new entities from their
attributes at inference.

**The canonical answer (lands in ~half of interviews):** "I would represent each
user and each video as a 64-dim embedding. The user tower takes demographics and
recent watch history; the video tower takes content features. I train them so
the dot product of a user and video embedding predicts whether the user watched
the video."

**How they are trained.** Pick what "similar" means (it is whatever your
objective says, not necessarily semantic), then pull positives together and push
negatives apart. Four approaches:

- **Matrix factorization** - decompose the interaction matrix R into U and V
  with U . V^T ~ R. Cheap, well-understood, a fine baseline for retrieval.
- **Two-tower / contrastive** - two encoders output into the same vector space;
  train with triplet loss or InfoNCE using in-batch negatives. Hard-negative
  mining drives quality. Pre-compute candidate embeddings, ANN lookup at serve.
- **Graph embeddings** - transductive (node2vec, retrain for new nodes) vs
  inductive (GraphSAGE, embeds new nodes without retraining -> cold start free).
- **Pre-trained + fine-tune** - start from BERT/CLIP/sentence-transformers,
  fine-tune on your data. 80% of the way with 10% of the compute.

**Used for:** features (transfer/share across teams), clustering (k-means on the
space), and retrieval (ANN search; semantic search and RAG). Dimensionality
64-1024, with 128-512 the sweet spot.

**In interviews, be specific:** name the loss, your negative-sampling strategy,
a dimensionality with a reason, and a story for refresh + cold start.

## Sources

- `references/raw/learn/ml-system-design/core-concepts/embeddings.md` - one-hot
  failures, training methods, usage patterns, dimensionality, and serving.
