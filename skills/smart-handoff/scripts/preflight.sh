#!/usr/bin/env bash
# preflight.sh - one read-only snapshot of everything a handoff doc needs.
#
# Usage:
#   bash scripts/preflight.sh [--slug <descriptive-name>]
#
# Prints shell-style KEY=value lines plus fenced blocks. Read-only: creates
# no directories and writes no files. The agent decides temp vs persist and
# runs its own `mkdir -p` on the chosen dir.
#
# Exit codes: 0 always (absence of git/humblskills is reported, not fatal).

set -uo pipefail

SLUG=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --slug) SLUG="${2:-}"; shift 2 ;;
    --help|-h)
      sed -n '2,12p' "$0" | sed 's/^# \{0,1\}//'
      exit 0 ;;
    *) echo "Unknown option: $1" >&2; exit 0 ;;
  esac
done

DATE="$(date +%F)"
TEMP_DIR="/tmp/.humblskills/handoffs"

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || true)"
if [[ -n "$REPO_ROOT" ]]; then
  PERSIST_DIR="$REPO_ROOT/.humblskills/handoffs"
else
  PERSIST_DIR="$PWD/.humblskills/handoffs"
fi

echo "HANDOFF_DATE=$DATE"
echo "TEMP_DIR=$TEMP_DIR"
echo "PERSIST_DIR=$PERSIST_DIR"
if [[ -n "$SLUG" ]]; then
  echo "TEMP_FILE=$TEMP_DIR/$SLUG-handoff-$DATE.md"
  echo "PERSIST_FILE=$PERSIST_DIR/$SLUG-handoff-$DATE.md"
fi

# ---------------------------------------------------------------- git state
if [[ -n "$REPO_ROOT" ]]; then
  echo "REPO_ROOT=$REPO_ROOT"
  echo "BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null)"
  echo "HEAD_SHA=$(git rev-parse --short HEAD 2>/dev/null)"
  echo "REMOTE_URL=$(git remote get-url origin 2>/dev/null || echo none)"
  UPSTREAM="$(git rev-parse --abbrev-ref '@{upstream}' 2>/dev/null || true)"
  echo "UPSTREAM=${UPSTREAM:-none}"
  if [[ -n "$UPSTREAM" ]]; then
    echo "UNPUSHED_COMMITS=$(git rev-list --count "$UPSTREAM"..HEAD 2>/dev/null || echo 0)"
  fi
  DIRTY="$(git status --porcelain 2>/dev/null | wc -l | tr -d ' ')"
  echo "DIRTY_FILES=$DIRTY"

  echo
  echo '```git-status'
  git status --short --branch 2>/dev/null | head -40
  echo '```'

  echo
  echo '```git-diffstat'
  git diff --stat HEAD 2>/dev/null | tail -30
  echo '```'

  echo
  echo '```recent-commits'
  git log --oneline -n 10 2>/dev/null
  echo '```'

  echo
  echo '```instruction-files'
  for f in CLAUDE.md AGENTS.md .cursorrules .cursor/rules README.md; do
    [[ -e "$REPO_ROOT/$f" ]] && echo "$f"
  done
  echo '```'
else
  echo "REPO_ROOT=none"
  echo "# not inside a git repo - Artifacts section must use absolute paths"
fi

# --------------------------------------------------- persist-dir git status
if [[ -n "$REPO_ROOT" ]]; then
  if git -C "$REPO_ROOT" check-ignore -q .humblskills 2>/dev/null; then
    echo "PERSIST_DIR_GITIGNORED=yes"
  else
    echo "PERSIST_DIR_GITIGNORED=no"
  fi
fi

# ------------------------------------------------------- existing handoffs
echo
echo '```existing-handoffs'
for d in "$TEMP_DIR" "$PERSIST_DIR"; do
  [[ -d "$d" ]] && find "$d" -maxdepth 1 -name '*-handoff-*.md' 2>/dev/null | sort
done
echo '```'

# -------------------------------------------------------- installed skills
echo
echo '```installed-skills'
if command -v humblskills >/dev/null 2>&1; then
  humblskills list 2>/dev/null | head -80
else
  echo "# humblskills CLI not on PATH - falling back to directory scan"
  for d in "$HOME/.claude/skills" "$HOME/.cursor/skills" ".claude/skills" ".cursor/skills" "skills"; do
    [[ -d "$d" ]] || continue
    echo "## $d"
    find "$d" -maxdepth 2 -name SKILL.md 2>/dev/null \
      | sed -E 's#/SKILL.md$##; s#.*/##' | sort
  done
fi
echo '```'
