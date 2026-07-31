"""Remove MkDocs built-in fenced_code so pymdownx.superfences handles fenced blocks.

Also renders the skill catalog into docs/skills.md from registry.json at build
time, so the published list can't drift from the released registry.
"""

from __future__ import annotations

import json
import re
import shutil
from pathlib import Path
from typing import Any

_AGENT_SKILL_SRC = Path("getting_started") / "installation-agent" / "SKILL.md"
_AGENT_SKILL_DEST = Path("getting_started") / "installation" / "SKILL.md"

_CATALOG_MARKER = "<!-- SKILL_CATALOG -->"

# Human labels for the closed category set in cli/internal/frontmatter.
_CATEGORY_LABELS = {
    "development": "Development",
    "design": "Design",
    "writing": "Writing",
    "meta": "Meta (skill authoring & project setup)",
}


def on_config(config: Any, **kwargs: Any) -> Any:
    config["markdown_extensions"] = [
        ext for ext in config["markdown_extensions"] if ext != "fenced_code"
    ]
    return config


def _first_sentence(description: str) -> str:
    """The lead sentence of a skill description, for a table cell.

    Skill descriptions are long trigger paragraphs written for the agent's
    router, not for humans reading a list.
    """
    text = " ".join(description.split())
    # Cut at whichever comes first: the end of the lead sentence, or the
    # trigger/anti-trigger boilerplate every description tails off into. Taking
    # the earliest of the two keeps one-sentence output even when a sentence end
    # is missed (abbreviations) or the boilerplate is several sentences down.
    cuts = [text.find(m) for m in (" Use when", " Use this", " Do NOT", " Trigger")]
    if match := re.search(r"\.\s", text):
        cuts.append(match.start() + 1)
    cuts = [c for c in cuts if c > 0]
    return text[: min(cuts)].rstrip() if cuts else text


def _render_catalog(registry_path: Path) -> str:
    """A category-grouped Markdown table of every skill in registry.json."""
    data = json.loads(registry_path.read_text(encoding="utf-8"))
    skills = data.get("skills", [])
    by_category: dict[str, list[dict[str, Any]]] = {}
    for skill in skills:
        by_category.setdefault(skill.get("category", "other"), []).append(skill)

    lines: list[str] = []
    # Known categories first, in taxonomy order, then anything unexpected.
    ordered = list(_CATEGORY_LABELS) + sorted(set(by_category) - set(_CATEGORY_LABELS))
    for category in ordered:
        entries = by_category.get(category)
        if not entries:
            continue
        lines.append(f"## {_CATEGORY_LABELS.get(category, category.title())}")
        lines.append("")
        lines.append("| Skill | Version | What it does |")
        lines.append("|-------|---------|--------------|")
        for skill in sorted(entries, key=lambda s: s["name"]):
            mirrored = " ↗" if skill.get("upstream") else ""
            summary = _first_sentence(skill.get("description", ""))
            lines.append(
                f"| `{skill['name']}`{mirrored} | {skill.get('version', '-')} | {summary} |"
            )
        lines.append("")
    lines.append(f"**{len(skills)} skills** in the registry.")
    return "\n".join(lines)


def on_page_markdown(markdown: str, page: Any, config: Any, **kwargs: Any) -> str:
    if _CATALOG_MARKER not in markdown:
        return markdown
    registry = Path(config["config_file_path"]).parent / "registry.json"
    if not registry.is_file():
        raise FileNotFoundError(f"registry.json missing, cannot build catalog: {registry}")
    return markdown.replace(_CATALOG_MARKER, _render_catalog(registry))


def on_post_build(config: Any, **kwargs: Any) -> None:
    docs_dir = Path(config["docs_dir"])
    site_dir = Path(config["site_dir"])
    src = docs_dir / _AGENT_SKILL_SRC
    if not src.is_file():
        raise FileNotFoundError(f"Agent SKILL source missing: {src}")
    dest = site_dir / _AGENT_SKILL_DEST
    dest.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(src, dest)
