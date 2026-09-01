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

Tap and formulas live in [`jjfantini/homebrew-humbl`](https://github.com/jjfantini/homebrew-humbl). **Stable** releases on `main` bump `Formula/humblskills.rb`. Pre-releases from `develop` (`vX.Y.Z-pre.N`) bump a **second** formula, `humblskills-pre`, and never replace the stable one — `brew upgrade humblskills` stays on the last real version.

Homebrew rejects `humblskills@beta` (`@` is only legal with a numeric version). The pre formula is therefore `humblskills-pre`.

```sh
brew install jjfantini/humbl/humblskills-pre   # pre-releases only
brew upgrade humblskills-pre
```

The two formulas `conflicts_with` each other. Install one, not both.

Upgrade the stable formula later with:

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
humblskills upgrade                    # self-update on the profile channel (stable by default)
humblskills upgrade --channel beta     # this run only; does not write the profile
humblskills upgrade --dry-run          # just show the version you'd move to
```

The install channel is one field in `~/.humblskills/profile.json` (`channel`:
`stable` or `beta`; unset means stable). Homebrew and `humblskills upgrade`
both read it — no extra config file.

```sh
humblskills profile get channel
humblskills profile set channel beta     # persist; default / "" / stable all mean stable
humblskills profile show
```

Or the same TUI editor that already exists (do not look for a second settings
app): run `humblskills` and open **Profile**, or `humblskills profile`. The
**install channel** row shows the current value; enter to switch stable ↔ beta.

| Channel | GitHub | Homebrew |
|---------|--------|----------|
| `stable` (default) | `/releases/latest` only | `brew upgrade humblskills` |
| `beta` | higher semver of latest stable vs latest prerelease | formula follows the winner (see below) |

**beta** means “always the newest version,” not “prereleases only.” After a
stable graduates (`v2.52.0` > `v2.52.0-pre.1`) beta picks the stable. A newer
pre (`v2.53.0-pre.1` > `v2.52.0`) stays on the pre.

On a Homebrew-managed install, `upgrade` detects that, asks for confirmation,
and runs brew itself so Cellar bookkeeping stays correct:

- Winner is a pre, already on `humblskills-pre`: `brew upgrade humblskills-pre`
- Winner is a stable, currently on `humblskills-pre`: `brew uninstall humblskills-pre && brew install humblskills`
- Winner is a stable, already on `humblskills`: `brew upgrade humblskills`

It only asks you to run brew yourself if `brew` isn't on `PATH`. You can of
course still do it directly:

```sh
brew upgrade humblskills                                          # already on the stable formula
brew upgrade humblskills-pre                                      # beta, winner is still a pre
brew uninstall humblskills-pre && brew install humblskills        # beta, winner is a graduated stable
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
