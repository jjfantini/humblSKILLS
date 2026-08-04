# smart-handoff tests

```bash
bash tests/run.sh
bash tests/run.sh --verbose
```

Runs under `/bin/bash` (macOS bash 3.2) deliberately — that is the floor both
scripts must clear. Uses `mktemp -d` for every fixture, so the host tree is
never touched.

**Not wired into CI.** Run it after editing `scripts/scan-secrets.sh` or
`scripts/preflight.sh`, after adding a detection pattern, or when porting to a
new shell environment.

## What is covered

`scan-secrets.sh`

- clean doc passes — `op://`, `security find-generic-password`, `gh auth token`,
  `<PLACEHOLDER>` and `$ENV_VAR` forms are not leaks
- leaky doc exits 1
- detection of AWS access key, GitHub token, Anthropic key, JWT, private-key
  block, and credentials embedded in a URL
- email address reported as `WARN`, not `FAIL`
- `WARN`-only doc exits 0; the same doc exits 1 under `--strict`
- missing file and no-arguments both exit 2
- **the report masks the match** — the raw secret never appears in output

`preflight.sh`

- exits 0 and prints `HANDOFF_DATE`, `TEMP_DIR`, `PERSIST_DIR`, `TEMP_FILE`,
  `PERSIST_FILE`
- `TEMP_FILE` matches `<slug>-handoff-<YYYY-MM-DD>.md` under
  `/tmp/.humblskills/handoffs/`
- emits the `installed-skills` block

## Fixture rule

Every fake credential is **assembled at runtime from string parts**
(`AWS="AKIA""ZZTESTKEY000000Q"`), so no literal credential shape is committed
to this repo and secret scanners in CI have nothing to flag. Keep it that way
when adding cases.

## Two traps worth remembering

1. **Never pipe the scanner into `grep -q`.** `grep -q` exits on first match,
   the scanner takes SIGPIPE, and `pipefail` turns a passing assertion into a
   failure. Capture the output into a variable first.
2. **Patterns contain `|`, `/` and `-`.** `sed` cannot mask them (every
   delimiter collides with an alternation) and `grep` reads a leading
   `-----BEGIN` as flags — hence literal bash substitution for masking and
   `grep -E -e "$re"` everywhere.
