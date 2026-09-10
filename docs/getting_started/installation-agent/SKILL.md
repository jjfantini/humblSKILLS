---
name: humblskills-cli
description: Install the humblskills CLI on Linux, macOS, or Windows, then use registry, install, and non-interactive flags in scripts and agents.
compatibility: Requires a supported platform build (see releases). Shell installer targets Unix-like systems; Windows uses release archives.
---

# humblskills CLI (install and use)

Use this skill when the user needs **humblskills** on their machine, in CI, or when an agent must run **`humblskills` with `--json` / `--yes`** instead of the TUI.

Canonical human docs (HTML): [Installation](https://jjfantini.github.io/humblSKILLS/getting_started/installation/) and [Quickstart](https://jjfantini.github.io/humblSKILLS/getting_started/quickstart/).

## Install

### Homebrew (recommended on Linux and macOS)

```sh
brew install jjfantini/humbl/humblskills
brew upgrade humblskills
```

Pre-releases use a second formula (`humblskills-pre`). Do not use
`humblskills@beta` — Homebrew only allows `@` + digits. The stable formula is
never rewritten on a `-pre.N` tag.

```sh
brew install jjfantini/humbl/humblskills-pre
brew upgrade humblskills-pre
```

Tap: [jjfantini/homebrew-humbl](https://github.com/jjfantini/homebrew-humbl).

### Shell installer (Linux and macOS)

For scripted installs or when not using Homebrew:

```sh
curl -fsSL https://raw.githubusercontent.com/jjfantini/humblSKILLS/main/scripts/install.sh | sh
```

Optional:

```sh
curl -fsSL https://raw.githubusercontent.com/jjfantini/humblSKILLS/main/scripts/install.sh | INSTALL_DIR=$HOME/.local/bin sh
curl -fsSL https://raw.githubusercontent.com/jjfantini/humblSKILLS/main/scripts/install.sh | VERSION=2.45.0 sh
```

### Go

```sh
go install github.com/jjfantini/humblSKILLS/cli/v2/cmd/humblskills@latest
```

### Direct download (including Windows)

Download the archive for your OS and architecture from [GitHub releases](https://github.com/jjfantini/humblSKILLS/releases/latest). Artifacts follow the pattern `humblskills_<version>_<os>_<arch>.<tar.gz|zip>`. Verify with `checksums.txt` in the release.

### Verify

```sh
humblskills doctor
```

### Upgrade the CLI later

```sh
humblskills upgrade                 # profile channel (stable if unset)
humblskills upgrade --channel beta  # this run only
humblskills profile get channel
humblskills profile set channel beta
brew upgrade humblskills                                          # stable formula
brew upgrade humblskills-pre                                      # beta, winner is still a pre
brew uninstall humblskills-pre && brew install humblskills        # beta, winner is a graduated stable
```

On a brew install, `upgrade` runs the brew commands for the resolved channel
winner (beta = newer of latest stable vs latest prerelease). Same `channel`
field as `profile.json`. In the TUI: `humblskills` or `humblskills profile` →
**install channel**.

`upgrade` updates the CLI binary; `update` updates installed skills.
`humblskills install` installs **skills**, not the CLI.

## CLI behavior

- **Interactive TTY:** `humblskills` or `humblskills start` opens the dashboard. Use **`--fullscreen`** for full-screen TUI when supported.
- **Non-interactive (CI, pipes, agents):** no TUI; the binary prints a short summary. Use explicit subcommands plus **`--json`** and **`--yes`**.

### Core commands

```sh
humblskills doctor
humblskills search                       # --category=<c> / --role=<r> to narrow
humblskills install smart-skill          # --platform, --scope/--global, --from, --force
humblskills list
humblskills update                       # --check to dry-run
humblskills update --all --yes
humblskills uninstall smart-skill
```

Every command accepts **`--json`** (machine-readable output) and **`--yes`** (skip prompts).

### Platforms

Installs go to every detected platform: `claude-code`, `cursor`, `codex`
(symlinks into `~/.claude/skills`, `~/.cursor/skills`, `~/.agents/skills`) and
`claude-desktop`, which writes an upload zip to `~/.humblskills/desktop/` because
Claude Desktop and claude.ai can't read skills from disk. Restrict with
`--platform`, or persist a default with
`humblskills profile set platforms claude-code,cursor`.

### Private / additional registries

Naming **any** registry makes the CLI use **only** its named set — the hosted public
default stops being consulted, and `--registry` / `HUMBLSKILLS_REGISTRY` /
`profile set registry` are ignored. Because that replaces rather than adds, the **first**
`registry add` seeds whatever was already in effect (`public`, or `default` for a
`profile set registry` URL) and reports it. So you do not need to add `public` yourself;
for private-only, run `humblskills registry remove public` afterwards.

```sh
humblskills registry add work my-company/our-skills   # owner/repo shorthand; seeds public
humblskills registry login --name work                # token → OS keychain (or pipe it on stdin)
humblskills registry list                             # confirm what you ended up with
humblskills install <skill> --from work               # if a name exists in several registries
```

Tokens for named registries come from the keychain (`registry login --name`), with
`HUMBLSKILLS_TOKEN` as a shared fallback. The global `--token` and `--registry` flags
and `HUMBLSKILLS_REGISTRY` apply **only** when no registry is named. Details:
[Registries](https://jjfantini.github.io/humblSKILLS/using_humblskills/registries/#the-mental-model).

### Team skillsets

```sh
humblskills init --from-installed        # scaffold ./humblskills.json
humblskills export -o humblskills.json   # snapshot installed skills (+ their named registries)
humblskills sync                         # install everything the file lists
humblskills sync --prune --yes           # make the local set match the file exactly
```

## Deeper topics

- [Skill catalog](https://jjfantini.github.io/humblSKILLS/skills/)
- [Platforms and where skills land](https://jjfantini.github.io/humblSKILLS/using_humblskills/platforms/)
- [Registries](https://jjfantini.github.io/humblSKILLS/using_humblskills/registries/)
- [Updating skills](https://jjfantini.github.io/humblSKILLS/using_humblskills/updating/)
- [Registry and skill format](https://jjfantini.github.io/humblSKILLS/using_humblskills/registry_and_format/)
- [Preserving user content](https://jjfantini.github.io/humblSKILLS/using_humblskills/preserving_user_content/)
- [Sharing skillsets](https://jjfantini.github.io/humblSKILLS/using_humblskills/sharing_skillsets/)
- [Eval quickstart](https://jjfantini.github.io/humblSKILLS/eval/quickstart/)
