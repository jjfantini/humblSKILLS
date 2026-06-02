---
title: "Relentless One-at-a-Time Interview Methodology"
context: claudeinit
category: interview
concept: methodology
description: "How to grill the user one question at a time, recommend an answer per question, walk the decision tree, and explore the codebase instead of asking when the answer is discoverable."
tags: interview, grill-me, questioning, decision-tree, recommended-answer
sources:
  - "references/raw/grill-me-skill.md"
  - "references/raw/user-brief.md"
last_ingested: 2026-06-02
---

## Relentless interview methodology

The quality of the generated `CLAUDE.md` is bounded by the quality of the
interview. A vague interview produces a vague file that the agent then ignores.
The job is to extract concrete, opinionated preferences — not to collect
shrugs. This is modelled on the `grill-me` skill: relentless, one question at a
time, until shared understanding is reached.

### The five rules

1. **One question per turn.** Never batch. Batching lets the user skim and
   answer the easy one while dropping the hard one. One question forces a real
   answer and lets the next question depend on the last.
2. **Recommend an answer for every question.** State your recommended default
   and a one-line reason, then ask. The user can accept with a word ("yes") or
   correct you. This is faster than open-ended prompts and surfaces your
   assumptions so they can be challenged.
3. **Walk the decision tree; resolve dependencies in order.** Project type
   gates everything. Stack gates testing and performance questions. Team size
   gates task-management questions. Ask the gating question first, then descend.
4. **Explore the codebase instead of asking when the answer is discoverable.**
   If a repo already exists, read `package.json` / `pyproject.toml` / `go.mod` /
   `Cargo.toml`, the test config, the CI files, the directory layout. Do not
   ask "what language is this?" when the manifest answers it. Ask only what the
   code cannot tell you — intent, philosophy, non-goals, preferences.
5. **Don't stop until every one of the 8 sections is resolved.** Track which
   sections are filled. Keep grilling the unfilled ones. The escape hatch is
   the user explicitly opting out (see below) — not the agent getting tired.

### Recommended-answer question shape

**Incorrect (open-ended, no recommendation, batched):**

```markdown
<!-- Forces the user to do all the thinking; invites a vague answer -->
What's your stack, how do you test, and what are your performance goals?
```

**Correct (one question, recommended default, reasoned):**

```markdown
<!-- One decision, a default to react to, a reason, room to correct -->
For the test runner I'd recommend **pytest** — it's the default for a Python
service this size and gives you fixtures + parametrize out of the box.
Is pytest right, or are you on unittest / nose / something else?
```

### The escape hatch

If the user says "use your best guesses", "just fill it in", "stop asking", or
similar, **stop interviewing immediately**. Fill every remaining section from
sensible defaults inferred from what you already know, and mark each inferred
choice inline so the user can scan and correct later:

```markdown
- Test runner: pytest  <!-- assumed: Python project default -->
```

Never hold the file hostage to more questions once the user has opted out.

### When to stop grilling a single section

A section is "resolved" when you could write its content without inventing a
preference. If you would have to guess a value the user cares about, ask one
more question. If the remaining unknowns are cosmetic, write it and move on —
the file is iterated later, not carved in stone.

## Sources

- `references/raw/grill-me-skill.md` — the one-at-a-time, recommend-an-answer,
  explore-the-codebase rules are taken directly from the grill-me skill.
- `references/raw/user-brief.md` — the "relentless until every section is
  resolved" requirement and the escape-hatch behaviour.
