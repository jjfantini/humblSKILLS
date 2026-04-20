# Installation

## Using an AI agent (recommended)

Paste this into your coding agent (Claude Code, Cursor, Codex, or similar). It loads the published install + CLI [`SKILL.md`](https://jjfantini.github.io/humblSKILLS/getting_started/installation/SKILL.md) so the model can follow OS-specific steps and verify the binary.

```text
Read https://jjfantini.github.io/humblSKILLS/getting_started/installation/SKILL.md and install humblskills on this machine following those instructions. When finished, run humblskills doctor and fix anything it reports until it passes.
```

## Homebrew (Linux and macOS)

If you use [Homebrew](https://brew.sh), this is the simplest way to install and upgrade `humblskills` yourself in a terminal:

```sh
brew install jjfantini/humbl/humblskills
```

Tap and formula live in [`jjfantini/homebrew-humbl`](https://github.com/jjfantini/homebrew-humbl); new releases bump the formula automatically.

Upgrade later with:

```sh
brew upgrade humblskills
```

## Shell installer (Linux/macOS)

Use this when you do not use Homebrew, or for scripted installs (for example in CI):

```sh
curl -fsSL https://raw.githubusercontent.com/jjfantini/humblSKILLS/main/scripts/install.sh | sh
```

Installs to `/usr/local/bin` by default (uses `sudo` if needed). Override the destination with `INSTALL_DIR`:

```sh
curl -fsSL https://raw.githubusercontent.com/jjfantini/humblSKILLS/main/scripts/install.sh | INSTALL_DIR=$HOME/.local/bin sh
```

Pin a version (example: `0.1.0`):

```sh
curl -fsSL https://raw.githubusercontent.com/jjfantini/humblSKILLS/main/scripts/install.sh | VERSION=0.1.0 sh
```

## Go

```sh
go install github.com/jjfantini/humblSKILLS/cli/cmd/humblskills@latest
```

## Direct download

Grab the archive for your platform from the [releases page](https://github.com/jjfantini/humblSKILLS/releases/latest) (including **Windows** builds):

- `humblskills_<version>_linux_amd64.tar.gz`
- `humblskills_<version>_linux_arm64.tar.gz`
- `humblskills_<version>_macos_amd64.tar.gz`
- `humblskills_<version>_macos_arm64.tar.gz`
- `humblskills_<version>_windows_amd64.zip`
- `humblskills_<version>_windows_arm64.zip`

Each release publishes `checksums.txt` with SHA-256 sums.

## Verify

```sh
humblskills doctor
```

See [Quickstart](quickstart.md) for everyday commands.
