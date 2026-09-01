#!/usr/bin/env bash
# Decide whether a release-please PR is safe to auto-merge.
#
# Minors and patches (including vX.Y.Z-pre.N) auto-merge. A major does not:
# it is rare, and on 2026-07-30 a phantom 3.0.0 shipped to the Homebrew tap
# 16 seconds after the release PR opened.
#
# Reads the version from `.release-please-manifest.json` on both refs rather
# than parsing the PR title, so a change to release-please's title format
# cannot quietly disable the guard. Prerelease suffixes are stripped before
# comparing majors (`2.52.0-pre.1` → `2.52.0`).
#
# Usage:
#   guard-major-bump.sh <repo> <head-sha> [base-ref]  # exit 0 = safe, 1 = human
#   guard-major-bump.sh --self-test
set -euo pipefail

MANIFEST=".release-please-manifest.json"

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
  gh api "repos/$1/contents/${MANIFEST}?ref=$2" \
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
  echo "guard-major-bump: all checks passed"
}

if [ "${1:-}" = "--self-test" ]; then
  self_test
  exit 0
fi

repo="${1:?usage: guard-major-bump.sh <repo> <head-sha> [base-ref]}"
head="${2:?usage: guard-major-bump.sh <repo> <head-sha> [base-ref]}"
base="${3:-main}"

current="$(manifest_version "$repo" "$base" || true)"
next="$(manifest_version "$repo" "$head" || true)"

if same_major "$current" "$next"; then
  echo "release ${current} -> ${next}: same major, safe to auto-merge"
  exit 0
fi

echo "release ${current:-<unreadable>} -> ${next:-<unreadable>}: NOT auto-merging."
echo "Major bumps and unreadable manifests need a human. Merge by hand if intended."
exit 1
