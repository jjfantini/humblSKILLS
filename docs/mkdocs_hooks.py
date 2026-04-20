"""Remove MkDocs built-in fenced_code so pymdownx.superfences handles fenced blocks."""

from __future__ import annotations

import shutil
from pathlib import Path
from typing import Any

_AGENT_SKILL_SRC = Path("getting_started") / "installation-agent" / "SKILL.md"
_AGENT_SKILL_DEST = Path("getting_started") / "installation" / "SKILL.md"


def on_config(config: Any, **kwargs: Any) -> Any:
    config["markdown_extensions"] = [
        ext for ext in config["markdown_extensions"] if ext != "fenced_code"
    ]
    return config


def on_post_build(config: Any, **kwargs: Any) -> None:
    docs_dir = Path(config["docs_dir"])
    site_dir = Path(config["site_dir"])
    src = docs_dir / _AGENT_SKILL_SRC
    if not src.is_file():
        raise FileNotFoundError(f"Agent SKILL source missing: {src}")
    dest = site_dir / _AGENT_SKILL_DEST
    dest.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(src, dest)
