# use-smart-commit tests

Opt-in test suite for `scripts/commit.sh` and `scripts/preflight.sh`. Not wired into any CI workflow — run explicitly after editing the scripts to confirm nothing regressed.

## How to run

```bash
bash skills/use-smart-commit/tests/run.sh
bash skills/use-smart-commit/tests/run.sh --verbose
```

The runner sets up a temporary git repo with `mktemp -d` and seeds it with a few scoped commits, so the host tree is never touched and the working directory is not polluted.

## What's covered

### `commit.sh`

- **Type validation**: each of the 10 allowed types (`feat`, `fix`, `perf`, `refactor`, `docs`, `test`, `build`, `ci`, `chore`, `style`) accepted; invalid type rejected; missing `--type` / `--subject` returns exit 2.
- **Subject validation**: under-72 accepted; over-72 rejected with helpful error; exactly-72 accepted; trailing-period rejected.
- **Body form validation**: subject-only allowed; partial labeled body (e.g. only `--changed`) rejected; multiple body forms rejected.
- **Skip-CI block**: all 5 token variants rejected case-insensitively, in both subject and body; the safe hyphenated `skip-ci` form is accepted.
- **Labeled body assembly**: all three sections rendered with correct headers and preserved content.
- **Breaking change**: `--breaking` produces `type(scope)!:` in the subject.
- **Footer matrix**: default-on; `--no-footer` flag; `HUMBLSKILLS_COMMIT_NO_FOOTER=1` env var; `.humblskills/no-footer` marker file. Each case asserts both the footer-absence and the state-reason output.
- **Real commits**: a labeled commit and a free-form `--no-footer` commit are actually committed and `git log -1` is inspected to confirm the body and footer landed correctly.

### `preflight.sh`

- Exits 0 in a git repo and emits all four expected sections.
- Detects the seeded scope from `git log`.
- Reports the default-on footer state.
- Exits 1 when run outside a git repo.

## When to run

- After editing `scripts/commit.sh` or `scripts/preflight.sh`
- After bumping the skill's version
- When porting to a new shell environment (different bash version, BSD vs GNU coreutils)

## Adding new cases

Test cases live inline in `run.sh`, grouped by `section "..."` headers. To add one:

1. Pick the right section header (or add a new one).
2. Invoke `commit.sh` (use the `dry()` helper for `--dry-run`, or call `bash "$COMMIT_SH"` directly for real commits with `fresh_change` first).
3. Use `expect_exit`, `expect_contains`, or `expect_not_contains` for assertions.

The runner exits 1 on any failure and prints a list of failed cases.
