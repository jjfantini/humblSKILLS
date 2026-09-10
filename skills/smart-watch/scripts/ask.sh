#!/usr/bin/env bash
# ask.sh — search indexed evidence.
#
#   bash scripts/ask.sh <id|all> "<fts5 query>" [--from S] [--to S] [--kind ocr|speech] [--limit N]
#   bash scripts/ask.sh <id> --timeline            dump every segment in time order
#   bash scripts/ask.sh --list                     list indexed sources
#
# FTS5 syntax: words are ANDed; "a OR b", "exact phrase", prefix*, NOT. Tokens are
# whole words (no stemming): use deploy* to match deployment/deployed.
# Output: <id>  <hh:mm:ss.cc>  <kind>  <frame|->  <snippet>

source "$(dirname "$0")/lib.sh"
[ -f "$SW_DB" ] || die "no index yet - run watch.sh first"

ID=""; Q=""; FROM=""; TO=""; KIND=""; LIMIT=50; TIMELINE=false
while [ $# -gt 0 ]; do
  case "$1" in
    --list)     db "SELECT id, printf('%8.1fs',duration), has_audio, created_at, path FROM sources ORDER BY created_at;"; exit 0 ;;
    --timeline) TIMELINE=true; shift ;;
    --from)     FROM="$2"; shift 2 ;;
    --to)       TO="$2"; shift 2 ;;
    --kind)     KIND="$2"; shift 2 ;;
    --limit)    LIMIT="$2"; shift 2 ;;
    -h|--help)  sed -n '2,10p' "$0"; exit 0 ;;
    *) if [ -z "$ID" ]; then ID="$1"; else Q="$1"; fi; shift ;;
  esac
done
[ -n "$ID" ] || { sed -n '2,10p' "$0"; exit 2; }
$TIMELINE || [ -n "$Q" ] || die "missing query"

WHERE="1=1"
[ "$ID" != all ] && WHERE="$WHERE AND s.source_id=$(sql_q "$ID")"
[ -n "$FROM" ] && WHERE="$WHERE AND s.t_start >= $FROM"
[ -n "$TO" ]   && WHERE="$WHERE AND s.t_start <= $TO"
[ -n "$KIND" ] && WHERE="$WHERE AND s.kind=$(sql_q "$KIND")"

if $TIMELINE; then
  db -separator '  ' "
    SELECT s.source_id, $(sql_hms s.t_start),
           s.kind, ifnull(s.frame,'-'), s.text
    FROM segments s WHERE $WHERE ORDER BY s.t_start LIMIT $LIMIT;"
else
  search() {
    db -separator '  ' "
    SELECT s.source_id, $(sql_hms s.t_start),
           s.kind, ifnull(s.frame,'-'), snippet(segments_fts, 0, '[', ']', '…', 24)
    FROM segments_fts f JOIN segments s ON s.rowid = f.rowid
    WHERE segments_fts MATCH $(sql_q "$1") AND $WHERE
    ORDER BY bm25(segments_fts), s.t_start LIMIT $LIMIT;"
  }
  # Natural-language queries ("what happened?") are FTS5 syntax errors; retry with
  # every token quoted (ANDed), and if that finds nothing, OR them.
  if ! out="$(search "$Q" 2>&1)"; then
    quoted="$(printf '%s' "$Q" | tr -s '[:space:]' '\n' | sed -e 's/"//g' -e '/^$/d' -e 's/.*/"&"/')"
    andq="$(printf '%s' "$quoted" | tr '\n' ' ')"
    orq="$(printf '%s' "$quoted" | paste -sd '|' - | sed 's/|/ OR /g')"
    out="$(search "$andq" 2>/dev/null)" || out=""
    [ -n "$out" ] || out="$(search "$orq" 2>/dev/null)" || die "query failed: $Q"
    log "(query rewritten: $andq / $orq)"
  fi
  [ -n "$out" ] && printf '%s\n' "$out" || log "no hits for: $Q"
fi
