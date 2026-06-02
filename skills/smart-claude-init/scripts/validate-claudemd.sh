#!/usr/bin/env bash
# validate-claudemd.sh - deterministic contract check for a generated CLAUDE.md.
#
# Verifies that a drafted CLAUDE.md is complete before it is written to a
# user's project:
#   1. every required `## ` section header is present
#   2. no unresolved `{{PLACEHOLDER}}` substitution token survives
#   3. no leftover `<!-- TODO ... -->` scaffold comment survives
#
# Modes:
#   (default)   8-section code contract
#   --general   4-section non-code contract
#
# Note: the placeholder check flags any `{{...}}` token. A CLAUDE.md that
# legitimately documents a templating language (Jinja, Handlebars, Vue, Hugo)
# may use `{{ var }}` in prose and trip this gate - that is a deliberate
# false-positive: review such a hit and, if intentional, the content is fine
# to write despite the non-zero exit.
#
# Usage:
#   bash validate-claudemd.sh <file>
#   bash validate-claudemd.sh --general <file>
#
# Exit codes:
#   0  contract satisfied - safe to write
#   1  one or more violations (each printed to stderr)
#   2  invocation error (missing arg, file not found, bad flag)

set -uo pipefail

MODE="code"
FILE=""

usage() {
  cat <<'EOF'
Usage: validate-claudemd.sh [--general] <file>

Validate a generated CLAUDE.md against the section/placeholder contract.

Options:
  --general   Use the 4-section non-code contract instead of the 8-section
              code contract.
  --help      Show this help.

Exit: 0 = valid, 1 = violations found, 2 = invocation error.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --general) MODE="general"; shift ;;
    --help|-h) usage; exit 0 ;;
    --) shift; FILE="${1:-}"; shift || true; break ;;
    -*) echo "validate-claudemd.sh: unknown flag: $1" >&2; usage >&2; exit 2 ;;
    *)
      if [[ -z "$FILE" ]]; then FILE="$1"; shift
      else echo "validate-claudemd.sh: unexpected extra argument: $1" >&2; exit 2; fi
      ;;
  esac
done

if [[ -z "$FILE" ]]; then
  echo "validate-claudemd.sh: missing <file> argument" >&2
  usage >&2
  exit 2
fi
if [[ ! -f "$FILE" ]]; then
  echo "validate-claudemd.sh: file not found: $FILE" >&2
  exit 2
fi

# Required `## ` section headers per mode (matched as fixed substrings).
if [[ "$MODE" == "general" ]]; then
  SECTIONS=(
    "Project Intent"
    "Working Preferences"
    "Quality Bar"
    "Core Principles"
  )
else
  SECTIONS=(
    "Project Intent"
    "Architecture & Stack"
    "Engineering Preferences"
    "Code Quality"
    "Testing"
    "Performance"
    "Bug Protocol"
    "Task Management & Core Principles"
  )
fi

violations=0

# 1. required section headers
for s in "${SECTIONS[@]}"; do
  if ! grep -Fq -- "## $s" "$FILE"; then
    echo "MISSING SECTION: '## $s'" >&2
    violations=$((violations + 1))
  fi
done

# 2. unresolved {{PLACEHOLDER}} tokens
if grep -nE '\{\{[^}]+\}\}' "$FILE" >/dev/null 2>&1; then
  while IFS= read -r line; do
    echo "UNRESOLVED PLACEHOLDER: $line" >&2
  done < <(grep -nE '\{\{[^}]+\}\}' "$FILE")
  count=$(grep -cE '\{\{[^}]+\}\}' "$FILE")
  violations=$((violations + count))
fi

# 3. leftover scaffold TODO comments (HTML-comment idiom used across this repo)
if grep -niF -- '<!-- TODO' "$FILE" >/dev/null 2>&1; then
  while IFS= read -r line; do
    echo "LEFTOVER TODO COMMENT: $line" >&2
  done < <(grep -niF -- '<!-- TODO' "$FILE")
  count=$(grep -ciF -- '<!-- TODO' "$FILE")
  violations=$((violations + count))
fi

if [[ "$violations" -gt 0 ]]; then
  echo "FAIL ($MODE): $violations issue(s) in $FILE" >&2
  exit 1
fi

echo "OK ($MODE): $FILE satisfies the CLAUDE.md contract."
exit 0
