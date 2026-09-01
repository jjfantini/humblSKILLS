# humblSKILLS

📖 **Full documentation:** [jjfantini.github.io/humblSKILLS](https://jjfantini.github.io/humblSKILLS/)

A personal skill registry and a single-binary Go CLI (`humblskills`) that
installs [agentskills.io](https://agentskills.io)-format skills into whichever
agent platform you use — Claude Code, Cursor, Codex, and friends.

## What's in this repo

1. **Skill registry** — a monorepo of agent skills authored in the
   agentskills.io format with light humblSKILLS frontmatter extensions under
   `metadata:` (`version`, `category`, `role`, `requires`, `platforms`, `tags`,
   `previous_names`, `preserve`).
2. **`humblskills` CLI** — fetches a skill directory and drops it in the right
   place for your agent platform. Zero servers, zero accounts, zero telemetry.

## Install

**Recommended:** send this to your agent so it loads the published install + CLI
`SKILL.md` and walks through setup on your machine (works from any OS the docs
cover):

```text
Read https://jjfantini.github.io/humblSKILLS/getting_started/installation/SKILL.md and install humblskills on this machine following those instructions. When finished, run humblskills doctor and fix anything it reports until it passes.
```

### Homebrew (Linux and macOS)

If you use [Homebrew](https://brew.sh), install and upgrade with:

```sh
brew install jjfantini/humbl/humblskills
brew upgrade humblskills
```

Pre-releases from `develop` (`vX.Y.Z-pre.N`) ship a **second** formula and never
replace the stable one (`humblskills@beta` is an illegal Homebrew class name):

```sh
brew install jjfantini/humbl/humblskills-pre
brew upgrade humblskills-pre
```

The CLI install channel is one field in `~/.humblskills/profile.json`
(`channel`: `stable` or `beta`; unset means stable). **beta** is always the
newest version (higher of latest stable vs latest prerelease) — not
prereleases only. After a stable graduates, a Homebrew `humblskills-pre`
install is switched with `brew uninstall humblskills-pre && brew install
humblskills`. Read or change the field with `humblskills profile get channel`
/ `profile set channel beta`, the existing Profile TUI (`humblskills` →
Profile → **install channel**), or a one-shot `humblskills upgrade --channel
beta`.

Formulas live in [`jjfantini/homebrew-humbl`](https://github.com/jjfantini/homebrew-humbl)
and are bumped by GoReleaser: `humblskills` on stable tags, `humblskills-pre`
on pre tags.

### Shell installer (Linux/macOS)

For machines without Homebrew, or for scripted installs:

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
go install github.com/jjfantini/humblSKILLS/cli/v2/cmd/humblskills@latest
```

### Direct download

Grab the archive for your platform from the
[releases page](https://github.com/jjfantini/humblSKILLS/releases/latest)
(including Windows):

- `humblskills_<version>_linux_amd64.tar.gz`
- `humblskills_<version>_linux_arm64.tar.gz`
- `humblskills_<version>_macos_amd64.tar.gz`
- `humblskills_<version>_macos_arm64.tar.gz`
- `humblskills_<version>_windows_amd64.zip`
- `humblskills_<version>_windows_arm64.zip`

Each release also publishes `checksums.txt` with SHA-256 sums.

## Quickstart

In a terminal, run **`humblskills`** or **`humblskills start`** to open the
interactive **dashboard** (tile grid with fuzzy search into every command).
Use explicit subcommands below for scripts, CI, or non-TTY environments.

```sh
humblskills doctor                    # verify the environment
humblskills search                    # browse the registry (--category=, --role=)
humblskills install smart-skill
humblskills install smart-skill smart-commit   # several at once, one shared dep resolution
humblskills install                   # picker: space picks several, enter installs them all
humblskills list
humblskills update                    # pick which drifted skills to upgrade
humblskills update --all --yes        # non-interactive bulk upgrade
humblskills upgrade                   # upgrade the CLI binary itself
humblskills uninstall smart-skill
humblskills export                    # snapshot installed skills to humblskills.json
humblskills sync                      # install everything in humblskills.json
humblskills export desktop            # zips to upload at claude.ai / Claude Desktop
```

Full skill catalog, generated from `registry.json`:
[jjfantini.github.io/humblSKILLS/skills](https://jjfantini.github.io/humblSKILLS/skills/).

### Platforms

Installs land in one canonical store (`~/.humblskills/skills/<skill>` by default)
and are exposed per platform. `claude-code`, `cursor`, and `codex` get symlinks
(`~/.claude/skills`, `~/.cursor/skills`, `~/.agents/skills`); **`claude-desktop`**
gets an upload **zip** in `~/.humblskills/desktop/`, because Claude Desktop and
claude.ai can't read skills off the filesystem — upload it at *Settings →
Capabilities → Skills*, and re-upload after an update. Restrict targets with
`--platform`, or persist a default with `humblskills profile set platforms
claude-code,cursor`. Details:
[Platforms](https://jjfantini.github.io/humblSKILLS/using_humblskills/platforms/).

### Private registries

Point the CLI at any registry, and — when it's backed by a **private** GitHub repo —
give it a token (a GitHub PAT with read access). The token authenticates both the
`registry.json` fetch and each skill download, sent as a `Bearer` header; it's ignored
for `file://` registries and the public default.

**Durable, no env vars (recommended).** Persist the registry URL in your profile and
store the token in your OS keychain once — then plain `humblskills` commands just work:

```sh
humblskills profile set registry https://raw.githubusercontent.com/<owner>/<repo>/main/registry.json
humblskills registry login          # prompts for the token (masked); stored in the OS keychain
humblskills search                  # uses the saved registry + token automatically
humblskills install <skill>
humblskills registry logout         # remove the stored token
```

`registry login` also accepts the token via `--token <t>` or piped stdin
(`echo "$TOK" | humblskills registry login`), and falls back to a `0600` file if no
keychain is available.

**Or ad-hoc / CI**, via flags or env vars. Token precedence: `--token` → `HUMBLSKILLS_TOKEN`
→ keychain → file. Registry precedence **in single-registry mode only** (no named registries
configured): `--registry` → `HUMBLSKILLS_REGISTRY` → `profile set registry` → hosted default.
Once you configure named registries, `--registry` and `HUMBLSKILLS_REGISTRY` are ignored — see
[Multiple registries](#multiple-registries-at-once) below.

```sh
export HUMBLSKILLS_REGISTRY=https://raw.githubusercontent.com/<owner>/<repo>/main/registry.json
export HUMBLSKILLS_TOKEN=<github token with read access>
humblskills search
```

### Multiple registries at once

Register several registries and `search`/`browse` show them **together, grouped by
registry** (each skill tagged with its source). Each registry keeps its own token.

> **Naming a registry replaces the default — but the CLI keeps the old one for you.** As soon
> as one named registry exists in your profile, the CLI uses **exactly** that set for every
> command: the hosted public default is no longer consulted, and `--registry`/`HUMBLSKILLS_REGISTRY`
> are ignored. Because that replaces rather than adds, the **first** `registry add` seeds whatever
> you were already using — `public`, or `default` pointing at your `profile set registry` URL —
> and tells you it did. Private-only is still available: `humblskills registry remove public`.

```sh
# Shorthand: pass owner/repo (or a github.com URL) — it expands to the raw
# registry.json URL. owner/repo@branch selects a branch.
humblskills registry add happyrobot jenningsfantini-happyrobot/happySKILLS   # public is seeded automatically
humblskills registry add                       # no args → prompts for name, URL, and (optional) token
humblskills registry login --name happyrobot   # token for the private one; verifies it can read the registry
humblskills registry list                      # show configured registries + token state
humblskills registry rename happyrobot hr      # rename (moves its stored token too)
humblskills registry add happyrobot <new-url>  # re-add with an existing name to change its URL
humblskills search                             # grouped: "── happyrobot ──" then "── public ──"
humblskills install <skill>                    # resolves to whichever registry has it
humblskills install <skill> --from public      # disambiguate when a name is in several registries
humblskills registry remove happyrobot         # drop it (and its stored token)
```

Tab-completion (skill names, registry names, `--from`/`--name`) works after installing the
completion script — e.g. `humblskills completion zsh` (see `humblskills completion --help`).

Run bare **`humblskills registry`** (or the dashboard's **registry** tile) to open the
interactive **registry manager** — add, rename, login, logout, remove, and refresh from
one screen (falls back to `registry list` when not on a TTY).

Configured registries also show up in `humblskills profile show` (under **Registries**).
`list` tags each installed skill with the registry it came from, `update` checks each
skill against **its** registry, and `doctor` reports each registry's reachability, token,
and skill count. When none are configured, everything falls back to the single registry
above (`--registry`/`HUMBLSKILLS_REGISTRY`/`profile set registry`/hosted default).

Beginner-friendly walkthrough:
[Registries](https://jjfantini.github.io/humblSKILLS/using_humblskills/registries/).

### Sharing skill sets across a team

`humblskills export` snapshots the skills you have installed into a
`humblskills.json` file (override with `-o`). Commit it to a repo, and every
teammate runs `humblskills sync` to install the same set — a single,
version-controlled source of truth for "which skills does this project want".

```sh
humblskills init                         # scaffold an empty ./humblskills.json to fill in
humblskills init --from-installed        # scaffold it from the skills you already have
humblskills export -o humblskills.json   # write the skillset
humblskills sync                         # install missing skills from ./humblskills.json
humblskills sync path/to/set.json --force  # reinstall everything from a specific file
humblskills sync https://example.com/humblskills.json  # sync from a hosted skillset
humblskills sync --prune                 # also uninstall skills not in the file
```

`init` bootstraps a new skillset file (default `./humblskills.json`); it writes
an empty set to fill in, or seeds it from your installed skills with
`--from-installed`, and refuses to clobber an existing file unless you pass
`--force`.

`sync` accepts a local path, a `file://` URL, or an `http(s)://` URL, so a team
can host one canonical skillset and everyone runs
`humblskills sync https://…/humblskills.json`. It pulls the current registry version of each skill (like `install`),
skips skills already up-to-date, and warns (without failing) about any skill in
the file that the registry doesn't know about. Add `--prune` to make your local
set match the file exactly — any installed skill the skillset doesn't list is
uninstalled (you're asked to confirm unless `--yes`). Platforms/scope follow the
same rules as `install`.

If any of your skills come from a **named** registry, `export` records that
registry in the file too (under `"registries"`), and `sync` on another machine
adds it to that user's profile and walks them through a token if it's private.
Registries you already have keep your own URL. Everything there is non-fatal — an
unreachable registry degrades to per-skill "not found" warnings.

Every command accepts `--json` for machine-readable output and `--yes` to
skip confirmation prompts.

## Benchmarking skills: `humblskills eval`

`eval` runs an up-to-four-arm benchmark of any skill — `no_skill` vs
`flat_skill` vs `flat_skill_wiki` vs `smart_skill` — grades the outputs, and
emits a single-file HTML dashboard. For smart skills the same harness runs
sessions in order so the brain state (patterns, decisions, log, wiki)
carries across sessions and you get a longitudinal trajectory that proves
the skill compounds over time. Drop `flat_skill_wiki` if you only want the
3-arm baseline; include it to separate "brain value" from "static wiki value"
in an ablation.

**Latest published showcase:** [adaptive-brand-voice-discovery · 2026-04-20](https://jjfantini.github.io/humblSKILLS/eval/reports/) — a 6-session compounding scenario over 10 idiosyncratic brand-voice rules. On cursor-agent, `smart_skill` scored pass_rate **0.935** vs `no_skill` **0.740** (**+26.3%**) and `flat_skill` **0.679** (**+37.7%**), while using **67% fewer tokens** than `no_skill`. Reproduce locally with `humblskills eval brand-voice`. Full index: [live docs](https://jjfantini.github.io/humblSKILLS/eval/reports/) · [source](docs/eval/reports/).

**4-arm ablation showcase:** [indie-launch-copy-iteration · 2026-04-27](https://jjfantini.github.io/humblSKILLS/eval/reports/smart-humanize-text/indie-launch-copy-iteration/) — 6 sessions over 13 indie-launch voice rules, three runs per arm (72 sessions total) on cursor-agent. Separates **brain value** (`smart_skill` vs `flat_skill_wiki`) from **wiki value** (`flat_skill_wiki` vs `flat_skill`) with identical preamble and scaffolding. `smart_skill` is the only arm above 0.9 pass rate on **both** no-feedback sessions (S5 0.922, S6 0.944); the brain-only delta over `flat_skill_wiki` is **+9.4% on S5 and +13.3% on S6**, and it beats `flat_skill` on quality while using **29% fewer tokens**. Surprising finding surfaced by the ablation: `flat_skill_wiki` lands *below* `no_skill` on aggregate and is the weakest arm on S6 — static wiki knowledge adjacent to the task can distract without helping. Reproduce with `humblskills eval run smart-humanize-text --scenario indie-launch-copy-iteration --runner cursor-agent`. Full write-up: [docs/eval/reports/smart-humanize-text/indie-launch-copy-iteration.md](docs/eval/reports/smart-humanize-text/indie-launch-copy-iteration.md).

Six runners ship behind one interface - pick whichever agent you already
use, or point an API key directly at the hosted model:

| Runner         | Auth                            | Notes                                                  |
|----------------|---------------------------------|--------------------------------------------------------|
| `claudecode`   | Claude Code login               | Wraps `claude -p --output-format stream-json`          |
| `cursor-agent` | Cursor login                    | Wraps `cursor-agent` headless CLI                      |
| `codex`        | Codex login                     | Wraps the OpenAI `codex` CLI                           |
| `anthropic-api`| `ANTHROPIC_API_KEY` / keyring   | Pure-Go Read/Write/Bash/Glob/Grep tool loop            |
| `openai-api`   | `OPENAI_API_KEY` / keyring      | Pure-Go tool loop                                      |
| `mock`         | none                            | For CI and dev - deterministic, zero tokens            |

### Quickstart

```sh
humblskills doctor                          # check runner availability
humblskills eval set-key anthropic          # store key in the OS keyring
humblskills eval runners                    # one-liner per-runner status
humblskills eval                            # dashboard entry → Eval Home TUI
humblskills eval run smart-skill        # non-TUI run
humblskills eval showcase                   # the canonical smart-skill demo
humblskills eval brand-voice                # the adaptive-brand-voice-discovery showcase (3-arm compounding)
humblskills eval ls                         # iterations per skill
humblskills eval prune smart-skill --keep-last 5
```

Secrets never land in the profile JSON. `eval set-key` resolves env >
OS keyring > `$XDG_CONFIG_HOME/humblskills/secrets.json` (perm 0600) in
that order, and the TUI prompts with a masked input.

### What lands on disk

Iteration artifacts under `$XDG_STATE_HOME/humblskills/evals/<skill>/iteration-N/`:

```
iteration-N/
├── benchmark.json      cross-section stats + deltas
├── trajectory.json     per-session time series (smart arm compounds here)
├── report.html         single-file Plotly dashboard
├── report.md           plaintext mirror (PR-friendly)
├── report.json         machine-readable
├── smart_skill/
│   └── session-NN/
│       ├── outputs/           files the agent wrote
│       ├── transcript.txt     full agent transcript
│       ├── timing.json        tokens + duration + cost
│       ├── metrics.json       tool-call counts + brain reads
│       ├── brain-snapshot-before/   brain state seeded into this session
│       └── brain-snapshot-after/    brain state after this session — feeds N+1
├── flat_skill/...
└── no_skill/...
```

Iterations are persistent and append-only. `humblskills eval prune` is the
retention knob.

### Authoring scenarios

Each skill ships an `evals/scenarios.json`. Sessions run in order; assertions
are either `llm` (sent to a judge model) or scripted (`path_exists`, `exec`,
`regex`, `script`, `json_valid`) - scripted beats LLM-judge for determinism.
`humblskills eval init <skill>` scaffolds a template. See
[`skills/smart-skill/evals/`](skills/smart-skill/evals/) for the
canonical example with retention checks across sessions.

## Preserving user content across updates

Smart skills often accumulate user-owned content over time - raw sources,
append-only memory (`log.md`, `decisions.md`, `patterns.md`), LLM-curated
wiki pages. By default `humblskills update` and `humblskills install`
overwrite the skill directory with whatever the registry ships. Skills that
need to keep user content around on update declare a preserve list under
`metadata:` in their `SKILL.md` frontmatter.

Entries are relative paths inside the skill directory. A trailing `/` makes
the entry a directory; anything else is a file. Globs are not supported.

| Entry form      | Example              | Meaning on update                                        |
| --------------- | -------------------- | -------------------------------------------------------- |
| **File**        | `references/log.md`  | User wins. User's bytes survive the update.              |
| **Directory**   | `references/wiki/`   | Deep merge. Staging wins per-file; user-only files kept. |

Fresh installs always seed everything from the registry - preserve only kicks
in when replacing an existing install. Running `humblskills uninstall` wipes
everything, including preserved content.

```yaml
---
name: my-smart-skill
description: ...
metadata:
  version: 0.2.0
  preserve:
    - references/log.md
    - references/patterns.md
    - references/decisions.md
    - references/raw/
    - references/wiki/
---
```

Skill authors who declare a preserve *directory* should note in their skill
docs that any files shipped inside that directory may be overwritten on
update - that's the deep-merge contract.

### You own the preserve list after install

The preserve list under **`metadata.preserve`** in the registry is the **seed**
- what ships on first install. After that, the list belongs to you.
`humblskills update` reads **`metadata.preserve`** from the **installed**
`SKILL.md` on disk (per target, so each platform + scope is independent), not
from the upstream registry entry.

That means:

- Add an entry locally -> that path survives your next `humblskills update`,
  even if upstream never listed it.
- Remove an entry locally -> that path gets overwritten by upstream bytes on
  the next update.
- Empty **`metadata.preserve`** -> the update is a clean overwrite for every
  path.

Use this to pin author-shipped files in place, protect notes you stash
inside the skill directory, or stop preserving a directory the author
reorganized.

Only **`metadata.preserve`** is treated as user-owned. Top-level agent-skills
fields (`name`, `description`, and the rest), every other key under
`metadata:` (`version`, `requires`, `platforms`, `tags`, and so on), and the
full markdown body flow through from upstream on every update. So when the
author ships a new description, version bump, or prose rewrite, you get it;
your preserve list rides along untouched. This also means your preserve edits
survive indefinitely - you don't need to re-edit after each update, because the
rewritten `SKILL.md` carries your list forward.

A few nuances:

- If you'd rather freeze the entire `SKILL.md` (maybe you've made prose
  edits you don't want overwritten), add `SKILL.md` itself to
  **`metadata.preserve`**. That makes user-wins on the file, so upstream changes
  to the description/version/body stop flowing - opt-in only.
- If the installed `SKILL.md` is missing, unparseable, or carries an
  invalid preserve list (e.g. a `..` traversal), the engine falls back to
  the registry's list and prints a warning. It won't wipe your data over
  broken YAML.
- The YAML round-trip on update normalizes whitespace and drops comments
  inside the frontmatter block. Keys and their values stay intact; only
  formatting inside the YAML mapping is rewritten.

### Getting clean upstream: `--force` or reinstall

```sh
humblskills update --force <skill>          # bypass local preserve, reinstall cleanly
humblskills install --force <skill>         # same effect outside update flow
humblskills uninstall <skill> && humblskills install <skill>   # equivalent
```

`--force` ignores your local preserve edits and replaces the on-disk skill
with exactly what the registry ships. This is the escape hatch for "throw
away my customizations and give me the author's version."

### When a skill is renamed upstream

`update` **follows** renames instead of skipping them. A skill that declares the
old name in **`metadata.previous_names`** is installed under its new name, your
**preserved files are carried across** from the old install, and the old
installation is retired. The picker and summary show it as
`use-smart-commit → smart-commit`, and the detail pane carries a `renamed from`
row, so nothing happens invisibly.

`humblskills update <new-name>` also reaches an install still recorded under the
old name. A rename is never guessed: a name still published in its own right is
never treated as a rename target, a name claimed by two skills is ignored, and a
skill that simply left the registry is left alone. (The `use-*` → `smart-*`
namespace change in v2.41 is the reason this exists.)

## Developing the CLI

The CLI source lives under [`cli/`](cli) as a nested Go module.

```sh
make build           # builds ./bin/humblskills
make test            # runs go test ./...
make registry        # regenerates registry.json from skills/ + embedded adapters
make registry-check  # local gate: fail if a rebuild would change registry.json
```

After adding or editing anything under `skills/`, regenerate (or let the
Registry workflow rewrite) so `source.sha` pins a commit that serves the
**current** skill trees — not merely one that still has the directory path.
Install fetches at that SHA and rejects `dir_sha` mismatches. Details:
[Registry & skill format](docs/using_humblskills/registry_and_format.md#how-registryjson-is-built-contributors).

### Mirrored skills

Skills distilled from an upstream source (the `better-*` family, and
`smart-frontend-design`) declare a top-level `upstream:` block naming where the
source lives, which file preserves the copy the distillation was written
against, and any deliberate `deltas:`. Because the preserved copy *is* the
baseline, drift needs no hashes or lockfile:

```sh
humblskills mirrors check         # which mirrors drifted, and how badly
humblskills mirrors plan <skill>  # emit a re-sync work order
```

`check` reports `current` / `drifted` / `rewritten` / `unknown` (a failed fetch is
never reported as healthy). `plan` writes the two files to diff, the wiki
concepts that cite the preserved copy, the declared deltas that must survive, and
a completion checklist. Detection is deterministic and automated;
re-distillation is judgment and is never automated. Both read a source checkout,
not installed copies.

Releases follow the two-branch path in
[`.github/workflows/release.yml`](.github/workflows/release.yml):

- **`develop`** — release-please opens a pre-release PR against
  `.release-please-manifest.develop.json`. Merging it tags `vX.Y.Z-pre.N` and
  publishes a GitHub **pre-release** (archives + checksums). GoReleaser updates
  `Formula/humblskills-pre.rb` only. The stable `humblskills` formula is left
  alone.
- **`main`** — merge `develop` with a merge commit (never squash). That promote
  does not tag. release-please then opens a stable PR against
  `.release-please-manifest.json` (last stable, never a `-pre`). Merging *that*
  PR tags `vX.Y.Z`, publishes the GitHub Release, and GoReleaser updates
  `Formula/humblskills.rb` so `brew upgrade humblskills` gets that version. The
  pre formula is not rewritten.

Both release PRs auto-merge on green for same-major bumps. A major bump waits
for a human.

The same job also pushes a sibling `cli/vX.Y.Z` tag, which is what `go install`
resolves against the nested module. Go's semantic import versioning requires the
module path to carry the major version, so `cli/go.mod` declares
`github.com/jjfantini/humblSKILLS/cli/v2` and the install path includes `/v2`.
**At the next major, the module path and every import must move to `/v3`** — a
mismatch is silent: the proxy simply ignores the tags and `@latest` keeps
serving the last version whose path matched.

## License

Content is licensed under [CC-BY-4.0](LICENSE). If Go source code licensing
becomes a concern later, the CLI code under `cli/` may be dual-licensed MIT —
but that has not been done yet.
