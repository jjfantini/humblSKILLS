#!/usr/bin/env bash
# Decide whether a release-please PR is safe to auto-merge.
#
# Minors and patches (including vX.Y.Z-pre.N) auto-merge. A major does not:
# it is rare, and on 2026-07-30 a phantom 3.0.0 shipped to the Homebrew tap
# 16 seconds after the release PR opened.
#
# Reads the version from the branch's own manifest on both refs rather
# than parsing the PR title, so a change to release-please's title format
# cannot quietly disable the guard. Prerelease suffixes are stripped before
# comparing majors (`2.52.0-pre.1` → `2.52.0`).
#
#   develop → .release-please-manifest.develop.json
#   main    → .release-please-manifest.json
#
# Usage:
#   guard-major-bump.sh <repo> <head-sha> [base-ref]  # exit 0 = safe, 1 = human
#   guard-major-bump.sh --self-test
set -euo pipefail

# Which manifest release-please updates on this branch. Must stay in sync
# with .github/workflows/release.yml.
manifest_for_base() {
  case "${1:-main}" in
    develop) printf '%s' ".release-please-manifest.develop.json" ;;
    *) printf '%s' ".release-please-manifest.json" ;;
  esac
}

# Numeric core of a semver: strip a -prerelease suffix, then require dotted
# digits only. Anything else is unreadable → unsafe (fail closed).
version_core() {
  local v="${1%%-*}"
  case "$v" in '' | *[!0-9.]*) return 1 ;; esac
  printf '%s' "$v"
}

# Exits 0 when cur -> next is a same-major bump.
same_major() {
  local cur next
  cur="$(version_core "${1:-}" || true)"
  next="$(version_core "${2:-}" || true)"
  [ -n "$cur" ] && [ -n "$next" ] && [ "${cur%%.*}" = "${next%%.*}" ]
}

manifest_version() {
  local repo="$1" ref="$2" file="$3"
  gh api "repos/${repo}/contents/${file}?ref=${ref}" \
    -H "Accept: application/vnd.github.raw" --jq '."."'
}

self_test() {
  # shellcheck disable=SC2317
  check() {
    if same_major "$2" "$3"; then got=safe; else got=blocked; fi
    [ "$got" = "$1" ] || { echo "FAIL: same_major('$2','$3') = $got, want $1"; exit 1; }
  }
  check safe    2.44.0 2.44.1        # patch
  check safe    2.44.1 2.45.0        # minor
  check safe    0.1.0  0.2.0         # pre-1.0 minor
  check safe    2.51.0 2.52.0-pre.1  # develop pre
  check safe    2.52.0-pre.1 2.52.0  # graduate pre → stable
  check safe    2.52.0-pre.1 2.52.0-pre.2
  check blocked 2.44.1 3.0.0         # the incident
  check blocked 2.51.0 3.0.0-pre.1   # major pre
  check blocked 3.0.0  2.42.0        # rollback
  check blocked 2.44.1 ''            # unreadable next
  check blocked ''     3.0.0         # unreadable current
  check blocked 2.44.1 'v3.0.0'      # unexpected format

  [ "$(manifest_for_base develop)" = ".release-please-manifest.develop.json" ] \
    || { echo "FAIL: develop must use the develop manifest"; exit 1; }
  [ "$(manifest_for_base main)" = ".release-please-manifest.json" ] \
    || { echo "FAIL: main must use the stable manifest"; exit 1; }
  [ "$(manifest_for_base)" = ".release-please-manifest.json" ] \
    || { echo "FAIL: default base must use the stable manifest"; exit 1; }

  # Stable manifest is last *stable*. A -pre here is the #270 skip.
  if [ -f .release-please-manifest.json ]; then
    stable="$(python3 -c "import json; print(json.load(open('.release-please-manifest.json'))['.'])")"
    case "$stable" in
      *-*) echo "FAIL: .release-please-manifest.json is prerelease: ${stable}"; exit 1 ;;
      '') echo "FAIL: .release-please-manifest.json has empty version"; exit 1 ;;
    esac
  fi
  if [ -f .github/workflows/release.yml ]; then
    grep -q ".release-please-manifest.develop.json" .github/workflows/release.yml \
      || { echo "FAIL: release.yml must point develop at the develop manifest"; exit 1; }
    grep -q "sync-develop-pre-after-stable.sh" .github/workflows/release.yml \
      || { echo "FAIL: release.yml must reset develop pre line after a stable"; exit 1; }
  fi
  echo "guard-major-bump: all checks passed"
}

if [ "${1:-}" = "--self-test" ]; then
  self_test
  exit 0
fi

repo="${1:?usage: guard-major-bump.sh <repo> <head-sha> [base-ref]}"
head="${2:?usage: guard-major-bump.sh <repo> <head-sha> [base-ref]}"
base="${3:-main}"
manifest="$(manifest_for_base "$base")"

current="$(manifest_version "$repo" "$base" "$manifest" || true)"
next="$(manifest_version "$repo" "$head" "$manifest" || true)"

if same_major "$current" "$next"; then
  echo "release ${current} -> ${next}: same major, safe to auto-merge"
  exit 0
fi

echo "release ${current:-<unreadable>} -> ${next:-<unreadable>}: NOT auto-merging."
echo "Major bumps and unreadable manifests need a human. Merge by hand if intended."
exit 1
