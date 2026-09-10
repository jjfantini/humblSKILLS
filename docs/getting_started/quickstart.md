# Quickstart

New to the CLI? Do these five things in order and you're done.

```sh
humblskills doctor                    # 1. is my environment OK?
humblskills search                    # 2. what can I install?
humblskills install smart-commit      # 3. install one
humblskills list                      # 4. what do I have?
humblskills update                    # 5. keep them current
```

That's the whole loop. Everything below is detail you can come back for.

!!! tip "Prefer clicking to typing?"
    Just run **`humblskills`** with no arguments — it opens a dashboard with a
    tile for every command. See [Interactive dashboard](#interactive-dashboard-tui).

## 1. Check your environment

```sh
humblskills doctor
```

`doctor` reports which agent platforms it detected (Claude Code, Cursor, Codex,
Claude Desktop), whether each install target is writable (`rw` / `ro`), the
health of every configured registry, and anything wrong with your install
manifest. If something later doesn't work, run this first.

## 2. Find a skill

```sh
humblskills search                    # no query → open the interactive browser
humblskills search commit             # match name, description, and tags
humblskills search --category=design  # narrow to one category
humblskills search --role=fde         # narrow to one target role
```

Or read the full [skill catalog](../skills.md).

## 3. Install it

```sh
humblskills install smart-commit
humblskills install smart-commit smart-skill     # several at once
humblskills install                              # no name → pick from a list
humblskills install smart-commit --yes           # skip the confirmation
humblskills install smart-commit --platform claude-code
humblskills install smart-commit --scope project # this repo only
```

Installing several at once is not a loop over single installs: the whole batch
shares one dependency resolution (so a dep two skills both need is fetched once),
one platform/scope prompt, and one progress screen. In the picker, **space** ticks
a row — the footer counts what you've picked — and **enter** (or `i`) installs
everything ticked. Naming a skill that isn't in any registry fails the batch
before anything is written, so you never end up half-installed.

By default a skill installs once into `~/.humblskills/skills/<skill>` and is
linked into every agent platform found on your machine. See
[Platforms & where skills land](../using_humblskills/platforms.md) for the paths,
scopes, and the Claude Desktop zip flow.

Save your preferences so you can stop passing flags:

```sh
humblskills profile set platforms claude-code,cursor
humblskills profile set scope global
humblskills profile show
```

## 4. See what you have

```sh
humblskills list          # installed skills, versions, and available updates
humblskills uninstall smart-commit
```

`list` tags each skill with the registry it came from and flags any that have a
newer version waiting.

## 5. Keep things current

```sh
humblskills update              # pick which drifted skills to upgrade
humblskills update --check      # dry run: show what would change
humblskills update --all --yes  # upgrade everything, no prompts
humblskills upgrade                  # CLI binary; profile channel (stable if unset)
humblskills upgrade --channel beta   # this run only
humblskills profile set channel beta # persist; TUI: humblskills → Profile → install channel
```

`update` handles skills; `upgrade` handles the binary. If a newer CLI build is
on your channel, `doctor` / `start` / `version` print
`newer version available: … — run \`humblskills upgrade\`` (or
`brew upgrade humblskills` / `humblskills-pre`) on stderr, and the dashboard
shows a **Newer version available** banner. `--json` stays machine-readable.

Your customizations
inside a skill survive an update, and renamed skills are followed automatically
— see [Updating skills](../using_humblskills/updating.md).

## Interactive dashboard (TUI)

On a normal terminal, **`humblskills`** with no subcommand opens the same
experience as **`humblskills start`**: a full-screen **dashboard** (tile grid
with fuzzy search) that routes into install, list, update, search, uninstall,
profile, eval, doctor, registry, and version. Press **ESC** from any sub-screen
to return to the grid.

```sh
humblskills          # TTY: open dashboard; non-TTY: print command summary
humblskills start    # always explicit
```

Optional global flag:

- **`--fullscreen`** - use full-screen TUI mode (also valid on `start`; requires a TTY).

Skill lists nest under collapsible **category** headings. Turn that off for one
flat list:

```sh
humblskills profile set group_by_category off
```

## Scripts, CI, and agents

Non-interactive environments (pipes, CI, agents) don't get the TUI — the binary
prints a short command summary instead. Use explicit subcommands plus:

- **`--json`** - machine-readable output
- **`--yes`** - skip confirmation prompts

```sh
humblskills install smart-commit --yes --json
humblskills update --all --yes --json
```

## Shell completions

Skill names, registry names, and flag values tab-complete once the completion
script is installed:

```sh
humblskills completion zsh --help    # also: bash, fish, powershell
```

## Command reference at a glance

| Command | What it does |
|---------|--------------|
| `humblskills` / `start` | Interactive dashboard |
| `doctor` | Check platforms, targets, registries, manifest |
| `search` | Find skills (`--category`, `--role`, `--json`) |
| `install` | Install a skill (`--platform`, `--scope`, `--global`, `--from`, `--force`) |
| `list` | Installed skills + available updates |
| `update` | Upgrade installed skills (`--check`, `--all`, `--force`) |
| `upgrade` | Upgrade the CLI binary (`--dry-run`) |
| `uninstall` | Remove a skill |
| `migrate` | Adopt hand-installed skills into humblskills |
| `registry` | Add / list / rename / remove registries, login, refresh |
| `profile` | Defaults: platforms, scope, registry (single-registry setups only), TUI options |
| `init` / `export` / `sync` | Share a skillset with a team |
| `export desktop` | Write Claude Desktop / claude.ai upload zips |
| `eval` | Benchmark a skill across arms |
| `mirrors` | Check mirrored skills against upstream (maintainers) |
| `completion` | Generate a shell completion script |
| `version` | Print the CLI version |

Add `--help` to any of them for the full flag list.

## Related topics

- [Skill catalog](../skills.md)
- [Platforms & where skills land](../using_humblskills/platforms.md)
- [Registries](../using_humblskills/registries.md) - private and multiple registries
- [Updating skills](../using_humblskills/updating.md)
- [Registry & skill format](../using_humblskills/registry_and_format.md)
- [Preserving user content](../using_humblskills/preserving_user_content.md)
- [Sharing skillsets](../using_humblskills/sharing_skillsets.md)
- [Eval quickstart](../eval/quickstart.md)
