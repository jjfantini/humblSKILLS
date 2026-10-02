#!/usr/bin/env bash
# promote.sh - open a pull request that adds a net-new smart skill to
# humblSKILLS (public) or happySKILLS (HappyRobot-internal).
#
# Flow: find the local skill -> write category/role/tags into its SKILL.md ->
# quality gate -> clone the target (fork it when you can't push) -> worktree
# off the base branch -> net-new check -> stage a clean copy -> lint ->
# build-registry validation -> commit -> push -> open the PR.
# Re-running for the same skill pushes a follow-up commit to the same branch
# and comments on the open PR, so review feedback is fixed locally and re-run.
#
# Net-new only: a skill the target already publishes is refused. Merging
# learnings back into an existing skill is a separate path that isn't built.
#
# Exit codes: 0 ok, 1 gate/git/GitHub failure, 2 usage error.
# Needs bash 3.2+ (the macOS default), git, and python3; gh to push and open
# the PR. --check needs only python3.
#
# Test seams (tests/run.sh): HUMBLSKILLS_PROMOTE_HOME relocates the clone
# cache; HUMBLSKILLS_PROMOTE_GIT_BASE replaces https://github.com in clone URLs.

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENGINE="$SCRIPT_DIR/lib/promote.py"

if [[ -t 1 && -z "${NO_COLOR:-}" ]]; then
  BOLD=$'\033[1m' RED=$'\033[31m' GRN=$'\033[32m' DIM=$'\033[2m' RST=$'\033[0m'
else
  BOLD="" RED="" GRN="" DIM="" RST=""
fi

usage() {
  cat <<'EOF'
Usage: promote.sh <skill> --to humblskills|happyskills [options]

Open a pull request that adds a net-new smart skill from your machine to a
skill registry, for the maintainer to review.

  <skill>    a skill name - looked up in ~/.humblskills/skills first, then each
             agent's skills dir - or the path to a skill directory

Target (required):
  --to humblskills   public registry jjfantini/humblSKILLS (PR into develop)
  --to happyskills   HappyRobot-internal registry happySKILLS (repo taken from
                     your `humblskills registry` config when it lists one)

Taxonomy - written into the skill's SKILL.md, then validated:
  --category <c>     development | design | writing | meta
  --role <r>         fde | ds | sdr, or none for a general-purpose skill
                     (happySKILLS needs a real role)
  --tags <a,b,c>     kebab-case keywords; added to the existing tags

Modes:
  --check            quality gate only - no git, no network
  --dry-run          build the commit in a throwaway worktree; push nothing

Options:
  --repo <owner/name>  target repository (overrides the default/discovered one)
  --base <branch>      base branch (default: develop when the repo has one)
  --branch <name>      PR branch (default: feat/add-<skill>)
  --checkout <dir>     use your existing clone of the target instead of a cache
  --draft              open the PR as a draft
  --allow-internal     allow HappyRobot mentions in a public promotion
  --no-build-registry  skip build-registry validation (it needs Go 1.23+)
  -h, --help           show this help

Examples:
  promote.sh smart-pr-review --to humblskills --check \
    --category development --role none --tags git,code-review,pull-requests
  promote.sh smart-load-matching --to happyskills \
    --category development --role fde --tags freight,loads,matching
  promote.sh ./my-skill --to humblskills --dry-run
EOF
}

say() { printf '%s\n' "$*"; }
ok() { printf '  %s✓%s %s\n' "$GRN" "$RST" "$*"; }
note() { printf '  %s%s%s\n' "$DIM" "$*" "$RST"; }
heading() { printf '\n%s%s%s\n' "$BOLD" "$*" "$RST"; }
die() {
  local code="$1"
  shift
  printf '%spromote: %s%s\n' "$RED" "$*" "$RST" >&2
  exit "$code"
}
engine() { python3 "$ENGINE" "$@"; }
lower() { printf '%s' "$1" | tr '[:upper:]' '[:lower:]'; }
HOME_PHYS=$(cd "$HOME" 2>/dev/null && pwd -P || printf '%s' "$HOME")
tildify() {
  case "$1" in
    "$HOME"/*) printf '~%s' "${1#"$HOME"}" ;;
    "$HOME_PHYS"/*) printf '~%s' "${1#"$HOME_PHYS"}" ;;
    *) printf '%s' "$1" ;;
  esac
}

# owner/name from any git URL form: https, ssh (git@host:owner/name), file path.
repo_of_url() {
  local u="${1%/}"
  u="${u%.git}"
  u="${u//://}"
  local name="${u##*/}"
  u="${u%/*}"
  printf '%s/%s\n' "${u##*/}" "$name"
}

promote_home() {
  if [[ -n "${HUMBLSKILLS_PROMOTE_HOME:-}" ]]; then
    printf '%s\n' "$HUMBLSKILLS_PROMOTE_HOME"
  elif [[ -n "${XDG_CACHE_HOME:-}" ]]; then
    printf '%s\n' "$XDG_CACHE_HOME/humblskills/promote"
  elif [[ "$(uname -s)" == Darwin ]]; then
    printf '%s\n' "$HOME/Library/Caches/humblskills/promote"
  else
    printf '%s\n' "$HOME/.cache/humblskills/promote"
  fi
}

# ---------------------------------------------------------------- arguments
args=()
for a in "$@"; do
  case "$a" in
    --*=*) args+=("${a%%=*}" "${a#*=}") ;;
    *) args+=("$a") ;;
  esac
done
set -- ${args[@]+"${args[@]}"}

SKILL_ARG="" TO="" CATEGORY="" ROLE="" TAGS="" MODE="promote"
REPO_FLAG="" BASE_FLAG="" BRANCH="" CHECKOUT=""
DRAFT=0 ALLOW_INTERNAL=0 BUILD_REGISTRY=1
HAVE_CATEGORY=0 HAVE_ROLE=0 HAVE_TAGS=0

value_of() {
  [[ $# -ge 2 && -n "$2" && "$2" != --* ]] || die 2 "$1 needs a value (see --help)"
  printf '%s' "$2"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h | --help) usage; exit 0 ;;
    --to) TO=$(value_of "$@") || exit 2; shift 2 ;;
    --category) CATEGORY=$(value_of "$@") || exit 2; HAVE_CATEGORY=1; shift 2 ;;
    --role) ROLE=$(value_of "$@") || exit 2; HAVE_ROLE=1; shift 2 ;;
    --tags) TAGS=$(value_of "$@") || exit 2; HAVE_TAGS=1; shift 2 ;;
    --repo) REPO_FLAG=$(value_of "$@") || exit 2; shift 2 ;;
    --base) BASE_FLAG=$(value_of "$@") || exit 2; shift 2 ;;
    --branch) BRANCH=$(value_of "$@") || exit 2; shift 2 ;;
    --checkout) CHECKOUT=$(value_of "$@") || exit 2; shift 2 ;;
    --check) MODE="check"; shift ;;
    --dry-run) MODE="dry-run"; shift ;;
    --draft) DRAFT=1; shift ;;
    --allow-internal) ALLOW_INTERNAL=1; shift ;;
    --no-build-registry) BUILD_REGISTRY=0; shift ;;
    -*) die 2 "unknown option $1 (see --help)" ;;
    *)
      [[ -z "$SKILL_ARG" ]] || die 2 "one skill per run (got $SKILL_ARG and $1)"
      SKILL_ARG="$1"
      shift
      ;;
  esac
done

if [[ -z "$SKILL_ARG" ]]; then
  usage >&2
  exit 2
fi
[[ -n "$TO" ]] || die 2 "--to is required: humblskills (public) or happyskills (HappyRobot-internal)"
command -v python3 >/dev/null 2>&1 || die 2 "python3 is required"
[[ -f "$ENGINE" ]] || die 2 "missing $ENGINE - reinstall smart-skill"

# ------------------------------------------------------------------- target
target_args=(target-info --to "$TO")
[[ -n "$REPO_FLAG" ]] && target_args+=(--repo "$REPO_FLAG")
target_info=$(engine "${target_args[@]}") || exit 2
TARGET_KEY="" LABEL="" SLUG="" SLUG_SOURCE="" DEFAULT_BASE=""
while IFS='=' read -r key value; do
  case "$key" in
    key) TARGET_KEY="$value" ;;
    label) LABEL="$value" ;;
    repo) SLUG="$value" ;;
    repo_source) SLUG_SOURCE="$value" ;;
    base) DEFAULT_BASE="$value" ;;
  esac
done <<<"$target_info"

if [[ "$TARGET_KEY" == happyskills && "$HAVE_ROLE" -eq 1 && "$ROLE" == none ]]; then
  die 2 "$LABEL organizes skills by role - pass --role fde, ds, or sdr"
fi

SKILL_DIR=$(engine resolve "$SKILL_ARG") || exit 2
NAME="$(basename "$SKILL_DIR")"
[[ -n "$BRANCH" ]] || BRANCH="feat/add-$NAME"

WORK=$(mktemp -d "${TMPDIR:-/tmp}/promote.XXXXXX") || die 1 "could not create a temp dir"
REPORT="$WORK/report.json"
KEEP=0 CLONE="" WT=""
cleanup() {
  if [[ -n "$CLONE" ]]; then
    git -C "$CLONE" update-ref -d "refs/promote/$BRANCH" >/dev/null 2>&1 || true
  fi
  if [[ "$KEEP" -eq 0 ]]; then
    if [[ -n "$WT" && -n "$CLONE" ]]; then
      git -C "$CLONE" worktree remove --force "$WT" >/dev/null 2>&1 || true
    fi
    rm -rf "$WORK"
  fi
}
trap cleanup EXIT

printf '%spromote%s %s -> %s %s(%s)%s\n' "$BOLD" "$RST" "$NAME" "$LABEL" "$DIM" "$SLUG" "$RST"
note "source  $(tildify "$SKILL_DIR")"
note "repo    $SLUG ($SLUG_SOURCE)"
note "mode    $MODE"

# ----------------------------------------------------------------- taxonomy
if [[ $((HAVE_CATEGORY + HAVE_ROLE + HAVE_TAGS)) -gt 0 ]]; then
  heading "Taxonomy -> SKILL.md"
  tax=(taxonomy "$SKILL_DIR")
  [[ "$HAVE_CATEGORY" -eq 1 ]] && tax+=(--category "$CATEGORY")
  [[ "$HAVE_ROLE" -eq 1 ]] && tax+=(--role "$ROLE")
  [[ "$HAVE_TAGS" -eq 1 ]] && tax+=(--tags "$TAGS")
  changed=$(engine "${tax[@]}") || exit 2
  if [[ -n "$changed" ]]; then
    printf '%s\n' "$changed"
  else
    note "already set"
  fi
fi

# --------------------------------------------------------------------- gate
say ""
gate=(check "$SKILL_DIR" --to "$TARGET_KEY" --report "$REPORT")
[[ "$ALLOW_INTERNAL" -eq 1 ]] && gate+=(--allow-internal)
[[ "$HAVE_ROLE" -eq 1 ]] && gate+=(--role-decided)
engine "${gate[@]}" || die 1 "fix the failures above, then re-run (the gate runs again before anything is pushed)"

if [[ "$MODE" == check ]]; then
  if [[ -n "$CHECKOUT" ]]; then
    say ""
    engine check "$SKILL_DIR" --to "$TARGET_KEY" --report "$REPORT" --repo-dir "$CHECKOUT" \
      || die 1 "fix the registry findings above, then re-run"
  fi
  say ""
  ok "passes the quality gate for $LABEL"
  note "next: re-run without --check to open the PR (or with --dry-run to preview the commit)"
  exit 0
fi

# ---------------------------------------------------------------- git setup
command -v git >/dev/null 2>&1 || die 2 "git is required"
HAVE_GH=0
command -v gh >/dev/null 2>&1 && HAVE_GH=1
if [[ "$MODE" == promote && "$HAVE_GH" -eq 0 ]]; then
  die 2 "the GitHub CLI is required to push and open the PR: brew install gh && gh auth login (or use --dry-run)"
fi

heading "Repository"
GIT_BASE="${HUMBLSKILLS_PROMOTE_GIT_BASE:-https://github.com}"
if [[ -n "$CHECKOUT" ]]; then
  CLONE=$(cd "$CHECKOUT" 2>/dev/null && git rev-parse --show-toplevel 2>/dev/null) \
    || die 2 "--checkout $CHECKOUT is not a git checkout"
else
  CLONE="$(promote_home)/$SLUG"
  if [[ -d "$CLONE/.git" ]]; then
    :
  else
    mkdir -p "$(dirname "$CLONE")" || die 1 "could not create $(dirname "$CLONE")"
    note "cloning $SLUG into $(tildify "$CLONE")"
    # gh handles private-repo auth and the user's protocol preference; plain
    # git still works for a public repo when gh isn't logged in.
    cloned=0
    if [[ "$HAVE_GH" -eq 1 && -z "${HUMBLSKILLS_PROMOTE_GIT_BASE:-}" ]]; then
      gh repo clone "$SLUG" "$CLONE" -- --quiet >/dev/null 2>&1 && cloned=1
    fi
    if [[ "$cloned" -eq 0 ]]; then
      rm -rf "$CLONE"
      git clone --quiet "$GIT_BASE/$SLUG.git" "$CLONE" 2>/dev/null \
        || die 1 "could not clone $SLUG - check \`gh auth status\` and that you can read the repo"
    fi
  fi
fi

TARGET_REMOTE=""
want="$(lower "$SLUG")"
for remote in $(git -C "$CLONE" remote); do
  url=$(git -C "$CLONE" remote get-url "$remote" 2>/dev/null) || continue
  if [[ "$(lower "$(repo_of_url "$url")")" == "$want" ]]; then
    TARGET_REMOTE="$remote"
    break
  fi
done
[[ -n "$TARGET_REMOTE" ]] || die 2 "$(tildify "$CLONE") has no remote pointing at $SLUG"
git -C "$CLONE" fetch --quiet --prune "$TARGET_REMOTE" || die 1 "could not fetch $SLUG"

BASE="$BASE_FLAG"
if [[ -z "$BASE" ]]; then
  if [[ -n "$DEFAULT_BASE" ]]; then
    BASE="$DEFAULT_BASE"
  elif git -C "$CLONE" rev-parse --verify --quiet "refs/remotes/$TARGET_REMOTE/develop" >/dev/null; then
    BASE="develop"
  else
    BASE=$(git -C "$CLONE" symbolic-ref --quiet --short "refs/remotes/$TARGET_REMOTE/HEAD" 2>/dev/null)
    BASE="${BASE#"$TARGET_REMOTE"/}"
    if [[ -z "$BASE" ]]; then
      BASE=$(git -C "$CLONE" ls-remote --symref "$TARGET_REMOTE" HEAD 2>/dev/null \
        | sed -n 's|^ref: refs/heads/\([^[:space:]]*\)[[:space:]]*HEAD$|\1|p')
    fi
  fi
fi
[[ -n "$BASE" ]] || die 1 "could not tell which branch $SLUG merges into - pass --base"
git -C "$CLONE" rev-parse --verify --quiet "refs/remotes/$TARGET_REMOTE/$BASE" >/dev/null \
  || die 1 "$SLUG has no $BASE branch - pass --base"
ok "$SLUG, base $BASE, branch $BRANCH"

# Where the branch goes: the target itself when you can push to it, otherwise
# your fork. The fork is addressed by URL so your clone's remotes stay as-is.
PUSH_TARGET="$TARGET_REMOTE"
HEAD_OWNER="${SLUG%%/*}"
FORKED=0
# A dry run only previews, so it never forks and never fails on GitHub access.
if [[ "$HAVE_GH" -eq 1 ]]; then
  perm=$(gh repo view "$SLUG" --json viewerPermission --jq .viewerPermission 2>/dev/null || true)
  case "$perm" in
    ADMIN | MAINTAIN | WRITE) ;;
    *)
      login=$(gh api user --jq .login 2>/dev/null || true)
      if [[ -z "$login" ]]; then
        [[ "$MODE" == promote ]] && die 1 "could not read your GitHub login - run gh auth login"
      else
        [[ "$MODE" == promote ]] && { gh repo fork "$SLUG" --clone=false >/dev/null 2>&1 || true; }
        fork_url=""
        for attempt in 1 2 3 4 5; do
          fork_url=$(gh api "repos/$SLUG/forks?per_page=100" --paginate \
            --jq ".[] | select(.owner.login == \"$login\") | .clone_url" 2>/dev/null | head -n 1)
          [[ -n "$fork_url" || "$MODE" != promote ]] && break
          sleep "$attempt"
        done
        if [[ -n "$fork_url" ]]; then
          PUSH_TARGET="$fork_url"
          HEAD_OWNER="$login"
          FORKED=1
          ok "no push access to $SLUG - using your fork $(repo_of_url "$fork_url")"
        elif [[ "$MODE" == promote ]]; then
          die 1 "no push access to $SLUG and no fork could be created - ask the maintainer for access"
        fi
      fi
      ;;
  esac
fi

WT="$WORK/worktree"
git -C "$CLONE" worktree prune >/dev/null 2>&1 || true
git -C "$CLONE" worktree add --quiet --detach "$WT" "refs/remotes/$TARGET_REMOTE/$BASE" >/dev/null 2>"$WORK/worktree.err" \
  || die 1 "could not create a worktree in $(tildify "$CLONE"): $(tail -n 1 "$WORK/worktree.err")"

say ""
engine check "$SKILL_DIR" --to "$TARGET_KEY" --report "$REPORT" --repo-dir "$WT" --base "$BASE" \
  || die 1 "fix the registry findings above, then re-run"

BRANCH_EXISTS=0
if git -C "$CLONE" ls-remote --exit-code --heads "$PUSH_TARGET" "$BRANCH" >/dev/null 2>&1; then
  git -C "$CLONE" fetch --quiet "$PUSH_TARGET" "+refs/heads/$BRANCH:refs/promote/$BRANCH" \
    || die 1 "could not fetch the existing $BRANCH branch"
  git -C "$WT" checkout --quiet --detach "refs/promote/$BRANCH" || die 1 "could not check out $BRANCH"
  BRANCH_EXISTS=1
fi

# -------------------------------------------------------------------- stage
heading "Stage"
DEST="$WT/skills/$NAME"
excluded=$(engine stage "$SKILL_DIR" "$DEST") || die 1 "could not stage $NAME"
ok "copied into skills/$NAME on top of $([[ $BRANCH_EXISTS -eq 1 ]] && echo "$BRANCH" || echo "$BASE")"
[[ -n "$excluded" ]] && note "left out: $(printf '%s' "$excluded" | tr '\n' ' ')"

ignored=$(git -C "$WT" ls-files --others --ignored --exclude-standard -- "skills/$NAME")
if [[ -n "$ignored" ]]; then
  engine record --report "$REPORT" --level fail --check ignored \
    --msg "$SLUG's .gitignore would silently drop: $(printf '%s' "$ignored" | sed "s|^skills/$NAME/||" | tr '\n' ' ')- rename or remove them"
  die 1 "the PR would be missing files the skill ships"
fi

engine lint "$DEST" --report "$REPORT" || die 1 "scripts/lint.sh fails on the staged copy"

# build-registry is what CI and every install trust; running it here turns a
# would-be broken registry into a failure on the author's machine instead.
if [[ "$BUILD_REGISTRY" -eq 1 ]]; then
  if ! command -v go >/dev/null 2>&1; then
    engine record --report "$REPORT" --level info --check build-registry \
      --msg "Go is not installed - skipped; the registry build validates after merge"
  else
    gominor=$(go env GOVERSION 2>/dev/null | sed -n 's/^go1\.\([0-9]*\).*/\1/p')
    if [[ -z "$gominor" || "$gominor" -lt 23 ]]; then
      engine record --report "$REPORT" --level info --check build-registry \
        --msg "Go 1.23+ needed (found $(go env GOVERSION 2>/dev/null)) - skipped"
    else
      note "validating with build-registry (the first run downloads Go modules)"
      if [[ -f "$WT/cli/go.mod" && -d "$WT/cli/cmd/build-registry" ]]; then
        out=$(cd "$WT" && go -C cli run ./cmd/build-registry --skills-dir="$WT/skills" \
          --out="$WORK/registry.json" --ref="$BASE" 2>&1)
      else
        out=$(cd "$WT" && go run github.com/jjfantini/humblSKILLS/cli/v2/cmd/build-registry@latest \
          --skills-dir="$WT/skills" --out="$WORK/registry.json" --repo="github.com/$SLUG" --ref="$BASE" 2>&1)
      fi
      rc=$?
      if [[ "$rc" -eq 0 ]]; then
        placement=$(engine placement "$WORK/registry.json" "$NAME" --label "$LABEL") || exit 1
        engine record --report "$REPORT" --level pass --check build-registry \
          --msg "registry builds with $NAME; the CLI lists it at $placement" --placement "$placement"
      elif printf '%s' "$out" | grep -q -- "$NAME"; then
        engine record --report "$REPORT" --level fail --check build-registry \
          --msg "$(printf '%s' "$out" | grep -- "$NAME" | head -n 3 | tr '\n' ' ')"
        die 1 "build-registry rejects $NAME"
      else
        engine record --report "$REPORT" --level warn --check build-registry \
          --msg "could not validate ($(printf '%s' "$out" | tail -n 1)); the registry build checks again after merge"
      fi
    fi
  fi
fi

# ------------------------------------------------------------------- commit
heading "Commit"
git -C "$WT" add -A -- "skills/$NAME"
COMMIT_MODE="new"
git -C "$WT" cat-file -e "HEAD:skills/$NAME/SKILL.md" 2>/dev/null && COMMIT_MODE="update"
CHANGED=1
if git -C "$WT" diff --cached --quiet; then
  CHANGED=0
  ok "$BRANCH already carries this exact copy of $NAME"
else
  [[ -n "$(git -C "$WT" config user.email)" && -n "$(git -C "$WT" config user.name)" ]] \
    || die 1 "set your git identity first: git config --global user.name \"Your Name\" && git config --global user.email you@example.com"
  engine render --report "$REPORT" --kind commit --mode "$COMMIT_MODE" >"$WORK/commit-msg" || exit 1
  git -C "$WT" commit --quiet -F "$WORK/commit-msg" || die 1 "git commit failed"
  ok "$(head -n 1 "$WORK/commit-msg") ($(git -C "$WT" rev-parse --short HEAD))"
  note "$(git -C "$WT" show --stat --format= HEAD | tail -n 1 | sed 's/^ *//')"
fi
engine render --report "$REPORT" --kind pr --mode "$COMMIT_MODE" --base "$BASE" >"$WORK/pr-body.md" || exit 1
TITLE=$(engine render --report "$REPORT" --kind title --mode new) || exit 1

if [[ "$MODE" == dry-run ]]; then
  KEEP=1
  heading "Dry run - nothing pushed"
  note "worktree  $WT"
  note "PR title  $TITLE"
  note "PR body   $WORK/pr-body.md"
  note "inspect:  git -C $WT show --stat HEAD"
  note "clean up: git -C $CLONE worktree remove --force $WT && rm -rf $WORK"
  exit 0
fi

# --------------------------------------------------------------- push + PR
heading "Pull request"
HEAD_REF="$BRANCH"
[[ "$FORKED" -eq 1 ]] && HEAD_REF="$HEAD_OWNER:$BRANCH"
if [[ "$CHANGED" -eq 1 ]]; then
  pushed=0
  for attempt in 1 2 3; do
    if git -C "$WT" push --quiet "$PUSH_TARGET" "HEAD:refs/heads/$BRANCH" 2>"$WORK/push.err"; then
      pushed=1
      break
    fi
    # A brand-new fork can take a few seconds before it accepts pushes.
    sleep $((attempt * 3))
  done
  [[ "$pushed" -eq 1 ]] \
    || die 1 "push to $BRANCH failed: $(tail -n 1 "$WORK/push.err") (for HTTPS remotes, \`gh auth setup-git\` lets git use your gh login)"
  ok "pushed $BRANCH"
fi

existing=$(gh pr list --repo "$SLUG" --head "$BRANCH" --state open --json url,headRepositoryOwner \
  --jq ".[] | select(.headRepositoryOwner.login == \"$HEAD_OWNER\") | .url" 2>/dev/null | head -n 1)
if [[ -n "$existing" ]]; then
  if [[ "$CHANGED" -eq 1 ]]; then
    summary=$(engine render --report "$REPORT" --kind summary) || exit 1
    gh pr comment "$existing" \
      --body "Updated from the author's local copy by \`promote.sh\` ($(git -C "$WT" rev-parse --short HEAD)): $summary." \
      >/dev/null 2>&1 || true
    ok "updated $existing"
  else
    ok "$existing is already up to date"
  fi
  PR_URL="$existing"
else
  pr_args=(pr create --repo "$SLUG" --base "$BASE" --head "$HEAD_REF" --title "$TITLE" --body-file "$WORK/pr-body.md")
  [[ "$DRAFT" -eq 1 ]] && pr_args+=(--draft)
  PR_URL=$(gh "${pr_args[@]}" 2>"$WORK/pr.err" | tail -n 1)
  [[ -n "$PR_URL" ]] || die 1 "pushed $BRANCH but could not open the PR ($(tail -n 1 "$WORK/pr.err")) - open it from $HEAD_REF into $BASE"
  ok "opened $PR_URL"
fi

say ""
say "${BOLD}Next${RST}"
say "  - Review happens on $PR_URL. Fix feedback in $(tildify "$SKILL_DIR") and re-run this command to update it."
if [[ "$TARGET_KEY" == humblskills ]]; then
  say "  - Once merged into $BASE it ships in the next pre-release, and in the registry when develop reaches main."
else
  say "  - Once merged and its registry is rebuilt, teammates install it from $LABEL."
fi
say "  - After that, \`humblskills install $NAME\` takes over your copy in ~/.humblskills/skills (or an agent's"
say "    skills dir) and keeps its preserved brain files, so \`humblskills update\` manages it from then on."
