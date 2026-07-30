---
title: "Ingest a New Raw File into Wiki + Brain"
context: brain
category: ingest
concept: workflow
description: "Turns unstructured dumps into traceable, queryable knowledge"
tags: ingest, raw, wiki, workflow, brain
sources: []
last_ingested: 2026-04-16
---

## Ingest Workflow

The user drops a file into `references/raw/`. The agent distills it into
one or more wiki concepts and updates the brain. No script required -
this is a pure agent operation.

**Incorrect (shallow ingest - one file in, one file out, no brain update):**

```
User: <drops article.md into references/raw/>
Agent: writes references/wiki/content/article.md with a summary
Agent: done
```

No `sources:` cited, no `log.md` entry, no category/concept split, no
`last_ingested` date, no lint run. After 20 ingests the brain has no
idea what it knows.

**Correct (full ingest):**

```
User: <drops references/raw/karpathy-brain-idea.md>

Agent:
1. Reads references/_index.md, patterns.md, decisions.md, last 5 of log.md
2. Reads references/raw/karpathy-brain-idea.md
3. Identifies 3 distinct concepts spanning 2 contexts:
   - brain/protocol/three-layer-architecture
   - brain/patterns/how-to-log-results
   - smart/structure/directory-layout (contributes to existing concept)
4. For each new concept, writes references/wiki/<ctx>/<cat>/<concept>.md
   with frontmatter:
     context: <ctx>
     category: <cat>
     concept: <concept>          # matches filename stem
     sources:
       - "references/raw/karpathy-brain-idea.md"
     last_ingested: 2026-04-16
5. Adds karpathy-brain-idea.md to the sources: of the existing
   smart/structure/directory-layout concept
6. Appends to references/log.md:
     [INGEST 2026-04-16] Processed karpathy-brain-idea.md.
       Produced 2 new concepts, updated 1.
7. If the material includes metrics, appends to patterns.md
8. If splitting concepts required a non-obvious call, appends to decisions.md
9. Runs scripts/lint.sh to regenerate _index.md and verify
```

## Steps

### 1. Read the brain

Mandatory (Brain Protocol). Skip = regressions.

### 2. Read the raw file

Full read. If large, read in chunks but process the entire file.

### 3. Decide scope

Ask: "How many distinct concepts does this file seed?" A single raw file
often spans multiple contexts/categories. Do NOT cram multiple concepts
into one wiki file - split.

### 4. For each concept

- Exists already? Update the concept file in place. Add the raw path to
  its `sources:` if not already listed. Refresh `last_ingested`.
- New concept? Create `references/wiki/<context>/<category>/<concept>.md`
  with full frontmatter + body per `references/_template.md`.

### 5. Write log.md entry

Always. At minimum:

```
[INGEST <YYYY-MM-DD>] Processed <raw-file>.
  Produced <N> new concepts, updated <M>.
```

### 6. Conditional writes

- If the source contains metrics/outcomes -> `patterns.md` entry
- If splitting/merging required a non-obvious choice -> `decisions.md` entry

### 7. Lint

```bash
bash scripts/lint.sh
```

Regenerates `_index.md`, validates path/frontmatter triples, flags
orphans and broken sources. Fix findings before finishing.

## Sources

- (none)
