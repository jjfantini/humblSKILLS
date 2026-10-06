#!/usr/bin/env bash
# tests/run.sh - test suite for smart-skill's scaffold.sh and promote.sh.
#
# Hermetic: HOME, the clone cache, and "GitHub" all live under a temp dir.
# GitHub is a set of bare repos reached through promote.sh's
# HUMBLSKILLS_PROMOTE_GIT_BASE seam, and `gh` is a stub on PATH that records
# every call. Nothing touches the network or a real repository.
#
# Inside the humblSKILLS repo it also checks the frontmatter parser against
# registry.json and the credential scanner against every published skill.
#
# Usage:
#   bash tests/run.sh
#   bash tests/run.sh --verbose
#
# Runs under bash 3.2, so `/bin/bash tests/run.sh` on macOS exercises the
# shell FDEs actually have.
#
# Exit codes:
#   0  all tests passed
#   1  one or more tests failed

set -uo pipefail

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
SKILL_ROOT=$(dirname "$SCRIPT_DIR")
PROMOTE_SH="$SKILL_ROOT/scripts/promote.sh"
SCAFFOLD_SH="$SKILL_ROOT/scripts/scaffold.sh"
ENGINE="$SKILL_ROOT/scripts/lib/promote.py"
REPO_ROOT=$(cd "$SKILL_ROOT/../.." && pwd)
SH="${BASH:-bash}"

VERBOSE=0
[[ "${1:-}" == "--verbose" ]] && VERBOSE=1

RED='\033[31m'
GRN='\033[32m'
BOLD='\033[1m'
RST='\033[0m'
PASS=0
FAIL=0
FAILED=()

pass() {
  PASS=$((PASS + 1))
  [[ "$VERBOSE" -eq 1 ]] && printf "  ${GRN}PASS${RST} %s\n" "$1"
  return 0
}
fail() {
  FAIL=$((FAIL + 1))
  FAILED+=("$1")
  printf "  ${RED}FAIL${RST} %s\n" "$1"
}
section() { printf "\n${BOLD}== %s ==${RST}\n" "$1"; }
expect_exit() {
  if [[ "$3" -eq "$1" ]]; then pass "$2"; else fail "$2 (expected exit $1, got $3)"; fi
}
expect_contains() {
  if printf '%s' "$3" | grep -qF -- "$2"; then
    pass "$1"
  else
    fail "$1 (output missing '$2')"
    [[ "$VERBOSE" -eq 1 ]] && printf '    --- output ---\n%s\n    ---\n' "$3"
  fi
}
expect_not_contains() {
  if printf '%s' "$3" | grep -qF -- "$2"; then
    fail "$1 (output unexpectedly contained '$2')"
    [[ "$VERBOSE" -eq 1 ]] && printf '    --- output ---\n%s\n    ---\n' "$3"
  else
    pass "$1"
  fi
}
expect_eq() {
  if [[ "$3" == "$2" ]]; then pass "$1"; else fail "$1 (expected '$2', got '$3')"; fi
}

# ------------------------------------------------------------------ sandbox
# Physical path: macOS temp dirs sit behind /var -> /private/var.
T=$(cd "$(mktemp -d)" && pwd -P)
trap 'cd / && rm -rf "$T"' EXIT
export HOME="$T/home"
mkdir -p "$HOME" "$T/bin" "$T/skills" "$T/gh-state"
git config --global user.name "Test Author"
git config --global user.email "author@test.invalid"
git config --global init.defaultBranch main
git config --global advice.detachedHead false
export NO_COLOR=1
# Importing the engine must not leave __pycache__ inside the skill.
export PYTHONDONTWRITEBYTECODE=1
export HUMBLSKILLS_PROMOTE_HOME="$T/cache"
export HUMBLSKILLS_PROMOTE_GIT_BASE="file://$T/github"
export GH_FAKE_ROOT="$T/github"
export GH_STATE="$T/gh-state"
unset HUMBLSKILLS_PROFILE HAPPYSKILLS_REPO XDG_CACHE_HOME XDG_DATA_HOME

cat >"$T/bin/gh" <<'STUB'
#!/usr/bin/env bash
# Records calls; emulates the handful of gh commands promote.sh uses.
echo "gh $*" >>"$GH_STATE/calls"
login="${GH_LOGIN:-forkuser}"
case "$1" in
  repo)
    case "$2" in
      view) echo "${GH_PERMISSION:-WRITE}" ;;
      fork)
        slug="$3"
        dst="$GH_FAKE_ROOT/$login/${slug#*/}.git"
        [[ -d "$dst" ]] || git clone --quiet --bare "$GH_FAKE_ROOT/$slug.git" "$dst"
        ;;
    esac
    ;;
  api)
    case "$2" in
      user) echo "$login" ;;
      repos/*/forks*)
        slug="${2#repos/}"
        slug="${slug%%/forks*}"
        [[ -d "$GH_FAKE_ROOT/$login/${slug#*/}.git" ]] && echo "file://$GH_FAKE_ROOT/$login/${slug#*/}.git"
        ;;
    esac
    ;;
  pr)
    case "$2" in
      list) [[ -f "$GH_STATE/pr-url" ]] && cat "$GH_STATE/pr-url" ;;
      create)
        while [[ $# -gt 0 ]]; do
          [[ "$1" == --body-file ]] && cp "$2" "$GH_STATE/pr-body.md"
          shift
        done
        echo "https://github.test/pull/1" >"$GH_STATE/pr-url"
        echo "https://github.test/pull/1"
        ;;
      comment) : ;;
    esac
    ;;
esac
exit 0
STUB
chmod +x "$T/bin/gh"
export PATH="$T/bin:$PATH"

reset_gh() { rm -f "$GH_STATE"/*; unset GH_PERMISSION GH_LOGIN; }

# A registry repo shaped like humblSKILLS: skills/, registry.json, a .gitignore
# that drops bin/ and .env, and a published skill with a previous name.
make_registry_repo() { # make_registry_repo <owner/name> <branches...>
  local slug="$1"
  shift
  local seed="$T/seed-${slug//\//-}"
  mkdir -p "$seed/skills/existing-skill"
  (
    cd "$seed" || exit 1
    git init -q
    git checkout -q -b main
    printf 'bin/\n.env\n' >.gitignore
    printf '{"schema_version": 1, "skills": []}\n' >registry.json
    printf -- '---\nname: existing-skill\ndescription: Exists. Use when testing.\nmetadata:\n  version: "1.2.0"\n  category: meta\n  previous_names: [old-name]\n---\n\nbody\n' \
      >skills/existing-skill/SKILL.md
    git add -A
    git commit -q -m "chore: seed registry"
    for b in "$@"; do git branch -q "$b"; done
  )
  mkdir -p "$T/github/${slug%/*}"
  git clone -q --bare "$seed" "$T/github/$slug.git"
}

# A finished smart skill, as an agent leaves it after following the create workflow.
make_skill() { # make_skill <parent> <name>
  local d="$1/$2"
  mkdir -p "$d/references/wiki/review/risk" "$d/references/raw" "$d/scripts"
  cp "$SKILL_ROOT/scripts/lint.sh" "$d/scripts/lint.sh"
  cp "$SKILL_ROOT/references/_brain.md" "$SKILL_ROOT/references/_template.md" "$d/references/"
  printf '# Index\n\n<!-- GENERATED:START -->\n<!-- GENERATED:END -->\n' >"$d/references/_index.md"
  printf '# Patterns\n\n---\n\n(no entries yet)\n' >"$d/references/patterns.md"
  printf '# Decisions\n\n---\n\n(no entries yet)\n' >"$d/references/decisions.md"
  printf '# Log\n\n---\n\n[INGEST 2026-10-01] created\n' >"$d/references/log.md"
  touch "$d/references/raw/.gitkeep"
  cat >"$d/references/wiki/review/risk/hotspots.md" <<'EOF'
---
title: "Where risk hides in a diff"
context: review
category: risk
concept: hotspots
description: "Migrations, auth, and config changes deserve a second look"
tags: review, risk
sources: []
last_ingested: 2026-10-01
---

## Hotspots

Migrations, auth, and config changes.
EOF
  cat >"$d/SKILL.md" <<EOF
---
name: $2
description: >
  Review a pull request diff for risky changes. Use when the user says
  "review this PR". Do NOT use for writing commit messages.
license: MIT
metadata:
  author: Test Author
  version: "0.1.0"
  category: development
  tags: [code-review, git, pull-requests]
  platforms: [claude-code, cursor, codex]
  preserve:
    - references/raw/
    - references/wiki/
    - references/decisions.md
    - references/log.md
    - references/patterns.md
---

# $2

## Brain Protocol

Read references/_index.md before acting; append to references/log.md after.

## Examples

### Example 1: review a diff

User says: "review this PR"

Result: a checklist of risky hunks.
EOF
  bash "$d/scripts/lint.sh" "$d" >/dev/null 2>&1
  printf '%s\n' "$d"
}

fresh() { # fresh <name> - a new finished skill under $T/skills
  rm -rf "${T:?}/skills/$1"
  make_skill "$T/skills" "$1"
}

gate() { # gate <dir> [promote.sh args...] - quality gate only
  local dir="$1"
  shift
  OUT=$("$SH" "$PROMOTE_SH" "$dir" --check "$@" 2>&1)
  RC=$?
}

set_desc() { # set_desc <skill dir> <description line> - replace the folded description
  python3 - "$1/SKILL.md" "$2" <<'EOF'
import re, sys
p, desc = sys.argv[1], sys.argv[2]
s = open(p).read()
s = re.sub(r"description: >\n(?:  .*\n)+", "description: " + desc + "\n", s, count=1)
open(p, "w").write(s)
EOF
}

frontmatter_of() { awk 'NR==1{next} /^---$/{exit} {print}' "$1/SKILL.md"; }

make_registry_repo jjfantini/humblSKILLS develop
make_registry_repo acme-corp/happySKILLS

# ==========================================================================
section "promote.sh: usage"
# ==========================================================================
OUT=$("$SH" "$PROMOTE_SH" --help 2>&1); RC=$?
expect_exit 0 "--help exits 0" "$RC"
expect_contains "--help documents both targets" "--to happyskills" "$OUT"

OUT=$("$SH" "$PROMOTE_SH" 2>&1); RC=$?
expect_exit 2 "no arguments is a usage error" "$RC"

OUT=$("$SH" "$PROMOTE_SH" anything 2>&1); RC=$?
expect_exit 2 "missing --to is a usage error" "$RC"
expect_contains "missing --to names both registries" "humblskills (public) or happyskills" "$OUT"

OUT=$("$SH" "$PROMOTE_SH" anything --to nowhere 2>&1); RC=$?
expect_exit 2 "unknown target rejected" "$RC"

OUT=$("$SH" "$PROMOTE_SH" no-such-skill --to humblskills --check 2>&1); RC=$?
expect_exit 2 "unknown skill name rejected" "$RC"
# shellcheck disable=SC2088 # the output abbreviates HOME as a literal ~
expect_contains "lookup starts at the canonical store" "~/.humblskills/skills/no-such-skill" "$OUT"

S=$(fresh smart-usage)
OUT=$("$SH" "$PROMOTE_SH" "$S" --to happyskills --role none --check 2>&1); RC=$?
expect_exit 2 "happySKILLS refuses --role none" "$RC"
OUT=$("$SH" "$PROMOTE_SH" "$S" --to humblskills --category tools --check 2>&1); RC=$?
expect_exit 2 "--category outside the closed set rejected" "$RC"
expect_contains "--category error lists the allowed values" "development, design, writing, meta" "$OUT"
OUT=$("$SH" "$PROMOTE_SH" "$S" --to humblskills --role ceo --check 2>&1); RC=$?
expect_exit 2 "--role outside the closed set rejected" "$RC"

# ==========================================================================
section "scaffold.sh: create flow"
# ==========================================================================
OUT=$(cd "$T" && "$SH" "$SCAFFOLD_SH" smart-new-thing 2>&1); RC=$?
expect_exit 0 "scaffold.sh exits 0" "$RC"
NEW="$HOME/.cursor/skills/smart-new-thing"
if [[ -x "$NEW/scripts/lint.sh" ]]; then pass "scaffold ships an executable scripts/lint.sh"; else fail "scaffold ships an executable scripts/lint.sh"; fi
FM=$(frontmatter_of "$NEW")
expect_contains "scaffold leaves a role placeholder" "role: TODO" "$FM"
expect_contains "scaffold preserves references/raw/" "- references/raw/" "$FM"
expect_contains "scaffold preserves references/wiki/" "- references/wiki/" "$FM"
expect_not_contains "decisions.md does not inherit smart-skill's history" "### 2026-" "$(cat "$NEW/references/decisions.md")"
expect_contains "next steps end with the promote question" "humblSKILLS (public)" "$OUT"

gate smart-new-thing --to humblskills
expect_exit 1 "a fresh scaffold fails the gate" "$RC"
expect_contains "gate flags the category placeholder" "category: still the scaffold TODO" "$OUT"
expect_contains "gate flags the role placeholder" "role: still the scaffold TODO" "$OUT"
expect_contains "gate flags the tags placeholder" "tags: still the scaffold TODO" "$OUT"
expect_contains "gate flags leftover TODO blocks" "scaffold TODO line(s) left in SKILL.md" "$OUT"
expect_contains "gate flags the author placeholder" "metadata.author is still the scaffold TODO" "$OUT"

# ==========================================================================
section "promote.sh: taxonomy write-back"
# ==========================================================================
OUT=$("$SH" "$PROMOTE_SH" smart-new-thing --to humblskills --check \
  --category development --role fde --tags "Code Review, pull_requests ,git,#git" 2>&1)
FM=$(frontmatter_of "$NEW")
expect_contains "category written" "category: development" "$FM"
expect_contains "role written" "role: fde" "$FM"
expect_contains "tags normalized, deduped, and written" "tags: [code-review, pull-requests, git]" "$FM"
expect_not_contains "scaffold comment dropped with the placeholder" "# required, one of" "$FM"
expect_contains "untouched keys survive" "platforms: [claude-code, cursor, codex]" "$FM"
expect_contains "write-back is reported" "set category: TODO -> development" "$OUT"

OUT=$("$SH" "$PROMOTE_SH" smart-new-thing --to humblskills --check --category development 2>&1)
expect_contains "unchanged taxonomy reported as already set" "already set" "$OUT"

OUT=$("$SH" "$PROMOTE_SH" smart-new-thing --to humblskills --check --tags security 2>&1)
expect_contains "--tags adds to existing tags" "tags: [code-review, pull-requests, git, security]" "$(frontmatter_of "$NEW")"

OUT=$("$SH" "$PROMOTE_SH" smart-new-thing --to humblskills --check --role none 2>&1)
expect_not_contains "--role none removes the role" "role:" "$(frontmatter_of "$NEW")"

S=$(fresh smart-no-meta)
python3 - "$S/SKILL.md" <<'EOF'
import re, sys
p = sys.argv[1]
s = open(p).read()
s = re.sub(r"metadata:\n(?:  .*\n)+", "", s, count=1)
open(p, "w").write(s)
EOF
OUT=$("$SH" "$PROMOTE_SH" "$S" --to humblskills --check --category meta --role ds --tags a-b 2>&1)
FM=$(frontmatter_of "$S")
expect_contains "missing metadata block is created" "metadata:" "$FM"
expect_contains "category lands in the new metadata block" "  category: meta" "$FM"

# ==========================================================================
section "promote.sh: quality gate"
# ==========================================================================
S=$(fresh smart-gate)
gate "$S" --to humblskills
expect_exit 0 "a finished skill passes for humblSKILLS" "$RC"
expect_contains "unscoped role is a warning on humblSKILLS" "role: unscoped" "$OUT"
expect_contains "gate reports where the CLI files it" "humblSKILLS › development › smart-gate" "$OUT"

gate "$S" --to happyskills
expect_exit 1 "happySKILLS requires a role" "$RC"
expect_contains "role failure explains why" "happySKILLS organizes skills by role" "$OUT"
gate "$S" --to happyskills --role sdr
expect_exit 0 "with a role it passes for happySKILLS" "$RC"
expect_contains "placement includes the role" "happySKILLS › development › sdr › smart-gate" "$OUT"

S=$(fresh smart-gate)
set_desc "$S" "Reviews pull request diffs for risky changes."
gate "$S" --to humblskills
expect_exit 1 "description without a trigger clause fails" "$RC"
expect_contains "trigger failure says what to add" "no trigger clause" "$OUT"

S=$(fresh smart-gate)
set_desc "$S" "\"$(printf 'Use when reviewing. %.0s' $(seq 1 60))\""
gate "$S" --to humblskills
expect_exit 1 "description over 1024 characters fails" "$RC"

S=$(fresh smart-gate)
set_desc "$S" "\"Maps a diff -> a checklist. Use when reviewing.\""
gate "$S" --to humblskills
expect_exit 1 "angle bracket in the description fails" "$RC"
expect_contains "angle bracket failure names the field" "'<' or '>' in description" "$OUT"

S=$(fresh smart-gate)
set_desc "$S" "Use when: the user asks"
gate "$S" --to humblskills
expect_exit 1 "unquoted ': ' in a value fails like the CLI's YAML parser" "$RC"
expect_contains "parse failure suggests quoting" "wrap it in quotes" "$OUT"

S=$(fresh smart-gate)
mv "$S" "$T/skills/smart-renamed"
gate "$T/skills/smart-renamed" --to humblskills
expect_exit 1 "name that differs from its directory fails" "$RC"

S=$(fresh claude-helper)
gate "$S" --to humblskills
expect_exit 1 "reserved claude- prefix fails" "$RC"

S=$(fresh smart-gate)
sed -i.bak 's/tags: \[code-review, git, pull-requests\]/tags: [Code Review]/' "$S/SKILL.md" && rm -f "$S/SKILL.md.bak"
gate "$S" --to humblskills
expect_exit 1 "non-kebab-case tag in SKILL.md fails" "$RC"

S=$(fresh smart-gate)
sed -i.bak 's/category: development/category: tools/' "$S/SKILL.md" && rm -f "$S/SKILL.md.bak"
gate "$S" --to humblskills
expect_exit 1 "category outside the closed set fails" "$RC"

S=$(fresh smart-gate)
rm "$S/scripts/lint.sh"
gate "$S" --to humblskills
expect_exit 1 "missing scripts/lint.sh fails" "$RC"
expect_contains "lint failure says how to fix it" "cp <smart-skill>/scripts/lint.sh" "$OUT"

S=$(fresh smart-gate)
sed -i.bak '/    - references\/wiki\//d' "$S/SKILL.md" && rm -f "$S/SKILL.md.bak"
gate "$S" --to humblskills
expect_exit 1 "unpreserved brain path fails" "$RC"
expect_contains "preserve failure names the path" "metadata.preserve must list references/wiki/" "$OUT"

S=$(fresh smart-gate)
printf '%s\n' "$(cat "$S/SKILL.md")" "" "<!-- TODO:START - fill me -->" >"$S/SKILL.md"
gate "$S" --to humblskills
expect_exit 1 "leftover scaffold marker fails" "$RC"

S=$(fresh smart-gate)
for i in $(seq 1 520); do echo "line $i"; done >>"$S/SKILL.md"
gate "$S" --to humblskills
expect_exit 1 "SKILL.md over 500 lines fails" "$RC"

S=$(fresh smart-gate)
printf 'token: ghp_%s\n' "abcdefghijklmnopqrstuvwxyz0123456789" >"$S/references/raw/notes.md"
gate "$S" --to humblskills
expect_exit 1 "GitHub token in a file fails" "$RC"
expect_contains "credential failure points at the file" "references/raw/notes.md:1 (GitHub token)" "$OUT"
expect_not_contains "credential value is never echoed" "abcdefghijklmnopqrstuvwxyz0123456789" "$OUT"

S=$(fresh smart-gate)
printf 'export GITHUB_TOKEN=ghp_%s\n' "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" >"$S/references/raw/notes.md"
gate "$S" --to humblskills
expect_exit 0 "placeholder token does not fail" "$RC"

S=$(fresh smart-gate)
printf 'header only: -----BEGIN RSA PRIVATE KEY-----\n' >"$S/references/raw/notes.md"
gate "$S" --to humblskills
expect_exit 0 "a PEM header with no key material does not fail" "$RC"

S=$(fresh smart-gate)
printf 'API_KEY=1\n' >"$S/.env"
gate "$S" --to humblskills
expect_exit 1 ".env file fails" "$RC"

S=$(fresh smart-gate)
ln -s ../SKILL.md "$S/references/link.md"
gate "$S" --to humblskills
expect_exit 1 "symlink fails" "$RC"

S=$(fresh smart-gate)
mkdir -p "$S/scripts/node_modules/x" && echo "{}" >"$S/scripts/node_modules/x/package.json"
gate "$S" --to humblskills
expect_exit 1 "node_modules fails" "$RC"

S=$(fresh smart-gate)
dd if=/dev/zero of="$S/references/raw/big.bin" bs=1048576 count=6 >/dev/null 2>&1
gate "$S" --to humblskills
expect_exit 1 "file over 5 MiB fails" "$RC"

S=$(fresh smart-gate)
echo "Learned from the HappyRobot dispatch team." >"$S/references/raw/notes.md"
gate "$S" --to humblskills
expect_exit 1 "HappyRobot mention fails for the public registry" "$RC"
expect_contains "internal-mention failure points at happySKILLS" "--to happyskills" "$OUT"
gate "$S" --to humblskills --allow-internal
expect_exit 0 "--allow-internal turns it into a warning" "$RC"
gate "$S" --to happyskills --role fde
expect_exit 0 "HappyRobot mention is fine for happySKILLS" "$RC"

S=$(fresh smart-gate)
echo "contact jane@customer-co.com" >"$S/references/raw/notes.md"
gate "$S" --to humblskills
expect_exit 0 "email address only warns" "$RC"
expect_contains "email warning names the file" "email address at references/raw/notes.md:1" "$OUT"

S=$(fresh smart-gate)
touch "$S/.DS_Store" "$S/references/.DS_Store"
gate "$S" --to humblskills
expect_exit 0 "junk files do not fail" "$RC"
expect_contains "junk files are reported as left out" "left out of the PR: .DS_Store" "$OUT"

S=$(fresh smart-gate)
printf -- '---\ntitle: broken\n---\n' >"$S/references/wiki/review/risk/broken.md"
gate "$S" --to humblskills
expect_exit 1 "lint.sh failure fails the gate" "$RC"
expect_contains "lint failure carries lint's finding" "missing frontmatter fields" "$OUT"

# ==========================================================================
section "promote.sh: dry run"
# ==========================================================================
reset_gh
S=$(fresh smart-review)
touch "$S/.DS_Store"
echo "stale" >>"$S/references/_index.md"
OUT=$("$SH" "$PROMOTE_SH" "$S" --to humblskills --dry-run --no-build-registry 2>&1); RC=$?
expect_exit 0 "dry run exits 0" "$RC"
expect_contains "dry run targets develop" "base develop" "$OUT"
expect_contains "dry run commit is a conventional feat" "feat(skills): add smart-review skill" "$OUT"
WT=$(printf '%s\n' "$OUT" | sed -n 's/^ *worktree *//p')
if [[ -d "$WT/skills/smart-review" ]]; then pass "dry run keeps the worktree for inspection"; else fail "dry run keeps the worktree for inspection"; fi
if [[ ! -e "$WT/skills/smart-review/.DS_Store" ]]; then pass "junk is not staged"; else fail "junk is not staged"; fi
expect_not_contains "staged index is regenerated by lint" "stale" "$(cat "$WT/skills/smart-review/references/_index.md")"
if git --git-dir="$T/github/jjfantini/humblSKILLS.git" rev-parse --verify --quiet refs/heads/feat/add-smart-review >/dev/null; then
  fail "dry run pushes nothing"
else
  pass "dry run pushes nothing"
fi
expect_not_contains "dry run opens no PR" "pr create" "$(cat "$GH_STATE/calls" 2>/dev/null)"

# ==========================================================================
section "promote.sh: open a PR (push access)"
# ==========================================================================
reset_gh
UP="$T/github/jjfantini/humblSKILLS.git"
OUT=$("$SH" "$PROMOTE_SH" "$S" --to humblskills --no-build-registry 2>&1); RC=$?
expect_exit 0 "promote exits 0" "$RC"
expect_eq "branch is pushed with a feat commit" "feat(skills): add smart-review skill" \
  "$(git --git-dir="$UP" log -1 --format=%s refs/heads/feat/add-smart-review 2>/dev/null)"
if git --git-dir="$UP" merge-base --is-ancestor refs/heads/develop refs/heads/feat/add-smart-review 2>/dev/null; then
  pass "branch is cut from develop"
else
  fail "branch is cut from develop"
fi
FILES=$(git --git-dir="$UP" ls-tree -r --name-only refs/heads/feat/add-smart-review -- skills/smart-review)
expect_contains "skill lands under skills/<name>" "skills/smart-review/SKILL.md" "$FILES"
expect_not_contains "junk stays out of the commit" ".DS_Store" "$FILES"
if git --git-dir="$UP" diff --quiet refs/heads/develop refs/heads/feat/add-smart-review -- registry.json; then
  pass "registry.json is left to the Registry workflow"
else
  fail "registry.json is left to the Registry workflow"
fi
expect_eq "commit is authored by the user's git identity" "Test Author" \
  "$(git --git-dir="$UP" log -1 --format=%an refs/heads/feat/add-smart-review)"
CALLS=$(cat "$GH_STATE/calls")
expect_contains "PR targets develop from the branch" "pr create --repo jjfantini/humblSKILLS --base develop --head feat/add-smart-review" "$CALLS"
expect_contains "PR title is the feat subject" "--title feat(skills): add smart-review skill" "$CALLS"
BODY=$(cat "$GH_STATE/pr-body.md")
expect_contains "PR body shows where the CLI files it" "## Where it lands in the CLI" "$BODY"
expect_contains "PR body carries the category" "| Category | \`development\`" "$BODY"
expect_contains "PR body carries the tags" "\`code-review\`, \`git\`, \`pull-requests\`" "$BODY"
expect_contains "PR body has the reviewer checklist" "## Reviewer checklist" "$BODY"
expect_contains "PR body reminds to merge with a merge commit" "gh pr merge --merge" "$BODY"
expect_contains "promote prints the PR URL" "opened https://github.test/pull/1" "$OUT"

OUT=$("$SH" "$PROMOTE_SH" "$S" --to humblskills --no-build-registry 2>&1); RC=$?
expect_exit 0 "re-run with no changes exits 0" "$RC"
expect_contains "re-run with no changes adds no commit" "already carries this exact copy" "$OUT"
expect_eq "branch still has exactly one promote commit" "1" \
  "$(git --git-dir="$UP" rev-list --count refs/heads/develop..refs/heads/feat/add-smart-review)"

echo "- a second hotspot" >>"$S/references/wiki/review/risk/hotspots.md"
OUT=$("$SH" "$PROMOTE_SH" "$S" --to humblskills --no-build-registry 2>&1); RC=$?
expect_exit 0 "re-run after a local edit exits 0" "$RC"
expect_eq "follow-up commit syncs without a second feat" "chore(skills): sync smart-review from its local copy" \
  "$(git --git-dir="$UP" log -1 --format=%s refs/heads/feat/add-smart-review)"
expect_contains "open PR is updated, not duplicated" "updated https://github.test/pull/1" "$OUT"
expect_contains "update leaves a comment for the reviewer" "pr comment https://github.test/pull/1" "$(cat "$GH_STATE/calls")"
expect_eq "only one PR was ever created" "1" "$(grep -c 'pr create' "$GH_STATE/calls")"

# ==========================================================================
section "promote.sh: registry rules"
# ==========================================================================
reset_gh
S=$(fresh existing-skill)
OUT=$("$SH" "$PROMOTE_SH" "$S" --to humblskills --no-build-registry 2>&1); RC=$?
expect_exit 1 "a published skill is refused" "$RC"
expect_contains "refusal explains net-new only" "only adds net-new skills" "$OUT"

S=$(fresh old-name)
OUT=$("$SH" "$PROMOTE_SH" "$S" --to humblskills --no-build-registry 2>&1); RC=$?
expect_exit 1 "a former name of a published skill is refused" "$RC"
expect_contains "former-name refusal names the owner" "former name of existing-skill" "$OUT"

S=$(fresh smart-bin)
mkdir -p "$S/bin" && echo "#!/bin/sh" >"$S/bin/tool"
OUT=$("$SH" "$PROMOTE_SH" "$S" --to humblskills --no-build-registry 2>&1); RC=$?
expect_exit 1 "files the target's .gitignore drops are refused" "$RC"
expect_contains "ignored-file failure names the file" "bin/tool" "$OUT"
expect_not_contains "nothing pushed when staging fails" "feat/add-smart-bin" \
  "$(git --git-dir="$UP" for-each-ref --format='%(refname)' refs/heads)"

# ==========================================================================
section "promote.sh: fork when there is no push access"
# ==========================================================================
reset_gh
export GH_PERMISSION=READ
S=$(fresh smart-forked)
OUT=$("$SH" "$PROMOTE_SH" "$S" --to humblskills --no-build-registry 2>&1); RC=$?
expect_exit 0 "fork promote exits 0" "$RC"
FORK="$T/github/forkuser/humblSKILLS.git"
expect_eq "branch lands on the fork" "feat(skills): add smart-forked skill" \
  "$(git --git-dir="$FORK" log -1 --format=%s refs/heads/feat/add-smart-forked 2>/dev/null)"
if git --git-dir="$UP" rev-parse --verify --quiet refs/heads/feat/add-smart-forked >/dev/null; then
  fail "nothing is pushed to the upstream repo"
else
  pass "nothing is pushed to the upstream repo"
fi
expect_contains "PR head is owner-qualified" "--head forkuser:feat/add-smart-forked" "$(cat "$GH_STATE/calls")"
unset GH_PERMISSION

# ==========================================================================
section "promote.sh: happySKILLS"
# ==========================================================================
reset_gh
mkdir -p "$HOME/.humblskills"
cat >"$HOME/.humblskills/profile.json" <<'EOF'
{"schema_version": 1, "registries": [
  {"name": "public", "url": "https://raw.githubusercontent.com/jjfantini/humblSKILLS/main/registry.json"},
  {"name": "happyrobot", "url": "https://raw.githubusercontent.com/acme-corp/happySKILLS/main/registry.json"}
]}
EOF
mkdir -p "$HOME/.humblskills/skills"
make_skill "$HOME/.humblskills/skills" smart-load-match >/dev/null
OUT=$("$SH" "$PROMOTE_SH" smart-load-match --to happyskills --role fde --no-build-registry 2>&1); RC=$?
expect_exit 0 "happySKILLS promote exits 0" "$RC"
expect_contains "skill found by name in the canonical store" "source  ~/.humblskills/skills/smart-load-match" "$OUT"
expect_contains "repo discovered from the humblskills profile" "acme-corp/happySKILLS (registry in ~/.humblskills/profile.json)" "$OUT"
expect_contains "base falls back to the default branch without develop" "base main" "$OUT"
expect_eq "branch is pushed to happySKILLS" "feat(skills): add smart-load-match skill" \
  "$(git --git-dir="$T/github/acme-corp/happySKILLS.git" log -1 --format=%s refs/heads/feat/add-smart-load-match 2>/dev/null)"
expect_contains "happySKILLS PR carries the role" "| Role | \`fde\` (forward-deployed engineer)" "$(cat "$GH_STATE/pr-body.md")"
expect_contains "role is written back to the local skill" "role: fde" "$(frontmatter_of "$HOME/.humblskills/skills/smart-load-match")"

# ==========================================================================
section "promote.sh: --checkout leaves your clone alone"
# ==========================================================================
reset_gh
git clone -q "$UP" "$T/myclone"
(cd "$T/myclone" && git checkout -q develop && echo dirty >>registry.json && echo scratch >scratch.txt)
BEFORE=$(cd "$T/myclone" && git status --porcelain && git rev-parse --abbrev-ref HEAD)
S=$(fresh smart-local-clone)
OUT=$("$SH" "$PROMOTE_SH" "$S" --to humblskills --checkout "$T/myclone" --no-build-registry 2>&1); RC=$?
expect_exit 0 "--checkout promote exits 0" "$RC"
AFTER=$(cd "$T/myclone" && git status --porcelain && git rev-parse --abbrev-ref HEAD)
expect_eq "working tree and branch untouched" "$BEFORE" "$AFTER"
expect_eq "no worktree left behind" "1" "$(git -C "$T/myclone" worktree list | wc -l | tr -d ' ')"
expect_eq "no private refs left behind" "" "$(git -C "$T/myclone" for-each-ref refs/promote)"
expect_eq "remotes untouched" "origin" "$(git -C "$T/myclone" remote | tr '\n' ' ' | sed 's/ $//')"

# ==========================================================================
section "calibration against this repo's skills"
# ==========================================================================
if [[ -f "$REPO_ROOT/registry.json" && -d "$REPO_ROOT/skills" ]]; then
  OUT=$(python3 - "$ENGINE" "$REPO_ROOT" <<'EOF'
import importlib.util, json, sys
from pathlib import Path
spec = importlib.util.spec_from_file_location("promote", sys.argv[1])
P = importlib.util.module_from_spec(spec)
spec.loader.exec_module(P)
root = Path(sys.argv[2])
reg = {s["name"]: s for s in json.loads((root / "registry.json").read_text())["skills"]}
mismatch = []
for d in sorted((root / "skills").iterdir()):
    fm, _, _ = P.read_skill_md(d / "SKILL.md")
    meta = fm.get("metadata") or {}
    r = reg.get(fm["name"])
    if r is None:
        continue
    pairs = [("description", fm["description"], r["description"]),
             ("version", str(meta.get("version")), r["version"]),
             ("category", meta.get("category"), r.get("category")),
             ("tags", P.as_list(meta.get("tags")) or None, r.get("tags")),
             ("preserve", P.as_list(meta.get("preserve")) or None, r.get("preserve")),
             ("previous_names", P.as_list(meta.get("previous_names")) or None, r.get("previous_names"))]
    mismatch += [f"{d.name}.{k}" for k, a, b in pairs if a != b]
print("parse-mismatches=" + ",".join(mismatch))
secrets = []
for d in sorted((root / "skills").iterdir()):
    rep = P.Report()
    import io, contextlib
    with contextlib.redirect_stdout(io.StringIO()):
        P.file_checks(d, "happyskills", rep, False)
    secrets += [r["msg"] for r in rep.data["results"] if r["check"] == "secrets"]
print("secret-findings=" + str(len(secrets)))
EOF
)
  expect_eq "frontmatter parser matches registry.json for every skill" "parse-mismatches=" "$(printf '%s\n' "$OUT" | grep parse-mismatches)"
  expect_eq "credential scanner is silent on every published skill" "secret-findings=0" "$(printf '%s\n' "$OUT" | grep secret-findings)"
else
  printf '  (skipped: not inside the humblSKILLS repo)\n'
fi

# ==========================================================================
printf "\n${BOLD}Summary:${RST}\n"
printf "  passed: ${GRN}%d${RST}\n" "$PASS"
printf "  failed: ${RED}%d${RST}\n" "$FAIL"
if [[ "$FAIL" -gt 0 ]]; then
  printf "\n${RED}FAILED CASES:${RST}\n"
  for c in "${FAILED[@]}"; do printf "  - %s\n" "$c"; done
  exit 1
fi
printf "\n${GRN}OK: all tests passed.${RST}\n"
exit 0
