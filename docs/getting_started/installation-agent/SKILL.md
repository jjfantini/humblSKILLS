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

Tap: [jjfantini/homebrew-humbl](https://github.com/jjfantini/homebrew-humbl).

### Shell installer (Linux and macOS)

For scripted installs or when not using Homebrew:

```sh
curl -fsSL https://raw.githubusercontent.com/jjfantini/humblSKILLS/main/scripts/install.sh | sh
```

Optional:

```sh
curl -fsSL https://raw.githubusercontent.com/jjfantini/humblSKILLS/main/scripts/install.sh | INSTALL_DIR=$HOME/.local/bin sh
curl -fsSL https://raw.githubusercontent.com/jjfantini/humblSKILLS/main/scripts/install.sh | VERSION=0.1.0 sh
```

### Go

```sh
go install github.com/jjfantini/humblSKILLS/cli/cmd/humblskills@latest
```

### Direct download (including Windows)

Download the archive for your OS and architecture from [GitHub releases](https://github.com/jjfantini/humblSKILLS/releases/latest). Artifacts follow the pattern `humblskills_<version>_<os>_<arch>.<tar.gz|zip>`. Verify with `checksums.txt` in the release.

### Verify

```sh
humblskills doctor
```

## CLI behavior

- **Interactive TTY:** `humblskills` or `humblskills start` opens the dashboard. Use **`--fullscreen`** for full-screen TUI when supported.
- **Non-interactive (CI, pipes, agents):** no TUI; the binary prints a short summary. Use explicit subcommands plus **`--json`** and **`--yes`**.

### Core commands

```sh
humblskills doctor
humblskills search
humblskills install use-smart-skill
humblskills list
humblskills update
humblskills update --all --yes
humblskills uninstall use-smart-skill
```

Every command accepts **`--json`** (machine-readable output) and **`--yes`** (skip prompts).

## Deeper topics

- [Registry and skill format](https://jjfantini.github.io/humblSKILLS/using_humblskills/registry_and_format/)
- [Preserving user content](https://jjfantini.github.io/humblSKILLS/using_humblskills/preserving_user_content/)
- [Eval quickstart](https://jjfantini.github.io/humblSKILLS/eval/quickstart/)
