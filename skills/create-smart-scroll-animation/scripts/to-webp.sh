#!/usr/bin/env bash
# to-webp.sh — batch convert PNG frames to WebP (cwebp preferred, ffmpeg fallback).
# Usage: bash scripts/to-webp.sh <in-dir> <out-dir> [quality=80]
#
# Writes frame_NNNN.webp to out-dir. Prints a size report on completion.
# If cwebp is on PATH, uses it (best quality-per-byte). Otherwise falls
# through to ffmpeg -c:v libwebp.
set -euo pipefail

if [[ $# -lt 2 ]]; then
  echo "Usage: $0 <in-dir> <out-dir> [quality=80]" >&2
  exit 2
fi

INDIR="$1"
OUTDIR="$2"
QUALITY="${3:-80}"

if [[ ! -d "$INDIR" ]]; then
  echo "Error: in-dir not found: $INDIR" >&2
  exit 3
fi

# Safety check: refuse non-empty non-/tmp out-dirs we didn't create.
if [[ -e "$OUTDIR" && "$OUTDIR" != /tmp/* && "$OUTDIR" != /var/folders/* ]]; then
  if [[ -n "$(ls -A "$OUTDIR" 2>/dev/null)" ]]; then
    # Allow if the directory only contains .webp files we're about to overwrite.
    NON_WEBP=$(find "$OUTDIR" -maxdepth 1 -type f ! -name '*.webp' ! -name '.*' | head -1 || true)
    if [[ -n "$NON_WEBP" ]]; then
      echo "Error: out-dir $OUTDIR contains non-webp files. Pick a fresh path or clean it manually." >&2
      exit 4
    fi
  fi
fi
mkdir -p "$OUTDIR"

USE_CWEBP=false
if command -v cwebp >/dev/null 2>&1; then
  USE_CWEBP=true
elif ! command -v ffmpeg >/dev/null 2>&1; then
  echo "Error: neither cwebp nor ffmpeg on PATH. Install one:" >&2
  echo "  brew install webp    # cwebp (preferred)" >&2
  echo "  brew install ffmpeg  # fallback encoder" >&2
  exit 5
fi

COUNT=0
for PNG in "$INDIR"/frame_*.png; do
  [[ -e "$PNG" ]] || continue
  BASE=$(basename "$PNG" .png)
  OUT="$OUTDIR/${BASE}.webp"
  if [[ "$USE_CWEBP" == true ]]; then
    cwebp -q "$QUALITY" -quiet "$PNG" -o "$OUT"
  else
    ffmpeg -v error -y -i "$PNG" -c:v libwebp -quality "$QUALITY" "$OUT"
  fi
  COUNT=$((COUNT + 1))
done

if [[ "$COUNT" -eq 0 ]]; then
  echo "Error: no frame_*.png files found in $INDIR" >&2
  exit 6
fi

# Size report.
TOTAL_BYTES=$(find "$OUTDIR" -name 'frame_*.webp' -type f -exec stat -f%z {} \; 2>/dev/null | awk '{s+=$1} END {print s+0}')
if [[ -z "$TOTAL_BYTES" || "$TOTAL_BYTES" == "0" ]]; then
  # Linux stat fallback.
  TOTAL_BYTES=$(find "$OUTDIR" -name 'frame_*.webp' -type f -exec stat -c%s {} \; 2>/dev/null | awk '{s+=$1} END {print s+0}')
fi

TOTAL_MB=$(awk "BEGIN { printf \"%.2f\", $TOTAL_BYTES / 1048576 }")
AVG_KB=$(awk "BEGIN { printf \"%.1f\", ($TOTAL_BYTES / $COUNT) / 1024 }")

ENCODER="cwebp"
if [[ "$USE_CWEBP" != true ]]; then ENCODER="ffmpeg libwebp (fallback)"; fi

echo "to-webp.sh: $COUNT frames @ q=$QUALITY via $ENCODER → $TOTAL_MB MB total (avg ${AVG_KB} KB/frame)"
