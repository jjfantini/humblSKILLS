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
- Registry: `make registry` regenerates `registry.json`; `make registry-check` fails if it's stale — run it
  after editing anything under `skills/`. It is a **local** gate: the Registry workflow's self-heal job
  replaced the old pre-merge check, so nothing in CI runs `--check` today. "Stale" means *anything a rebuild
  would change* — drifted content, or a `source.sha` whose skill tree ids disagree with HEAD (skill missing
  or serving an older revision; see below).
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

### Release path (develop pre-release → main stable + brew)
`.github/workflows/release.yml` is the only release entry point:

- Push/merge to **`develop`** → release-please (`release-please-config.develop.json` +
  `.release-please-manifest.develop.json`) cuts a GitHub **pre-release** tagged
  `vX.Y.Z-pre.N`. GoReleaser publishes archives and the `humblskills-pre` formula.
  The stable `humblskills` formula is not touched (`skip_upload: auto`).
- Merge **`develop` → `main`** (merge commit, never squash) is the promote, not
  the tag. release-please on main (`release-please-config.json` +
  `.release-please-manifest.json`, last **stable** only) then opens the stable
  PR. Merging that PR tags `vX.Y.Z` and GoReleaser updates
  `jjfantini/homebrew-humbl` `Formula/humblskills.rb` so `brew upgrade humblskills`
  gets that version. The pre formula is not rewritten on a stable tag.
  Do not put a `-pre` version in the main manifest: release-please will treat
  `vX.Y.Z-pre` as the last release and skip (run 33548192550: last-saw
  `v2.52.0-pre` at `375eb41`; the #270 merge subject is not conventional).

  **After a stable graduate, rewrite the develop manifest to that stable.**
  `versioning: prerelease` on `2.52.0-pre.3` only cuts `2.52.0-pre.4`.
  Semver: `2.52.0-pre.N` < `2.52.0`, so beta (`max(stable, pre)`) stays on
  stable and never shows the update-notice banner. The `record_stable_on_develop`
  job in `release.yml` runs `scripts/sync-develop-pre-after-stable.sh` after a
  non-`-pre` tag and records `X.Y.Z` in `.release-please-manifest.develop.json`.
  The next conventional commit on develop then opens `X.Y.(Z+1)-pre.1` (fix)
  or `X.(Y+1).0-pre.1` (feat). Do not hand-tag. Do not re-cut `X.Y.Z-pre.N`.
  If the job is missing and develop is stuck, a real `fix:`/`feat:` commit
  with footer `Release-As: X.Y.(Z+1)-pre.1` is the fallback.

Both release PRs auto-merge on green for same-major bumps (`scripts/guard-major-bump.sh`
blocks majors). Secrets live on the `release` environment: `RELEASE_PLEASE_TOKEN` (repo PAT)
and `HOMEBREW_TAP_TOKEN` (write to the tap). Do not add extra patch workflows around this
path; if a release fails, fix the two configs or the one workflow.

### Commit messages MUST be Conventional Commits (non-negotiable)
`release-please-config.json` / `release-please-config.develop.json` set `release-type: "go"`,
which cut releases **only** from commits that follow
[Conventional Commits](https://www.conventionalcommits.org) syntax on the branch that
releases (`develop` for pre, `main` for stable). A commit that doesn't match is silently
invisible to release-please — no changelog entry, no version bump, no release, and
`humblskills upgrade` never sees the change. This already happened once (commits like `profile:`,
`adapters:`, `install:`, `tui:` merged to `main` with zero release effect) and had to be fixed forward with
an extra empty `feat:` commit — don't repeat it.

Every commit message, on every branch that will reach `main` (not just the PR title), must be:

```
<type>[(optional-scope)]: <description>
```

- `feat` → minor bump. `fix` / `perf` → patch bump. Add `!` after the type/scope (`feat!:`) or a
  `BREAKING CHANGE:` footer for a major bump.
- `chore`, `docs`, `refactor`, `test`, `build`, `ci`, `style`, `revert` are valid types but do **not**
  trigger a release on their own — use them for exactly what they mean, don't reach for `feat`/`fix` just to
  force a release.
- The scope is the affected area, in parentheses, e.g. `feat(install):`, `fix(tui):`, `chore(registry):`.
  Never use the area name as the type itself (`install: ...`, `tui: ...` are invalid — that's the mistake
  that caused the missed release).
- If a PR/branch has no `feat`/`fix`/`perf` commit but ships a user-visible change, add one (or make the
  final merge commit one) — don't rely on non-conventional prose to describe user-facing work.

**Never write the literal string `[skip ci]` (or `[ci skip]`, `[no ci]`, `[skip actions]`) in a commit
message body, even when describing it.** GitHub scans every commit message in a push, not just the subject
line, and skips *all* workflow runs for that push. A commit body explaining the marker suppressed CI for an
entire branch here — no failing check, no queued run, nothing to notice. Write it unbracketed (`skip-ci`)
when you need to refer to it in prose.

### Merge PRs with `--merge`. Never `--squash`, never `--rebase`.
Two independent things break on a squash, both silently:

1. **release-please reads the individual commit messages on `develop` and `main`.** A squash collapses the whole PR into
   its title, so a branch carrying `fix:` plus two `feat:` commits ships as a patch and loses both feature
   changelog entries. A merge commit preserved all four and correctly produced the 2.43.0 minor bump.
2. **`registry.json` pins `source.sha` to the commit where skill content last changed, and `install`
   fetches the skill tarball at exactly that SHA.** Under a merge commit that SHA stays reachable forever.
   A squash replaces it with a new commit and orphans the original, so the recorded SHA becomes
   unreachable and installs of those skills start failing — long after the PR is forgotten.

```sh
gh pr merge <n> --merge      # correct
gh pr merge <n> --squash     # breaks both of the above
```

### `registry.json`'s `source.sha` must serve every skill it lists (current bytes)
`install` fetches each skill's tarball at `source.sha`, then recomputes `dir_sha` over what it extracted
and rejects a mismatch. Two failure modes share that path:

| Recorded `source.sha` vs skill content | What install sees |
|---|---|
| Skill directory absent at that commit | `extract: no files found under "skills/<name>" in tarball` |
| Skill directory present but an **older revision** | Extract succeeds, then `dir_sha` mismatch rejects the install |

Presence alone is not enough. `sourceSHAVerdict` (`cli/cmd/build-registry/main.go`) compares **git tree
object ids** per skill path (`git ls-tree`) between the recorded SHA and the candidate stamp — so a commit
that still has `skills/<name>/` but with stale bytes is rewritten the same way as a missing skill
(`serves outdated` vs `does not contain` in the log line).

`make registry` reads the working tree but stamps `source.sha` from git HEAD, so **the run that first adds
or edits a skill always records a SHA that cannot yet serve the new bytes**. That is expected locally; the
Registry workflow repairs it on the next push, once the change is committed. Editing skills is the common
case — the older presence-only check left those SHAs frozen after CI reported "already in sync".

Do not "fix" this by making `semanticDiff` compare the `source` block — that comparison is zeroed on purpose,
and reinstating it makes the workflow push on every trigger forever. The repair bypasses the "already in sync"
exit only when the recorded SHA is provably broken **and** the replacement provably contains every listed
skill tree, so it converges after one write. It needs real history, which is why the workflow checks out
with `fetch-depth: 0`.

`make registry-check` fails on the same condition, so `--check` never disagrees with `make registry`. It stays
green when *no* rebuild could fix it — the uncommitted-skill case above — because that state is normal and
self-heals on the next push; a gate that goes red on the expected path just teaches people to ignore it.

If you ever need to repair it by hand:

```sh
go -C cli run ./cmd/build-registry --skills-dir=$PWD/skills --out=$PWD/registry.json --ref=main --sha=<sha>

# Verify before pushing: equal tree ids mean <sha> serves the same bytes as HEAD.
git rev-parse "<sha>:skills/<name>" "HEAD:skills/<name>"
```

A git tree id is **not** the `dir_sha` install checks — that is a sha256 over
canonicalized `(rel_path, mode, content_sha)` tuples (`registry.DirSHA`), while
a tree id is a sha1 over git's own serialization. The two are never equal and
comparing them is meaningless. Tree ids are only useful *between commits*,
which is exactly how `sourceSHAVerdict` uses them: same id at two commits means
identical bytes, so whatever `dir_sha` one produces the other produces too.

### A conflicted PR gets no CI at all
GitHub cannot build the merge ref for a PR with conflicts, so it dispatches **no** `pull_request` workflows.
The PR shows no checks rather than failing ones, and with nothing gating `main` it stays merge-able by hand.
If CI "didn't trigger", check `gh pr view <n> --json mergeable` before suspecting Actions.
