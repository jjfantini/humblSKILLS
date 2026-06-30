#!/usr/bin/env bash
# api.sh - generic UploadThing REST caller. The escape hatch for ANY endpoint.
#
# POSTs to https://api.uploadthing.com/<path> with the auth header and the given
# JSON body, then pretty-prints the JSON response.
#
# Usage:
#   bash scripts/api.sh <path> [json-body]
# Examples:
#   bash scripts/api.sh v7/getAppInfo
#   bash scripts/api.sh v6/listFiles '{"limit":10}'
#   bash scripts/api.sh v6/getUsageInfo
#   bash scripts/api.sh v6/deleteFiles '{"fileKeys":["KEY_1"]}'
#   bash scripts/api.sh v6/renameFiles '{"updates":[{"fileKey":"KEY","newName":"x.png"}]}'
#   bash scripts/api.sh v6/updateACL '{"updates":[{"fileKey":"KEY","acl":"public-read"}]}'
#
# See the full endpoint catalog: https://docs.uploadthing.com/api-reference/openapi-spec
# Exit codes: mirrors the HTTP result (0 if < 400, 1 otherwise) | 2 missing dependency
set -uo pipefail
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# shellcheck source=lib.sh
. "$SCRIPT_DIR/lib.sh"
ut_preflight

[ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ] && { sed -n '2,20p' "$SCRIPT_DIR/api.sh" | sed 's/^# \{0,1\}//'; exit 0; }

PATH_ARG="${1:-}"
[ -n "$PATH_ARG" ] || ut_die "usage: api.sh <path> [json-body]"
BODY="${2:-}"
[ -z "$BODY" ] && BODY='{}'

if ! printf '%s' "$BODY" | jq -e . >/dev/null 2>&1; then
  ut_die "body is not valid JSON: $BODY"
fi

RESP=$(ut_api "$PATH_ARG" "$BODY"); RC=$?
printf '%s' "$RESP" | jq . 2>/dev/null || printf '%s\n' "$RESP"
exit $RC
