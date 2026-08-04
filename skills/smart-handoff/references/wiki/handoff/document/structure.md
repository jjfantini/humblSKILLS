---
title: "Handoff Document Structure and File Naming"
context: handoff
category: document
concept: structure
description: "A fixed section order means the receiving agent can skim to what it needs instead of reading prose"
tags: document, structure, naming, template, sections
sources: []
last_ingested: 2026-08-03
---

## File Name

```
<descriptive-name>-handoff-<YYYY-MM-DD>.md
```

`<descriptive-name>` is a kebab-case slug naming the *work*, not the session:
`oauth-token-refresh`, `registry-private-install`, `flaky-parser-test`.
Never `session`, `context`, `notes`, or `handoff` — the word `handoff` is
already in the name.

```
oauth-token-refresh-handoff-2026-08-03.md      correct
handoff-2026-08-03.md                          no slug, collides
session-notes-2026-08-03.md                    wrong shape
oauth_token_refresh_handoff_2026_08_03.md      wrong separators
```

## Section Order (Fixed)

The skeleton lives at `assets/handoff-template.md`. Copy it, fill it, delete
sections that are genuinely empty — but do not reorder, because the order is
what makes the doc skimmable.

| # | Section | Purpose |
|---|---------|---------|
| 1 | Title + lifecycle line | temp vs persist, date, from → to harness |
| 2 | Objective | one paragraph, what the next session must accomplish |
| 3 | Repository State | repo, remote, branch, HEAD, dirty/unpushed counts |
| 4 | What Is Already Done | `[verified]` / `[assumed]` labelled |
| 5 | Next Steps | ordered, each naming a file |
| 6 | Environment & Commands | exact build / test / lint commands |
| 7 | Artifacts | paths and URLs — links only, never contents |
| 8 | Access & Secrets | retrieval instructions, never values |
| 9 | Suggested Skills | relevant skills + install commands |
| 10 | Open Questions | what blocks which step |
| 11 | Gotchas & Dead Ends | what was tried and failed, so it isn't retried |
| 12 | Pickup Prompt | paste-ready paragraph pointing at this file's path |

## The No-Duplication Rule

A handoff is an **index plus the reasoning that lives nowhere else**. Anything
already captured in a durable artifact gets referenced, not restated.

### Incorrect

```markdown
## What Is Already Done
The design doc says we chose PostgREST over a custom API layer because
row-level security is enforced from the token, which means... [400 words
pasted from docs/adr/0007-postgrest.md]
```

### Correct

```markdown
## Artifacts
| Artifact | Location | Why it matters |
|----------|----------|----------------|
| PostgREST ADR | `docs/adr/0007-postgrest.md` | binding decision on the data layer |
| Open PR | https://github.com/org/repo/pull/412 | the diff under review |
```

## Length Ceiling

Aim under 150 lines. A handoff longer than the code it describes is a sign
that artifact contents leaked in. Cut back to links.

## Sources

- (none) - authored from the smart-handoff design session.
