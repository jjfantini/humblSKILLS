#!/usr/bin/env bash
# lib.sh - shared helpers for use-upload-thing. SOURCE this file; do not run it.
#
# Provides:
#   ut_preflight            - assert curl + jq are present
#   ut_resolve_key          - echo the API key to stdout (for $() capture); never logs it
#   ut_key_source           - echo where the key resolves from: env | token | file:<path> | none
#   ut_mask <secret>        - echo a masked form (first 7 + last 4) for safe display
#   ut_api <path> [json]    - POST to https://api.uploadthing.com/<path>; prints body; rc>=1 on HTTP >= 400
#   ut_app_id               - echo the app id resolved from /v7/getAppInfo
#
# Key resolution order (first match wins):
#   1. $UPLOADTHING_API_KEY                 (env; never persisted by this skill)
#   2. $UPLOADTHING_TOKEN                    (v7 base64 token; its .apiKey field is decoded)
#   3. $UT_CONFIG_FILE                       (0600 file outside the repo, written by auth.sh --save)
#
# The key is NEVER printed to logs, embedded in URLs, or passed as a process arg.

UT_API_BASE="${UPLOADTHING_API_URL:-https://api.uploadthing.com}"
UT_CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/humblskills"
UT_CONFIG_FILE="${UT_CONFIG_FILE:-$UT_CONFIG_DIR/uploadthing.env}"

ut_die() { printf 'error: %s\n' "$*" >&2; exit 1; }

ut_preflight() {
  command -v curl >/dev/null 2>&1 || ut_die "curl is required but not on PATH"
  command -v jq   >/dev/null 2>&1 || ut_die "jq is required but not on PATH"
}

ut_resolve_key() {
  if [ -n "${UPLOADTHING_API_KEY:-}" ]; then
    printf '%s' "$UPLOADTHING_API_KEY"
    return 0
  fi
  if [ -n "${UPLOADTHING_TOKEN:-}" ]; then
    local decoded key
    decoded=$(printf '%s' "$UPLOADTHING_TOKEN" | base64 -d 2>/dev/null) || decoded=""
    key=$(printf '%s' "$decoded" | jq -r '.apiKey // empty' 2>/dev/null)
    if [ -n "$key" ]; then
      printf '%s' "$key"
      return 0
    fi
  fi
  if [ -f "$UT_CONFIG_FILE" ]; then
    local k
    k=$( . "$UT_CONFIG_FILE" >/dev/null 2>&1; printf '%s' "${UPLOADTHING_API_KEY:-}" )
    if [ -n "$k" ]; then
      printf '%s' "$k"
      return 0
    fi
  fi
  return 1
}

ut_key_source() {
  if [ -n "${UPLOADTHING_API_KEY:-}" ]; then echo "env"; return; fi
  if [ -n "${UPLOADTHING_TOKEN:-}" ]; then echo "token"; return; fi
  if [ -f "$UT_CONFIG_FILE" ]; then echo "file:$UT_CONFIG_FILE"; return; fi
  echo "none"
}

ut_mask() {
  local k="$1" n=${#1}
  if [ "$n" -le 11 ]; then printf '****'; else printf '%s...%s' "${k:0:7}" "${k: -4}"; fi
}

# ut_api <path> [json-body]
# Prints the raw response body to stdout. Returns 0 on HTTP < 400, 1 otherwise.
ut_api() {
  local path="$1"
  # NB: do not write ${2:-{}} - bash ends the expansion at the first '}', which
  # appends a stray '}' to populated bodies and corrupts the JSON.
  local body="${2:-}"
  [ -z "$body" ] && body='{}'
  local key
  key=$(ut_resolve_key) || ut_die "no API key found - run 'bash scripts/auth.sh --status' for guidance"
  local tmp status
  tmp=$(mktemp)
  status=$(curl -sS -o "$tmp" -w '%{http_code}' \
    -X POST "$UT_API_BASE/${path#/}" \
    -H "x-uploadthing-api-key: $key" \
    -H "content-type: application/json" \
    --data "$body" 2>/dev/null) || { rm -f "$tmp"; ut_die "network request to $path failed"; }
  cat "$tmp"
  rm -f "$tmp"
  [ "${status:-0}" -lt 400 ] 2>/dev/null
}

ut_app_id() {
  local resp
  resp=$(ut_api "v7/getAppInfo" '{}') || return 1
  printf '%s' "$(printf '%s' "$resp" | jq -r '.appId // empty')"
}
