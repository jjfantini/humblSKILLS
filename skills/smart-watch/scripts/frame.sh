#!/usr/bin/env bash
# frame.sh — path of the indexed frame nearest a timestamp, for the agent to Read.
#
#   bash scripts/frame.sh <id> <seconds|hh:mm:ss>
#
# Falls back to extracting the exact frame with ffmpeg when the nearest sampled
# frame is more than one sample interval away.

source "$(dirname "$0")/lib.sh"
ID="${1:-}"; T="${2:-}"
[ -n "$ID" ] && [ -n "$T" ] || { sed -n '2,7p' "$0"; exit 2; }
DIR="$SW_SOURCES/$ID"; [ -d "$DIR" ] || die "unknown source id: $ID"

T="$(parse_t "$T")" || die "bad time (use seconds or hh:mm:ss)"

row="$(db -separator '|' "SELECT frame, t_start, abs(t_start-$T) d FROM segments
  WHERE source_id=$(sql_q "$ID") AND kind='ocr' ORDER BY d LIMIT 1;")"
frame="${row%%|*}"; rest="${row#*|}"; t="${rest%%|*}"; dist="${rest#*|}"
fps="$(db "SELECT fps FROM sources WHERE id=$(sql_q "$ID");")"
tol="$(awk -v f="${fps:-1}" 'BEGIN{print 1/f}')"

if [ -n "$frame" ] && awk -v d="$dist" -v tol="$tol" 'BEGIN{exit !(d<=tol)}'; then
  log "nearest indexed frame at $(fmt_t "$t")"
  echo "$DIR/$frame"
else
  mkdir -p "$DIR/extra"
  out="$DIR/extra/$(printf '%09.3f' "$T").jpg"
  [ -f "$out" ] || ffmpeg -v error -y -ss "$T" -i "$DIR"/source.* -frames:v 1 -q:v 3 "$out" || die "ffmpeg seek failed"
  log "extracted exact frame at $(fmt_t "$T")"
  echo "$out"
fi
