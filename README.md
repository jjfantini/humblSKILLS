# humblSKILLS

A personal skill registry and a single-binary Go CLI (`humblskills`) that
installs [agentskills.io](https://agentskills.io)-format skills into whichever
agent platform you use — Claude Code, Cursor, Codex, and friends.

## What's in this repo

1. **Skill registry** — a monorepo of agent skills authored in the
   agentskills.io format with light humblSKILLS frontmatter extensions
   (`requires`, `platforms`, `post_install`, `tags`).
2. **`humblskills` CLI** — fetches a skill directory and drops it in the right
   place for your agent platform. Zero servers, zero accounts, zero telemetry.

## Install

### Shell installer (Linux/macOS)

```sh
curl -fsSL https://raw.githubusercontent.com/jjfantini/humblSKILLS/main/scripts/install.sh | sh
```

Installs to `/usr/local/bin` by default (uses `sudo` if needed). Override
with `INSTALL_DIR`:

```sh
curl -fsSL https://raw.githubusercontent.com/jjfantini/humblSKILLS/main/scripts/install.sh | INSTALL_DIR=$HOME/.local/bin sh
```

Pin a specific version with `VERSION=0.1.0 sh`.

### Go

```sh
go install github.com/jjfantini/humblSKILLS/cli/cmd/humblskills@latest
```

### Direct download

Grab the archive for your platform from the
[releases page](https://github.com/jjfantini/humblSKILLS/releases/latest):

- `humblskills_<version>_linux_amd64.tar.gz`
- `humblskills_<version>_linux_arm64.tar.gz`
- `humblskills_<version>_macos_amd64.tar.gz`
- `humblskills_<version>_macos_arm64.tar.gz`
- `humblskills_<version>_windows_amd64.zip`
- `humblskills_<version>_windows_arm64.zip`

Each release also publishes `checksums.txt` with SHA-256 sums.

### Homebrew (Linux/macOS)

```sh
brew install jjfantini/humbl/humblskills
```

Formulas live in [`jjfantini/homebrew-humbl`](https://github.com/jjfantini/homebrew-humbl)
and are bumped automatically by the release workflow.

## Quickstart

```sh
humblskills doctor                    # verify the environment
humblskills search                    # browse the registry
humblskills install skill-example-hello
humblskills list
humblskills update                    # pick which drifted skills to upgrade
humblskills update --all --yes        # non-interactive bulk upgrade
humblskills uninstall skill-example-hello
```

Every command accepts `--json` for machine-readable output and `--yes` to
skip confirmation prompts.

## Developing the CLI

The CLI source lives under [`cli/`](cli) as a nested Go module.

```sh
make build           # builds ./bin/humblskills
make test            # runs go test ./...
make registry        # regenerates registry.json from skills/ + adapters/
make sync-adapters   # mirrors adapters/ into the CLI's embed directory
```

Releases are cut by pushing a semver tag (e.g. `git tag v0.1.0 && git push
origin v0.1.0`); the workflow in [`.github/workflows/release.yml`](.github/workflows/release.yml)
runs GoReleaser, uploads archives + checksums to GitHub Releases, and also
pushes a sibling `cli/v0.1.0` tag so `go install` works against the nested
module.

## License

Content is licensed under [CC-BY-4.0](LICENSE). If Go source code licensing
becomes a concern later, the CLI code under `cli/` may be dual-licensed MIT —
but that has not been done yet.
