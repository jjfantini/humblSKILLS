#!/usr/bin/env bash
# upload.sh - upload a local file to UploadThing using the v7 flow.
#
# Flow: POST /v7/prepareUpload (get presigned ingest URL) -> PUT the file to it.
# No file router / slug is used (server-side upload), so no callback registration
# is needed.
#
# Usage:
#   bash scripts/upload.sh <path> [options]
# Options:
#   --name NAME                 override the stored file name (default: basename)
#   --acl public-read|private   ACL (only honored if the app allows overrides)
#   --content-disposition inline|attachment   (default: inline)
#   --custom-id ID              set a custom identifier
#   --expires-in SECONDS        auto-delete after N seconds
#
# Prints JSON: { "key": ..., "ingestUrl": ..., "ufsUrl": ... }
# Exit codes: 0 ok | 1 failure | 2 missing dependency
set -uo pipefail
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# shellcheck source=lib.sh
. "$SCRIPT_DIR/lib.sh"
ut_preflight

[ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ] && { sed -n '2,20p' "$SCRIPT_DIR/upload.sh" | sed 's/^# \{0,1\}//'; exit 0; }

FILE="${1:-}"
[ -n "$FILE" ] || ut_die "usage: upload.sh <path> [options]"
shift || true
[ -f "$FILE" ] || ut_die "file not found: $FILE"

NAME=""; ACL=""; CD="inline"; CUSTOM_ID=""; EXPIRES_IN=""
while [ $# -gt 0 ]; do
  case "$1" in
    --name) NAME="$2"; shift 2 ;;
    --acl) ACL="$2"; shift 2 ;;
    --content-disposition) CD="$2"; shift 2 ;;
    --custom-id) CUSTOM_ID="$2"; shift 2 ;;
    --expires-in) EXPIRES_IN="$2"; shift 2 ;;
    *) ut_die "unknown option: $1" ;;
  esac
done

NAME="${NAME:-$(basename "$FILE")}"
SIZE=$(wc -c < "$FILE" | tr -d '[:space:]')
TYPE=$(file --mime-type -b "$FILE" 2>/dev/null || echo "application/octet-stream")

REQ=$(jq -n \
  --arg fn "$NAME" --argjson fs "$SIZE" --arg ft "$TYPE" --arg cd "$CD" \
  '{fileName:$fn, fileSize:$fs, fileType:$ft, contentDisposition:$cd}')
[ -n "$ACL" ]        && REQ=$(printf '%s' "$REQ" | jq --arg a "$ACL" '. + {acl:$a}')
[ -n "$CUSTOM_ID" ]  && REQ=$(printf '%s' "$REQ" | jq --arg c "$CUSTOM_ID" '. + {customId:$c}')
[ -n "$EXPIRES_IN" ] && REQ=$(printf '%s' "$REQ" | jq --argjson e "$EXPIRES_IN" '. + {expiresIn:$e}')

RESP=$(ut_api "v7/prepareUpload" "$REQ") \
  || ut_die "prepareUpload failed: $(printf '%s' "$RESP" | jq -r '.error // .' 2>/dev/null || printf '%s' "$RESP")"

KEY=$(printf '%s' "$RESP" | jq -r '.key // empty')
URL=$(printf '%s' "$RESP" | jq -r '.url // empty')
[ -n "$KEY" ] && [ -n "$URL" ] || ut_die "prepareUpload returned no key/url: $RESP"

PUT_STATUS=$(curl -sS -o /dev/null -w '%{http_code}' \
  -X PUT "$URL" -F "file=@$FILE;type=$TYPE;filename=$NAME" 2>/dev/null) \
  || ut_die "ingest PUT request failed (network)"
[ "${PUT_STATUS:-0}" -lt 400 ] 2>/dev/null \
  || ut_die "ingest upload failed with HTTP $PUT_STATUS for key $KEY"

APP_ID=$(ut_app_id || true)
UFS=""
[ -n "$APP_ID" ] && UFS="https://$APP_ID.ufs.sh/f/$KEY"

jq -n --arg key "$KEY" --arg ingestUrl "$URL" --arg ufsUrl "$UFS" \
  '{key:$key, ingestUrl:$ingestUrl, ufsUrl:$ufsUrl}'
