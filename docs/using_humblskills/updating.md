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
