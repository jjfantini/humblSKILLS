# Updating skills and the CLI

Two different things get updated, with two different commands. Mixing them up is
the most common confusion:

| You want to update… | Command |
|---------------------|---------|
| The **skills** you installed | `humblskills update` |
| The **`humblskills` binary** itself | `humblskills upgrade` |

## Updating installed skills

```sh
humblskills update                    # pick from the list of skills that drifted
humblskills update --check            # print what would change, change nothing
humblskills update --all --yes        # update everything, no prompts (CI-friendly)
humblskills update smart-commit       # just one skill
```

Run with no arguments on a normal terminal and you get a picker listing every
skill whose installed copy differs from what its registry now publishes. Nothing
is written until you confirm.

A skill counts as **drifted** when its version or its registry content hash has
changed. The humblSKILLS repo's own commit SHA is deliberately *not* used — it
moves on every commit, which would flag every skill as stale after each CLI
release.

`update` re-checks each skill against **the registry it was installed from**, so
a mixed public + private setup updates correctly.

### Your customizations survive

By default `update` keeps the files a skill declares in `metadata.preserve` —
memory files, curated notes, raw sources. This is what makes a "smart" skill's
accumulated knowledge survive an upgrade. See
[Preserving user content](preserving_user_content.md) for the full contract.

To throw your local changes away and take exactly what the registry ships:

```sh
humblskills update --force <skill>
```

`--force` is the one update mode that destroys content, so it now **lists the
user-owned files it would overwrite and asks you to confirm**. In a pipe, in CI,
or with `--json` there is nobody to ask, so it stops instead of guessing:

```console
$ humblskills update --force smart-commit --json
humblskills: update --force: overwrites user-owned files: smart-commit
(references/log.md, references/decisions.md) — re-run with --yes to confirm
```

The same gate covers `install --force`, `sync --force`, `sync --prune` and
`uninstall`. In the TUI, `f` on the update list and the **force reinstall**
toggle in the install modal reach the same behaviour, so neither surface can do
something the other can't.

## Adding a platform to skills you already have

Installed a new agent (Codex, say) and want your existing skills on it too? Add
it to your profile's `default_platforms`, then:

```sh
humblskills update --platforms          # add missing platforms + apply any drift
humblskills update --check --platforms  # preview, change nothing
```

Every platform target is a symlink into one canonical store, so covering a new
platform is a symlink plus a manifest entry. When the skill is already current
nothing is downloaded and **no file in the store is touched** — the summary says
`linked` rather than `installed` to make that explicit:

```console
$ humblskills update --check --platforms
1 skill to act on:
  smart-commit  link only  (+codex, no content change)
```

If a skill *is* drifted, the refresh and the new platform happen in one pass, so
you don't need two commands. Skills that declare a `platforms` allow-list are
only offered the platforms they support.

## When a skill gets renamed

Skills occasionally get renamed upstream (for example, the `use-*` family was
renamed to `smart-*`). You don't have to do anything special — `update` follows
the rename instead of skipping it:

The picker and the update summary show the rename explicitly, as
`use-smart-commit → smart-commit`, so you can see what's happening before you
confirm.

What happens:

1. The new skill is installed under its **new** name.
2. Your **preserved files are carried across** from the old install — memory,
   notes, and raw sources move with the skill.
3. The old installation is retired.

`humblskills update <new-name>` also reaches an install still recorded under the
old name, so you can update by the name you now know the skill by.

!!! note "Renames are never guessed"
    A rename is only followed when a registry skill explicitly claims the old
    name via `previous_names` in its frontmatter. A name that is still published
    in its own right is never treated as a rename target, and a name claimed by
    two skills is ignored rather than resolved by guessing. If a skill simply
    disappeared from the registry, `update` leaves your copy alone — there is
    nothing to upgrade to.

## Upgrading the CLI

```sh
humblskills upgrade              # download + verify + swap in the latest release
humblskills upgrade --dry-run    # just report the version you'd move to
```

The upgrade checks GitHub releases, verifies the checksum, and replaces the
running binary in place.

If you installed via Homebrew, `upgrade` tells you to use Homebrew instead, so
Homebrew's own bookkeeping stays correct:

```sh
brew upgrade humblskills
```

## Related topics

- [Preserving user content](preserving_user_content.md)
- [Registries](registries.md)
- [Quickstart](../getting_started/quickstart.md)
