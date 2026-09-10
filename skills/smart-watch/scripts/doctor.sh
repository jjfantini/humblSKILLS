#!/usr/bin/env bash
# doctor.sh — report smart-watch dependencies; install only with approval.
#
#   bash scripts/doctor.sh                 report only (exit 1 if a required dep is missing)
#   bash scripts/doctor.sh --install       install missing brew deps; needs a TTY or --yes
#   bash scripts/doctor.sh --install --yes non-interactive install (agent must have user approval first)
#
# The dependency list is the `metadata.dependencies:` block in SKILL.md — one
# flow mapping per line — so the frontmatter and this script cannot drift.

source "$(dirname "$0")/lib.sh"

INSTALL=false; YES=false
for a in "$@"; do
  case "$a" in
    --install) INSTALL=true ;;
    --yes|-y)  YES=true ;;
    -h|--help) sed -n '2,9p' "$0"; exit 0 ;;
    *) die "unknown option: $a" ;;
  esac
done

field() { # field <key> <line>  -> value (unquoted) or ""
  printf '%s\n' "$2" | sed -nE "s/.*[{,] *$1: *\"?([^,\"}]*)\"?.*/\1/p"
}

deps() {
  awk '/^  dependencies:/{f=1;next} f&&/^  [a-z]/{f=0} f&&/^    - \{/{print}' "$SW_SKILL_ROOT/SKILL.md"
}

is_present() { # is_present <name> <kind>
  case "$1" in
    swiftc)        xcode-select -p >/dev/null 2>&1 && have swiftc ;;
    screencapture) [ -x /usr/sbin/screencapture ] ;;
    whisper-cpp)   have whisper-cli ;;
    *)             have "$1" ;;
  esac
}

applies() { # applies <when>
  case "$1" in
    "")           return 0 ;;
    macos_lt_26)  macos_lt_26 ;;
    *)            return 0 ;;
  esac
}

MISSING_REQ=(); MISSING_OPT=(); MISSING_BREW=()
printf '%-14s %-7s %-9s %-8s %s\n' "tool" "kind" "status" "needed" "provides"
while IFS= read -r line; do
  name="$(field name "$line")"; kind="$(field kind "$line")"
  req="$(field required "$line")"; prov="$(field provides "$line")"
  inst="$(field install "$line")"; when="$(field when "$line")"
  if ! applies "$when"; then status="n/a"; needed="no (macOS $(macos_major))"
  elif is_present "$name" "$kind"; then status="ok"; needed="$req"
  else
    status="MISSING"; needed="$req"
    [ "$kind" = brew ] && MISSING_BREW+=("$name")
    if [ "$req" = true ]; then MISSING_REQ+=("$name"); else MISSING_OPT+=("$name"); fi
  fi
  printf '%-14s %-7s %-9s %-8s %s\n' "$name" "$kind" "$status" "$needed" "$prov"
  if [ "$status" = MISSING ]; then
    [ "$kind" = brew ] && printf '%-14s install: brew install %s\n' "" "$name"
    [ -n "$inst" ]     && printf '%-14s install: %s\n' "" "$inst"
  fi
done < <(deps)

echo
echo "macOS $(sw_vers -productVersion)  transcription: $(macos_lt_26 && echo 'whisper-cpp fallback' || echo 'native SpeechAnalyzer')"
echo "state: $SW_STATE   cache: $SW_CACHE"

if [ ${#MISSING_BREW[@]} -gt 0 ]; then
  echo
  echo "missing brew packages: ${MISSING_BREW[*]}"
  echo "  brew install ${MISSING_BREW[*]}"
  if $INSTALL; then
    have brew || die "Homebrew not found: https://brew.sh"
    if ! $YES; then
      [ -t 0 ] || die "--install without a TTY requires --yes (get the user's approval first)"
      printf 'Install %s with Homebrew? [y/N] ' "${MISSING_BREW[*]}"; read -r ans
      case "$ans" in y|Y|yes) ;; *) die "declined" ;; esac
    fi
    brew install "${MISSING_BREW[@]}"
    for p in "${MISSING_BREW[@]}"; do
      if [ "$p" = whisper-cpp ] && [ ! -f "$SW_WHISPER_MODEL" ]; then
        mkdir -p "$SW_MODELS"
        log "downloading whisper model ggml-base.en.bin (~150 MB) to $SW_MODELS"
        curl -L --fail --progress-bar -o "$SW_WHISPER_MODEL" "$SW_WHISPER_MODEL_URL"
      fi
    done
  else
    echo "  (re-run with --install to install these; nothing is installed without approval)"
  fi
fi

if [ ${#MISSING_REQ[@]} -gt 0 ] && ! $INSTALL; then
  log; log "required tools missing: ${MISSING_REQ[*]}"
  exit 1
fi

# Build/refresh the Swift helper and smoke-test OCR on a generated frame.
if have swiftc; then
  ensure_watchkit
  if have ffmpeg; then
    tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
    ffmpeg -v error -f lavfi -i color=c=white:s=64x64 -frames:v 1 -y "$tmp/px.png"
    "$SW_BIN" ocr "$tmp/px.png" >/dev/null && echo "watchkit: ok ($SW_BIN)"
  fi
fi
