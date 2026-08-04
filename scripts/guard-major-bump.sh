#!/usr/bin/env bash
# Decide whether a release-please PR is safe to auto-merge.
#
# Auto-merge lands release PRs unattended, which is fine for minors and
# patches and wrong for a major: on 2026-07-30 release-please lost track of
# the latest release, re-derived the version from a `feat!` months old, and
# a phantom 3.0.0 was merged 16 seconds after it opened and shipped to the
# Homebrew tap before anyone saw it. A major is rare and worth a human.
#
# Reads the version from `.release-please-manifest.json` on both refs rather
# than parsing the PR title, so a change to release-please's title format
# cannot quietly disable the guard.
#
# Usage:
#   guard-major-bump.sh <repo> <head-sha>   # exit 0 = safe, 1 = needs a human
#   guard-major-bump.sh --self-test
set -euo pipefail

MANIFEST=".release-please-manifest.json"

# Exits 0 when cur -> next is a same-major bump. Anything it cannot read as
# two dotted numbers is treated as unsafe: a guard that fails open is not a
# guard, and "releases paused" is a loud failure while "major shipped" is not.
same_major() {
  local cur="${1:-}" next="${2:-}"
  case "$cur" in '' | *[!0-9.]*) return 1 ;; esac
  case "$next" in '' | *[!0-9.]*) return 1 ;; esac
  [ "${cur%%.*}" = "${next%%.*}" ]
}

manifest_version() {
  gh api "repos/$1/contents/${MANIFEST}?ref=$2" \
    -H "Accept: application/vnd.github.raw" --jq '."."'
}

self_test() {
  # shellcheck disable=SC2317
  check() {
    if same_major "$2" "$3"; then got=safe; else got=blocked; fi
    [ "$got" = "$1" ] || { echo "FAIL: same_major('$2','$3') = $got, want $1"; exit 1; }
  }
  check safe    2.44.0 2.44.1     # patch
  check safe    2.44.1 2.45.0     # minor
  check safe    0.1.0  0.2.0      # pre-1.0 minor
  check blocked 2.44.1 3.0.0      # the incident
  check blocked 3.0.0  2.42.0     # rollback, also a human's call
  check blocked 2.44.1 ''         # unreadable next -> fail closed
  check blocked ''     3.0.0      # unreadable current -> fail closed
  check blocked 2.44.1 'v3.0.0'   # unexpected format -> fail closed
  echo "guard-major-bump: all checks passed"
}

if [ "${1:-}" = "--self-test" ]; then
  self_test
  exit 0
fi

repo="${1:?usage: guard-major-bump.sh <repo> <head-sha>}"
head="${2:?usage: guard-major-bump.sh <repo> <head-sha>}"

current="$(manifest_version "$repo" main || true)"
next="$(manifest_version "$repo" "$head" || true)"

if same_major "$current" "$next"; then
  echo "release ${current} -> ${next}: same major, safe to auto-merge"
  exit 0
fi

echo "release ${current:-<unreadable>} -> ${next:-<unreadable>}: NOT auto-merging."
echo "Major bumps and unreadable manifests need a human. Merge by hand if intended."
exit 1
