#!/usr/bin/env bash
# tests/run.sh - opt-in test suite for smart-claude-init's validator.
#
# Exercises scripts/validate-claudemd.sh: section presence, placeholder and
# leftover-TODO detection, the --general contract, and invocation errors. All
# fixtures are built under a throwaway mktemp dir so the host tree is untouched.
#
# This is NOT wired into any CI workflow - run it explicitly after editing the
# validator or the bundled templates to confirm nothing regressed.
#
# Usage:
#   bash tests/run.sh
#   bash tests/run.sh --verbose
#
# Exit codes:
#   0  all tests passed
#   1  one or more tests failed
#   2  setup error (validator/templates missing)

set -uo pipefail

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
SKILL_ROOT=$(dirname "$SCRIPT_DIR")
VALIDATE="$SKILL_ROOT/scripts/validate-claudemd.sh"
CODE_TMPL="$SKILL_ROOT/assets/claude-code.md.tmpl"
GEN_TMPL="$SKILL_ROOT/assets/claude-general.md.tmpl"

for f in "$VALIDATE" "$CODE_TMPL" "$GEN_TMPL"; do
  if [[ ! -f "$f" ]]; then
    echo "ERROR: required file not found: $f" >&2
    exit 2
  fi
done

VERBOSE=0
[[ "${1:-}" == "--verbose" ]] && VERBOSE=1

RED='\033[31m'; GRN='\033[32m'; BOLD='\033[1m'; RST='\033[0m'
PASS=0; FAIL=0; FAILED=()

pass() { PASS=$((PASS + 1)); [[ "$VERBOSE" -eq 1 ]] && printf "  ${GRN}PASS${RST} %s\n" "$1"; return 0; }
fail() { FAIL=$((FAIL + 1)); FAILED+=("$1"); printf "  ${RED}FAIL${RST} %s\n" "$1"; }

expect_exit() {  # <expected> <label> <actual>
  local expected="$1" label="$2" actual="$3"
  if [[ "$actual" -eq "$expected" ]]; then pass "$label"
  else fail "$label (expected exit $expected, got $actual)"; fi
}
expect_contains() {  # <label> <needle> <haystack>
  local label="$1" needle="$2" haystack="$3"
  if echo "$haystack" | grep -qF -- "$needle"; then pass "$label"
  else
    fail "$label (output missing '$needle')"
    [[ "$VERBOSE" -eq 1 ]] && echo "    --- output ---"$'\n'"$haystack"$'\n'"    ---"
  fi
}
section() { printf "\n${BOLD}== %s ==${RST}\n" "$1"; }

# ---- isolated fixture dir -------------------------------------------------
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

# A fully-filled, valid code CLAUDE.md (every {{TOKEN}} replaced).
GOOD_CODE="$TMPDIR/good-code.md"
sed -E 's/\{\{[^}]+\}\}/a concrete project-specific value/g' "$CODE_TMPL" > "$GOOD_CODE"

# A fully-filled, valid general CLAUDE.md.
GOOD_GEN="$TMPDIR/good-general.md"
sed -E 's/\{\{[^}]+\}\}/a concrete project-specific value/g' "$GEN_TMPL" > "$GOOD_GEN"

# ==========================================================================
section "happy path"
# ==========================================================================

bash "$VALIDATE" "$GOOD_CODE" >/dev/null 2>&1
expect_exit 0 "filled code file passes (default mode)" "$?"

bash "$VALIDATE" --general "$GOOD_GEN" >/dev/null 2>&1
expect_exit 0 "filled general file passes (--general)" "$?"

out=$(bash "$VALIDATE" "$GOOD_CODE" 2>&1); rc=$?
expect_exit 0 "valid file reports OK" "$rc"
expect_contains "OK message names code mode" "OK (code)" "$out"

# ==========================================================================
section "missing section"
# ==========================================================================

NO_PERF="$TMPDIR/no-perf.md"
grep -v '^## Performance$' "$GOOD_CODE" > "$NO_PERF"
out=$(bash "$VALIDATE" "$NO_PERF" 2>&1); rc=$?
expect_exit 1 "code file missing Performance rejected" "$rc"
expect_contains "missing-section error names the section" "MISSING SECTION: '## Performance'" "$out"

NO_BUG="$TMPDIR/no-bug.md"
grep -v '^## Bug Protocol$' "$GOOD_CODE" > "$NO_BUG"
out=$(bash "$VALIDATE" "$NO_BUG" 2>&1); rc=$?
expect_exit 1 "code file missing Bug Protocol rejected" "$rc"
expect_contains "missing Bug Protocol reported" "## Bug Protocol" "$out"

# ==========================================================================
section "unresolved placeholder"
# ==========================================================================

LEFT_TOKEN="$TMPDIR/left-token.md"
cp "$GOOD_CODE" "$LEFT_TOKEN"
printf '\n- leftover: {{STACK}}\n' >> "$LEFT_TOKEN"
out=$(bash "$VALIDATE" "$LEFT_TOKEN" 2>&1); rc=$?
expect_exit 1 "leftover {{placeholder}} rejected" "$rc"
expect_contains "placeholder error reported" "UNRESOLVED PLACEHOLDER" "$out"

# raw template itself must fail (placeholders present) - the workflow's sanity check
out=$(bash "$VALIDATE" "$CODE_TMPL" 2>&1); rc=$?
expect_exit 1 "raw code template fails (placeholders present)" "$rc"

# ==========================================================================
section "leftover TODO comment"
# ==========================================================================

LEFT_TODO="$TMPDIR/left-todo.md"
cp "$GOOD_CODE" "$LEFT_TODO"
printf '\n<!-- TODO: fill in the deploy story -->\n' >> "$LEFT_TODO"
out=$(bash "$VALIDATE" "$LEFT_TODO" 2>&1); rc=$?
expect_exit 1 "leftover <!-- TODO --> comment rejected" "$rc"
expect_contains "TODO error reported" "LEFTOVER TODO COMMENT" "$out"

# prose mentioning TODO without the HTML-comment form is NOT flagged
PROSE_TODO="$TMPDIR/prose-todo.md"
cp "$GOOD_CODE" "$PROSE_TODO"
printf '\n- Resolve all TODO items in the tracker before release.\n' >> "$PROSE_TODO"
bash "$VALIDATE" "$PROSE_TODO" >/dev/null 2>&1
expect_exit 0 "prose 'TODO' (no HTML comment) is not flagged" "$?"

# ==========================================================================
section "--general contract"
# ==========================================================================

# A code file lacks the general-only sections (Working Preferences, Quality
# Bar) so validating it as general must fail.
out=$(bash "$VALIDATE" --general "$GOOD_CODE" 2>&1); rc=$?
expect_exit 1 "code file fails the general contract" "$rc"
expect_contains "general mode flags missing Working Preferences" "## Working Preferences" "$out"

NO_QBAR="$TMPDIR/no-qbar.md"
grep -v '^## Quality Bar$' "$GOOD_GEN" > "$NO_QBAR"
out=$(bash "$VALIDATE" --general "$NO_QBAR" 2>&1); rc=$?
expect_exit 1 "general file missing Quality Bar rejected" "$rc"

# ==========================================================================
section "invocation errors"
# ==========================================================================

bash "$VALIDATE" >/dev/null 2>&1
expect_exit 2 "no file argument returns exit 2" "$?"

bash "$VALIDATE" "$TMPDIR/does-not-exist.md" >/dev/null 2>&1
expect_exit 2 "nonexistent file returns exit 2" "$?"

bash "$VALIDATE" --bogus "$GOOD_CODE" >/dev/null 2>&1
expect_exit 2 "unknown flag returns exit 2" "$?"

bash "$VALIDATE" --help >/dev/null 2>&1
expect_exit 0 "--help returns exit 0" "$?"

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
