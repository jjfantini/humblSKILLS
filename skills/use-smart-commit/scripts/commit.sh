#!/usr/bin/env bash
# commit.sh - canonical commit invoker for use-smart-commit.
#
# Standardizes commit creation: validates subject format, blocks skip-CI
# tokens, assembles the Changed/Why/Impact body (or a free-form body for
# trivial commits), and appends the humblSKILLS authorship footer with
# default-on opt-out semantics.
#
# Usage:
#   commit.sh --type <type> [--scope <scope>] [--breaking]
#             --subject "<subject>"
#             ( --changed "<...>" --why "<...>" --impact "<...>"
#             | --body "<one-paragraph>"
#             | --body-file <path> )
#             [--no-footer] [--dry-run]
#
# Validations:
#   - --type in: feat|fix|perf|refactor|docs|test|build|ci|chore|style
#   - subject is non-empty, ≤ 72 chars including type/scope/!/colon
#   - subject has no trailing period
#   - no skip-CI tokens anywhere in the assembled message
#   - exactly one body form is provided
#   - if --changed/--why/--impact: all three must be present
#
# Footer (default-on, precedence high → low):
#   1. --no-footer CLI flag
#   2. HUMBLSKILLS_COMMIT_NO_FOOTER=1 environment variable
#   3. .humblskills/no-footer marker file at repo root
#   4. default: include the authorship footer
#
# Exit codes:
#   0  commit created (or dry-run printed) successfully
#   1  validation failure (no commit made)
#   2  invocation error (bad flags, missing required args)

set -uo pipefail

FOOTER='Authored by humblSKILLS; "use-smart-commit"'
SUBJECT_MAX=72
VALID_TYPES='feat|fix|perf|refactor|docs|test|build|ci|chore|style'
SKIP_CI_PATTERNS='\[skip ci\]|\[ci skip\]|\[no ci\]|\[skip actions\]|\[actions skip\]'

TYPE=""
SCOPE=""
BREAKING=0
SUBJECT=""
BODY=""
CHANGED=""
WHY=""
IMPACT=""
BODY_FILE=""
NO_FOOTER=0
DRY_RUN=0

usage() {
  sed -n '2,40p' "$0" | sed 's/^# \{0,1\}//'
  exit 2
}

err() { echo "ERROR: $*" >&2; }

while [[ $# -gt 0 ]]; do
  case "$1" in
    --type)       TYPE="${2:-}"; shift 2 ;;
    --scope)      SCOPE="${2:-}"; shift 2 ;;
    --breaking)   BREAKING=1; shift ;;
    --subject)    SUBJECT="${2:-}"; shift 2 ;;
    --body)       BODY="${2:-}"; shift 2 ;;
    --changed)    CHANGED="${2:-}"; shift 2 ;;
    --why)        WHY="${2:-}"; shift 2 ;;
    --impact)     IMPACT="${2:-}"; shift 2 ;;
    --body-file)  BODY_FILE="${2:-}"; shift 2 ;;
    --no-footer)  NO_FOOTER=1; shift ;;
    --dry-run)    DRY_RUN=1; shift ;;
    -h|--help)    usage ;;
    *)            err "unknown option: $1"; usage ;;
  esac
done

# --- required args ---------------------------------------------------------
if [[ -z "$TYPE" ]]; then err "--type is required"; exit 2; fi
if [[ -z "$SUBJECT" ]]; then err "--subject is required"; exit 2; fi

# --- type validation -------------------------------------------------------
if ! [[ "$TYPE" =~ ^(${VALID_TYPES})$ ]]; then
  err "--type='$TYPE' is not one of: ${VALID_TYPES//|/, }"
  exit 1
fi

# --- body form validation: exactly one -------------------------------------
BODY_FORMS=0
[[ -n "$BODY" ]] && BODY_FORMS=$((BODY_FORMS + 1))
[[ -n "$BODY_FILE" ]] && BODY_FORMS=$((BODY_FORMS + 1))
HAVE_LABELED=0
if [[ -n "$CHANGED$WHY$IMPACT" ]]; then
  if [[ -z "$CHANGED" || -z "$WHY" || -z "$IMPACT" ]]; then
    err "--changed, --why, and --impact must all be provided together"
    exit 1
  fi
  HAVE_LABELED=1
  BODY_FORMS=$((BODY_FORMS + 1))
fi
if [[ "$BODY_FORMS" -gt 1 ]]; then
  err "pick exactly one body form: --changed/--why/--impact triple, --body, or --body-file"
  exit 1
fi

# --- subject assembly + validation -----------------------------------------
SCOPE_PART=""
[[ -n "$SCOPE" ]] && SCOPE_PART="($SCOPE)"
BREAK_PART=""
[[ "$BREAKING" -eq 1 ]] && BREAK_PART="!"
FULL_SUBJECT="${TYPE}${SCOPE_PART}${BREAK_PART}: ${SUBJECT}"

SUBJ_LEN=${#FULL_SUBJECT}
if [[ "$SUBJ_LEN" -gt "$SUBJECT_MAX" ]]; then
  err "subject line is $SUBJ_LEN chars; max is $SUBJECT_MAX"
  err "        '$FULL_SUBJECT'"
  exit 1
fi
case "$FULL_SUBJECT" in
  *.) err "subject must not end with a period"; exit 1 ;;
esac

# --- footer resolution -----------------------------------------------------
FOOTER_STATE=""
INCLUDE_FOOTER=0
ROOT=$(git rev-parse --show-toplevel 2>/dev/null) || ROOT="."
if [[ "$NO_FOOTER" -eq 1 ]]; then
  FOOTER_STATE="off (--no-footer flag)"
elif [[ "${HUMBLSKILLS_COMMIT_NO_FOOTER:-0}" == "1" ]]; then
  FOOTER_STATE="off (HUMBLSKILLS_COMMIT_NO_FOOTER env)"
elif [[ -f "$ROOT/.humblskills/no-footer" ]]; then
  FOOTER_STATE="off (.humblskills/no-footer marker)"
else
  FOOTER_STATE="on (default)"
  INCLUDE_FOOTER=1
fi

# --- body assembly ---------------------------------------------------------
BODY_TEXT=""
if [[ "$HAVE_LABELED" -eq 1 ]]; then
  BODY_TEXT="Changed:
$CHANGED

Why:
$WHY

Impact:
$IMPACT"
elif [[ -n "$BODY" ]]; then
  BODY_TEXT="$BODY"
elif [[ -n "$BODY_FILE" ]]; then
  if [[ ! -f "$BODY_FILE" ]]; then
    err "--body-file not found: $BODY_FILE"
    exit 1
  fi
  BODY_TEXT=$(cat "$BODY_FILE")
fi

# --- message assembly ------------------------------------------------------
MESSAGE="$FULL_SUBJECT"
if [[ -n "$BODY_TEXT" ]]; then
  MESSAGE="$MESSAGE

$BODY_TEXT"
fi
if [[ "$INCLUDE_FOOTER" -eq 1 ]]; then
  MESSAGE="$MESSAGE

$FOOTER"
fi

# --- skip-CI token check ---------------------------------------------------
if echo "$MESSAGE" | grep -iE "$SKIP_CI_PATTERNS" >/dev/null; then
  err "commit message contains a skip-CI token"
  err "        (matching ${SKIP_CI_PATTERNS//\\/}, case-insensitive)"
  err "        GitHub Actions would suppress every workflow for this push."
  err "        If you must discuss the mechanism, write 'skip-ci' (hyphenated, no brackets)."
  exit 1
fi

# --- dry-run or commit -----------------------------------------------------
echo "--- subject ($SUBJ_LEN/$SUBJECT_MAX chars) ---"
echo "$FULL_SUBJECT"
echo "--- footer: $FOOTER_STATE ---"
if [[ "$DRY_RUN" -eq 1 ]]; then
  echo ""
  echo "$MESSAGE"
  echo ""
  echo "(dry-run; no commit made)"
  exit 0
fi

echo "$MESSAGE" | git commit -F -
