# Brain Spec

This skill uses the humblSKILLS Smart Skill pattern: a thin `SKILL.md` router,
distilled wiki concepts, append-only memory, and deterministic linting.

## Read Before Work

Read these files before using the skill:

1. `references/_index.md`
2. `references/patterns.md`
3. `references/decisions.md`
4. The last 5 entries in `references/log.md`
5. Relevant files under `references/wiki/`

## Write After Work

Always append a session entry to `references/log.md`.

When the task produces measured outcomes, append them to
`references/patterns.md`.

When the task involves a non-obvious approval, privacy, or routing choice,
append it to `references/decisions.md`.

When a wiki concept changes, run:

```bash
bash scripts/lint.sh
```

## File Roles

- `SKILL.md`: router and trigger guidance.
- `references/_index.md`: generated map of wiki concepts.
- `references/patterns.md`: measured lessons.
- `references/decisions.md`: durable routing and approval decisions.
- `references/log.md`: append-only session history.
- `references/wiki/`: distilled concepts.
- `references/raw/`: source material the operator provides.

Do not store API keys, cookies, account secrets, or raw session material in this
skill.
