#!/usr/bin/env bash
# preflight.sh - pre-commit inspection helper for use-smart-commit.
#
# Emits a structured snapshot the agent reads at workflow step 1:
#   1. git status (porcelain)
#   2. git diff --stat (unstaged + staged)
#   3. top 10 scopes from the last 100 commits
#   4. footer state with reason
#
# Usage:
#   bash scripts/preflight.sh
#
# Exit codes:
#   0  always (informational helper; failures emit to stderr but don't exit)
#   1  not in a git repository

set -uo pipefail

ROOT=$(git rev-parse --show-toplevel 2>/dev/null) || {
  echo "ERROR: not in a git repository" >&2
  exit 1
}

echo "=== STATUS (git status --porcelain) ==="
STATUS_OUT=$(git status --porcelain)
if [[ -z "$STATUS_OUT" ]]; then
  echo "(working tree clean - nothing to commit)"
else
  echo "$STATUS_OUT"
fi
echo ""

echo "=== DIFF STAT ==="
echo "--- unstaged ---"
UNSTAGED=$(git diff --stat)
[[ -z "$UNSTAGED" ]] && echo "(none)" || echo "$UNSTAGED"
echo "--- staged ---"
STAGED=$(git diff --stat --staged)
[[ -z "$STAGED" ]] && echo "(none)" || echo "$STAGED"
echo ""

echo "=== TOP SCOPES (last 100 commits, by frequency) ==="
SCOPES=$(git log --oneline -100 2>/dev/null \
  | sed -nE 's/^[a-f0-9]+ [a-z]+\(([^)]+)\)!?:.*$/\1/p' \
  | sort | uniq -c | sort -rn | head -10)
if [[ -z "$SCOPES" ]]; then
  echo "(no scoped commits found in last 100; pick a fresh scope)"
else
  echo "$SCOPES"
fi
echo ""

echo "=== FOOTER STATE ==="
if [[ "${HUMBLSKILLS_COMMIT_NO_FOOTER:-0}" == "1" ]]; then
  echo "off (HUMBLSKILLS_COMMIT_NO_FOOTER env var is set)"
elif [[ -f "$ROOT/.humblskills/no-footer" ]]; then
  echo "off (.humblskills/no-footer marker file present)"
else
  echo "on (default - agent may still apply per-conversation or memory opt-out)"
fi
