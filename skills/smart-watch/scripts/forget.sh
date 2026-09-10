#!/usr/bin/env bash
# forget.sh — delete an indexed source: its frames, audio, OCR text, transcript
# and index rows. Recordings capture whatever is on screen, secrets included,
# and everything sits in plain text under ~/.local/state/smart-watch.
#
#   bash scripts/forget.sh <id> [<id>...]

source "$(dirname "$0")/lib.sh"
[ $# -gt 0 ] || { sed -n '2,6p' "$0"; exit 2; }
for id in "$@"; do
  [ -d "$SW_SOURCES/$id" ] || log "warn: no files for $id"
  rm -rf "$SW_SOURCES/$id"
  [ -f "$SW_DB" ] && db "DELETE FROM segments WHERE source_id=$(sql_q "$id"); DELETE FROM sources WHERE id=$(sql_q "$id");"
  log "forgot $id"
done
