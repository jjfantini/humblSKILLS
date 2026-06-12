# User Request

Date: 2026-06-12

Create a smart skill called `/use-worktree-flow` that details the coding
development workflow:

1. Start in a local repo.
2. Create a new worktree and a new branch with similar, directed names that
   start with conventional prefixes. Example: worktree `feat-add-data`,
   branch `feat/add-data`.
3. PR the worktree branch into `develop`. On green CI/CD and green tests, merge
   it into `develop`.
4. Once in `develop`, open a PR into `main` or `master`. Ask whether the user
   wants auto-merging, which allows merge on green tests, linting,
   verifications, and CI/CD. If the user wants manual review, respect that.
5. If merging into `main` or `master` generates a release PR, ask whether the
   user wants the agent to merge it or whether they want to manually cut and
   review the release.
6. Once the version is cut, make the local repo reflect upstream on `develop`
   and `main` or `master`, then clean up the worktree and feature branch. Keep
   commit history, but leave no stale local or remote branches or worktrees.

Additional requirements:

- Ask whether alternative parallel work is happening and briefly research that
  locally, including other Codex, Claude, or Cursor agents.
- Default to worktrees for the most reliable isolation.
- Support two paths: Vibe mode, where the agent has autonomy and the user
  reviews after the work; and HITL mode, where the user verifies gates.
- Gather intent upfront with a clear question set and default fallbacks when
  the user says "I defer to you".
- Use best practices from `/use-smart-skill`.
