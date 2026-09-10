#!/usr/bin/env bash
# brief.sh — one markdown digest per source: errors, pinned moments, screen changes,
# transcript, plus a contact sheet (sheet.jpg) of the change-point frames so the
# agent can see the whole recording in one Read.
#
#   bash scripts/brief.sh <id> [--threshold 0.5] [--max-changes 40]
#
# Writes <source dir>/brief.md and sheet.jpg, prints brief.md to stdout.
# threshold: Jaccard similarity of the OCR *word* sets of consecutive frames below
# which a frame counts as a screen change. Measured: a static screen re-OCRs at
# 0.75-0.96 (h264 + Vision jitter), so exact-line matching reads static as 100%
# changed; 0.5 separates real changes from jitter. Lower = fewer, bigger changes.

source "$(dirname "$0")/lib.sh"
ID="${1:-}"; shift || true
THRESH=0.5; MAXC=40
while [ $# -gt 0 ]; do
  case "$1" in
    --threshold)   THRESH="$2"; shift 2 ;;
    --max-changes) MAXC="$2"; shift 2 ;;
    *) die "unknown option: $1" ;;
  esac
done
[ -n "$ID" ] || { sed -n '2,10p' "$0"; exit 2; }
DIR="$SW_SOURCES/$ID"; [ -d "$DIR" ] || die "unknown source id: $ID"
have python3 || die "python3 not found (ships with Xcode Command Line Tools)"

python3 - "$SW_DB" "$ID" "$DIR" "$THRESH" "$MAXC" <<'PY' > "$DIR/brief.md"
import json, re, sqlite3, sys
db, sid, d, thresh, maxc = sys.argv[1], sys.argv[2], sys.argv[3], float(sys.argv[4]), int(sys.argv[5])
meta = json.load(open(f"{d}/meta.json"))
con = sqlite3.connect(db)
rows = con.execute("SELECT kind,t_start,t_end,frame,text,pinned FROM segments WHERE source_id=? ORDER BY t_start", (sid,)).fetchall()
ocr = [r for r in rows if r[0] == "ocr" and not r[5]]
pins = [r for r in rows if r[5]]
speech = [r for r in rows if r[0] == "speech"]

def hms(t):
    t = float(t); h = int(t // 3600); m = int(t % 3600 // 60); s = t % 60
    return f"{h:02d}:{m:02d}:{s:05.2f}"
def lines(text): return [l.strip() for l in text.split(" | ") if l.strip()]
def words(text): return set(re.findall(r"[a-z0-9]{3,}", text.lower()))
def jaccard(a, b):
    return 1.0 if not a and not b else len(a & b) / len(a | b)
def novel(ls, other_words, k=2):
    # lines carrying >= k words absent from the other frame (drops OCR jitter)
    return [l for l in ls if len(words(l) - other_words) >= k]

ERR = re.compile(r"\b(error|errors|exception|traceback|failed|failure|fatal|panic|warning|warn|"
                 r"undefined|cannot read|not found|denied|forbidden|unauthori[sz]ed|timeout|timed out|"
                 r"refused|crash|crashed|segfault|stack trace|unhandled|rejected|invalid|"
                 r"(?-i:E[A-Z]{3,}[A-Z0-9]*)|[45]\d\d (?:not found|bad|internal|forbidden|unauthorized|error|gateway|unavailable))\b", re.I)

out = []
P = out.append
P(f"# Brief: {sid}")
P("")
P(f"- source: `{meta.get('source')}`  duration: {hms(meta.get('duration') or 0)}  detail: {meta.get('detail')} (every {meta.get('interval')}s)")
P(f"- frames: {meta.get('frames_indexed')} unique of {meta.get('frames')} sampled, {meta.get('pins')} pinned  |  speech: {meta.get('speech_segments')} segments ({meta.get('engine')})  |  indexed {meta.get('indexed_at')}")
P(f"- files: `{d}/` (frames/, brief.md, sheet.jpg, ocr.tsv, transcript.tsv)")
P("")

# --- errors
hits = []
for k, t, te, fr, text, pin in rows:
    m = ERR.search(text)
    if m:
        ls = lines(text) if k == "ocr" else [text]
        ctx = next((l for l in ls if ERR.search(l)), ls[0] if ls else text)
        hits.append((t, k, fr, ctx[:160]))
P(f"## Errors and warnings ({len(hits)})")
P("")
if not hits: P("none matched (pattern: error/exception/failed/timeout/denied/4xx-5xx text…)")
seen = set()
for t, k, fr, ctx in hits:
    key = ctx.lower()
    if key in seen: continue
    seen.add(key)
    P(f"- {hms(t)}  {k}  {fr or '-'}  `{ctx}`")
P("")

# --- pins
if pins:
    P(f"## Pinned frames ({len(pins)}) — exact frame at each --at time plus ±2s")
    P("")
    for k, t, te, fr, text, pin in pins:
        ls = lines(text)
        P(f"- {hms(t)}  {fr}")
        for l in ls[:8]: P(f"    {l[:140]}")
    P("")

# --- screen changes
changes = []
prev = None  # (lines, words) of the last frame
for k, t, te, fr, text, pin in ocr:
    cur, cw = lines(text), words(text)
    if prev is None:
        changes.append((t, fr, 0.0, cur[:6], []))
    else:
        pl, pw = prev
        sim = jaccard(pw, cw)
        if sim < thresh:
            changes.append((t, fr, sim, novel(cur, pw)[:6], novel(pl, cw)[:3]))
    prev = (cur, cw)
P(f"## Screen changes ({len(changes)}) — frames whose visible words differ from the previous frame (similarity < {thresh})")
P("")
if not ocr: P("no OCR text indexed (no video stream, or no readable text on screen)")
shown = changes if len(changes) <= maxc else [changes[int(i * len(changes) / maxc)] for i in range(maxc)]
for t, fr, sim, new, gone in shown:
    tag = "start" if sim == 0.0 and t == shown[0][0] else f"changed {int((1 - sim) * 100)}%"
    P(f"- {hms(t)}  {tag}  {fr}")
    for l in new: P(f"    + {l[:140]}")
    for l in gone: P(f"    - {l[:100]}")
if len(changes) > maxc: P(f"- … {len(changes) - maxc} more; raise --max-changes or use ask.sh --timeline")
P("")

# --- transcript
P(f"## Transcript ({len(speech)} segments)")
P("")
if not speech: P("no speech indexed")
for k, t, te, fr, text, pin in speech:
    P(f"- [{hms(t)}–{hms(te or t)}] {text}")
P("")

# --- sheet manifest
sheet = [c for c in changes]
if len(sheet) > 12: sheet = [sheet[int(i * len(sheet) / 12)] for i in range(12)]
json.dump([{"t": c[0], "frame": c[1]} for c in sheet], open(f"{d}/sheet.json", "w"))
P("## Contact sheet")
P("")
if sheet:
    P(f"`{d}/sheet.jpg` — {len(sheet)} change-point frames, left→right then top→bottom: " + ", ".join(hms(c[0]) for c in sheet))
else:
    P("none (no frames)")
P("")
P("## Next")
P("")
P(f"- search: `bash scripts/ask.sh {sid} \"<word OR phrase>\"`   timeline: `bash scripts/ask.sh {sid} --timeline --from S --to S`")
P(f"- look:   `bash scripts/frame.sh {sid} <hh:mm:ss>` then Read the jpeg")
P(f"- more detail at a moment: `bash scripts/watch.sh <file> --force --at 1:23,4:56`")
print("\n".join(out))
PY

# Contact sheet: same-size frames tiled 4 wide, 640px per cell. No drawtext in brew
# ffmpeg (no libfreetype), so timestamps live in brief.md, not on the image.
n="$(python3 -c "import json;print(len(json.load(open('$DIR/sheet.json'))))")"
if [ "$n" -gt 0 ]; then
  python3 -c "import json;[print(\"file '$DIR/\"+x['frame']+\"'\") for x in json.load(open('$DIR/sheet.json'))]" > "$DIR/sheet.txt"
  rows=$(( (n + 3) / 4 )); cols=4; [ "$n" -lt 4 ] && cols="$n"
  ffmpeg -v error -y -f concat -safe 0 -i "$DIR/sheet.txt" -vf "scale=640:-2,tile=${cols}x${rows}" -frames:v 1 -q:v 4 "$DIR/sheet.jpg" \
    || log "warn: contact sheet failed"
  rm -f "$DIR/sheet.txt"
fi
cat "$DIR/brief.md"
