# AGENTS.md

## Cursor Cloud specific instructions

This repo is the **humblskills** project: a single-binary Go CLI (`cli/`) that installs
[agentskills.io](https://agentskills.io)-format skills, a markdown skill registry (`skills/` → `registry.json`),
and an MkDocs docs site (`docs/`). The CLI is fully local ("zero servers, zero accounts, zero telemetry"),
so **no backend/database/network services are required** to build, test, or run it.

### Toolchain notes (non-obvious)
- The CLI requires **Go 1.23+** (`cli/go.mod` pins `go 1.23.0`). The base VM image ships an older Go, so a
  newer Go (1.26.x) is installed at `/usr/local/go` and symlinked to `/usr/local/bin/go`. The startup update
  script keeps Go module deps fetched; if `go version` ever reports < 1.23, re-install Go before building.
- `cli/` is a **nested Go module**. The root `Makefile` drives everything via `go -C cli ...` — run make
  targets from the repo root, not from inside `cli/`.

### Commands (all from repo root; see `Makefile`)
- Build: `make build` → binary at `bin/humblskills`.
- Lint/vet: `make vet` (CI uses `go vet`; there is no golangci-lint config).
- Test: `make test`, or CI-equivalent `go -C cli test -race -count=1 ./...`.
- Registry: `make registry` regenerates `registry.json`; `make registry-check` fails if it's stale
  (enforced in CI — run it after editing anything under `skills/`).
- Eval (no external deps): `make eval-mock` runs the eval harness with the deterministic `mock` runner,
  writing artifacts to `.eval-workspace/` (gitignored). Real eval runners (`claudecode`, `cursor-agent`,
  `codex`, `anthropic-api`, `openai-api`) are optional and need their respective agent CLI or
  `ANTHROPIC_API_KEY` / `OPENAI_API_KEY`.

### Running the CLI
- `./bin/humblskills doctor` — shows detected agent platforms + registry/eval readiness.
- `./bin/humblskills search <q>` / `install <skill>` / `list` / `update`.
- To exercise the **local** code against the in-repo registry (instead of the hosted one), pass
  `--registry file:///workspace/registry.json`. Use `--yes` (and optionally `--platform`/`--scope`) to run
  non-interactively; with no args many commands open an interactive TUI.
- Install writes to platform skill dirs (e.g. `.cursor/skills/`, `~/.humblskills/`). In this repo
  `.cursor/`, `.claude/`, `bin/`, `site/`, `.eval-workspace/` are gitignored, but `~/.humblskills/` (created
  by `--global` or default installs) is NOT — clean up stray `.humblskills/` if it appears in the worktree.

### Docs site (optional)
- Needs Python venv tooling (`python3.12-venv`). Build with the venv at `~/.venvs/humblskills-docs`:
  `~/.venvs/humblskills-docs/bin/mkdocs build --strict` (config: `mkdocs.yml`). `mkdocs serve` for preview.
  Note `mkdocs build` drops a `docs/__pycache__/` (not gitignored) — remove it after building.
