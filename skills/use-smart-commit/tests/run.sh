#!/usr/bin/env bash
# tests/run.sh - opt-in test suite for use-smart-commit scripts.
#
# Exercises commit.sh validation, happy-path assembly, and footer matrix,
# plus a basic preflight.sh sanity check. Runs entirely inside a temporary
# git repo so the host tree is never touched.
#
# This is NOT wired into any CI workflow - run it explicitly after editing
# the scripts to confirm nothing regressed.
#
# Usage:
#   bash tests/run.sh
#   bash tests/run.sh --verbose
#
# Exit codes:
#   0  all tests passed
#   1  one or more tests failed

set -uo pipefail

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
SKILL_ROOT=$(dirname "$SCRIPT_DIR")
COMMIT_SH="$SKILL_ROOT/scripts/commit.sh"
PREFLIGHT_SH="$SKILL_ROOT/scripts/preflight.sh"

if [[ ! -x "$COMMIT_SH" ]]; then
  echo "ERROR: $COMMIT_SH not found or not executable" >&2
  exit 2
fi
if [[ ! -x "$PREFLIGHT_SH" ]]; then
  echo "ERROR: $PREFLIGHT_SH not found or not executable" >&2
  exit 2
fi

VERBOSE=0
[[ "${1:-}" == "--verbose" ]] && VERBOSE=1

RED='\033[31m'; GRN='\033[32m'; YEL='\033[33m'; BOLD='\033[1m'; RST='\033[0m'

PASS=0
FAIL=0
FAILED=()

# ---- isolated test repo --------------------------------------------------
TMPDIR=$(mktemp -d)
trap 'cd / && rm -rf "$TMPDIR"' EXIT
cd "$TMPDIR"
git init -q -b main >/dev/null
git config user.email "test@example.com"
git config user.name "Test User"
echo seed > seed.txt
git add seed.txt
git commit -q -m "chore: seed initial repo" 2>/dev/null

# Seed two scoped commits so preflight.sh has top-scopes content.
echo a > a.txt; git add a.txt; git commit -q -m "feat(foo): first scoped commit"
echo b > b.txt; git add b.txt; git commit -q -m "fix(foo): another scoped commit"

# ---- helpers -------------------------------------------------------------
pass() { PASS=$((PASS + 1)); [[ "$VERBOSE" -eq 1 ]] && printf "  ${GRN}PASS${RST} %s\n" "$1"; }
fail() { FAIL=$((FAIL + 1)); FAILED+=("$1"); printf "  ${RED}FAIL${RST} %s\n" "$1"; }

# fresh_change <id>: stage a unique file so a real commit has something to commit.
fresh_change() {
  echo "change-$1" > "stage-$1.txt"
  git add "stage-$1.txt"
}

# dry: runs commit.sh --dry-run with the given args; captures stdout+stderr and exit code.
dry() {
  bash "$COMMIT_SH" "$@" --dry-run 2>&1
}

# expect_exit <expected> <label> <output> <actual>
expect_exit() {
  local expected="$1" label="$2" actual="$3"
  if [[ "$actual" -eq "$expected" ]]; then
    pass "$label"
  else
    fail "$label (expected exit $expected, got $actual)"
  fi
}

# expect_contains <label> <needle> <haystack>
expect_contains() {
  local label="$1" needle="$2" haystack="$3"
  if echo "$haystack" | grep -qF -- "$needle"; then
    pass "$label"
  else
    fail "$label (output missing '$needle')"
    [[ "$VERBOSE" -eq 1 ]] && echo "    --- output ---"$'\n'"$haystack"$'\n'"    ---"
  fi
}

# expect_not_contains <label> <needle> <haystack>
expect_not_contains() {
  local label="$1" needle="$2" haystack="$3"
  if echo "$haystack" | grep -qF -- "$needle"; then
    fail "$label (output unexpectedly contained '$needle')"
    [[ "$VERBOSE" -eq 1 ]] && echo "    --- output ---"$'\n'"$haystack"$'\n'"    ---"
  else
    pass "$label"
  fi
}

section() { printf "\n${BOLD}== %s ==${RST}\n" "$1"; }

# ==========================================================================
# commit.sh : type validation
# ==========================================================================
section "commit.sh: type validation"

for t in feat fix perf refactor docs test build ci chore style; do
  out=$(dry --type "$t" --subject "valid subject for $t" --body "b" 2>&1)
  expect_exit 0 "valid type accepted: $t" "$?"
done

out=$(dry --type wibble --subject "x" --body "y" 2>&1); rc=$?
expect_exit 1 "invalid type rejected (wibble)" "$rc"
expect_contains "invalid type produces helpful error" "not one of" "$out"

out=$(bash "$COMMIT_SH" --subject "x" --body "y" --dry-run 2>&1); rc=$?
expect_exit 2 "missing --type returns exit 2" "$rc"

out=$(bash "$COMMIT_SH" --type fix --body "y" --dry-run 2>&1); rc=$?
expect_exit 2 "missing --subject returns exit 2" "$rc"

# ==========================================================================
# commit.sh : subject validation
# ==========================================================================
section "commit.sh: subject validation"

out=$(dry --type fix --scope parser --subject "handle empty input without panicking" --body "b" 2>&1); rc=$?
expect_exit 0 "subject 49 chars (under 72) accepted" "$rc"
expect_contains "subject line echoed in dry-run" "fix(parser): handle empty input without panicking" "$out"

out=$(dry --type feat --scope auth --subject "this subject is intentionally very long and overflows the 72-character limit" --body "b" 2>&1); rc=$?
expect_exit 1 "subject > 72 chars rejected" "$rc"
expect_contains "subject-too-long error mentions char count" "chars; max is 72" "$out"

out=$(dry --type fix --subject "trailing period bad." --body "b" 2>&1); rc=$?
expect_exit 1 "subject with trailing period rejected" "$rc"
expect_contains "trailing-period error message" "must not end with a period" "$out"

# 72-char subject exactly (with scope and ! it's already 9 chars; need 63 chars of subject text)
exact72="x: $(printf 'a%.0s' {1..69})"  # build subject so total = 72
out=$(dry --type chore --subject "$(printf 'a%.0s' {1..65})" --body "b" 2>&1); rc=$?
expect_exit 0 "subject at exactly 72 chars accepted" "$rc"

# ==========================================================================
# commit.sh : body form validation
# ==========================================================================
section "commit.sh: body form validation"

out=$(dry --type fix --subject "x" 2>&1); rc=$?
expect_exit 0 "no body is allowed (subject-only commit)" "$rc"

out=$(dry --type fix --subject "x" --changed "c" 2>&1); rc=$?
expect_exit 1 "partial labeled body (only --changed) rejected" "$rc"
expect_contains "partial labeled error message" "must all be provided together" "$out"

out=$(dry --type fix --subject "x" --changed "c" --why "w" 2>&1); rc=$?
expect_exit 1 "partial labeled body (missing --impact) rejected" "$rc"

out=$(dry --type fix --subject "x" --body "free-form" --changed "c" --why "w" --impact "i" 2>&1); rc=$?
expect_exit 1 "multiple body forms (--body + labeled) rejected" "$rc"
expect_contains "multiple-forms error message" "pick exactly one body form" "$out"

# ==========================================================================
# commit.sh : skip-CI token block
# ==========================================================================
section "commit.sh: skip-CI token block"

for token in '[skip ci]' '[ci skip]' '[no ci]' '[skip actions]' '[actions skip]'; do
  out=$(dry --type fix --subject "valid" --body "body contains $token literally" 2>&1); rc=$?
  expect_exit 1 "skip-CI token rejected in body: $token" "$rc"
done

# token in subject
out=$(dry --type fix --subject "subject with [skip ci] in it" --body "b" 2>&1); rc=$?
expect_exit 1 "skip-CI token rejected in subject" "$rc"

# case-insensitive
out=$(dry --type fix --subject "x" --body "body has [SKIP CI] uppercase" 2>&1); rc=$?
expect_exit 1 "skip-CI token rejected case-insensitively" "$rc"

# hyphenated form NOT rejected
out=$(dry --type docs --subject "explain the skip-ci marker behaviour" --body "Use skip-ci with a hyphen safely." 2>&1); rc=$?
expect_exit 0 "hyphenated skip-ci form accepted" "$rc"

# ==========================================================================
# commit.sh : labeled body assembly
# ==========================================================================
section "commit.sh: labeled body assembly"

out=$(dry --type fix --scope parser \
  --subject "handle empty input without panicking" \
  --changed "ParseError::Empty replaces unwrap()." \
  --why "Empty stdin was crashing." \
  --impact "Resolves #214." 2>&1); rc=$?
expect_exit 0 "labeled body accepted" "$rc"
expect_contains "labeled body has Changed: section" "Changed:" "$out"
expect_contains "labeled body has Why: section" "Why:" "$out"
expect_contains "labeled body has Impact: section" "Impact:" "$out"
expect_contains "labeled body preserves Changed content" "ParseError::Empty replaces unwrap()." "$out"
expect_contains "labeled body preserves Why content" "Empty stdin was crashing." "$out"
expect_contains "labeled body preserves Impact content" "Resolves #214." "$out"

# ==========================================================================
# commit.sh : breaking change
# ==========================================================================
section "commit.sh: breaking change"

out=$(dry --type feat --scope api --breaking --subject "drop /v1 endpoint" --body "b" 2>&1); rc=$?
expect_exit 0 "breaking change accepted" "$rc"
expect_contains "breaking change produces ! in subject" "feat(api)!: drop /v1 endpoint" "$out"

# ==========================================================================
# commit.sh : footer matrix
# ==========================================================================
section "commit.sh: footer matrix"

# default-on
out=$(dry --type fix --subject "x" --body "b" 2>&1)
expect_contains "footer on by default" 'Authored by humblSKILLS; "use-smart-commit"' "$out"
expect_contains "footer state reports 'on (default)'" "on (default)" "$out"

# --no-footer flag
out=$(dry --type fix --subject "x" --body "b" --no-footer 2>&1)
expect_not_contains "--no-footer suppresses footer" 'Authored by humblSKILLS;' "$out"
expect_contains "--no-footer state reports flag" "off (--no-footer flag)" "$out"

# env var
out=$(HUMBLSKILLS_COMMIT_NO_FOOTER=1 dry --type fix --subject "x" --body "b" 2>&1)
expect_not_contains "env var suppresses footer" 'Authored by humblSKILLS;' "$out"
expect_contains "env var state reports env" "HUMBLSKILLS_COMMIT_NO_FOOTER env" "$out"

# marker file
mkdir -p .humblskills
touch .humblskills/no-footer
out=$(dry --type fix --subject "x" --body "b" 2>&1)
expect_not_contains "marker file suppresses footer" 'Authored by humblSKILLS;' "$out"
expect_contains "marker file state reports marker" ".humblskills/no-footer marker" "$out"
rm -rf .humblskills

# precedence: --no-footer beats absence of marker (sanity)
out=$(HUMBLSKILLS_COMMIT_NO_FOOTER=0 dry --type fix --subject "x" --body "b" 2>&1)
expect_contains "env=0 falls back to default-on" "on (default)" "$out"

# ==========================================================================
# commit.sh : real commit (not dry-run)
# ==========================================================================
section "commit.sh: real commit assembly"

fresh_change real-1
bash "$COMMIT_SH" --type feat --scope test \
  --subject "real commit with labeled body" \
  --changed "Added a test file." \
  --why "To verify commit.sh produces a real commit." \
  --impact "Confirms end-to-end script behaviour." >/dev/null 2>&1
rc=$?
expect_exit 0 "real commit (labeled) exits 0" "$rc"
body=$(git log -1 --format=%B)
expect_contains "real commit body has subject" "feat(test): real commit with labeled body" "$body"
expect_contains "real commit body has Changed section" "Changed:" "$body"
expect_contains "real commit body has footer" 'Authored by humblSKILLS; "use-smart-commit"' "$body"

fresh_change real-2
bash "$COMMIT_SH" --type docs --subject "trivial commit free-form body" \
  --body "Single paragraph body for a trivial commit." --no-footer >/dev/null 2>&1
rc=$?
expect_exit 0 "real commit (free-form, no footer) exits 0" "$rc"
body=$(git log -1 --format=%B)
expect_contains "free-form body preserved verbatim" "Single paragraph body for a trivial commit." "$body"
expect_not_contains "no-footer commit omits footer" 'Authored by humblSKILLS;' "$body"

# ==========================================================================
# preflight.sh sanity
# ==========================================================================
section "preflight.sh: sanity"

out=$(bash "$PREFLIGHT_SH" 2>&1); rc=$?
expect_exit 0 "preflight.sh exits 0 in a git repo" "$rc"
expect_contains "preflight has STATUS section" "=== STATUS" "$out"
expect_contains "preflight has DIFF STAT section" "=== DIFF STAT" "$out"
expect_contains "preflight has TOP SCOPES section" "=== TOP SCOPES" "$out"
expect_contains "preflight has FOOTER STATE section" "=== FOOTER STATE" "$out"
expect_contains "preflight reports scope from seeded history" "foo" "$out"
expect_contains "preflight reports footer on by default" "on (default" "$out"

# preflight outside a git repo
out=$(cd /tmp && bash "$PREFLIGHT_SH" 2>&1); rc=$?
expect_exit 1 "preflight.sh exits 1 outside a git repo" "$rc"

# ==========================================================================
# summary
# ==========================================================================
printf "\n${BOLD}Summary:${RST}\n"
printf "  passed: ${GRN}%d${RST}\n" "$PASS"
printf "  failed: ${RED}%d${RST}\n" "$FAIL"

if [[ "$FAIL" -gt 0 ]]; then
  printf "\n${RED}FAILED CASES:${RST}\n"
  for c in "${FAILED[@]}"; do printf "  - %s\n" "$c"; done
  exit 1
fi

printf "\n${GRN}OK: all tests passed.${RST}\n"
exit 0
