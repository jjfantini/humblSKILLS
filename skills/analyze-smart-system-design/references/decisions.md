# Decisions

Reasoning memory. Each entry records a non-obvious choice: the context,
the options considered, what was chosen, why, and the observed result.
Never delete entries - if a decision is reversed, add a new entry that
references the old one.

Entry shape:

```
### <YYYY-MM-DD> | <short title>
- Context: <the situation that required a choice>
- Options: (A) <opt>, (B) <opt>, (C) <opt>
- Chose: <letter and name>
- Why: <the rationale, ideally citing evidence>
- Result: <what happened after, or "TBD">
```

---

### 2026-06-01 | Wiki taxonomy: six context buckets
- Context: migrating 60 Hello Interview raw pages plus 6 whiteboards into the wiki, needed a stable `context/category` taxonomy the lint derives from the filesystem.
- Options: (A) one flat context, (B) mirror the raw URL folders (in-a-hurry/core-concepts/patterns/deep-dives/problem-breakdowns), (C) intent-based buckets reflecting how a candidate reaches for material.
- Chose: (C) six contexts - `delivery` (framework), `concepts` (network/data/scaling/theory), `patterns` (core), `tech` (core/advanced), `breakdown` (core), `examples` (whiteboards).
- Why: routing in SKILL.md maps cleanly to the 6 framework steps; concepts group by topic rather than by source page; `tech` splits core stores from niche/advanced. All contexts are lowercase, 3-10 chars, hyphen-free per `_brain.md`.
- Result: 66 wiki concepts; SKILL.md "How to Use" points each framework step at a context/category.

### 2026-06-01 | Text-only raw: assumed-flow notes for omitted diagrams
- Context: the staged markdown captured prose but not the Excalidraw architecture diagrams; every problem breakdown's high-level design lived in an image we could not ingest.
- Options: (A) omit the high-level design, (B) describe it as if the diagram were present, (C) reconstruct a plausible flow from the prose and clearly label it as inferred.
- Chose: (C) - each breakdown's high-level design carries a blockquote `> Assumed flow (diagram in source omitted; inferred): ...` reconstructed from the written walkthrough.
- Why: preserves the design signal without fabricating that we saw the diagram; the framing is honest and consistent so future passes can verify against the raw prose or a re-captured image.
- Result: all 29 breakdowns plus 5 whiteboard examples cite their raw page (and the whiteboard PNG where one exists); whiteboard pngs use the actual filenames (e.g. `whatsapp_chat.png`).

### 2026-06-01 | Ingest user whiteboards for 7 full final designs
- Context: user supplied 7 reference whiteboard images (payment, metrics, job scheduler, youtube, youtube top-k, auction, web crawler) to replace inferred "assumed flow" notes with actual diagram-backed summaries.
- Options: (A) whiteboard PNGs only in raw/, (B) whiteboard wiki + update paired breakdown concepts, (C) replace Hello Interview assumed-flow entirely without keeping prose sources.
- Chose: (B) - PNGs in `raw/whiteboards/`, new `wiki/examples/whiteboards/*` concepts, paired `breakdown/core/*` updated to cite PNGs and document deep-dive boxes from diagrams.
- Why: keeps Hello Interview prose as secondary source while giving the skill honest diagram-backed flows; separates simplicity-bar sketches from full final designs in `overview.md`.
- Result: 7 new PNGs, 7 new whiteboard wiki concepts, 7 breakdown updates, `overview.md` index; lint TBD.
