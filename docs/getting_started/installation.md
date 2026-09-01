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

Tap and formula live in [`jjfantini/homebrew-humbl`](https://github.com/jjfantini/homebrew-humbl); **stable** releases on `main` bump the formula automatically. Pre-releases from `develop` (`vX.Y.Z-pre.N`) do not.

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
curl -fsSL https://raw.githubusercontent.com/jjfantini/humblSKILLS/main/scripts/install.sh | VERSION=2.45.0 sh
```

## Go

```sh
go install github.com/jjfantini/humblSKILLS/cli/v2/cmd/humblskills@latest
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

`doctor` prints the agent platforms it found, whether each install target is
writable, and the health of every configured registry. If it reports a problem,
fix that before installing skills.

## Staying up to date

```sh
humblskills upgrade              # self-update: download, verify, swap the binary
humblskills upgrade --dry-run    # just show the version you'd move to
```

On a Homebrew-managed install, `upgrade` detects that, asks for confirmation, and
runs `brew update && brew upgrade humblskills` for you, so Homebrew's own Cellar
bookkeeping stays correct. It only asks you to run brew yourself if `brew` isn't on
`PATH`. You can of course still do it directly:

```sh
brew upgrade humblskills
```

`upgrade` updates the **CLI**; `humblskills update` updates your installed
**skills**. See [Updating skills](../using_humblskills/updating.md).

## Optional: shell completions

Tab-complete skill names, registry names, and flag values:

```sh
humblskills completion zsh --help    # also: bash, fish, powershell
```

Each shell's `--help` prints the exact setup steps for that shell — for zsh and bash
that's a one-time command to run (writing the completion file), not a line to paste
into your shell config.

## Next

See [Quickstart](quickstart.md) for everyday commands, or jump straight to the
[skill catalog](../skills.md).
