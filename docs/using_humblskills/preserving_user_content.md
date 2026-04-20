# Preserving user content

Smart skills often accumulate user-owned content: raw sources, append-only memory (`log.md`, `decisions.md`, `patterns.md`), wiki pages, and so on. By default, `humblskills update` and `humblskills install` **overwrite** the skill directory with what the registry ships.

Skills that must keep user content across updates declare a **`preserve:`** list in `SKILL.md` frontmatter.

## Rules

Entries are **paths relative to the skill directory**. A trailing **`/`** means a **directory**; otherwise the entry is a **file**. Globs are not supported.

| Entry form | Example | Meaning on update |
|------------|---------|-------------------|
| **File** | `references/log.md` | User wins: your bytes survive the update. |
| **Directory** | `references/wiki/` | Deep merge: staging wins per file; files only on your side are kept. |

Fresh installs always seed everything from the registry. `preserve` applies when **replacing** an existing install. `humblskills uninstall` removes everything, including preserved paths.

### Example `SKILL.md` frontmatter

```yaml
---
name: my-smart-skill
description: ...
version: 0.2.0
preserve:
  - references/log.md
  - references/patterns.md
  - references/decisions.md
  - references/raw/
  - references/wiki/
---
```

Authors who ship a **preserve directory** should document that files the author ships inside that directory may still be overwritten on update (deep-merge contract).

## You own the preserve list after install

The registry’s `preserve:` list is the **seed** on first install. After that, **`humblskills update` reads `preserve:` from the installed `SKILL.md` on disk** (per platform and scope), not from upstream.

- Add an entry locally → that path survives the next update, even if upstream did not list it.
- Remove an entry locally → that path is overwritten from upstream on the next update.
- Clear `preserve:` → the next update does a full overwrite for paths that are not listed.

Only **`preserve:`** is treated as user-owned. Other frontmatter (`name`, `description`, `version`, `requires`, `platforms`, `tags`) and the markdown body are refreshed from upstream on update, while your `preserve:` list is carried forward.

### Freeze all of `SKILL.md`

To stop upstream from changing description, version, or body, add **`SKILL.md`** itself to `preserve:`. That is opt-in and stops those upstream updates until you remove it.

### Broken or missing frontmatter

If the installed `SKILL.md` is missing, unparseable, or has an invalid preserve list (for example path traversal with `..`), the engine falls back to the registry list and prints a warning.

YAML round-trip on update may normalize whitespace and drop comments inside the frontmatter block; keys and values stay intact.

## Clean reinstall

Bypass local preserve and match the registry exactly:

```sh
humblskills update --force <skill>
humblskills install --force <skill>
humblskills uninstall <skill> && humblskills install <skill>
```

`--force` ignores your local preserve edits and replaces the on-disk skill with the registry version.
