# smart-skill tests

Test suite for `scripts/promote.sh` (with its engine `scripts/lib/promote.py`) and the create-flow parts of `scripts/scaffold.sh`. CI runs it on Linux and on macOS's stock bash 3.2.

## How to run

```bash
bash skills/smart-skill/tests/run.sh
bash skills/smart-skill/tests/run.sh --verbose
/bin/bash skills/smart-skill/tests/run.sh     # macOS: the bash 3.2 FDEs actually have
```

Everything happens under a `mktemp -d` sandbox: `HOME`, git identity, the clone cache, and "GitHub". GitHub is a set of bare repos that `promote.sh` reaches through its `HUMBLSKILLS_PROMOTE_GIT_BASE` seam, and `gh` is a stub on `PATH` that records every call and emulates `repo view`, `repo fork`, `api`, and `pr list/create/comment`. Nothing touches the network or a real repository. `--no-build-registry` keeps Go out of the loop; the real `build-registry` path is exercised by a manual `--dry-run` against the live repo.

## What's covered

- **Usage**: missing `--to`, unknown targets and skills, closed-set validation of `--category` / `--role`, `--role none` refused for happySKILLS.
- **Create flow**: `scaffold.sh` ships `scripts/lint.sh`, a `role:` placeholder, all five brain paths in `metadata.preserve`, header-only `decisions.md`, and ends by asking the promote question; a fresh scaffold fails the gate on every placeholder.
- **Taxonomy write-back**: flags write category/role/tags into `SKILL.md`, normalize and merge tags, remove the role for `--role none`, create a missing `metadata:` block, and leave every other key and the body untouched.
- **Quality gate**: one case per rule - trigger clause, 1024-char cap, angle brackets, unquoted `: `, name/dir mismatch, reserved prefix, kebab-case tags, closed-set category, missing `lint.sh`, unpreserved brain paths, leftover TODO markers, 500-line cap, credentials (never echoed), placeholder tokens and bare PEM headers (allowed), `.env`, symlinks, `node_modules`, oversized files, HappyRobot mentions in a public promotion (and `--allow-internal`), email warnings, junk files, and `lint.sh` failures.
- **Dry run**: worktree kept for inspection, junk not staged, `_index.md` regenerated, nothing pushed, no PR.
- **Promote with push access**: `feat/add-<skill>` cut from `develop`, a `feat(skills): add <skill> skill` commit by the user's identity, only `skills/<skill>/` in the diff (no `registry.json`), PR base/head/title/body; re-run without changes adds nothing; re-run after an edit adds a `chore(skills): sync ...` commit and comments on the existing PR instead of opening another.
- **Registry rules**: published names and former names (`previous_names`) refused; files the target's `.gitignore` would drop refused before anything is pushed.
- **Fork**: without push access the branch goes to the user's fork and the PR head is `owner:branch`.
- **happySKILLS**: lookup by name in `~/.humblskills/skills`, repo discovered from the humblskills profile's registries, base falls back to the default branch, role required and written back.
- **`--checkout`**: a user's own clone keeps its working tree, branch, remotes, and refs; no worktree is left behind.
- **Calibration** (inside this repo only): the stdlib frontmatter parser matches `registry.json` for every published skill, and the credential scanner reports nothing on any of them.

## When to run

- After editing `scripts/promote.sh`, `scripts/lib/promote.py`, or `scripts/scaffold.sh`
- After changing the CLI's categories, roles, or adapters (the Go test `promote_sync_test.go` also guards those)
- When porting to a new shell environment (BSD vs GNU tools, bash 3.2)

## Adding new cases

Cases live inline in `run.sh`, grouped by `section "..."` headers. `fresh <name>` builds a finished skill to mutate, `gate <dir> [args]` runs the gate and sets `OUT`/`RC`, and `expect_exit`, `expect_contains`, `expect_not_contains`, and `expect_eq` assert. The runner exits 1 on any failure and lists the failed cases.
