#!/usr/bin/env bash
# extract-frames.sh — FFmpeg frame extraction with computed target_fps.
# Usage: bash scripts/extract-frames.sh <video-path> <out-dir> [target-count=100]
#
# Computes target_fps = round(target_count / duration). Clamps target_fps
# to source_fps (never duplicates source frames). Emits frame_%04d.png
# plus a manifest.json describing the run.
set -euo pipefail

if [[ $# -lt 2 ]]; then
  echo "Usage: $0 <video-path> <out-dir> [target-count=100]" >&2
  exit 2
fi

VIDEO="$1"
OUTDIR="$2"
TARGET_COUNT="${3:-100}"

if [[ ! -f "$VIDEO" ]]; then
  echo "Error: file not found: $VIDEO" >&2
  exit 3
fi

if ! command -v ffmpeg >/dev/null 2>&1; then
  echo "Error: ffmpeg not on PATH. Install with 'brew install ffmpeg' or 'apt install ffmpeg'." >&2
  exit 4
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROBE_JSON=$(bash "$SCRIPT_DIR/probe.sh" "$VIDEO")

# Parse duration + fps from the probe JSON without jq.
DURATION=$(echo "$PROBE_JSON" | awk -F'[:,]' '/"duration":/ { gsub(/ /, "", $2); print $2 }')
SOURCE_FPS=$(echo "$PROBE_JSON" | awk -F'[:,]' '/"fps":/ { gsub(/ /, "", $2); print $2 }')

if [[ -z "$DURATION" || -z "$SOURCE_FPS" ]]; then
  echo "Error: could not parse duration/fps from probe output." >&2
  echo "$PROBE_JSON" >&2
  exit 5
fi

# Compute target_fps = round(target_count / duration).
TARGET_FPS=$(awk "BEGIN {
  raw = $TARGET_COUNT / $DURATION;
  src = $SOURCE_FPS;
  # Clamp to source fps (never duplicate frames).
  if (raw > src) raw = src;
  printf \"%d\", int(raw + 0.5);
}")

if [[ "$TARGET_FPS" -le 0 ]]; then
  echo "Error: computed target_fps is 0. Check duration ($DURATION) and target_count ($TARGET_COUNT)." >&2
  exit 6
fi

# Safety check: only clear out-dir if it's under /tmp/ or the user confirmed.
# For v0.1.0: refuse non-/tmp non-empty dirs to avoid data loss.
if [[ -e "$OUTDIR" ]]; then
  if [[ "$OUTDIR" != /tmp/* && "$OUTDIR" != /var/folders/* ]]; then
    if [[ -n "$(ls -A "$OUTDIR" 2>/dev/null)" ]]; then
      echo "Error: out-dir $OUTDIR is non-empty and not under /tmp/. Remove it manually or pick a /tmp/* path." >&2
      exit 7
    fi
  else
    rm -rf "$OUTDIR"
  fi
fi
mkdir -p "$OUTDIR"

echo "extract-frames.sh: fps=$TARGET_FPS target=$TARGET_COUNT duration=$DURATION → $OUTDIR" >&2

ffmpeg -v error -y -i "$VIDEO" \
  -vf "fps=${TARGET_FPS},scale=1920:-2" \
  "$OUTDIR/frame_%04d.png"

# Count the actual frames emitted.
ACTUAL=$(find "$OUTDIR" -name 'frame_*.png' -type f | wc -l | tr -d ' ')

cat > "$OUTDIR/manifest.json" <<EOF
{
  "video": "$VIDEO",
  "duration": $DURATION,
  "source_fps": $SOURCE_FPS,
  "target_fps": $TARGET_FPS,
  "target_count": $TARGET_COUNT,
  "actual_count": $ACTUAL,
  "format": "png",
  "resolution": "1920xauto",
  "outdir": "$OUTDIR"
}
EOF

echo "extract-frames.sh: emitted $ACTUAL PNG frames to $OUTDIR (manifest.json)" >&2
cat "$OUTDIR/manifest.json"
