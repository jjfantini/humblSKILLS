#!/usr/bin/env bash
# selftest.sh — the one runnable check. Generates fixtures locally (ffmpeg lavfi
# burned-in timecode + `say` speech), indexes them into a throwaway state dir,
# and asserts: PTS alignment, OCR, transcript, search, and verify verdicts.
#
#   bash scripts/selftest.sh

HERE="$(cd "$(dirname "$0")" && pwd)"
export XDG_STATE_HOME="$(mktemp -d "${TMPDIR:-/tmp}/smart-watch-selftest.XXXXXX")"
trap 'rm -rf "$XDG_STATE_HOME"' EXIT
source "$HERE/lib.sh"
have ffmpeg || die "ffmpeg missing - run doctor.sh"

fail=0
check() { if eval "$2"; then echo "ok    $1"; else echo "FAIL  $1"; fail=1; fi; }

W="$XDG_STATE_HOME/fixtures"; mkdir -p "$W"
# 12s test pattern with burned-in hh:mm:ss.mmm + frame counter, then drop frames
# irregularly so the clip is VFR like a real screencapture.
ffmpeg -v error -f lavfi -i testsrc2=size=1280x720:rate=30 -t 12 -c:v libx264 -pix_fmt yuv420p -y "$W/cfr.mp4"
ffmpeg -v error -i "$W/cfr.mp4" -vf "select='not(mod(n,7))+gt(random(1),0.7)'" -fps_mode vfr -c:v libx264 -pix_fmt yuv420p -y "$W/vfr.mp4"
if have say; then
  say -o "$W/speech.aiff" "Build number eight eight four two passed all checks. Then the migration failed with a timeout."
  ffmpeg -v error -y -i "$W/vfr.mp4" -i "$W/speech.aiff" -c:v copy -c:a aac "$W/fixture.mp4"
else
  cp "$W/vfr.mp4" "$W/fixture.mp4"
fi

ID="$(bash "$HERE/watch.sh" "$W/fixture.mp4" --detail fine --at 4.5,99 2>"$W/watch.log" | tail -1)"
check "watch.sh produced an id"            '[ -n "$ID" ]'
check "frames extracted"                    '[ "$(ls "$SW_SOURCES/$ID/frames" | wc -l | tr -d " ")" -ge 8 ]'

# PTS alignment: the OCR text of each indexed frame must contain the burned-in
# clock at the same second as its t_start. `fps=1` would break this.
misaligned="$(db "SELECT count(*) FROM segments WHERE source_id=$(sql_q "$ID") AND kind='ocr'
  AND instr(text, printf('00:00:%02d.', cast(t_start as int))) = 0;")"
check "every OCR frame's t_start matches its burned-in clock (misaligned=$misaligned)" '[ "$misaligned" = 0 ]'

if have say; then
  hits="$(bash "$HERE/ask.sh" "$ID" "8842" --kind speech | wc -l | tr -d ' ')"
  check "transcript indexed and searchable (8842)"   '[ "$hits" -ge 1 ]'
  check "ask.sh OR query"                             '[ "$(bash "$HERE/ask.sh" "$ID" "timeout OR nonsenseword" | wc -l | tr -d " ")" -ge 1 ]'
fi
check "pins indexed (3 frames for one in-range pin, out-of-range skipped)" '[ "$(db "SELECT count(*) FROM segments WHERE source_id=$(sql_q "$ID") AND pinned=1;")" = 3 ]'
check "brief.md written with all sections"   'grep -q "^## Errors" "$SW_SOURCES/$ID/brief.md" && grep -q "^## Screen changes" "$SW_SOURCES/$ID/brief.md" && grep -q "^## Pinned" "$SW_SOURCES/$ID/brief.md"'
check "contact sheet rendered"               '[ -s "$SW_SOURCES/$ID/sheet.jpg" ]'
check "natural-language ask is rewritten"    '[ "$(bash "$HERE/ask.sh" "$ID" "what failed?" 2>/dev/null | wc -l | tr -d " ")" -ge 1 ]'
check "pins-only re-run adds without reindex" 'bash "$HERE/watch.sh" "$W/fixture.mp4" --at 2 --no-brief 2>&1 >/dev/null | grep -c "adding pins only" | grep -q 1'
check "frame.sh returns an existing jpeg"  '[ -f "$(bash "$HERE/frame.sh" "$ID" 3 2>/dev/null)" ]'
check "re-watch is a no-op"                '[ "$(bash "$HERE/watch.sh" "$W/fixture.mp4" 2>/dev/null)" = "$ID" ]'

if have jq; then
  H="$(shasum -a 256 "$W/fixture.mp4" | cut -d" " -f1)"
  cat > "$W/contract.json" <<JSON
{"checks":[
 {"id":"good","type":"file_sha256","path":"fixture.mp4","expect":"$H"},
 {"id":"bad","type":"command","run":"exit 3","expect_exit":0},
 {"id":"unv","type":"http","url":"http://127.0.0.1:1/"}
]}
JSON
  rc=0; out="$(bash "$HERE/verify.sh" "$W/contract.json")" || rc=$?
  check "verify.sh exit 1 on mixed contract"          '[ "$rc" = 1 ]'
  check "verify.sh VERIFIED/FAILED/UNVERIFIED verdicts" 'grep -q "^VERIFIED *good" <<<"$out" && grep -q "^FAILED *bad" <<<"$out" && grep -q "^UNVERIFIED *unv" <<<"$out"'
fi

[ "$fail" = 0 ] && echo "selftest: all passed" || { echo "selftest: FAILED (see $XDG_STATE_HOME, kept)"; trap - EXIT; exit 1; }
