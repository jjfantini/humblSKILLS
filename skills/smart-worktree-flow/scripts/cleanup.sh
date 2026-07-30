#!/usr/bin/env bash
set -euo pipefail

usage() {
  printf 'Usage: %s <feature-branch> [worktree-path] [production-branch] [integration-branch]\n' "$0"
  printf 'Example: %s feat/add-data ../feat-add-data main develop\n' "$0"
}

if [[ "${1:-}" == "--help" ]] || [[ $# -lt 1 ]] || [[ $# -gt 4 ]]; then
  usage
  exit 0
fi

FEATURE_BRANCH="$1"
WORKTREE_PATH="${2:-}"
PRODUCTION_BRANCH="${3:-main}"
INTEGRATION_BRANCH="${4:-develop}"

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  printf 'ERROR: not inside a git worktree\n' >&2
  exit 1
fi

CURRENT_BRANCH="$(git branch --show-current)"
if [[ "$CURRENT_BRANCH" == "$FEATURE_BRANCH" ]]; then
  printf 'ERROR: checkout another branch before deleting %s\n' "$FEATURE_BRANCH" >&2
  exit 1
fi

if [[ -n "$(git status --porcelain)" ]]; then
  printf 'ERROR: working tree is dirty; cleanup requires a clean repo\n' >&2
  exit 1
fi

git fetch origin --prune --quiet

for BRANCH in "$PRODUCTION_BRANCH" "$INTEGRATION_BRANCH"; do
  if git show-ref --verify --quiet "refs/heads/${BRANCH}"; then
    git checkout "$BRANCH" >/dev/null
  else
    git checkout -b "$BRANCH" "origin/${BRANCH}" >/dev/null
  fi
  git merge --ff-only "origin/${BRANCH}"
done

git checkout "$PRODUCTION_BRANCH" >/dev/null

if [[ -n "$WORKTREE_PATH" ]]; then
  WORKTREE_ABS="$(cd "$(dirname "$WORKTREE_PATH")" && pwd)/$(basename "$WORKTREE_PATH")"
  WORKTREE_REGISTERED=false
  while IFS= read -r LINE; do
    case "$LINE" in
      worktree\ *)
        if [[ "${LINE#worktree }" == "$WORKTREE_ABS" ]]; then
          WORKTREE_REGISTERED=true
        fi
        ;;
    esac
  done < <(git worktree list --porcelain)

  if [[ "$WORKTREE_REGISTERED" == true ]]; then
    git worktree remove "$WORKTREE_ABS"
  elif [[ -d "$WORKTREE_ABS" ]]; then
    printf 'ERROR: path exists but is not a registered worktree: %s\n' "$WORKTREE_PATH" >&2
    exit 1
  fi
fi

if git show-ref --verify --quiet "refs/heads/${FEATURE_BRANCH}"; then
  git branch -d "$FEATURE_BRANCH"
fi

if git ls-remote --exit-code --heads origin "$FEATURE_BRANCH" >/dev/null 2>&1; then
  git push origin --delete "$FEATURE_BRANCH"
fi

git worktree prune

printf 'Cleanup complete for branch: %s\n' "$FEATURE_BRANCH"
