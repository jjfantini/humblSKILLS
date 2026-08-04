#!/usr/bin/env bash
# run.sh - isolated checks for smart-handoff's two scripts.
#
#   bash tests/run.sh [--verbose]
#
# Every fixture secret is assembled at runtime from harmless parts so no
# literal credential shape is ever committed to this repo.

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_ROOT="$(dirname "$SCRIPT_DIR")"
SCAN="$SKILL_ROOT/scripts/scan-secrets.sh"
PREFLIGHT="$SKILL_ROOT/scripts/preflight.sh"

VERBOSE=false
[[ "${1:-}" == "--verbose" ]] && VERBOSE=true

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

pass=0
fail=0

ok()   { pass=$((pass + 1)); echo "  ok   $1"; }
bad()  { fail=$((fail + 1)); echo "  FAIL $1"; }

# assert_exit <expected-code> <label> <file> [extra-args...]
assert_exit() {
  local want="$1" label="$2"; shift 2
  local out; out="$("$SHELL_BIN" "$SCAN" "$@" 2>&1)"; local got=$?
  if [[ "$got" == "$want" ]]; then ok "$label"; else
    bad "$label (want exit $want, got $got)"
    echo "$out" | sed 's/^/       /'
  fi
  $VERBOSE && echo "$out" | sed 's/^/       /'
}

# assert_reports <pattern-name> <label> <file>
# Capture first, then grep: piping straight into `grep -q` kills the scanner
# with SIGPIPE, and pipefail turns that into a bogus failure.
assert_reports() {
  local name="$1" label="$2" file="$3" out
  out="$("$SHELL_BIN" "$SCAN" "$file" 2>&1)"
  if grep -q "$name" <<< "$out"; then
    ok "$label"
  else
    bad "$label (expected report of '$name')"
    echo "$out" | sed 's/^/       /'
  fi
}

SHELL_BIN=/bin/bash   # macOS bash 3.2 - the floor the scripts must clear

echo "smart-handoff tests ($SHELL_BIN)"
echo

# ---------------------------------------------------------------- fixtures
# Assembled from parts: nothing here is a scannable literal in the repo.
AWS="AKIA""ZZTESTKEY000000Q"           # assembled at runtime; never a literal here
GH="ghp_""0123456789abcdefghij0123456789abcd"
ANT="sk-ant-""api03-0123456789abcdefghij"
JWT="eyJhbGciOiJIUzI1NiJ9"".eyJzdWIiOiIxMjM0NTY3ODkwIn0"".abcdefghij"
PEM="-----BEGIN RSA PRIVATE KEY-----"

cat > "$TMP/clean.md" <<EOF
# Retry Backoff — Session Handoff

## Access & Secrets
- GitHub token: \`gh auth token --user someone\`
- Anthropic key: \`op read "op://Personal/anthropic/api-key"\`
- UploadThing: \`security find-generic-password -s uploadthing -a app-files -w\`
- Staging DB: postgres://app@db.example.com:5432/app (password in 1Password)
- Placeholder: API_TOKEN=<YOUR_TOKEN_HERE>
- Env form: export API_TOKEN=\$LINEAR_API_TOKEN
EOF

cat > "$TMP/leaky.md" <<EOF
# Bad Handoff
aws_access_key_id = $AWS
GH_TOKEN=$GH
ANTHROPIC_API_KEY=$ANT
Authorization: Bearer $JWT
$PEM
db: postgres://admin:hunter2@db.internal:5432/app
EOF

cat > "$TMP/warnonly.md" <<EOF
# Warn Only
Contact the reviewer at dev.person@company.io before step 3.
config: client_secret = "placeholder-but-long-enough"
EOF

# ------------------------------------------------------------------- cases
assert_exit 0 "clean doc passes (op:// and keychain refs are not leaks)" "$TMP/clean.md"
assert_exit 1 "leaky doc fails" "$TMP/leaky.md"
assert_exit 0 "warn-only doc passes without --strict" "$TMP/warnonly.md"
assert_exit 1 "warn-only doc fails with --strict" --strict "$TMP/warnonly.md"
assert_exit 2 "missing file is an invocation error" "$TMP/nope.md"
assert_exit 2 "no arguments is an invocation error"

assert_reports "aws-access-key"     "detects AWS access key"      "$TMP/leaky.md"
assert_reports "github-token"       "detects GitHub token"        "$TMP/leaky.md"
assert_reports "anthropic-key"      "detects Anthropic key"       "$TMP/leaky.md"
assert_reports "jwt"                "detects JWT"                 "$TMP/leaky.md"
assert_reports "private-key-block"  "detects private key block"   "$TMP/leaky.md"
assert_reports "basic-auth-url"     "detects credential in URL"   "$TMP/leaky.md"
assert_reports "email-address"      "warns on email address"      "$TMP/warnonly.md"

# The report must never echo the raw secret back.
report="$("$SHELL_BIN" "$SCAN" "$TMP/leaky.md" 2>&1)"
if grep -q "$AWS" <<< "$report"; then
  bad "report masks the matched secret"
else
  ok "report masks the matched secret"
fi

# ------------------------------------------------------------- preflight.sh
pf="$("$SHELL_BIN" "$PREFLIGHT" --slug demo-task 2>&1)"; pf_code=$?
if [[ $pf_code -eq 0 ]]; then ok "preflight.sh exits 0"; else bad "preflight.sh exits 0 (got $pf_code)"; fi

for key in HANDOFF_DATE TEMP_DIR PERSIST_DIR TEMP_FILE PERSIST_FILE; do
  if grep -q "^$key=" <<< "$pf"; then ok "preflight prints $key"; else bad "preflight prints $key"; fi
done

if grep -qE '^TEMP_FILE=/tmp/\.humblskills/handoffs/demo-task-handoff-[0-9]{4}-[0-9]{2}-[0-9]{2}\.md$' <<< "$pf"; then
  ok "TEMP_FILE matches the naming contract"
else
  bad "TEMP_FILE matches the naming contract"
  grep '^TEMP_FILE=' <<< "$pf" | sed 's/^/       /'
fi

if grep -q 'installed-skills' <<< "$pf"; then ok "preflight enumerates skills"; else bad "preflight enumerates skills"; fi

$VERBOSE && echo "$pf" | sed 's/^/       /'

echo
echo "$pass passed, $fail failed"
[[ $fail -eq 0 ]] || exit 1
