---
name: use-worktree-flow
description: >
  Run a worktree-first development workflow from intent gathering through
  feature PR, develop merge, main merge, release PR, brew verification, and
  cleanup. Use when the user says "worktree flow", "start a feature",
  "ship this feature", "PR into develop", "full dev workflow", or wants a
  feature isolated from parallel agents. Do NOT use for atomic commit authoring
  (use use-smart-commit) or merge-conflict repair.
license: MIT
compatibility: "Requires git 2.5+, GitHub CLI (`gh`) for PR/check operations, and network access for remote sync, CI, releases, and Homebrew verification."
metadata:
  author: jjfantini
  version: "1.0.0"
  tags: [git, worktree, pull-requests, release, workflow, humblskill]
  platforms: [claude-code, cursor]
  preserve:
    - references/decisions.md
    - references/log.md
    - references/patterns.md
---

# Worktree Flow

Use this when shipping code through an isolated worktree: feature branch to
`develop`, `develop` to `main` or `master`, release PR, local sync, and cleanup.

## Brain Protocol (read BEFORE creating anything)

1. `references/_index.md` - what this skill knows (map)
2. `references/patterns.md` - what worked, with numbers
3. `references/decisions.md` - past reasoning, don't repeat mistakes
4. `references/log.md` - last 5 session entries
5. Relevant `references/wiki/flow/<category>/` concepts per task

After completing work, UPDATE the brain:
- Performance data (if reported) -> `patterns.md`
- Non-obvious decisions -> `decisions.md`
- Session summary (always) -> append to `log.md`
- Run `scripts/lint.sh` to regenerate `_index.md` and verify structure

_Full spec: `references/_brain.md`._

## CCCCC Architecture

| Layer | Role | Location |
|---|---|---|
| Core | Root structure of the skill | `SKILL.md`, `references/`, `scripts/` |
| Context | Top-level taxonomy grouping | First segment under `references/wiki/` |
| Category | Specific topic within a context | Second segment under `references/wiki/` |
| Concept | One atomic idea per file | Filename stem and frontmatter field |
| Command | Deterministic executable script | `scripts/<command>.sh` |

## Mandatory Step 0: Gather Intent

Before implementation, ask the user the questions in
`references/wiki/flow/setup/intent-gathering.md`. If they defer, use these
defaults: Vibe mode, worktree isolation, auto-merge develop to main on green,
auto-merge release PR on green, and clean stale worktrees and branches.

Also inspect reality before choosing the path: run `git status`,
`git worktree list`, check current branch tracking, and look for other active
agent or terminal work. Default to a worktree when isolation is useful or when
the workspace is dirty.

## Workflow

1. Sync the local repo and identify the integration branch (`develop`) plus
   production branch (`main` or `master`).
2. Create a paired worktree and branch. Example: worktree `feat-add-data`,
   branch `feat/add-data`.
3. PR the feature branch into `develop`; merge only after tests, lint,
   verifications, and CI/CD are green.
4. PR `develop` into `main` or `master`; follow the upfront Vibe or HITL
   decision. Prefer merge commits when release tooling reads conventional
   commits from main.
5. Handle the generated release PR if the repo has release automation. Merge
   it on green only if the user chose that path.
6. Sync local branches with upstream, remove the worktree, and delete stale
   local and remote feature branches unless the user opted out.

## How to Use

**Live enumeration of categories and concepts:**
Read `references/_index.md` after running `scripts/lint.sh`.

**Ask the right questions upfront:**
Read `references/wiki/flow/setup/intent-gathering.md`.

**Name and create the worktree:**
Read `references/wiki/flow/setup/worktree-naming.md`, then run
`bash scripts/create-worktree.sh <type> <slug> [base-branch]`.

**Gate feature work into develop:**
Read `references/wiki/flow/pr/develop-gate.md`.

**Gate develop into main or master:**
Read `references/wiki/flow/pr/main-gate.md`.

**Handle release automation:**
Read `references/wiki/flow/release/release-pr.md`.

**Sync and cleanup after release:**
Read `references/wiki/flow/cleanup/sync-and-prune.md`, then run
`bash scripts/cleanup.sh <branch> [worktree-path]`.

## Examples

### Example 1: Vibe Mode Feature

User says: "Use the worktree flow and ship this feature. I defer to you."

Actions:
1. Apply default choices: Vibe mode, worktree isolation, auto-merge on green,
   release PR auto-merge, cleanup enabled.
2. Create `feat-add-data` and `feat/add-data` from `origin/develop`.
3. Implement, verify, PR into `develop`, merge on green, PR `develop` into
   `main`, handle release PR, verify brew upgrade, cleanup.

Result: The feature ships through release with no stale branch or worktree.

### Example 2: HITL Release Review

User says: "Start a worktree but I want to review before main and release."

Actions:
1. Use HITL gates for `develop` -> `main` and the release PR.
2. Still use automated checks to gather evidence before asking for approval.
3. Stop before merging `main` or release PR until the user approves.

Result: The user keeps the final review gates while work stays isolated.

## Success Signals

- Initial intent is explicit: Vibe or HITL, auto-merge policy, release policy,
  cleanup policy, and worktree choice.
- Worktree and branch names share the same conventional type and slug.
- No PR merges until tests, lint, local verification, and CI/CD are green.
- Release automation is monitored through tag/artifact/Homebrew completion.
- Local `develop` and `main` or `master` match upstream after cleanup.
