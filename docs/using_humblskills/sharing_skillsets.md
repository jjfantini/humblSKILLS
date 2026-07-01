# Sharing skillsets

A **skillset** is a small, version-controlled manifest (default `humblskills.json`) that lists the skills a project or team wants installed. Commit it to a repo and every teammate runs `humblskills sync` to land the same set - no shared registry account, no manual skill-by-skill setup.

```json
{
  "schema_version": 1,
  "skills": [
    { "name": "use-smart-commit", "version": "1.0.3" },
    { "name": "use-worktree-flow", "version": "0.4.0" }
  ]
}
```

`version` is informational only (the version captured when the file was written); `sync` always installs whatever the registry currently ships for that skill, matching `install` semantics.

## Create a skillset

```sh
humblskills init                     # scaffold an empty ./humblskills.json to fill in
humblskills init --from-installed    # scaffold it from the skills you already have
humblskills export -o humblskills.json   # snapshot your currently installed skills
```

`init` refuses to overwrite an existing file unless you pass `--force`. `export` always overwrites the target path. Both write a **sorted, pretty-printed** file for stable, diff-friendly commits.

## Install from a skillset

```sh
humblskills sync                                        # install missing skills from ./humblskills.json
humblskills sync path/to/set.json --force                # reinstall everything from a specific file
humblskills sync https://example.com/humblskills.json    # sync from a hosted skillset
humblskills sync --prune                                  # also uninstall skills not in the file
```

`sync` accepts a local path, a `file://` URL, or an `http(s)://` URL, so a team can host one canonical skillset (for example, alongside its docs site) and everyone runs `humblskills sync https://…/humblskills.json`. Remote fetches are capped at 1 MiB and time out after 15 seconds.

Skills already installed and up to date are skipped; pass `--force` to reinstall them anyway. A skill listed in the skillset that the registry doesn't know about is reported as a warning, not a hard failure - the rest of the sync still runs.

### Keep a local set in sync exactly (`--prune`)

By default `sync` only **adds** skills. Pass `--prune` to also **remove** any locally installed skill that the skillset doesn't list, so your machine ends up matching the file exactly:

```sh
humblskills sync --prune
```

Pruning is destructive, so it asks for confirmation (skip with `--yes`, or run with `--json` for a machine-readable summary instead of a prompt).

### Platforms and scope

`export`, `init`, and `sync` follow the same platform/scope rules as `install`: explicit `--platform`/`--scope`/`--global` flags win, otherwise your [profile](../getting_started/quickstart.md) defaults apply.

## Related topics

- [Registry & skill format](registry_and_format.md)
- [Preserving user content](preserving_user_content.md)
- [Quickstart](../getting_started/quickstart.md)
