# Smart skills & the brain

A **smart skill** is a skill that remembers. Alongside its instructions it carries a **brain** — a self-learning memory of markdown files the agent reads before every task and writes to after. Instructions tell the agent what to do; the brain tells it what happened the last thirty times it did it.

The whole system is plain markdown on disk: traceable (every distilled claim cites its raw source), transparent (you can open and read every file), and improvable (results compound session over session). Its one rule: **no data, no improvement.**

## Anatomy: the CCCCC ontology

Every smart skill is organized in five layers:

| Layer | Role | Location |
|-------|------|----------|
| **Core** | Root structure of the skill | `SKILL.md`, `references/`, `scripts/`, `assets/` |
| **Context** | Top-level knowledge domain | first segment under `references/wiki/` |
| **Category** | Topic within a context | second segment |
| **Concept** | One atomic idea per file | `references/wiki/<context>/<category>/<concept>.md` |
| **Command** | Deterministic executable a concept can route to | `scripts/<name>.sh\|.py` |

The layers are loaded with **progressive disclosure**: the frontmatter description is always in the agent's context, the `SKILL.md` body loads when the skill is invoked, and everything under `references/` loads on demand per concept. The brain lives at that third level — memory costs nothing until it's consulted.

## Three territories

The brain divides `references/` into three ownership zones:

| Territory | Owner | Location | Contents |
|-----------|-------|----------|----------|
| **Raw** | Human | `references/raw/` | Unfiltered sources: articles, PDFs, transcripts, exports |
| **Wiki** | LLM | `references/wiki/<ctx>/<cat>/<concept>.md` | Distilled one-concept-per-file articles, each citing its raw sources |
| **Meta** | LLM | `references/{_index,patterns,decisions,log}.md` | Operational memory read every session |

The agent never edits or renames anything in `raw/` — that is human territory, the ground truth. Wiki concepts link back to it through a `sources:` array in their frontmatter, so every distilled claim has a causal chain to its origin.

A wiki that has moved ahead of its raw source is the **normal state of a living
skill**, not drift to reconcile. Raw is a point-in-time snapshot (and the
provenance baseline `sources:` points at); rewriting it to match a later fact
erases that citation. Read the gap as "wiki is current, snapshot is historical"
— for example, a model-tier concept may name today's model while
`references/raw/` still records the upstream skill that said yesterday's.

## The loop: read before, write after

This is the self-learning mechanism. Every session — even a simple query — runs it:

```
           ┌── READ (before any task) ──────────────────────┐
           │ 1. _index.md      what the skill already knows │
           │ 2. patterns.md    what worked, with numbers    │
           │ 3. decisions.md   past reasoning               │
           │ 4. log.md         last 5 session entries       │
           │ 5. wiki/…         concepts relevant to task    │
           └────────────────────────────────────────────────┘
                              do the work
           ┌── WRITE (after) ───────────────────────────────┐
           │ always            append session to log.md     │
           │ new material      raw/ → distilled wiki files  │
           │ quantified result append to patterns.md        │
           │ non-obvious call  append to decisions.md       │
           │ any wiki change   lint.sh regenerates _index   │
           └────────────────────────────────────────────────┘
```

The protocol block sits in `SKILL.md` *before* the usage instructions on purpose: by the time an agent reaches "when to use", it has already decided what to do — memory has to come first or it is never consulted.

That's the compounding: every session leaves at least a log line, every measured outcome leaves a pattern, every judgment call leaves a decision. After 30 sessions, the agent starts every task with 30 sessions of context.

## What keeps it honest

Two conventions stop the brain from decaying into a junk drawer:

- **`patterns.md` requires evidence.** Every entry carries Context / Approach / Result / Worked / Didn't / Lesson, with numbers in the result. A session with no measurable outcome writes only to `log.md` — patterns are never invented. Entries are append-only: a later contradiction gets a new entry referencing the old one, never an edit.
- **`scripts/lint.sh` is the health check.** It verifies that every wiki file's path matches its frontmatter, that no raw file is orphaned, that every `sources:` entry resolves, and flags stale or duplicate concepts. It also regenerates `_index.md` between sentinel markers, and the taxonomy itself is derived from the filesystem — there is no registry file to drift.

!!! note "The brain survives updates"
    `log.md`, `patterns.md`, `decisions.md`, `raw/`, and the wiki are declared in the skill's `preserve` list, so `humblskills update` ships new instructions without wiping accumulated memory. See [Preserving user content](using_humblskills/preserving_user_content.md).

## Measured, not claimed

The [eval harness](eval/index.md) runs smart skills in ordered sessions so brain state carries across runs, then ablates the brain against flat variants of the same skill. In the [adaptive-brand-voice-discovery showcase](eval/reports/smart-humanize-text/adaptive-brand-voice-discovery.md), `smart_skill` scored a **0.935** pass rate vs **0.740** with no skill (**+26.3%**) while using **67% fewer tokens**. The [4-arm ablation](eval/reports/smart-humanize-text/indie-launch-copy-iteration.md) isolates the brain itself from static wiki knowledge: on the two no-feedback sessions the brain-only delta was **+9.4%** and **+13.3%**, at **29% fewer tokens** than the flat skill.

## Where to go next

| If you want to… | Read |
|-----------------|------|
| Author a smart skill | install the [`smart-skill`](https://github.com/jjfantini/humblSKILLS/tree/main/skills/smart-skill) skill — it scaffolds the layout, brain, and lint |
| Read the full protocol spec | [`references/_brain.md`](https://github.com/jjfantini/humblSKILLS/blob/main/skills/smart-skill/references/_brain.md) inside `smart-skill` |
| Keep the brain across updates | [Preserving user content](using_humblskills/preserving_user_content.md) |
| Prove a skill compounds | [Eval overview](eval/index.md) · [published reports](eval/reports/index.md) |
