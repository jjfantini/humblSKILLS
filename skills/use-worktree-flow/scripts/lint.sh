#!/usr/bin/env bash
set -euo pipefail

SKILL_ROOT="${1:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"

SKILL_ROOT="$SKILL_ROOT" python3 <<'PYEOF'
import datetime
import os
import re
import stat
import sys
from pathlib import Path

skill_root = Path(os.environ["SKILL_ROOT"]).resolve()
refs = skill_root / "references"
wiki = refs / "wiki"
raw = refs / "raw"
index = refs / "_index.md"
required = {
    "title",
    "context",
    "category",
    "concept",
    "description",
    "tags",
    "sources",
    "last_ingested",
}
scalar = re.compile(r"^([A-Za-z_][A-Za-z0-9_]*):\s*(.*)$")
failures = []
warnings = []


def parse_frontmatter(path: Path) -> tuple[dict[str, str], list[str]]:
    fields: dict[str, str] = {}
    sources: list[str] = []
    in_sources = False
    lines = path.read_text().splitlines()
    if not lines or lines[0].strip() != "---":
        return fields, sources

    for line in lines[1:]:
        if line.strip() == "---":
            break
        if in_sources:
            stripped = line.strip()
            if stripped.startswith("- "):
                value = stripped[2:].strip().strip('"').strip("'")
                sources.append(value)
                continue
            if stripped == "":
                continue
            in_sources = False
        if line.strip().startswith("sources:"):
            fields["sources"] = "true"
            rest = line.split("sources:", 1)[1].strip()
            if rest not in ("", "[]"):
                sources.extend(
                    item.strip().strip('"').strip("'")
                    for item in rest.strip("[]").split(",")
                    if item.strip()
                )
            else:
                in_sources = rest == ""
            continue
        match = scalar.match(line)
        if match:
            key, value = match.group(1), match.group(2).strip().strip('"').strip("'")
            if value:
                fields[key] = value
    return fields, sources


if not (skill_root / "SKILL.md").exists():
    print(f"ERROR: no SKILL.md at {skill_root}", file=sys.stderr)
    sys.exit(2)

wiki_files = sorted(path for path in wiki.rglob("*.md") if path.is_file())
raw_files = sorted(path for path in raw.rglob("*") if path.is_file() and path.name != ".gitkeep")
metadata: dict[Path, tuple[dict[str, str], list[str]]] = {}
contexts: dict[str, set[str]] = {}

for path in wiki_files:
    rel_under_wiki = path.relative_to(wiki).parts
    rel = path.relative_to(skill_root).as_posix()
    if len(rel_under_wiki) != 3:
        failures.append(f"{rel}: expected references/wiki/<context>/<category>/<concept>.md")
        continue

    context, category, filename = rel_under_wiki
    concept = filename.removesuffix(".md")
    contexts.setdefault(context, set()).add(category)
    fields, sources = parse_frontmatter(path)
    metadata[path] = (fields, sources)

    missing = sorted(required - fields.keys())
    if missing:
        failures.append(f"{rel}: missing frontmatter fields: {', '.join(missing)}")
    if (fields.get("context"), fields.get("category"), fields.get("concept")) != (
        context,
        category,
        concept,
    ):
        failures.append(f"{rel}: path/frontmatter mismatch")

    for source in sources:
        if not source.startswith("references/raw/"):
            failures.append(f"{rel}: source outside references/raw: {source}")
        elif not (skill_root / source).exists():
            failures.append(f"{rel}: broken source: {source}")
    if not sources:
        warnings.append(f"{rel}: empty sources")

    command = fields.get("command")
    if command:
        command_path = skill_root / command
        if not command_path.exists():
            failures.append(f"{rel}: missing command: {command}")
        elif not command_path.stat().st_mode & stat.S_IXUSR:
            failures.append(f"{rel}: command is not executable: {command}")

cited = {source for _, sources in metadata.values() for source in sources}
for path in raw_files:
    rel = path.relative_to(skill_root).as_posix()
    if rel not in cited:
        warnings.append(f"{rel}: raw file is not cited")

body: list[str] = []
body.append("## Summary\n\n")
if contexts:
    for context in sorted(contexts):
        categories = ", ".join(f"`{category}`" for category in sorted(contexts[context]))
        body.append(f"- **{context}** -> {categories}\n")
else:
    body.append("(no wiki concepts yet)\n")
body.append("\n## Wiki\n\n")

for context in sorted(contexts):
    body.append(f"### {context}\n\n")
    for category in sorted(contexts[context]):
        body.append(f"#### {category}\n\n")
        category_files = [
            path
            for path in wiki_files
            if path.relative_to(wiki).parts[:2] == (context, category)
        ]
        for path in category_files:
            fields, _ = metadata[path]
            link = path.relative_to(refs).as_posix()
            body.append(f"- [{path.stem}.md]({link}) - {fields.get('title', path.stem)}\n")
        body.append("\n")

body.append("---\n\n## Raw Sources\n\n")
if raw_files:
    for path in raw_files:
        rel = path.relative_to(skill_root).as_posix()
        link = path.relative_to(refs).as_posix()
        body.append(f"- [{path.name}]({link})\n")
else:
    body.append("(none yet)\n")

body.append("\n---\n\n## Scripts\n\n")
scripts = sorted((skill_root / "scripts").glob("*.sh"))
if scripts:
    for script in scripts:
        body.append(f"- [scripts/{script.name}](../scripts/{script.name})\n")
else:
    body.append("(none)\n")

existing = index.read_text() if index.exists() else "# Index\n\n"
start = "<!-- GENERATED:START -->"
end = "<!-- GENERATED:END -->"
if start not in existing or end not in existing:
    failures.append("references/_index.md: missing generated markers")
else:
    preamble = existing.split(start, 1)[0].rstrip() + "\n\n"
    index.write_text(preamble + start + "\n\n" + "".join(body) + "\n" + end + "\n")

log = refs / "log.md"
if log.exists():
    today = datetime.date.today().isoformat()
    log.open("a").write(
        f"\n[LINT {today}] {len(wiki_files)} wiki, {len(raw_files)} raw. "
        f"Hard: {len(failures)}, Soft: {len(warnings)}. Regenerated _index.md.\n"
    )

for warning in warnings:
    print(f"WARN: {warning}")
for failure in failures:
    print(f"FAIL: {failure}")

if failures:
    sys.exit(1)

print(f"OK: {len(wiki_files)} wiki files, {len(raw_files)} raw files")
PYEOF
