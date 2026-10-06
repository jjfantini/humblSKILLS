---
title: "What the Promotion Gate Checks and How to Fix It"
context: smart
category: promote
concept: quality-gate
description: "promote.sh refuses to open a PR while any check fails; each check exists because skipping it breaks the registry build, an install, or a user's data"
tags: promote, gate, validation, review, secrets, frontmatter, lint
sources:
  - "references/raw/user-request-promote.md"
  - "references/raw/agentskills-spec.md"
  - "references/raw/anthropic-skill-building-guide.pdf"
last_ingested: 2026-10-02
command: scripts/promote.sh
---

## Fail Locally, Not in Review

Every FAIL below would otherwise surface as a broken registry build, a failed
install, or a leak. The gate runs before anything is pushed and again on the
staged copy. WARNs never block; they are copied into the PR for the reviewer.

**Incorrect:** opening the PR and letting review find the problems.
**Correct:** `promote.sh <skill> --to <target> --check` until it reports 0 failures.

| Check | Fails when | Fix |
|---|---|---|
| frontmatter | no `name`/`description`/`license`/`metadata.author`/semver `metadata.version`; a list where a value belongs; top-level `version`/`tags`; an unquoted value with `: ` | match the scaffold shape; quote values containing `: ` |
| name | not kebab-case, over 64 chars, differs from the directory, starts with `claude`/`anthropic` | rename the skill and its directory together |
| description | over 1024 chars, `<` or `>` anywhere in frontmatter, no trigger clause ("Use when ...") | rewrite per `anthropic/description/trigger-design.md` |
| taxonomy | category/role/tags missing, TODO, or outside the closed sets | `--category`, `--role`, `--tags` (see `smart/promote/taxonomy.md`) |
| content | scaffold `TODO` blocks left, no `## Brain Protocol`, SKILL.md over 500 lines | finish the scaffold; move detail into `references/` |
| structure | brain files, `references/raw/`, `references/wiki/`, or `scripts/lint.sh` missing; brain paths not in `metadata.preserve` | copy `scripts/lint.sh` from smart-skill; preserve all five brain paths so `update` never wipes user data |
| files | symlinks, `node_modules/`/`.venv/`, a file over 5 MiB or skill over 15 MiB | inline the file; install deps at runtime; drop bulky raw sources |
| secrets | live-looking tokens or keys, `.env`, SSH keys | delete, and rotate anything that was real |
| public | "HappyRobot" in a humblSKILLS promotion | `--to happyskills`, or `--allow-internal` if public-safe |
| lint | `scripts/lint.sh` exits non-zero | fix what lint names (`brain/lint/checks.md`) |
| net-new | the target already publishes the name, or another skill lists it in `previous_names` | pick a new name; changing a published skill is not this path |
| ignored | the target's `.gitignore` would drop a file (e.g. `bin/`) | rename the path; otherwise the PR silently misses it |
| build-registry | the real registry build rejects the skill (when Go 1.23+ is installed) | fix the field it names |

Junk (`.DS_Store`, `__pycache__/`, editor folders) is left out of the PR
automatically. File modes are normalized to what git records.

WARNs to raise with the user: no negative trigger, no `## Examples`, fewer than
3 tags, unscoped role, SKILL.md over 200 lines, a stale `_index.md`, email
addresses, absolute home paths, JWT-shaped strings, an empty wiki.

## Sources

- `references/raw/agentskills-spec.md` - name and description limits.
- `references/raw/anthropic-skill-building-guide.pdf` - angle-bracket and
  reserved-name rules, 1024-character description cap.
- `references/raw/user-request-promote.md` - keep the bar high without being
  bureaucratic.

## Command

```bash
bash scripts/promote.sh <skill> --to humblskills|happyskills --check
```
