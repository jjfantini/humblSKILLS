#!/usr/bin/env bash
# verify.sh — evaluate a frozen contract and return verdicts, not opinions.
#
#   bash scripts/verify.sh <contract.json> [--json]
#
# Contract: {"checks":[{...}, ...]}. Check types (see assets/contract.example.json):
#   file_sha256  path, expect                         shasum -a 256
#   file_exists  path                                  test -e
#   json_path    file, path (jq expr), expect          jq -r
#   sql          db, query, expect                     sqlite3
#   http         url, expect_status, [expect_body_contains]   curl --max-time 10
#   command      run, expect_exit, [expect_stdout_contains]   bash -c
#
# Verdicts: VERIFIED (matched) / FAILED (ran, did not match) /
# UNVERIFIED (could not run: missing tool, absent file, unreachable host).
# Exit 0 only when every check is VERIFIED. UNVERIFIED is never a pass.

source "$(dirname "$0")/lib.sh"
CONTRACT="${1:-}"; JSON=false; [ "${2:-}" = --json ] && JSON=true
[ -f "$CONTRACT" ] || { sed -n '2,17p' "$0"; exit 2; }
have jq || die "verify.sh needs jq to read the contract: brew install jq (ask the user before installing)"

CDIR="$(cd "$(dirname "$CONTRACT")" && pwd)"
n="$(jq '.checks | length' "$CONTRACT")"
[ "$n" -gt 0 ] || die "contract has no checks"

# resolve paths relative to the contract file
rel() { case "$1" in /*|"") printf '%s' "$1" ;; *) printf '%s/%s' "$CDIR" "$1" ;; esac; }

# One check per call; prints "<VERDICT>|<detail>". Runs in a subshell from the
# loop so a crash or `exit` inside cannot take down the run. bash 3.2 cannot
# parse `case` inside $(...), hence a function.
# ponytail: no per-check timeout beyond curl --max-time; macOS has no `timeout`.
run_check() {
  set +e
  case "$type" in
    file_sha256)
      p="$(rel "$(jq -r .path <<<"$c")")"; want="$(jq -r .expect <<<"$c")"
      [ -f "$p" ] || { echo "UNVERIFIED|no such file: $p"; exit; }
      got="$(shasum -a 256 "$p" | cut -d' ' -f1)"
      [ "$got" = "$want" ] && echo "VERIFIED|sha256 $got" || echo "FAILED|sha256 $got != $want" ;;
    file_exists)
      p="$(rel "$(jq -r .path <<<"$c")")"
      [ -e "$p" ] && echo "VERIFIED|exists $p" || echo "FAILED|missing $p" ;;
    json_path)
      f="$(rel "$(jq -r .file <<<"$c")")"; path="$(jq -r .path <<<"$c")"; want="$(jq -r .expect <<<"$c")"
      [ -f "$f" ] || { echo "UNVERIFIED|no such file: $f"; exit; }
      got="$(jq -r "$path" "$f" 2>&1)" || { echo "UNVERIFIED|jq error: $got"; exit; }
      [ "$got" = "$want" ] && echo "VERIFIED|$path = $got" || echo "FAILED|$path = $got, expected $want" ;;
    sql)
      d="$(rel "$(jq -r .db <<<"$c")")"; q="$(jq -r .query <<<"$c")"; want="$(jq -r .expect <<<"$c")"
      [ -f "$d" ] || { echo "UNVERIFIED|no such db: $d"; exit; }
      got="$(sqlite3 -readonly "$d" "$q" 2>&1)" || { echo "UNVERIFIED|sqlite error: $got"; exit; }
      [ "$got" = "$want" ] && echo "VERIFIED|$got" || echo "FAILED|got $got, expected $want" ;;
    http)
      url="$(jq -r .url <<<"$c")"; ws="$(jq -r '.expect_status // 200' <<<"$c")"; wb="$(jq -r '.expect_body_contains // empty' <<<"$c")"
      body="$(mktemp)"; trap 'rm -f "$body"' EXIT
      code="$(curl -sS --max-time 10 -o "$body" -w '%{http_code}' "$url" 2>/dev/null)" || { echo "UNVERIFIED|unreachable: $url"; exit; }
      if [ "$code" != "$ws" ]; then echo "FAILED|status $code, expected $ws"
      elif [ -n "$wb" ] && ! grep -qF -- "$wb" "$body"; then echo "FAILED|status $code but body lacks '$wb'"
      else echo "VERIFIED|status $code"; fi ;;
    command)
      run="$(jq -r .run <<<"$c")"; we="$(jq -r '.expect_exit // 0' <<<"$c")"; wo="$(jq -r '.expect_stdout_contains // empty' <<<"$c")"
      so="$(cd "$CDIR" && bash -c "$run" 2>/dev/null)"; ec=$?
      if [ "$ec" != "$we" ]; then echo "FAILED|exit $ec, expected $we"
      elif [ -n "$wo" ] && ! grep -qF -- "$wo" <<<"$so"; then echo "FAILED|exit $ec but stdout lacks '$wo'"
      else echo "VERIFIED|exit $ec"; fi ;;
    *) echo "UNVERIFIED|unknown check type: $type" ;;
  esac
}

results="[]"; all_ok=true
printf '%-10s %-14s %-12s %s\n' "verdict" "id" "type" "detail"
i=0
while [ "$i" -lt "$n" ]; do
  c="$(jq -c ".checks[$i]" "$CONTRACT")"; i=$((i+1))
  id="$(jq -r --arg n "$i" '.id // ("check-" + $n)' <<<"$c")"
  type="$(jq -r '.type' <<<"$c")"
  out="$(run_check)"
  verdict="${out%%|*}"; detail="${out#*|}"
  [ "$verdict" = VERIFIED ] || all_ok=false
  printf '%-10s %-14s %-12s %s\n' "$verdict" "$id" "$type" "$detail"
  results="$(jq -c --arg id "$id" --arg t "$type" --arg v "$verdict" --arg d "$detail" \
    '. + [{id:$id,type:$t,verdict:$v,detail:$d}]' <<<"$results")"
done

overall="FAILED"; $all_ok && overall="VERIFIED"
echo; echo "overall: $overall"
if $JSON; then
  jq -n --arg o "$overall" --arg c "$CONTRACT" --argjson r "$results" \
    '{contract:$c, overall:$o, checked_at:(now|todate), checks:$r}'
fi
$all_ok
