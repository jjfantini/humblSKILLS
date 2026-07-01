# humblSKILLS

📖 **Full documentation:** [jjfantini.github.io/humblSKILLS](https://jjfantini.github.io/humblSKILLS/)

A personal skill registry and a single-binary Go CLI (`humblskills`) that
installs [agentskills.io](https://agentskills.io)-format skills into whichever
agent platform you use — Claude Code, Cursor, Codex, and friends.

## What's in this repo

1. **Skill registry** — a monorepo of agent skills authored in the
   agentskills.io format with light humblSKILLS frontmatter extensions
   (`requires`, `platforms`, `tags`, `preserve`).
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
```

Formulas live in [`jjfantini/homebrew-humbl`](https://github.com/jjfantini/homebrew-humbl)
and are bumped automatically by the release workflow. Upgrade with `brew upgrade humblskills`.

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
go install github.com/jjfantini/humblSKILLS/cli/cmd/humblskills@latest
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
humblskills search                    # browse the registry
humblskills install use-smart-skill
humblskills list
humblskills update                    # pick which drifted skills to upgrade
humblskills update --all --yes        # non-interactive bulk upgrade
humblskills uninstall use-smart-skill
humblskills export                    # snapshot installed skills to humblskills.json
humblskills sync                      # install everything in humblskills.json
```

### Sharing skill sets across a team

`humblskills export` snapshots the skills you have installed into a
`humblskills.json` file (override with `-o`). Commit it to a repo, and every
teammate runs `humblskills sync` to install the same set — a single,
version-controlled source of truth for "which skills does this project want".

```sh
humblskills export -o humblskills.json   # write the skillset
humblskills sync                         # install missing skills from ./humblskills.json
humblskills sync path/to/set.json --force  # reinstall everything from a specific file
```

`sync` pulls the current registry version of each skill (like `install`),
skips skills already up-to-date, and warns (without failing) about any skill in
the file that the registry doesn't know about. Platforms/scope follow the same
rules as `install`.

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

**4-arm ablation showcase:** [indie-launch-copy-iteration](https://jjfantini.github.io/humblSKILLS/eval/indie-launch-analysis/) — 6 sessions over 13 indie-launch voice rules, three runs per arm (72 sessions total) on claudecode. Separates **brain value** (`smart_skill` vs `flat_skill_wiki`) from **wiki value** (`flat_skill_wiki` vs `flat_skill`) with identical preamble and scaffolding. The cumulative-retention outcome assertion (S5 + S6 violations ≤ 1) passes **3/3 for `smart_skill` and 0/3 for every other arm**. `smart_skill` hits 43% fewer violations than `flat_skill_wiki` while using 2.6% fewer tokens and 8.9% less wall time. Surprising finding surfaced by the ablation: `flat_skill_wiki` is the **worst** of the four arms — static wiki knowledge adjacent to the task can distract without helping. Reproduce with `humblskills eval run use-smart-humanize-text --scenario indie-launch-copy-iteration`. Full analysis: [docs/eval/indie-launch-analysis.md](docs/eval/indie-launch-analysis.md).

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
humblskills eval run use-smart-skill        # non-TUI run
humblskills eval showcase                   # the canonical use-smart-skill demo
humblskills eval brand-voice                # the adaptive-brand-voice-discovery showcase (3-arm compounding)
humblskills eval ls                         # iterations per skill
humblskills eval prune use-smart-skill --keep-last 5
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
[`skills/use-smart-skill/evals/`](skills/use-smart-skill/evals/) for the
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

## Developing the CLI

The CLI source lives under [`cli/`](cli) as a nested Go module.

```sh
make build           # builds ./bin/humblskills
make test            # runs go test ./...
make registry        # regenerates registry.json from skills/ + embedded adapters
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
