#!/usr/bin/env bash
# export-pdf.sh — render a single-file HTML page to a US Letter PDF with
# headless Google Chrome. Paper size/orientation come from the file's @page
# rule; passing an orientation here injects an override into a temp copy so
# the source file is never modified.
#
# Usage: export-pdf.sh <file.html> [portrait|landscape] [out.pdf]
# Env:   CHROME_BIN — explicit browser binary (checked first)
set -euo pipefail

usage() { echo "Usage: export-pdf.sh <file.html> [portrait|landscape] [out.pdf]" >&2; exit 2; }

HTML="${1:-}"; [ -n "$HTML" ] && [ -f "$HTML" ] || usage
ORIENT="${2:-}"
OUT="${3:-${HTML%.html}.pdf}"

case "$ORIENT" in
  ""|portrait|landscape) ;;
  *) echo "error: orientation must be 'portrait' or 'landscape', got '$ORIENT'" >&2; usage ;;
esac

find_chrome() {
  if [ -n "${CHROME_BIN:-}" ]; then echo "$CHROME_BIN"; return; fi
  local mac_candidates=(
    "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
    "/Applications/Chromium.app/Contents/MacOS/Chromium"
  )
  local c
  for c in "${mac_candidates[@]}"; do
    [ -x "$c" ] && { echo "$c"; return; }
  done
  for c in google-chrome google-chrome-stable chromium chromium-browser chrome; do
    command -v "$c" >/dev/null 2>&1 && { command -v "$c"; return; }
  done
  echo "error: no Chrome/Chromium found — install Google Chrome or set CHROME_BIN" >&2
  exit 1
}

CHROME="$(find_chrome)"

# Orientation override: append a trailing @page rule (last rule wins) to a
# temp copy in the same directory, so any relative file:// references and
# the source file both stay intact.
SRC="$HTML"
CLEANUP=""
if [ -n "$ORIENT" ]; then
  DIR="$(cd "$(dirname "$HTML")" && pwd)"
  SRC="$(mktemp "$DIR/.export-XXXXXX.html")"
  CLEANUP="$SRC"
  trap '[ -n "$CLEANUP" ] && rm -f "$CLEANUP"' EXIT
  SIZE="letter"; [ "$ORIENT" = "landscape" ] && SIZE="letter landscape"
  cp "$HTML" "$SRC"
  printf '\n<style>@page { size: %s; }</style>\n' "$SIZE" >> "$SRC"
fi

ABS="$(cd "$(dirname "$SRC")" && pwd)/$(basename "$SRC")"

"$CHROME" \
  --headless \
  --disable-gpu \
  --no-pdf-header-footer \
  --virtual-time-budget=8000 \
  --print-to-pdf="$OUT" \
  "file://$ABS" 2>/dev/null

[ -s "$OUT" ] || { echo "error: Chrome produced no output at $OUT" >&2; exit 1; }
echo "wrote $OUT${ORIENT:+ ($ORIENT)}"
