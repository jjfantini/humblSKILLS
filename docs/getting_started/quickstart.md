# Quickstart

## Interactive dashboard (TUI)

On a normal terminal, **`humblskills`** with no subcommand opens the same experience as **`humblskills start`**: a full-screen **dashboard** (tile grid with fuzzy search) that routes into install, list, update, search, uninstall, profile, eval, doctor, registry refresh, and version. Press **ESC** from a sub-screen to return to the grid.

```sh
humblskills          # TTY: open dashboard; non-TTY: print command summary
humblskills start    # always explicit
```

Optional global flag:

- **`--fullscreen`** - use full-screen TUI mode (also valid on `start`; requires a TTY).

Non-interactive environments (pipes, CI, agents) do not get the TUI: the binary prints a short command summary instead. Use **`--json`**, **`--yes`**, and explicit subcommands (below) for scripts.

## Core commands (CLI)

```sh
humblskills doctor                    # verify the environment
humblskills search                    # browse the registry
humblskills install use-smart-skill
humblskills install use-smart-skill --global --yes
humblskills migrate claude-code --global --yes
humblskills list
humblskills update                    # pick which drifted skills to upgrade
humblskills update --all --yes        # non-interactive bulk upgrade
humblskills uninstall use-smart-skill
humblskills init --from-installed      # scaffold a shareable humblskills.json
humblskills sync                       # install everything a skillset lists
```

Use `install --global` when you want one canonical copy in
`~/.humblskills/skills/<skill>` with symlinks into every detected agent
platform. Codex reads the symlink from `$HOME/.agents/skills/<skill>`.

Use `migrate claude-code --global` to adopt existing registry-known skills from
`~/.claude/skills` into the canonical store, then fan them out with symlinks.
Unregistered personal Claude Code skills are reported and skipped.

## Machine-friendly output

Every command accepts:

- **`--json`** - machine-readable output
- **`--yes`** - skip confirmation prompts

Use these in scripts and CI.

## Related topics

- [Registry & skill format](../using_humblskills/registry_and_format.md)
- [Preserving user content](../using_humblskills/preserving_user_content.md)
- [Sharing skillsets](../using_humblskills/sharing_skillsets.md)
- [Eval quickstart](../eval/quickstart.md)
