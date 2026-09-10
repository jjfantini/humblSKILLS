---
title: "Verification Contracts: Freeze Before, Evaluate After"
context: verify
category: contracts
concept: schema
description: "The contract JSON shape, six check types, and the VERIFIED/FAILED/UNVERIFIED verdict rule that keeps a missing tool from reading as a pass"
tags: verify, contract, verdict, shasum, jq, sqlite, curl, proof-of-work
sources: []
last_ingested: 2026-09-10
command: scripts/verify.sh
---

## Contracts Are Frozen Before the Work

A contract is a JSON file the user or agent writes **before** doing the work, listing checks that
must hold afterwards. `verify.sh` evaluates it with native tools and prints verdicts. The model's
opinion never enters; the check either matched, didn't, or could not run.

```json
{"checks":[
 {"id":"bundle",  "type":"file_sha256", "path":"dist/app.js", "expect":"<hex>"},
 {"id":"readme",  "type":"file_exists", "path":"README.md"},
 {"id":"version", "type":"json_path",   "file":"package.json", "path":".version", "expect":"1.2.0"},
 {"id":"rows",    "type":"sql",         "db":"app.db", "query":"select count(*) from users", "expect":"3"},
 {"id":"health",  "type":"http",        "url":"http://localhost:3000/health", "expect_status":200, "expect_body_contains":"ok"},
 {"id":"tests",   "type":"command",     "run":"npm test --silent", "expect_exit":0, "expect_stdout_contains":"passing"}
]}
```

Relative paths resolve against the contract file's directory; `command` runs there too.

**Incorrect (reading absence as success):**

```text
health  curl failed  -> "endpoint fine, nothing to report"
```

**Correct:**

```text
UNVERIFIED  health  http  unreachable: http://localhost:3000/health
overall: FAILED   (exit 1)
```

Verdicts: `VERIFIED` (ran, matched), `FAILED` (ran, did not match), `UNVERIFIED` (could not run:
absent file, jq/sqlite error, unreachable host, unknown type). Exit 0 only when every check is
VERIFIED. `--json` appends a machine-readable result. Each check runs in a subshell, so a crash in one
cannot take down the run. There is no per-check timeout beyond `curl --max-time 10`; wrap long
commands yourself. DOM checks are deliberately absent - the agent already has browser tools.

## Sources

Synthesis from the upstream watch-skill verdict taxonomy plus this skill's implementation; no raw file.
