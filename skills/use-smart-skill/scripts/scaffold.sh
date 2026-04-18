#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_ROOT="$(dirname "$SCRIPT_DIR")"

usage() {
  cat <<'EOF'
Usage: scaffold.sh <skill-name> [OPTIONS]

Scaffold a new Smart Skill directory (CCCCC + brain).

Arguments:
  skill-name    Name of the skill (lowercase, hyphens, e.g. my-new-skill)

Options:
  --scripts     Create scripts/ directory
  --assets      Create assets/ directory
  --location    personal (default) or project
                  personal: ~/.cursor/skills/
                  project:  .cursor/skills/
  --help        Show this help

Creates:
  <target>/
    SKILL.md                        (with Brain Protocol injected)
    references/
      _template.md  _brain.md
      _index.md  patterns.md  decisions.md  log.md
      wiki/
      raw/.gitkeep
    scripts/   (if --scripts)
    assets/    (if --assets)

Migrating an existing flat skill? See:
  use-smart-skill/references/wiki/smart/migrate/workflow.md

Examples:
  scaffold.sh my-skill
  scaffold.sh my-skill --scripts --assets
  scaffold.sh my-skill --location project --scripts
EOF
  exit 0
}

if [[ $# -lt 1 ]] || [[ "$1" == "--help" ]]; then
  usage
fi

SKILL_NAME="$1"
shift

INCLUDE_SCRIPTS=false
INCLUDE_ASSETS=false
LOCATION="personal"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --scripts) INCLUDE_SCRIPTS=true; shift ;;
    --assets)  INCLUDE_ASSETS=true; shift ;;
    --location)
      LOCATION="${2:-personal}"
      shift 2
      ;;
    --help) usage ;;
    *) echo "Unknown option: $1"; usage ;;
  esac
done

if [[ "$LOCATION" == "personal" ]]; then
  BASE_DIR="$HOME/.cursor/skills"
elif [[ "$LOCATION" == "project" ]]; then
  BASE_DIR=".cursor/skills"
else
  echo "Error: --location must be 'personal' or 'project'"
  exit 1
fi

TARGET="$BASE_DIR/$SKILL_NAME"

create_file() {
  local filepath="$1"
  local content="$2"
  if [[ -f "$filepath" ]]; then
    echo "  skip: $filepath (exists)"
  else
    printf '%s' "$content" > "$filepath"
    echo "  created: $filepath"
  fi
}

copy_template() {
  local src="$1"
  local dest="$2"
  if [[ -f "$dest" ]]; then
    echo "  skip: $dest (exists)"
  elif [[ -f "$src" ]]; then
    cp "$src" "$dest"
    echo "  created: $dest (from template)"
  else
    echo "  warn: template not found at $src, skipping"
  fi
}

echo "Scaffolding Smart Skill: $SKILL_NAME"
echo "  location: $TARGET"
echo ""

mkdir -p "$TARGET/references/wiki"
mkdir -p "$TARGET/references/raw"
echo "  created: $TARGET/references/"
echo "  created: $TARGET/references/wiki/"
echo "  created: $TARGET/references/raw/"

if [[ "$INCLUDE_SCRIPTS" == true ]]; then
  mkdir -p "$TARGET/scripts"
  echo "  created: $TARGET/scripts/"
fi

if [[ "$INCLUDE_ASSETS" == true ]]; then
  mkdir -p "$TARGET/assets"
  echo "  created: $TARGET/assets/"
fi

SKILL_MD_CONTENT='---
name: '"$SKILL_NAME"'
description: >
  TODO: Describe what this skill does and when to use it. Include BOTH
  what the skill does AND the trigger scenarios that should invoke it.
license: MIT
compatibility: "TODO or delete - max 500 chars. Use when skill requires specific env (bash, git, docker, Python 3.14+, network access, specific product). Most skills do not need this."
metadata:
  author: TODO
  version: "1.0.0"
---

# '"$SKILL_NAME"'

<!-- TODO:START - one-sentence skill summary (2-3 lines max) -->
TODO: Brief description of what this skill is and what it does.
<!-- TODO:END -->

## Brain Protocol (read BEFORE creating anything)

1. `references/_index.md`       - what this skill knows (map)
2. `references/patterns.md`     - what worked, with numbers
3. `references/decisions.md`    - past reasoning, don'"'"'t repeat mistakes
4. `references/log.md`          - last 5 session entries
5. Relevant `references/wiki/<context>/<category>/` concepts per task

After completing work, UPDATE the brain:
- New user-provided material lives in `references/raw/` (LLM never renames or edits it)
- Distilled insights -> new/updated `references/wiki/<context>/<category>/<concept>.md`
  - Cite every raw file you used in the `sources:` frontmatter array
- Performance data (if reported) -> `patterns.md`
- Non-obvious decisions -> `decisions.md`
- Session summary (always) -> append to `log.md`
- Run `scripts/lint.sh` to regenerate `_index.md` and verify structure

Every 2 weeks OR when requested: run `scripts/lint.sh` for contradictions,
stale data, orphan wiki pages, orphan raw files, and broken `sources:` paths.

No data, no improvement.

_Full spec (territories, linking, ingest/lint/patterns worked examples, command routing): `references/_brain.md`._

## Brain Operations

| Operation | Driver            | Effect                                                              |
|-----------|-------------------|---------------------------------------------------------------------|
| Ingest    | Agent (no script) | Read new `raw/` file -> write/update wiki concepts + brain meta     |
| Query     | Agent (runtime)   | Read brain first -> synthesize from wiki -> write novel concepts    |
| Lint      | `scripts/lint.sh` | Regenerate `_index.md`; report orphans, contradictions, stale, broken |

## CCCCC Architecture

| Layer        | Role                                  | Location                                          |
|--------------|---------------------------------------|---------------------------------------------------|
| **Core**     | Root structure of the skill           | `SKILL.md`, `references/`, `scripts/`, `assets/`  |
| **Context**  | Top-level taxonomy grouping           | First segment under `references/wiki/`            |
| **Category** | Specific topic within a context       | Second segment under `references/wiki/`           |
| **Concept**  | One atomic idea per file              | Filename stem AND required frontmatter field      |
| **Command**  | Deterministic executable script       | `scripts/<command>.sh\|.py`, linked from wiki     |

## When to Use

<!-- TODO:START - enumerate concrete trigger scenarios for this skill -->
- TODO: trigger scenario 1
- TODO: trigger scenario 2
<!-- TODO:END -->

## How to Use

**Live enumeration of contexts, categories, and concepts:**
Read `references/_index.md` (auto-regenerated by `scripts/lint.sh`).

**Brain protocol, naming conventions, writing principles, linking contract, ingest workflow, lint checks, `patterns.md` entry shape:**
Read `references/_brain.md`.

**Wiki concept file shape:**
Read `references/_template.md`.

<!-- TODO:START - skill-specific routing. Add bullets pointing at the wiki
     concepts most relevant to this skill primary workflows. One bullet
     per distinct entry point; link to a specific concept file. -->

**TODO: primary workflow name:**
Read `references/wiki/<context>/<category>/<concept>.md`.

<!-- TODO:END -->
'

INDEX_MD_CONTENT='# Index

Auto-generated by `scripts/lint.sh` from the filesystem walk of
`references/wiki/`, `references/raw/`, and `scripts/`. Do not hand-edit
below the generated markers. To change the taxonomy, add or remove wiki
files and re-run lint.

<!-- GENERATED:START -->

(run `scripts/lint.sh` to populate)

<!-- GENERATED:END -->
'

create_file "$TARGET/SKILL.md" "$SKILL_MD_CONTENT"

copy_template "$SKILL_ROOT/references/_template.md" "$TARGET/references/_template.md"
copy_template "$SKILL_ROOT/references/_brain.md"    "$TARGET/references/_brain.md"
copy_template "$SKILL_ROOT/references/patterns.md"  "$TARGET/references/patterns.md"
copy_template "$SKILL_ROOT/references/decisions.md" "$TARGET/references/decisions.md"

create_file "$TARGET/references/_index.md" "$INDEX_MD_CONTENT"

TODAY="$(date +%Y-%m-%d)"
LOG_MD_CONTENT='# Log

Append-only session log. Every session MUST append at least one entry.
Never edit old entries - they are the historical record. Most recent
entries appear at the bottom.

Entry shape:

```
[INGEST|QUERY|LINT <YYYY-MM-DD>] <one-line summary>
  <optional indented detail line(s)>
```

---

[INGEST '"$TODAY"'] Scaffolded '"$SKILL_NAME"' via scripts/scaffold.sh.
  - Directory layout created: references/{wiki,raw}/, brain meta files, templates
  - Awaiting first raw material and wiki concepts
'

create_file "$TARGET/references/log.md" "$LOG_MD_CONTENT"

if [[ ! -f "$TARGET/references/wiki/.gitkeep" ]]; then
  touch "$TARGET/references/wiki/.gitkeep"
  echo "  created: $TARGET/references/wiki/.gitkeep"
else
  echo "  skip: $TARGET/references/wiki/.gitkeep (exists)"
fi

if [[ ! -f "$TARGET/references/raw/.gitkeep" ]]; then
  touch "$TARGET/references/raw/.gitkeep"
  echo "  created: $TARGET/references/raw/.gitkeep"
else
  echo "  skip: $TARGET/references/raw/.gitkeep (exists)"
fi

echo ""
echo "Done. Next steps:"
echo "  1. Edit $TARGET/SKILL.md - fill in description, triggers, How-to-Use"
echo "  2. Create references/wiki/<context>/<category>/<concept>.md files"
echo "  3. Drop any source material into references/raw/ (keep original filenames)"
echo "  4. Run scripts/lint.sh to populate references/_index.md"
echo "  5. See $SKILL_ROOT/references/_template.md for wiki concept shape"
echo "  6. See $SKILL_ROOT/references/_brain.md for the brain protocol"
