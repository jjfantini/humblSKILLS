---
title: "Promote a Finished Skill as a Pull Request"
context: smart
category: promote
concept: workflow
description: "The end-to-end path from a finished local skill to a reviewed PR in humblSKILLS or happySKILLS, without hand-copying files or guessing branches"
tags: promote, publish, pull-request, humblskills, happyskills, registry, review
sources:
  - "references/raw/user-request-promote.md"
last_ingested: 2026-10-02
command: scripts/promote.sh
---

## Promote a Finished Skill

A skill that only lives in `~/.humblskills/skills/` helps one person. Promotion
opens a pull request that adds it to a registry, where the maintainer reviews
it and every `humblskills install` can reach it. It is the last step of the
create workflow, and only for **net-new** skills.

**Incorrect (hand-copying into a clone):**

```bash
cp -R ~/.cursor/skills/my-skill ~/src/humblSKILLS/skills/
git checkout -b my-skill && git add . && git commit -m "add skill"
gh pr create --base main   # wrong base, no taxonomy, no gate, .DS_Store committed
```

**Correct (ask, gate, then promote):**

```bash
bash scripts/promote.sh my-skill --to humblskills --check \
  --category development --role none --tags git,code-review,pull-requests
bash scripts/promote.sh my-skill --to humblskills
```

## Steps

1. **Ask where it goes.** When the skill works, ask the user: promote to
   humblSKILLS (public), to happySKILLS (HappyRobot-specific), or keep it
   local? Recommend one using `smart/promote/targets.md`. Never promote unasked.
2. **Settle the taxonomy.** Propose category, role, and 3+ tags with
   `smart/promote/taxonomy.md`; confirm with the user. Pass them as flags:
   `promote.sh` writes them into the local `SKILL.md`, so re-runs need none.
3. **Run the gate:** `promote.sh <skill> --to <target> --check`. Fix every
   FAIL in the local skill and re-run; read the WARNs out to the user.
   `smart/promote/quality-gate.md` explains each finding.
4. **Promote:** drop `--check`. The script clones or forks the target, cuts
   `feat/add-<skill>` from the base branch, stages a clean copy, lints it,
   validates it with `build-registry` when Go 1.23+ is installed, commits,
   pushes, and opens the PR. Add `--dry-run` first if the user wants to see the
   commit before anything is pushed.
5. **Hand over the PR URL.** Review happens there; the maintainer merges.
6. **Iterate.** Apply review feedback to the local skill and run the same
   command again: it pushes a follow-up commit and comments on the PR.
7. **After the merge**, `humblskills install <skill>` takes over the local
   copy in place and keeps its preserved brain files.

Log the promotion (target, PR URL, warnings accepted) in this skill's
`references/log.md`; record a non-obvious target or taxonomy call in
`decisions.md`.

## Sources

- `references/raw/user-request-promote.md` - the request: promote from
  smart-skill, ask humblSKILLS vs happySKILLS after create, net-new only.

## Command

```bash
bash scripts/promote.sh <skill-or-path> --to humblskills|happyskills [--check|--dry-run] \
  [--category c] [--role r|none] [--tags a,b,c]
```
