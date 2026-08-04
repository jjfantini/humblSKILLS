# Sharing skillsets

A **skillset** is a small, version-controlled manifest (default `humblskills.json`) that lists the skills a project or team wants installed. Commit it to a repo and every teammate runs `humblskills sync` to land the same set - no shared registry account, no manual skill-by-skill setup.

```json
{
  "schema_version": 1,
  "skills": [
    { "name": "smart-commit", "version": "1.0.3" },
    { "name": "smart-worktree-flow", "version": "0.4.0" }
  ]
}
```

`version` is informational only (the version captured when the file was written); `sync` always installs whatever the registry currently ships for that skill, matching `install` semantics.

## Skillsets can carry their registries

If any of your skills come from a **named** registry (a private company one, say),
`export` records that registry in the file so `sync` can configure it on the
other machine for you:

```json
{
  "schema_version": 1,
  "registries": [
    { "name": "work", "url": "https://raw.githubusercontent.com/my-company/our-skills/main/registry.json" }
  ],
  "skills": [
    { "name": "internal-deploy-runbook", "version": "0.3.0" },
    { "name": "smart-commit", "version": "1.0.3" }
  ]
}
```

On `sync`, for each listed registry the CLI:

1. **Adds it** to your profile if you don't have it (a registry you already have
   under that name keeps *your* URL — the CLI tells you and moves on).
2. **Checks it's readable**, and if it isn't and no token is stored, walks you
   through creating and storing one — the same path as
   [`registry login`](registries.md#private-registries-add-a-token).

A **single** unreachable registry is not fatal: `sync` warns, continues, and reports
that registry's skills as not-found. But if **every** configured registry fails to load,
`sync` aborts and installs nothing. Seeding makes that much harder to hit — a machine
syncing a skillset that names only a private registry still ends up with the public one
alongside it — but it is still reachable if every registry in play is unreadable.

When a skill name exists in more than one registry, `sync` prefers the order the
skillset lists them in, on the grounds that the file's author knows where their
skills live.

!!! note "Your teammate keeps their public catalog automatically"
    `export` only records **named** registries, on the assumption that every CLI already has
    the public default. That assumption stops holding the moment anything is named, because
    naming a registry replaces the default rather than adding to it (see
    [Registries](registries.md#the-mental-model)).

    `sync` bootstraps skillset registries through the same path as `registry add`, so the
    first one it adds seeds whatever the receiving machine was already using. A teammate with
    no named registries who syncs a skillset carrying only a private `work` registry ends up
    with **both** `work` and `public`, and sees `also kept "public", the registry you were
    already using` in the output.

    You can still list `public` explicitly if you want the skillset to be self-describing
    rather than relying on that:

    ```json
    "registries": [
      { "name": "public", "url": "https://raw.githubusercontent.com/jjfantini/humblSKILLS/main/registry.json" },
      { "name": "work",   "url": "https://raw.githubusercontent.com/my-company/our-skills/main/registry.json" }
    ]
    ```

    After syncing, `humblskills registry list` shows what you actually ended up with.

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

Pruning is destructive, so it asks for confirmation. Pass `--yes` to consent up front. In a pipe, in CI, or with `--json` there is nobody to ask, so it **stops with an error** naming what it would delete rather than proceeding — the same gate as `update --force`.

### Platforms and scope

`sync` follows the same platform/scope rules as `install`: explicit `--platform`/`--scope`/`--global` flags win, otherwise your [profile](../getting_started/quickstart.md) defaults apply. `init` and `export` only write the skillset file, so they take no platform or scope flags.

## Related topics

- [Registries](registries.md)
- [Platforms & where skills land](platforms.md)
- [Registry & skill format](registry_and_format.md)
- [Preserving user content](preserving_user_content.md)
- [Quickstart](../getting_started/quickstart.md)
