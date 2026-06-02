---
title: "End-to-End CLAUDE.md Generation Workflow"
context: claudeinit
category: generate
concept: workflow
description: "The full pipeline: detect project type, interview to fill every section, substitute into the template, validate, write CLAUDE.md, and log the session."
tags: workflow, pipeline, generate, validate, write
sources:
  - "references/raw/user-brief.md"
  - "references/raw/grill-me-skill.md"
last_ingested: 2026-06-02
command: scripts/validate-claudemd.sh
---

## The pipeline

Six steps, in order. Steps 1-3 are agent judgment; step 4 is deterministic;
steps 5-6 finish the job.

1. **Detect project type.** Read the target directory for code signals (see
   `references/wiki/claudeinit/template/code-vs-general.md`). Pick
   `assets/claude-code.md.tmpl` (8 sections) or `assets/claude-general.md.tmpl`
   (4 sections). Confirm with the user only when ambiguous.

2. **Explore before asking.** If a repo exists, read the manifest, configs, CI,
   test setup, and layout. Pre-fill every answer the code already gives. This
   shortens the interview to only what the code cannot tell you.

3. **Interview to fill the gaps.** Grill the user one question at a time, with
   a recommended answer each, walking the decision tree until every section is
   resolved (see `references/wiki/claudeinit/interview/methodology.md` and
   `.../interview/question-bank.md`). Honour the "use your best guesses"
   escape hatch.

4. **Substitute into the template.** Read the chosen `.tmpl` from `assets/` and
   replace every `{{PLACEHOLDER}}` token with the resolved content. Remove any
   optional block the project doesn't need (e.g. a "no explicit budget"
   Performance section still gets a real sentence — never leave a placeholder).

5. **Validate, then write.** Run the validator against your drafted content
   **before** writing the file to the user's project:

   ```bash
   bash scripts/validate-claudemd.sh /path/to/draft-CLAUDE.md          # code mode
   bash scripts/validate-claudemd.sh --general /path/to/draft-CLAUDE.md # general mode
   ```

   It exits non-zero and lists the problem if a required section header is
   missing, a `{{...}}` placeholder survives, or a `TODO` is left behind. Fix
   and re-run until it exits 0, then write the final file to `CLAUDE.md` at the
   project root (or the path the user specified).

6. **Iterate and log.** Show the user the result and offer one refinement pass —
   a good CLAUDE.md is iterated, not one-shot. Then append a one-line entry to
   `references/log.md` recording the project type, sections filled, and any
   notable preference. If the user reported what worked, add a `patterns.md`
   entry; if you made a non-obvious template choice, add a `decisions.md` entry.

### Why validate before write

The user's `CLAUDE.md` is read by the agent every session. A leftover
`{{STACK}}` or a `TODO` silently degrades every future task in that project.
The deterministic check (step 5) is the guardrail that the interview was
actually completed — it is the same contract the `tests/run.sh` suite exercises.

### Don't

- Don't write the file with placeholders "to fill later" — that defeats the
  skill. Resolve or remove every token first.
- Don't skip the project-type detection and force eight code sections onto a
  research project.
- Don't one-shot and walk away — offer the refinement pass.

## Command

Validate a drafted CLAUDE.md before writing it to the user's project:

```bash
bash scripts/validate-claudemd.sh <file>            # 8-section code contract
bash scripts/validate-claudemd.sh --general <file>  # 4-section general contract
```

Exit 0 = ready to write. Non-zero = the listed section/placeholder/TODO must be
fixed first.

## Sources

- `references/raw/user-brief.md` — the detect / interview / fill / iterate
  shape and the validate-before-write quality bar.
- `references/raw/grill-me-skill.md` — the explore-before-asking and
  one-at-a-time interview steps.
