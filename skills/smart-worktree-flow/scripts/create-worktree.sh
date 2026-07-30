#!/usr/bin/env bash
set -euo pipefail

usage() {
  printf 'Usage: %s <type> <slug> [base-branch]\n' "$0"
  printf 'Example: %s feat add-data origin/develop\n' "$0"
}

if [[ "${1:-}" == "--help" ]] || [[ $# -lt 2 ]] || [[ $# -gt 3 ]]; then
  usage
  exit 0
fi

TYPE="$1"
SLUG="$2"
BASE_BRANCH="${3:-origin/develop}"

case "$TYPE" in
  feat|fix|docs|test|refactor|perf|build|ci|chore) ;;
  *)
    printf 'ERROR: unsupported conventional type: %s\n' "$TYPE" >&2
    exit 1
    ;;
esac

if [[ ! "$SLUG" =~ ^[a-z0-9]+(-[a-z0-9]+)*$ ]]; then
  printf 'ERROR: slug must be kebab-case, got: %s\n' "$SLUG" >&2
  exit 1
fi

WORKTREE_DIR="../${TYPE}-${SLUG}"
BRANCH_NAME="${TYPE}/${SLUG}"

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  printf 'ERROR: not inside a git worktree\n' >&2
  exit 1
fi

git fetch origin --quiet

if git show-ref --verify --quiet "refs/heads/${BRANCH_NAME}"; then
  printf 'ERROR: local branch already exists: %s\n' "$BRANCH_NAME" >&2
  exit 1
fi

if git ls-remote --exit-code --heads origin "$BRANCH_NAME" >/dev/null 2>&1; then
  printf 'ERROR: remote branch already exists: %s\n' "$BRANCH_NAME" >&2
  exit 1
fi

if [[ -e "$WORKTREE_DIR" ]]; then
  printf 'ERROR: worktree path already exists: %s\n' "$WORKTREE_DIR" >&2
  exit 1
fi

git rev-parse --verify "$BASE_BRANCH" >/dev/null
git worktree add "$WORKTREE_DIR" -b "$BRANCH_NAME" "$BASE_BRANCH"

printf 'Created worktree: %s\n' "$WORKTREE_DIR"
printf 'Created branch:   %s\n' "$BRANCH_NAME"
