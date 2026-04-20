"""Remove MkDocs built-in fenced_code so pymdownx.superfences handles fenced blocks."""

from __future__ import annotations

from typing import Any


def on_config(config: Any, **kwargs: Any) -> Any:
    config["markdown_extensions"] = [
        ext for ext in config["markdown_extensions"] if ext != "fenced_code"
    ]
    return config
