# Installation

## Shell installer (Linux/macOS)

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

Grab the archive for your platform from the [releases page](https://github.com/jjfantini/humblSKILLS/releases/latest):

- `humblskills_<version>_linux_amd64.tar.gz`
- `humblskills_<version>_linux_arm64.tar.gz`
- `humblskills_<version>_macos_amd64.tar.gz`
- `humblskills_<version>_macos_arm64.tar.gz`
- `humblskills_<version>_windows_amd64.zip`
- `humblskills_<version>_windows_arm64.zip`

Each release publishes `checksums.txt` with SHA-256 sums.

## Homebrew (Linux/macOS)

```sh
brew install jjfantini/humbl/humblskills
```

Formulas live in [`jjfantini/homebrew-humbl`](https://github.com/jjfantini/homebrew-humbl) and are bumped by the release workflow.

## Verify

```sh
humblskills doctor
```

See [Quickstart](quickstart.md) for everyday commands.
