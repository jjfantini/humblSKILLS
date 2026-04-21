#!/usr/bin/env bash
# probe.sh — inspect an MP4 and emit compact JSON with duration/fps/dims/codec.
# Usage: bash scripts/probe.sh <video-path>
#
# Exits non-zero if: ffprobe is missing, file doesn't exist, or the file
# isn't decodable as a video.
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "Usage: $0 <video-path>" >&2
  exit 2
fi

VIDEO="$1"

if [[ ! -f "$VIDEO" ]]; then
  echo "Error: file not found: $VIDEO" >&2
  exit 3
fi

if ! command -v ffprobe >/dev/null 2>&1; then
  echo "Error: ffprobe not on PATH. Install with 'brew install ffmpeg' or 'apt install ffmpeg'." >&2
  exit 4
fi

# Pull the first video stream's metadata in one ffprobe invocation.
RAW=$(ffprobe -v error \
  -select_streams v:0 \
  -show_entries stream=width,height,r_frame_rate,codec_name,nb_frames,duration \
  -show_entries format=duration \
  -of default=noprint_wrappers=1:nokey=0 \
  "$VIDEO")

# Parse (bash-only — no jq dependency).
WIDTH=""
HEIGHT=""
FPS_FRAC=""
CODEC=""
STREAM_DURATION=""
STREAM_FRAMES=""
FORMAT_DURATION=""

while IFS='=' read -r key val; do
  case "$key" in
    width)      WIDTH="$val" ;;
    height)     HEIGHT="$val" ;;
    r_frame_rate) FPS_FRAC="$val" ;;
    codec_name) CODEC="$val" ;;
    duration)
      # ffprobe emits `duration` both at stream and format level with our flags.
      # Take the stream-level one first; format-level overrides if stream is N/A.
      if [[ -z "$STREAM_DURATION" || "$STREAM_DURATION" == "N/A" ]]; then
        STREAM_DURATION="$val"
      else
        FORMAT_DURATION="$val"
      fi
      ;;
    nb_frames) STREAM_FRAMES="$val" ;;
  esac
done <<< "$RAW"

# Pick the best duration signal.
DURATION="$STREAM_DURATION"
if [[ -z "$DURATION" || "$DURATION" == "N/A" ]]; then
  DURATION="$FORMAT_DURATION"
fi

# Parse fractional fps (e.g. "30/1" or "30000/1001") into a decimal.
if [[ "$FPS_FRAC" == */* ]]; then
  NUM="${FPS_FRAC%/*}"
  DEN="${FPS_FRAC#*/}"
  if [[ "$DEN" -eq 0 ]]; then
    FPS="0"
  else
    FPS=$(awk "BEGIN { printf \"%.3f\", $NUM / $DEN }")
  fi
else
  FPS="$FPS_FRAC"
fi

# Total frames: prefer nb_frames; fall back to duration * fps if N/A.
if [[ -z "$STREAM_FRAMES" || "$STREAM_FRAMES" == "N/A" ]]; then
  TOTAL=$(awk "BEGIN { printf \"%d\", $DURATION * $FPS }")
else
  TOTAL="$STREAM_FRAMES"
fi

cat <<EOF
{
  "path": "$VIDEO",
  "duration": $DURATION,
  "fps": $FPS,
  "width": $WIDTH,
  "height": $HEIGHT,
  "codec": "$CODEC",
  "total_frames": $TOTAL
}
EOF
