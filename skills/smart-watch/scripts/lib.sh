#!/usr/bin/env bash
# Shared helpers for smart-watch scripts. Sourced, never executed.
# Targets /bin/bash 3.2 (macOS): no associative arrays, no `wait -n`.

set -euo pipefail

SW_SKILL_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SW_CACHE="${XDG_CACHE_HOME:-$HOME/.cache}/smart-watch"
SW_STATE="${XDG_STATE_HOME:-$HOME/.local/state}/smart-watch"
SW_BIN="$SW_CACHE/bin/watchkit"
SW_SRC="$SW_SKILL_ROOT/scripts/swift/watchkit.swift"
SW_DB="$SW_STATE/index.db"
SW_SOURCES="$SW_STATE/sources"
SW_MODELS="$SW_CACHE/models"
SW_WHISPER_MODEL="$SW_MODELS/ggml-base.en.bin"
SW_WHISPER_MODEL_URL="https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.en.bin"

log()  { printf '%s\n' "$*" >&2; }
die()  { log "error: $*"; exit 1; }
have() { command -v "$1" >/dev/null 2>&1; }

macos_major() { sw_vers -productVersion 2>/dev/null | cut -d. -f1; }
macos_lt_26() { [ "$(macos_major || echo 0)" -lt 26 ]; }

# SQL single-quote escaping. The quote must live in a variable: ${v//\'/\'\'}
# leaves the backslashes in and emits \'\'.
sql_q() { local q="'" v="$1"; printf "'%s'" "${v//$q/$q$q}"; }

# Run SQL against the index. busy_timeout returns a row, so it is silenced.
db() {
  sqlite3 "$SW_DB" ".output /dev/null" "PRAGMA busy_timeout=5000;" ".output stdout" "$@"
}
# sqlite3 ignores stdin once it has SQL arguments, so scripts go through .read
db_file() { db ".read $1"; }

db_init() {
  mkdir -p "$SW_STATE" "$SW_SOURCES"
  db "
    PRAGMA journal_mode=WAL;
    CREATE TABLE IF NOT EXISTS sources(
      id TEXT PRIMARY KEY, path TEXT NOT NULL, duration REAL, fps REAL,
      has_audio INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL);
    CREATE TABLE IF NOT EXISTS segments(
      rowid INTEGER PRIMARY KEY, source_id TEXT NOT NULL REFERENCES sources(id),
      kind TEXT NOT NULL CHECK(kind IN ('ocr','speech')),
      t_start REAL NOT NULL, t_end REAL, frame TEXT, text TEXT NOT NULL,
      pinned INTEGER NOT NULL DEFAULT 0);
    CREATE INDEX IF NOT EXISTS segments_src_t ON segments(source_id, t_start);
    CREATE VIRTUAL TABLE IF NOT EXISTS segments_fts USING fts5(
      text, content='segments', content_rowid='rowid', tokenize='unicode61 remove_diacritics 2');
    CREATE TRIGGER IF NOT EXISTS segments_ai AFTER INSERT ON segments BEGIN
      INSERT INTO segments_fts(rowid, text) VALUES (new.rowid, new.text); END;
    CREATE TRIGGER IF NOT EXISTS segments_ad AFTER DELETE ON segments BEGIN
      INSERT INTO segments_fts(segments_fts, rowid, text) VALUES ('delete', old.rowid, old.text); END;
  " >/dev/null
  # migrations for indexes created before a column existed
  db "PRAGMA table_info(segments);" | grep -q '|pinned|' \
    || db "ALTER TABLE segments ADD COLUMN pinned INTEGER NOT NULL DEFAULT 0;" >/dev/null
}

# SQL expression rendering a seconds column as hh:mm:ss.cc (sqlite % truncates REALs)
sql_hms() { printf "printf('%%02d:%%02d:%%05.2f', cast(%s/3600 as int), cast((%s - 3600*cast(%s/3600 as int))/60 as int), %s - 60*cast(%s/60 as int))" "$1" "$1" "$1" "$1" "$1"; }

# Seconds from "83", "83.5", "1:23", "01:23:45.5". Prints nothing on garbage.
parse_t() {
  printf '%s' "$1" | awk -F: '
    { for (i=1;i<=NF;i++) if ($i !~ /^[0-9]*\.?[0-9]+$/) exit 1
      s=0; for (i=1;i<=NF;i++) s=s*60+$i; printf "%.3f", s }'
}

# Sampling interval (seconds) for a detail level and a duration.
# auto: <=3min 1s, <=15min 3s, <=45min 5s, else 10s.
detail_interval() { # detail_interval <fine|normal|coarse|auto> <duration_s>
  case "$1" in
    fine)   echo 1 ;;
    normal) echo 3 ;;
    coarse) echo 10 ;;
    auto)   awk -v d="${2:-0}" 'BEGIN{ if (d<=180) print 1; else if (d<=900) print 3; else if (d<=2700) print 5; else print 10 }' ;;
    *)      return 1 ;;
  esac
}

# hh:mm:ss.cc from seconds
fmt_t() { awk -v s="$1" 'BEGIN{h=int(s/3600);m=int((s-h*3600)/60);x=s-h*3600-m*60;printf "%02d:%02d:%05.2f",h,m,x}'; }

# Build the Swift helper when missing or when its source hash changed.
ensure_watchkit() {
  have swiftc || die "swiftc not found. Install Xcode Command Line Tools: xcode-select --install"
  mkdir -p "$(dirname "$SW_BIN")"
  local want have_hash=""
  want="$(shasum -a 256 "$SW_SRC" | cut -c1-16)"
  [ -f "$SW_BIN.srchash" ] && have_hash="$(cat "$SW_BIN.srchash")"
  if [ ! -x "$SW_BIN" ] || [ "$want" != "$have_hash" ]; then
    log "building watchkit (swiftc -O, ~6s)..."
    swiftc -O -parse-as-library -o "$SW_BIN" "$SW_SRC" || die "watchkit build failed"
    printf '%s' "$want" > "$SW_BIN.srchash"
  fi
}
