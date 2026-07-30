#!/usr/bin/env bash
# auth.sh - authenticate the skill against an UploadThing project.
#
# An API key maps 1:1 to a single UploadThing app/project. This script resolves,
# validates, and (optionally) persists the key. The key is never printed.
#
# Usage:
#   bash scripts/auth.sh --status          # show where the key resolves from (masked)
#   bash scripts/auth.sh --check           # validate the key via POST /v7/getAppInfo
#   printf '%s' "$KEY" | bash scripts/auth.sh --save   # persist key (stdin) to a 0600 file
#   UPLOADTHING_API_KEY=sk_live_... bash scripts/auth.sh --save   # persist from env
#
# Exit codes: 0 ok | 1 validation/usage failure | 2 missing dependency
set -uo pipefail
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# shellcheck source=lib.sh
. "$SCRIPT_DIR/lib.sh"
ut_preflight

usage() {
  sed -n '2,16p' "$SCRIPT_DIR/auth.sh" | sed 's/^# \{0,1\}//'
  exit "${1:-0}"
}

guidance() {
  cat >&2 <<EOF

No API key resolved. To authenticate:
  1. Open the dashboard:  https://uploadthing.com/dashboard
  2. Select your app, go to the 'API Keys' tab, copy the key (sk_live_...).
     (The V7 tab shows a base64 UPLOADTHING_TOKEN; either works - its apiKey is decoded.)
  3. Provide it to this skill via ONE of:
       export UPLOADTHING_API_KEY=sk_live_...           # session only, not persisted
       printf '%s' 'sk_live_...' | bash scripts/auth.sh --save   # persisted to $UT_CONFIG_FILE (0600)
EOF
}

cmd_status() {
  local src; src=$(ut_key_source)
  echo "key source: $src"
  if [ "$src" = "none" ]; then
    guidance
    return 1
  fi
  local key; key=$(ut_resolve_key) || { guidance; return 1; }
  echo "key (masked): $(ut_mask "$key")"
  echo "config file: $UT_CONFIG_FILE"
  echo "Run 'bash scripts/auth.sh --check' to validate against the API."
}

cmd_check() {
  local resp
  if resp=$(ut_api "v7/getAppInfo" '{}'); then
    echo "auth OK"
    printf '%s' "$resp" | jq .
  else
    echo "auth FAILED" >&2
    printf '%s' "$resp" | jq . >&2 2>/dev/null || printf '%s\n' "$resp" >&2
    return 1
  fi
}

cmd_save() {
  local newkey="${UPLOADTHING_API_KEY:-}"
  if [ -z "$newkey" ]; then
    if [ -t 0 ]; then
      ut_die "no key on stdin. Pipe it: printf '%s' \"\$KEY\" | bash scripts/auth.sh --save"
    fi
    IFS= read -r newkey || true
  fi
  newkey=$(printf '%s' "$newkey" | tr -d '\r\n')
  [ -n "$newkey" ] || ut_die "refusing to save an empty key"

  local resp
  resp=$(UPLOADTHING_API_KEY="$newkey" ut_api "v7/getAppInfo" '{}') \
    || ut_die "key did not validate: $(printf '%s' "$resp" | jq -r '.error // .' 2>/dev/null || printf '%s' "$resp")"

  mkdir -p "$UT_CONFIG_DIR"
  chmod 700 "$UT_CONFIG_DIR" 2>/dev/null || true
  ( umask 177; printf 'UPLOADTHING_API_KEY=%s\n' "$newkey" > "$UT_CONFIG_FILE" )
  chmod 600 "$UT_CONFIG_FILE" 2>/dev/null || true
  echo "saved key to $UT_CONFIG_FILE (mode 600)"
  echo "validated against appId: $(printf '%s' "$resp" | jq -r '.appId')"
}

case "${1:---status}" in
  --status) cmd_status ;;
  --check)  cmd_check ;;
  --save)   cmd_save ;;
  -h|--help) usage 0 ;;
  *) echo "unknown option: $1" >&2; usage 1 ;;
esac
