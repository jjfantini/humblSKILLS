#!/usr/bin/env bash
# After main graduates vX.Y.Z, develop must not keep incrementing X.Y.Z-pre.N.
#
# release-please `versioning: prerelease` only starts a new pre line when the
# last-released version is *not* already a prerelease:
#
#   2.52.0-pre.3 + feat/fix → 2.52.0-pre.4   (still < 2.52.0)
#   2.52.0       + fix      → 2.52.1-pre
#   2.52.0       + feat     → 2.53.0-pre
#
# Semver: 2.52.0-pre.N < 2.52.0. Beta = max(stable, pre) therefore stays on
# the graduated stable until `.release-please-manifest.develop.json` records
# that stable (no -pre suffix). The next conventional commit then opens
# X.Y.(Z+1)-pre.1 or a feat minor. Intra-cycle 2.52.1-pre.2 stays on
# versioning: prerelease and does not need this rewrite.
#
# Usage:
#   sync-develop-pre-after-stable.sh --stable X.Y.Z [--manifest PATH] [--write]
#   sync-develop-pre-after-stable.sh --self-test
#
# Without --write, prints whether a rewrite would happen and exits 0.
# With --write, rewrites only when the manifest is exactly X.Y.Z-pre or
# X.Y.Z-pre.N. A newer line (2.52.1-pre.1 after 2.52.0) is left alone.
set -euo pipefail

manifest_default=".release-please-manifest.develop.json"

# True when $1 is a prerelease of exactly $2 (2.52.0-pre or 2.52.0-pre.3).
is_pre_of_stable() {
  local current="$1" stable="$2"
  case "$current" in
    "${stable}-pre" | "${stable}-pre."*) return 0 ;;
    *) return 1 ;;
  esac
}

read_manifest_version() {
  python3 -c "import json,sys; print(json.load(open(sys.argv[1]))['.'])" "$1"
}

write_manifest_version() {
  python3 -c '
import json, sys
path, ver = sys.argv[1], sys.argv[2]
with open(path) as f:
    data = json.load(f)
data["."] = ver
with open(path, "w") as f:
    json.dump(data, f, indent=2)
    f.write("\n")
' "$1" "$2"
}

# Core: print "reset" or "keep" given current and stable versions.
decide() {
  local current="$1" stable="$2"
  case "$stable" in
    '' | *[!0-9.]* | *..* | .* | *.) return 2 ;;
  esac
  case "$stable" in
    *.*.*) ;;
    *) return 2 ;;
  esac
  if [ -z "$current" ]; then
    return 2
  fi
  if is_pre_of_stable "$current" "$stable"; then
    printf '%s\n' reset
    return 0
  fi
  printf '%s\n' keep
}

self_test() {
  check() {
    local want="$1" current="$2" stable="$3" got
    if ! got="$(decide "$current" "$stable")"; then
      got=invalid
    fi
    [ "$got" = "$want" ] || {
      echo "FAIL: decide('$current','$stable') = $got, want $want"
      exit 1
    }
  }
  check reset   2.52.0-pre.3 2.52.0   # the stuck 2.52.0-pre line
  check reset   2.52.0-pre   2.52.0
  check reset   2.52.1-pre.1 2.52.1   # next graduate
  check keep    2.52.0       2.52.0   # already recorded
  check keep    2.52.1-pre.1 2.52.0   # already moved on
  check keep    2.53.0-pre.1 2.52.0
  check keep    2.51.0-pre.1 2.52.0   # do not rewrite an older line
  check invalid ''           2.52.0
  check invalid 2.52.0-pre.3 ''
  check invalid 2.52.0-pre.3 2.52
  check invalid 2.52.0-pre.3 v2.52.0

  tmp="$(mktemp)"
  printf '%s\n' '{ ".": "2.52.0-pre.3" }' >"$tmp"
  decide_out="$(decide "$(read_manifest_version "$tmp")" 2.52.0)"
  [ "$decide_out" = reset ] || { echo "FAIL: temp manifest decide"; exit 1; }
  write_manifest_version "$tmp" 2.52.0
  [ "$(read_manifest_version "$tmp")" = 2.52.0 ] || {
    echo "FAIL: write did not set 2.52.0"
    exit 1
  }
  rm -f "$tmp"

  echo "sync-develop-pre-after-stable: all checks passed"
}

if [ "${1:-}" = "--self-test" ]; then
  self_test
  exit 0
fi

stable=""
manifest="$manifest_default"
write=0
while [ $# -gt 0 ]; do
  case "$1" in
    --stable)
      stable="${2:-}"
      shift 2
      ;;
    --manifest)
      manifest="${2:-}"
      shift 2
      ;;
    --write)
      write=1
      shift
      ;;
    *)
      echo "usage: $0 --stable X.Y.Z [--manifest PATH] [--write]" >&2
      exit 2
      ;;
  esac
done

if [ -z "$stable" ]; then
  echo "usage: $0 --stable X.Y.Z [--manifest PATH] [--write]" >&2
  exit 2
fi
stable="${stable#v}"

current="$(read_manifest_version "$manifest")"
action="$(decide "$current" "$stable" || true)"
if [ "$action" != reset ]; then
  echo "keep ${manifest} at ${current} (stable ${stable})"
  exit 0
fi

echo "reset ${manifest}: ${current} -> ${stable}"
if [ "$write" -eq 1 ]; then
  write_manifest_version "$manifest" "$stable"
fi
