#!/usr/bin/env bash
# lint.sh - health-check the brain of a Smart Skill
#
# Usage:
#   cd my-skill && bash scripts/lint.sh
#   bash scripts/lint.sh /abs/path/to/my-skill
#
# Checks:
#   1. Every wiki file's path matches its context/category/concept frontmatter
#   2. Every wiki file has all required frontmatter fields
#   3. Every sources: path resolves to a real file under references/raw/
#   4. Raw files not cited by any wiki concept (orphan raw - warning)
#   5. Wiki concepts with empty sources: (orphan concept - warning)
#   6. Duplicate concept: values across different paths (contradiction)
#   7. last_ingested older than STALE_DAYS (default 180) - info only
#
# Side effect: rewrites references/_index.md between <!-- GENERATED:START -->
# and <!-- GENERATED:END --> markers. Preamble above :START is preserved.
# Appends a [LINT <date>] entry to references/log.md.
#
# Taxonomy: derived from the filesystem walk of references/wiki/. No
# separate registry file.
#
# Exit codes:
#   0  all hard checks passed
#   1  one or more hard findings
#   2  invocation error (missing SKILL.md, etc.)
#
# Requires python3 (pre-installed on macOS). Pure stdlib.

set -uo pipefail

SKILL_ROOT="${1:-$PWD}"

if ! command -v python3 >/dev/null 2>&1; then
  echo "ERROR: python3 is required but not found on PATH" >&2
  exit 2
fi

STALE_DAYS="${STALE_DAYS:-180}" \
SKILL_ROOT="$SKILL_ROOT" \
exec python3 - "$@" <<'PYEOF'
import os
import re
import sys
import datetime
from pathlib import Path

RED = "\033[31m"
YEL = "\033[33m"
GRN = "\033[32m"
BOLD = "\033[1m"
RST = "\033[0m"

SKILL_ROOT = Path(os.environ["SKILL_ROOT"]).resolve()
STALE_DAYS = int(os.environ.get("STALE_DAYS", "180"))

if not (SKILL_ROOT / "SKILL.md").exists():
    print(f"ERROR: {SKILL_ROOT} does not look like a skill (no SKILL.md)", file=sys.stderr)
    print("Usage: bash lint.sh [skill-root]", file=sys.stderr)
    sys.exit(2)

REFS = SKILL_ROOT / "references"
WIKI = REFS / "wiki"
RAW = REFS / "raw"
INDEX = REFS / "_index.md"

hard_findings = 0
soft_findings = 0

def bold(msg): print(f"{BOLD}{msg}{RST}")
def fail(msg):
    global hard_findings
    print(f"  {RED}[FAIL]{RST} {msg}")
    hard_findings += 1
def warn(msg):
    global soft_findings
    print(f"  {YEL}[WARN]{RST} {msg}")
    soft_findings += 1
def info(msg): print(f"  [INFO] {msg}")

bold(f"Linting skill: {SKILL_ROOT}")
print()

# --------------------------------------------------------------------------
# Derive taxonomy from filesystem walk (no registry file)
# --------------------------------------------------------------------------

context_cats = {}
if WIKI.is_dir():
    for wf in WIKI.rglob("*.md"):
        if not wf.is_file():
            continue
        parts = wf.relative_to(WIKI).parts
        if len(parts) >= 2:
            context_cats.setdefault(parts[0], set()).add(parts[1])

bold("Taxonomy (derived from filesystem):")
if context_cats:
    for ctx in sorted(context_cats):
        cats = ", ".join(sorted(context_cats[ctx]))
        print(f"  {ctx} -> {cats}")
else:
    print("  (no wiki concepts yet)")
print()

# --------------------------------------------------------------------------
# Walk wiki files, parse frontmatter
# --------------------------------------------------------------------------

REQUIRED = ["title", "context", "category", "concept",
            "description", "tags",
            "sources", "last_ingested"]

fm_scalar = re.compile(r"^([A-Za-z_][A-Za-z0-9_]*):\s*(.*)$")

def parse_frontmatter(path: Path):
    """Return (fields: dict, sources: list[str])."""
    fields = {}
    sources = []
    in_fm = False
    in_sources = False
    with path.open() as f:
        for i, raw in enumerate(f):
            line = raw.rstrip("\n")
            if i == 0:
                if line.strip() != "---":
                    return fields, sources
                in_fm = True
                continue
            if in_fm and line.strip() == "---":
                break
            if not in_fm:
                continue
            if in_sources:
                stripped = line.lstrip()
                if line.startswith((" ", "\t")) and stripped.startswith("- "):
                    v = stripped[2:].strip()
                    if v.startswith('"') and v.endswith('"'):
                        v = v[1:-1]
                    elif v.startswith("'") and v.endswith("'"):
                        v = v[1:-1]
                    sources.append(v)
                    continue
                elif line.strip() == "":
                    continue
                else:
                    in_sources = False
            if line.strip().startswith("sources:"):
                rest = line.split("sources:", 1)[1].strip()
                if rest in ("", "[]"):
                    fields["sources"] = True
                    in_sources = (rest == "")
                else:
                    fields["sources"] = True
                continue
            m = fm_scalar.match(line)
            if m:
                k, v = m.group(1), m.group(2).strip()
                if v.startswith('"') and v.endswith('"'):
                    v = v[1:-1]
                elif v.startswith("'") and v.endswith("'"):
                    v = v[1:-1]
                if v != "":
                    fields[k] = v
    if "sources" not in fields and sources:
        fields["sources"] = True
    return fields, sources

wiki_files = []
if WIKI.is_dir():
    wiki_files = sorted(p for p in WIKI.rglob("*.md") if p.is_file())

bold("Checking wiki files...")

wiki_meta = {}  # path -> (fields, sources)
concept_seen = {}  # concept_value -> first "ctx/cat/concept" string

for wf in wiki_files:
    rel = wf.relative_to(SKILL_ROOT).as_posix()
    try:
        rel_under_wiki = wf.relative_to(WIKI).as_posix()
    except ValueError:
        fail(f"{rel}: not under references/wiki/")
        continue

    parts = rel_under_wiki.split("/")
    if len(parts) != 3 or not parts[2].endswith(".md"):
        fail(f"{rel}: invalid wiki path (expected references/wiki/<ctx>/<cat>/<concept>.md)")
        continue

    path_ctx, path_cat, fname = parts
    path_concept = fname[:-3]

    fields, sources = parse_frontmatter(wf)
    wiki_meta[wf] = (fields, sources)

    missing = [k for k in REQUIRED if k not in fields]
    if missing:
        fail(f"{rel}: missing frontmatter fields: {' '.join(missing)}")

    fm_ctx = fields.get("context", "")
    fm_cat = fields.get("category", "")
    fm_concept = fields.get("concept", "")

    if (fm_ctx, fm_cat, fm_concept) != (path_ctx, path_cat, path_concept):
        fail(f"{rel}: path/frontmatter mismatch")
        print(f"         path:  context={path_ctx}, category={path_cat}, concept={path_concept}")
        print(f"         front: context={fm_ctx}, category={fm_cat}, concept={fm_concept}")

    for src in sources:
        if not src.startswith("references/raw/"):
            fail(f"{rel}: source not under references/raw/: '{src}'")
            continue
        if not (SKILL_ROOT / src).exists():
            fail(f"{rel}: broken source path: '{src}'")

    if not sources:
        warn(f"{rel}: empty sources: (orphan concept - synthesis-only; audit if unintentional)")

    if fm_concept:
        key = f"{fm_ctx}/{fm_cat}/{fm_concept}"
        prev = concept_seen.get(fm_concept)
        if prev and prev != key:
            warn(f"{rel}: concept '{fm_concept}' also used at {prev} (possible contradiction)")
        else:
            concept_seen[fm_concept] = key

    last = fields.get("last_ingested", "")
    if last:
        try:
            d = datetime.date.fromisoformat(last)
            age = (datetime.date.today() - d).days
            if age > STALE_DAYS:
                info(f"{rel}: last_ingested={last} is {age} days old (> {STALE_DAYS})")
        except ValueError:
            warn(f"{rel}: last_ingested='{last}' is not a valid ISO date")

print(f"  scanned {len(wiki_files)} wiki files")
print()

# --------------------------------------------------------------------------
# Raw orphans
# --------------------------------------------------------------------------

bold("Checking raw files...")

cited = set()
for _, srcs in wiki_meta.values():
    cited.update(srcs)

raw_files = []
if RAW.is_dir():
    for p in sorted(RAW.rglob("*")):
        if p.is_file() and p.name != ".gitkeep":
            raw_files.append(p)

for rf in raw_files:
    rel = rf.relative_to(SKILL_ROOT).as_posix()
    if rel not in cited:
        warn(f"{rel}: orphan raw (not cited by any wiki concept)")

print(f"  scanned {len(raw_files)} raw files")
print()

# --------------------------------------------------------------------------
# Regenerate _index.md between sentinel markers
# --------------------------------------------------------------------------

bold(f"Regenerating {INDEX} ...")

START_MARKER = "<!-- GENERATED:START -->"
END_MARKER = "<!-- GENERATED:END -->"

DEFAULT_PREAMBLE = """# Index

Auto-generated by `scripts/lint.sh` from the filesystem walk of
`references/wiki/`, `references/raw/`, and `scripts/`. Do not hand-edit
below the generated markers. To change the taxonomy, add or remove wiki
files and re-run lint.

"""

preamble = DEFAULT_PREAMBLE
if INDEX.exists():
    existing = INDEX.read_text()
    if START_MARKER in existing:
        preamble = existing.split(START_MARKER, 1)[0]
        # ensure trailing blank line between preamble and marker
        if not preamble.endswith("\n\n"):
            preamble = preamble.rstrip("\n") + "\n\n"

# Build generated body
body = []
body.append("## Summary\n")
body.append("\n")
body.append("Context -> categories. See `## Wiki` below for the concept enumeration.\n")
body.append("\n")
if context_cats:
    for ctx in sorted(context_cats):
        cats = ", ".join(f"`{c}`" for c in sorted(context_cats[ctx]))
        body.append(f"- **{ctx}** -> {cats}\n")
else:
    body.append("(no wiki concepts yet)\n")
body.append("\n")

body.append("## Wiki\n")
body.append("\n")
if wiki_files:
    # Group by (ctx, cat) -> list of (concept, title, rel-path-under-references)
    grouped = {}
    for wf in wiki_files:
        fields, _ = wiki_meta[wf]
        ctx = fields.get("context", "unknown")
        cat = fields.get("category", "unknown")
        concept = fields.get("concept", wf.stem)
        title = fields.get("title", "(no title)")
        # Link is relative to references/_index.md
        link = wf.relative_to(REFS).as_posix()
        grouped.setdefault((ctx, cat), []).append((concept, title, link))

    current_ctx = None
    for (ctx, cat) in sorted(grouped):
        if ctx != current_ctx:
            body.append(f"### {ctx}\n")
            body.append("\n")
            current_ctx = ctx
        body.append(f"#### {cat}\n")
        body.append("\n")
        for concept, title, link in sorted(grouped[(ctx, cat)]):
            body.append(f"- [{concept}.md]({link}) - {title}\n")
        body.append("\n")
else:
    body.append("(none yet)\n")
    body.append("\n")

body.append("---\n")
body.append("\n")
body.append("## Raw Sources\n")
body.append("\n")
if raw_files:
    reverse = {}
    for wf, (_, srcs) in wiki_meta.items():
        for s in srcs:
            reverse.setdefault(s, []).append(wf.relative_to(SKILL_ROOT).as_posix())
    for rf in raw_files:
        rel = rf.relative_to(SKILL_ROOT).as_posix()
        link = rf.relative_to(REFS).as_posix()
        citers = reverse.get(rel, [])
        if citers:
            body.append(f"- [{rf.name}]({link}) - cited by: {', '.join(sorted(citers))}\n")
        else:
            body.append(f"- [{rf.name}]({link}) - (not cited; orphan)\n")
    body.append("\n")
else:
    body.append("(none yet)\n")
    body.append("\n")

body.append("---\n")
body.append("\n")
body.append("## Scripts\n")
body.append("\n")
scripts_dir = SKILL_ROOT / "scripts"
if scripts_dir.is_dir():
    found = sorted(
        p for p in scripts_dir.iterdir()
        if p.is_file() and p.suffix in (".sh", ".py")
    )
    if found:
        for p in found:
            # Link is relative to references/_index.md -> one level up
            link = f"../scripts/{p.name}"
            body.append(f"- [scripts/{p.name}]({link})\n")
    else:
        body.append("(none)\n")
else:
    body.append("(none)\n")
body.append("\n")

generated = "".join(body)

final = (
    preamble
    + START_MARKER
    + "\n\n"
    + generated
    + END_MARKER
    + "\n"
)

INDEX.write_text(final)
print(f"  {GRN}wrote{RST} {INDEX}")
print()

# --------------------------------------------------------------------------
# Append log entry
# --------------------------------------------------------------------------

log = REFS / "log.md"
if log.exists():
    today = datetime.date.today().isoformat()
    with log.open("a") as f:
        f.write(
            f"\n[LINT {today}] {len(wiki_files)} wiki, {len(raw_files)} raw. "
            f"Hard: {hard_findings}, Soft: {soft_findings}. Regenerated _index.md.\n"
        )

# --------------------------------------------------------------------------
# Summary
# --------------------------------------------------------------------------

bold("Summary:")
print(f"  wiki files:    {len(wiki_files)}")
print(f"  raw files:     {len(raw_files)}")
print(f"  hard findings: {hard_findings}")
print(f"  soft findings: {soft_findings}")
print()

if hard_findings:
    print(f"{RED}FAIL: {hard_findings} hard finding(s). Fix before shipping.{RST}")
    sys.exit(1)

print(f"{GRN}OK: all hard checks passed.{RST}")
sys.exit(0)
PYEOF
