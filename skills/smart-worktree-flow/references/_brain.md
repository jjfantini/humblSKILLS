# Brain Spec

This skill uses the humblSKILLS smart-skill brain pattern: `SKILL.md` is a
router, `references/raw/` stores source material, `references/wiki/` stores
distilled concepts, and `references/{_index,patterns,decisions,log}.md` stores
operational memory.

## Read Before Any Task

Read, in order:

1. `references/_index.md`
2. `references/patterns.md`
3. `references/decisions.md`
4. Last 5 entries of `references/log.md`
5. Relevant `references/wiki/flow/<category>/` files

## Write After Any Task

- Append one session entry to `references/log.md`.
- Add quantified workflow outcomes to `references/patterns.md`.
- Add non-obvious decisions to `references/decisions.md`.
- When new source material arrives, keep it in `references/raw/` and cite it
  from wiki frontmatter.
- Run `scripts/lint.sh` after wiki changes to regenerate `_index.md`.

## Territory Rules

- Raw files are source material. Do not rename or rewrite them.
- Wiki files are distilled concepts owned by the skill.
- `_index.md` is generated between its sentinel markers only.

## Wiki Contract

Every concept must live at:

```text
references/wiki/<context>/<category>/<concept>.md
```

The frontmatter `context`, `category`, and `concept` fields must match that
path. `sources:` paths are relative to the skill root and must point under
`references/raw/`.
