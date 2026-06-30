#!/usr/bin/env bash
# download.sh - download a file from UploadThing by key or URL.
#
# If given a URL, downloads it directly. If given a file key, calls
# POST /v6/requestFileAccess to obtain a presigned ufsUrl (works for both public
# and private files) and downloads that.
#
# Usage:
#   bash scripts/download.sh <fileKeyOrUrl> [-o OUTPUT] [--expires-in SECONDS]
#
# Exit codes: 0 ok | 1 failure | 2 missing dependency
set -uo pipefail
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# shellcheck source=lib.sh
. "$SCRIPT_DIR/lib.sh"
ut_preflight

[ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ] && { sed -n '2,11p' "$SCRIPT_DIR/download.sh" | sed 's/^# \{0,1\}//'; exit 0; }

ARG="${1:-}"
[ -n "$ARG" ] || ut_die "usage: download.sh <fileKeyOrUrl> [-o OUTPUT] [--expires-in SECONDS]"
shift || true

OUT=""; EXPIRES_IN=""
while [ $# -gt 0 ]; do
  case "$1" in
    -o|--output) OUT="$2"; shift 2 ;;
    --expires-in) EXPIRES_IN="$2"; shift 2 ;;
    *) ut_die "unknown option: $1" ;;
  esac
done

case "$ARG" in
  http://*|https://*)
    URL="$ARG"
    DEFAULT_NAME=$(basename "${ARG%%\?*}")
    ;;
  *)
    REQ=$(jq -n --arg k "$ARG" '{fileKey:$k}')
    [ -n "$EXPIRES_IN" ] && REQ=$(printf '%s' "$REQ" | jq --argjson e "$EXPIRES_IN" '. + {expiresIn:$e}')
    RESP=$(ut_api "v6/requestFileAccess" "$REQ") \
      || ut_die "requestFileAccess failed: $(printf '%s' "$RESP" | jq -r '.error // .' 2>/dev/null || printf '%s' "$RESP")"
    URL=$(printf '%s' "$RESP" | jq -r '.ufsUrl // .url // empty')
    [ -n "$URL" ] || ut_die "no URL returned for key $ARG: $RESP"
    DEFAULT_NAME="$ARG"
    ;;
esac

OUT="${OUT:-$DEFAULT_NAME}"
STATUS=$(curl -sSL -o "$OUT" -w '%{http_code}' "$URL" 2>/dev/null) \
  || ut_die "download request failed (network)"
[ "${STATUS:-0}" -lt 400 ] 2>/dev/null || ut_die "download failed with HTTP $STATUS"
echo "downloaded -> $OUT ($(wc -c < "$OUT" | tr -d '[:space:]') bytes)"
