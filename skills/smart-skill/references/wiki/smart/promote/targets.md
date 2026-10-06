---
title: "Choose humblSKILLS or happySKILLS"
context: smart
category: promote
concept: targets
description: "Which registry a promoted skill belongs in, and the repo, base branch, and access each one uses - a wrong pick either leaks internal knowledge or hides a useful skill"
tags: promote, humblskills, happyskills, registry, public, private, happyrobot
sources:
  - "references/raw/user-request-promote.md"
last_ingested: 2026-10-02
---

## Pick the Registry Before Anything Is Pushed

A pull request branch on a public repo is public the moment it is pushed, so
the target has to be right before review, not during it.

| | humblSKILLS | happySKILLS |
|---|---|---|
| Flag | `--to humblskills` | `--to happyskills` |
| Audience | anyone with `humblskills` | HappyRobot only |
| Repo | `jjfantini/humblSKILLS` (public) | your `humblskills registry` entry for happySKILLS, else `jenningsfantini-happyrobot/happySKILLS` |
| Base branch | `develop` | `develop` if it exists, else the default branch |
| Role | optional (`--role none` = general-purpose) | required |

**Incorrect (an internal runbook sent to the public registry):**

```bash
bash scripts/promote.sh smart-load-matching --to humblskills   # cites customer lanes
```

**Correct:**

```bash
bash scripts/promote.sh smart-load-matching --to happyskills --role fde
```

## Decision Rule

Recommend **happySKILLS** when any of these hold:

- it names HappyRobot, a customer, or a teammate, or cites internal URLs
- it depends on internal systems, APIs, or credentials
- it encodes a HappyRobot process (deploy runbook, sales motion, support flow)
- `raw/` or the brain holds meeting notes, tickets, or customer data

Recommend **humblSKILLS** when an engineer outside HappyRobot could install it
and get value without any of the above. When unsure, recommend happySKILLS:
going public later is another net-new promotion, but a public push can't be
taken back. For humblSKILLS the gate fails on any "HappyRobot" mention unless
`--allow-internal` confirms it is public-safe.

## Access

The script pushes to the target when you can, and otherwise forks it and opens
the PR from your fork. If neither works (a private repo you can't fork), ask
the maintainer for access. Override the repo with `--repo owner/name` or
`HAPPYSKILLS_REPO`, the base with `--base`.

## Sources

- `references/raw/user-request-promote.md` - the two targets and the
  "HappyRobot-specific goes to happySKILLS" rule.
