#!/usr/bin/env bash
# watch.sh — index a recording into timestamped frames + OCR + transcript, then brief it.
#
#   bash scripts/watch.sh <file|url> [options]
#   bash scripts/watch.sh --record <seconds> [options]
#
# Options:
#   --detail <auto|fine|normal|coarse>  frame sampling: fine=1s, normal=3s, coarse=10s,
#                                       auto (default) picks by length: <=3m 1s, <=15m 3s, <=45m 5s, else 10s
#   --fps <n>       explicit frames per second (overrides --detail; 0.5 = one every 2s)
#   --at <t,t,...>  pin timestamps the user cares about (83, 1:23, 01:23:45); exact frames
#                   are extracted and OCR'd at each, plus one frame 2s before and after
#   --no-audio      skip audio extraction and transcription
#   --audio         with --record: also capture the default input device
#   --display <n>   with --record: display number (default 1)
#   --locale <id>   transcription locale (default en_US)
#   --no-brief      skip generating brief.md after indexing
#   --force         re-index even if this source id already exists
#
# Prints brief.md (unless --no-brief) then the source id on the last stdout line.
# Progress goes to stderr.

source "$(dirname "$0")/lib.sh"

SRC=""; RECORD=0; DETAIL=auto; FPS=""; PINS=""; AUDIO=true; REC_AUDIO=false; DISPLAY_N=1
LOCALE="en_US"; BRIEF=true; FORCE=false
while [ $# -gt 0 ]; do
  case "$1" in
    --record)   RECORD="$2"; shift 2 ;;
    --detail)   DETAIL="$2"; shift 2 ;;
    --fps)      FPS="$2"; shift 2 ;;
    --at)       PINS="$2"; shift 2 ;;
    --no-audio) AUDIO=false; shift ;;
    --audio)    REC_AUDIO=true; shift ;;
    --display)  DISPLAY_N="$2"; shift 2 ;;
    --locale)   LOCALE="$2"; shift 2 ;;
    --no-brief) BRIEF=false; shift ;;
    --force)    FORCE=true; shift ;;
    -h|--help)  sed -n '2,20p' "$0"; exit 0 ;;
    -*)         die "unknown option: $1" ;;
    *)          SRC="$1"; shift ;;
  esac
done
[ -n "$SRC" ] || [ "$RECORD" != 0 ] || { sed -n '2,20p' "$0"; exit 2; }
detail_interval "$DETAIL" 0 >/dev/null || die "--detail must be auto|fine|normal|coarse"

have ffmpeg && have ffprobe || die "ffmpeg not found - run scripts/doctor.sh"
ensure_watchkit
db_init

WORK="$(mktemp -d "${TMPDIR:-/tmp}/smart-watch.XXXXXX")"
trap 'rm -rf "$WORK"' EXIT

# ---- 1. acquire -------------------------------------------------------------
if [ "$RECORD" != 0 ]; then
  # Must stay /usr/sbin/screencapture: it is Apple-signed and inherits the
  # terminal's Screen Recording grant. A homegrown ScreenCaptureKit binary is
  # ad-hoc signed and loses TCC on every rebuild. -g records the default INPUT
  # only; system audio needs BlackHole + a Multi-Output Device.
  SRC="$WORK/recording-$(date +%Y%m%d-%H%M%S).mov"
  ga=(); $REC_AUDIO && ga=(-g)
  log "recording display $DISPLAY_N for ${RECORD}s -> $SRC"
  /usr/sbin/screencapture -v -V "$RECORD" -D "$DISPLAY_N" -k "${ga[@]+"${ga[@]}"}" "$SRC" \
    || die "screencapture failed (Screen Recording permission for this terminal?)"
  [ -s "$SRC" ] || die "screencapture wrote no file"
elif printf '%s' "$SRC" | grep -qE '^https?://'; then
  have yt-dlp || die "URL sources need yt-dlp: brew install yt-dlp (ask the user before installing)"
  log "downloading $SRC"
  yt-dlp -q --no-playlist -o "$WORK/download.%(ext)s" "$SRC" || die "yt-dlp failed"
  SRC="$(ls "$WORK"/download.* | head -1)"
else
  [ -f "$SRC" ] || die "no such file: $SRC"
  SRC="$(cd "$(dirname "$SRC")" && pwd)/$(basename "$SRC")"
fi

ID="$(shasum -a 256 "$SRC" | cut -c1-12)"
DIR="$SW_SOURCES/$ID"
PINS_ONLY=false
if [ -f "$DIR/.done" ] && ! $FORCE; then
  if [ -n "$PINS" ]; then
    PINS_ONLY=true; log "already indexed as $ID - adding pins only"
  else
    log "already indexed as $ID ($DIR). Use --force to redo."
    echo "$ID"; exit 0
  fi
fi
if ! $PINS_ONLY; then
  rm -rf "$DIR"; mkdir -p "$DIR/frames"   # a half-built dir (no .done) is redone, not trusted
  EXT="${SRC##*.}"
  case "$SRC" in
    "$WORK"/*) mv "$SRC" "$DIR/source.$EXT" ;;   # recordings and downloads live here
    *)         ln -s "$SRC" "$DIR/source.$EXT" ;; # user files stay where they are
  esac
fi
SRC="$(ls "$DIR"/source.* | head -1)"

DURATION="$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$SRC" | cut -d. -f1-2)"
HAS_AUDIO=0; ffprobe -v error -select_streams a:0 -show_entries stream=codec_type -of csv=p=0 "$SRC" | grep -q audio && HAS_AUDIO=1
# an mp3 cover image is a "video" stream with disposition attached_pic - not footage
HAS_VIDEO=0; ffprobe -v error -select_streams v:0 -show_entries stream=codec_type:stream_disposition=attached_pic -of csv=p=0 "$SRC" \
  | grep -q '^video,0' && HAS_VIDEO=1
[ -n "$DURATION" ] || die "ffprobe could not read $SRC"

if [ -n "$FPS" ]; then INTERVAL="$(awk -v f="$FPS" 'BEGIN{printf "%.4f", 1/f}')"
else INTERVAL="$(detail_interval "$DETAIL" "$DURATION")"; FPS="$(awk -v i="$INTERVAL" 'BEGIN{printf "%.4f", 1/i}')"; fi
EST="$(awk -v d="$DURATION" -v i="$INTERVAL" 'BEGIN{n=int(d/i)+1; printf "%d frames, ~%ds OCR", n, n*0.4}')"
$PINS_ONLY || log "source $ID: ${DURATION}s video=$HAS_VIDEO audio=$HAS_AUDIO  detail=$DETAIL every ${INTERVAL}s ($EST)"

SQL="$WORK/insert.sql"
echo "BEGIN;" > "$SQL"
if ! $PINS_ONLY; then
  echo "DELETE FROM segments WHERE source_id=$(sql_q "$ID"); DELETE FROM sources WHERE id=$(sql_q "$ID");" >> "$SQL"
  echo "INSERT INTO sources(id,path,duration,fps,has_audio,created_at) VALUES ($(sql_q "$ID"),$(sql_q "$SRC"),$DURATION,$FPS,$HAS_AUDIO,datetime('now'));" >> "$SQL"
fi

# OCR a list of frames (one path per line, time-ordered) paired with a pts file;
# appends INSERTs, deduping consecutive identical text. $1=frames.txt $2=pts.txt $3=pinned(0|1)
ocr_and_index() {
  local frames="$1" pts="$2" pinned="$3" prev_hash="" fpath text p h
  tr '\n' '\0' < "$frames" | xargs -0 "$SW_BIN" ocr > "$WORK/ocr.tsv"
  while IFS=$'\t' read -r fpath text && IFS= read -r p <&3; do
    [ -z "$text" ] && continue
    h="$(printf '%s' "$text" | shasum | cut -c1-16)"
    [ "$h" = "$prev_hash" ] && [ "$pinned" = 0 ] && continue
    prev_hash="$h"; OCR_CHARS=$((OCR_CHARS+${#text})); [ "$pinned" = 0 ] && N_KEPT=$((N_KEPT+1))
    echo "INSERT INTO segments(source_id,kind,t_start,t_end,frame,text,pinned) VALUES ($(sql_q "$ID"),'ocr',$p,NULL,$(sql_q "frames/$(basename "$fpath")"),$(sql_q "$text"),$pinned);" >> "$SQL"
  done < "$WORK/ocr.tsv" 3< "$pts"
  cat "$WORK/ocr.tsv" >> "$DIR/ocr.tsv"
}

# ---- 2. frames (PTS-preserving) + 3. OCR ------------------------------------
N_FRAMES=0; N_KEPT=0; OCR_CHARS=0; N_PINS=0
if [ "$HAS_VIDEO" = 1 ] && ! $PINS_ONLY; then
  # `fps=N` retimes VFR input onto a grid (measured 0.47s off). `select` on the
  # true timeline + passthrough keeps real PTS; showinfo (info level!) reports
  # them so the Nth frame pairs with the Nth pts_time line. Width is capped, never
  # reduced further: 1280px turned readable UI text into garbage.
  ffmpeg -hide_banner -loglevel info -nostats -i "$SRC" \
    -vf "select='isnan(prev_selected_t)+gte(t-prev_selected_t\,$INTERVAL)',scale='min(iw\,2560)':-2,showinfo" \
    -fps_mode passthrough -q:v 3 "$DIR/frames/%06d.jpg" 2> "$WORK/showinfo.log" \
    || die "ffmpeg frame extraction failed (see $WORK/showinfo.log)"
  grep -o 'pts_time:[0-9.]*' "$WORK/showinfo.log" | cut -d: -f2 > "$WORK/pts.txt"
  N_FRAMES="$(wc -l < "$WORK/pts.txt" | tr -d ' ')"
  [ "$N_FRAMES" -gt 0 ] || die "no frames extracted (no pts_time lines - was -loglevel lowered?)"
  log "extracted $N_FRAMES frames, running OCR..."
  find "$DIR/frames" -name '[0-9]*.jpg' | sort > "$WORK/frames.txt"
  ocr_and_index "$WORK/frames.txt" "$WORK/pts.txt" 0
fi
# Pinned timestamps: exact frame plus -2s/+2s context, never deduped.
if [ "$HAS_VIDEO" = 1 ] && [ -n "$PINS" ]; then
  : > "$WORK/pin_frames.txt"; : > "$WORK/pin_pts.txt"
  for raw in $(printf '%s' "$PINS" | tr ',' ' '); do
    t="$(parse_t "$raw")" || { log "warn: bad pin '$raw', skipped"; continue; }
    awk -v t="$t" -v d="$DURATION" 'BEGIN{exit !(t<=d)}' || { log "warn: pin $raw is past the end (${DURATION}s), skipped"; continue; }
    for off in -2 0 2; do
      tt="$(awk -v t="$t" -v o="$off" 'BEGIN{x=t+o; if (x<0) x=0; printf "%.3f", x}')"
      out="$DIR/frames/pin-$(printf '%09.3f' "$tt").jpg"
      [ -f "$out" ] && continue
      ffmpeg -v error -y -ss "$tt" -i "$SRC" -frames:v 1 -vf "scale='min(iw\,2560)':-2" -q:v 3 "$out" 2>/dev/null || continue
      echo "$out" >> "$WORK/pin_frames.txt"; echo "$tt" >> "$WORK/pin_pts.txt"
    done
    N_PINS=$((N_PINS+1))
  done
  if [ -s "$WORK/pin_frames.txt" ]; then
    log "OCR on $(wc -l < "$WORK/pin_frames.txt" | tr -d ' ') pinned frames..."
    ocr_and_index "$WORK/pin_frames.txt" "$WORK/pin_pts.txt" 1
  fi
fi

# ---- 4. audio + 5. transcript -----------------------------------------------
N_SPEECH=0; ENGINE="none"
if $AUDIO && [ "$HAS_AUDIO" = 1 ] && ! $PINS_ONLY; then
  ffmpeg -v error -y -i "$SRC" -vn -ac 1 -ar 16000 "$DIR/audio.wav" || die "audio extraction failed"
  if ! macos_lt_26; then
    ENGINE="speechanalyzer"
    "$SW_BIN" transcribe "$DIR/audio.wav" "$LOCALE" > "$DIR/transcript.tsv" || log "warn: transcription failed"
  elif have whisper-cli && [ -f "$SW_WHISPER_MODEL" ]; then
    ENGINE="whisper-cpp"
    whisper-cli -m "$SW_WHISPER_MODEL" -f "$DIR/audio.wav" -ocsv -of "$DIR/transcript" >/dev/null 2>&1 || log "warn: whisper failed"
    # csv: start_ms,end_ms,"text"  ->  tsv seconds
    awk -F, 'NR>1{s=$1/1000;e=$2/1000;$1="";$2="";t=substr($0,3);gsub(/^"|"$/,"",t);gsub(/""/,"\"",t);printf "%.2f\t%.2f\t%s\n",s,e,t}' \
      "$DIR/transcript.csv" > "$DIR/transcript.tsv" 2>/dev/null || true
  else
    log "warn: no transcription engine (macOS < 26 without whisper-cpp) - frames only"
  fi
  if [ -s "$DIR/transcript.tsv" ]; then
    while IFS=$'\t' read -r s e text; do
      [ -n "$text" ] || continue
      N_SPEECH=$((N_SPEECH+1))
      echo "INSERT INTO segments(source_id,kind,t_start,t_end,frame,text) VALUES ($(sql_q "$ID"),'speech',$s,$e,NULL,$(sql_q "$text"));" >> "$SQL"
    done < "$DIR/transcript.tsv"
  fi
fi

# ---- 6. index ---------------------------------------------------------------
echo "COMMIT;" >> "$SQL"
db_file "$SQL" >/dev/null || die "index write failed"
if $PINS_ONLY; then
  python3 - "$DIR/meta.json" "$N_PINS" <<'PY2' 2>/dev/null || true
import json,sys; p=sys.argv[1]; m=json.load(open(p)); m["pins"]=m.get("pins",0)+int(sys.argv[2]); json.dump(m,open(p,"w"))
PY2
else
cat > "$DIR/meta.json" <<EOF
{"id":"$ID","source":"$SRC","duration":$DURATION,"interval":$INTERVAL,"detail":"$DETAIL","frames":$N_FRAMES,"frames_indexed":$N_KEPT,"pins":$N_PINS,"ocr_chars":$OCR_CHARS,"speech_segments":$N_SPEECH,"engine":"$ENGINE","indexed_at":"$(date -u +%Y-%m-%dT%H:%M:%SZ)"}
EOF
fi
touch "$DIR/.done"

# ---- 7. brief + report ------------------------------------------------------
log ""
if $PINS_ONLY; then log "added $N_PINS pins to $ID"
else log "indexed $ID  ${DURATION}s  frames $N_KEPT/$N_FRAMES unique (+$N_PINS pins)  ocr ${OCR_CHARS} chars  speech $N_SPEECH segments ($ENGINE)"; fi
if $BRIEF; then
  bash "$SW_SKILL_ROOT/scripts/brief.sh" "$ID" && log "brief: $DIR/brief.md  sheet: $DIR/sheet.jpg"
fi
log "  ask:   bash $SW_SKILL_ROOT/scripts/ask.sh $ID \"<query>\""
log "  frame: bash $SW_SKILL_ROOT/scripts/frame.sh $ID <seconds>"
echo "$ID"
