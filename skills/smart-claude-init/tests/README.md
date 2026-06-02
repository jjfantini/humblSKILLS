# Tests for smart-claude-init

`run.sh` is an opt-in suite for `scripts/validate-claudemd.sh` — the
deterministic contract check the generation workflow runs before writing a
`CLAUDE.md` to a user's project. It builds all fixtures under a `mktemp -d`
directory, so the host tree is never touched.

**Not wired into CI.** Run it explicitly after editing `validate-claudemd.sh`
or the bundled `assets/*.tmpl` templates, or after bumping the skill version.

```bash
bash tests/run.sh            # quiet
bash tests/run.sh --verbose  # print every PASS/FAIL line
```

Exit `0` = all passed, `1` = one or more failed, `2` = setup error
(validator or templates missing).

## What it covers

| Group | Case |
|-------|------|
| happy path | filled code template passes (default mode) |
| happy path | filled general template passes (`--general`) |
| happy path | valid file prints `OK (code)` |
| missing section | code file missing `## Performance` → exit 1, names the section |
| missing section | code file missing `## Bug Protocol` → exit 1 |
| placeholder | a leftover `{{STACK}}` token → exit 1, `UNRESOLVED PLACEHOLDER` |
| placeholder | the raw `claude-code.md.tmpl` itself fails (sanity check) |
| TODO comment | leftover `<!-- TODO ... -->` → exit 1, `LEFTOVER TODO COMMENT` |
| TODO comment | prose containing the word `TODO` (no HTML comment) is **not** flagged |
| general contract | a code file fails `--general` (lacks Working Preferences / Quality Bar) |
| general contract | general file missing `## Quality Bar` → exit 1 |
| invocation | no file arg → exit 2 |
| invocation | nonexistent file → exit 2 |
| invocation | unknown flag → exit 2 |
| invocation | `--help` → exit 0 |

## Adding a case

Build the fixture from the bundled template (`sed` to fill or mutate it), run
`bash "$VALIDATE" ...`, capture `$?`, and assert with `expect_exit` /
`expect_contains`. Keep every fixture inside `$TMPDIR` so the suite stays
hermetic.
