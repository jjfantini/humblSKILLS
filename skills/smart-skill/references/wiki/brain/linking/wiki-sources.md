---
title: "The sources: Frontmatter Contract Linking Wiki to Raw"
context: brain
category: linking
concept: wiki-sources
description: "Only mechanism making wiki claims traceable to ground truth"
tags: sources, linking, frontmatter, traceability, wiki, raw
sources: []
last_ingested: 2026-04-16
---

## The sources: Contract

Every wiki concept carries a `sources:` array in its frontmatter. This is
the ONLY authoritative link from distilled knowledge back to the raw
material it came from. Direction is one-way: wiki points at raw, raw
never points at wiki.

**Incorrect (sources missing, or pointing elsewhere):**

```yaml
---
title: "Disaster hook formula"
context: content
category: hooks
concept: disaster-formula
sources: []                                 # empty: orphan; nothing supports the claims
---
```

OR:

```yaml
sources:
  - "https://karpathy.ai/wiki/"             # external URL, not a raw file
  - "../other-skill/references/raw/x.md"    # escapes the skill root
```

External URLs belong in a "Reference" link at the bottom of the body.
`sources:` is strictly for files under `references/raw/`.

**Correct:**

```yaml
---
title: "Disaster hook formula"
context: content
category: hooks
concept: disaster-formula
sources:
  - "references/raw/2026-04-12 linkedin analytics.csv"
  - "references/raw/karpathy-brain-idea.md"
  - "references/raw/Screenshot 2026-04-14 at 09.32.png"
last_ingested: 2026-04-14
---
```

## Rules

1. **Paths are relative to the skill root.** Always. `references/raw/...`
   never `../...` and never absolute.
2. **Always quoted.** Raw filenames may contain spaces, dots, emoji, or
   special characters. Quote every entry in the YAML list.
3. **Must resolve.** `scripts/lint.sh` flags broken paths. If a raw file
   is removed, prune its citations.
4. **Many-to-many is allowed.** One raw file may appear in `sources:` of
   N wiki concepts. One wiki concept may cite M raw files.
5. **Empty list is legal but flagged.** Pure synthesis concepts (derived
   from other wiki concepts, not from raw) may have empty `sources:`.
   Lint marks them as orphans for audit; they're not errors.
6. **Never edit raw to satisfy a wiki claim.** If a wiki concept
   over-claims what a raw file says, fix the wiki concept.

## Why This Asymmetry

Raw is immutable human territory. If the LLM could write back to raw,
the audit trail rots. Wiki is LLM territory, regenerable, validated by
lint. The `sources:` array lets any claim in the skill be traced to a
specific file the human owns, in one hop.

## Reverse Map

`references/_index.md` maintains the reverse map (which wiki concepts
cite a given raw file). This is derived, not hand-maintained - run
`scripts/lint.sh` to regenerate.

## Annotating Sources in the Body

When a concept cites multiple raw files, add a `## Sources` section
below the body explaining what each source contributed:

```markdown
## Sources

- `references/raw/analytics.csv` - impression/comment counts
- `references/raw/karpathy-brain-idea.md` - the hook-structure hypothesis
- `references/raw/Screenshot 2026-04-14.png` - post-level engagement rates
```

This enriches provenance for future ingests and makes lint contradiction
detection easier.

## Sources

- (none)
