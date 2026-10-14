#!/usr/bin/env python3
"""Load private pickup settings, select a handoff, or prepare Notion setup."""

import argparse
import json
import os
from pathlib import Path
import re
import sys
import tempfile
from uuid import UUID
from zoneinfo import ZoneInfo, ZoneInfoNotFoundError


def config_path():
    return Path(os.environ.get("XDG_CONFIG_HOME", str(Path.home() / ".config"))).expanduser() / "humblskills" / "smart-pickup.json"


def string(value, name):
    if not isinstance(value, str) or not value.strip():
        raise ValueError(f"{name} must be a nonempty string")


def string_list(value, name):
    if not isinstance(value, list):
        raise ValueError(f"{name} must be a list")
    for item in value:
        string(item, name)


def validate(config):
    if not isinstance(config, dict):
        raise ValueError("config must be an object")
    unknown = set(config) - {"version", "timezone", "project_roots", "todos"}
    if unknown:
        raise ValueError(f"unknown config keys: {', '.join(sorted(unknown))}")
    if type(config.get("version")) is not int or config["version"] != 1:
        raise ValueError("config version must be 1")
    if "todos" not in config:
        raise ValueError("configure a todos source or explicitly set todos to null")
    if "timezone" in config:
        string(config["timezone"], "timezone")
        try:
            ZoneInfo(config["timezone"])
        except ZoneInfoNotFoundError as exc:
            raise ValueError("timezone must be a known IANA timezone") from exc
    string_list(config.get("project_roots", []), "project_roots")
    todos = config["todos"]
    if todos is None:
        return config
    if not isinstance(todos, dict):
        raise ValueError("todos must be an object or null")
    unknown = set(todos) - {"provider", "source", "fields", "closed_statuses", "scope_values"}
    if unknown:
        raise ValueError(f"unknown todos keys: {', '.join(sorted(unknown))}")
    for key in ("provider", "source"):
        string(todos.get(key), f"todos.{key}")
    if not re.fullmatch(r"[a-z][a-z0-9-]*", todos["provider"]):
        raise ValueError("todos.provider must be a lowercase service name or file")
    if todos["provider"] == "notion":
        try:
            if not todos["source"].startswith("collection://"):
                raise ValueError("wrong source scheme")
            UUID(todos["source"][len("collection://"):])
        except ValueError as exc:
            raise ValueError("Notion source must be a verified collection:// data source ID") from exc
    fields = todos.get("fields", {})
    scopes = todos.get("scope_values", {})
    for mapping, name in ((fields, "fields"), (scopes, "scope_values")):
        if not isinstance(mapping, dict):
            raise ValueError(f"todos.{name} must be an object")
        for key, value in mapping.items():
            string(key, f"todos.{name} key")
            string(value, f"todos.{name} value")
    if set(fields) - {"title", "status", "done", "project", "priority", "due"}:
        raise ValueError("unknown todos.fields semantic key")
    if any("/" in key or "\\" in key or key in (".", "..") for key in scopes):
        raise ValueError("scope_values keys must be single directory names")
    if scopes and "project" not in fields:
        raise ValueError("scope_values requires a project field")
    string_list(todos.get("closed_statuses", []), "todos.closed_statuses")
    if todos.get("closed_statuses") and "status" not in fields:
        raise ValueError("closed_statuses requires a status field")
    if todos.get("closed_statuses") and "done" in fields:
        raise ValueError("use closed_statuses or a done checkbox field for completion, not both")
    return config


def save_config(config, destination):
    validate(config)
    destination.parent.mkdir(parents=True, exist_ok=True)
    temporary = None
    try:
        with tempfile.NamedTemporaryFile(mode="w", dir=destination.parent, prefix=".smart-pickup-", delete=False) as file:
            temporary = Path(file.name)
            file.write(json.dumps(config, indent=2) + "\n")
        os.replace(temporary, destination)
    finally:
        if temporary and temporary.exists():
            temporary.unlink()


PRIOR_ROW = re.compile(r"^\|\s*\**\s*(?:prior|previous) handoffs?\s*\**\s*\|(.*)$", re.IGNORECASE | re.MULTILINE)


def title(path):
    with path.open(errors="replace") as file:
        first = next((line.strip() for line in file if line.strip()), "")
    return first[2:].strip() if first.startswith("# ") else ""


def handoffs(directory, topic):
    """Newest first; every topic word must appear in the filename or the document title."""
    if not directory.is_dir():
        return []
    words = re.findall(r"[^\W_]+", topic.casefold())
    if topic.strip() and not words:
        return []
    files = [p for p in directory.iterdir() if p.is_file() and p.suffix.lower() == ".md"]
    if words:
        files = [p for p in files if all(w in f"{p.name} {title(p)}".casefold() for w in words)]
    return sorted(files, key=lambda p: (p.stat().st_mtime_ns, p.name), reverse=True)


def superseded_by(selected):
    """Handoffs in the same folder whose Prior handoff row names the selected file."""
    later = []
    if selected.parent.parts[-2:] != (".humblskills", "handoffs"):
        return later
    for path in handoffs(selected.parent, ""):
        if path != selected and any(selected.name in row for row in PRIOR_ROW.findall(path.read_text(errors="replace"))):
            later.append(str(path))
    return later


def alternatives(roots, topic):
    found = set()
    warnings = []
    for root in roots:
        root = Path(root).expanduser().resolve()
        if not root.is_dir():
            warnings.append(f"project root unavailable: {root}")
            continue
        for current, directories, _files in os.walk(root, onerror=lambda error: warnings.append(str(error))):
            current = Path(current)
            depth = len(current.relative_to(root).parts)
            directories[:] = [] if depth >= 4 else [d for d in directories if not d.startswith(".") and d not in ("node_modules", "venv", "dist")]
            candidates = handoffs(current / ".humblskills" / "handoffs", topic)
            if candidates:
                found.add(candidates[0].resolve())
    return sorted(found, key=lambda p: (p.stat().st_mtime_ns, str(p)), reverse=True), warnings


def context(cwd, config_file, topic="", explicit=None):
    cwd = Path(cwd).expanduser().resolve()
    if not cwd.is_dir():
        raise ValueError("starting directory does not exist")
    config = validate(json.loads(config_file.read_text())) if config_file.exists() else None
    selected = None
    if explicit:
        selected = Path(explicit).expanduser()
        if not selected.is_absolute():
            selected = cwd / selected
        selected = selected.resolve()
        if not selected.is_file() or selected.suffix.lower() != ".md":
            raise ValueError("explicit handoff must be an existing Markdown file")
        others = []
    else:
        files = handoffs(cwd / ".humblskills" / "handoffs", topic)
        selected = files[0] if files else None
        others = files[1:]
    candidates, warnings = ([], []) if selected else alternatives((config or {}).get("project_roots", []), topic)
    todos = (config or {}).get("todos")
    scope = None
    if todos:
        scopes = todos.get("scope_values", {})
        for part in reversed(cwd.parts):
            if part in scopes:
                scope = {"field": todos["fields"]["project"], "value": scopes[part]}
                break
    return {
        "starting_directory": str(cwd), "config_path": str(config_file),
        "needs_configuration": config is None, "settings": config,
        "selected_handoff": str(selected) if selected else None,
        "selected_title": title(selected) if selected else None,
        "superseded_by": superseded_by(selected) if selected else [],
        "other_handoffs": [{"path": str(p), "title": title(p)} for p in others[:5]],
        "other_handoff_count": len(others),
        "selection": "explicit" if explicit else ("local" if selected else "choose_workstream"),
        "candidates": [str(p) for p in candidates], "warnings": warnings,
        "todo_scope": scope,
        "todo_scope_unmapped": bool(todos and todos.get("scope_values") and scope is None),
    }


def notion_request(title, parent=None):
    string(title, "Notion database title")
    request = {"title": title, "schema": (Path(__file__).resolve().parents[1] / "assets" / "notion-tasks.sql").read_text()}
    if parent:
        if not re.fullmatch(r"[0-9a-fA-F]{32}|[0-9a-fA-F]{8}(?:-[0-9a-fA-F]{4}){3}-[0-9a-fA-F]{12}", parent):
            raise ValueError("parent must be a Notion page ID")
        request["parent"] = {"type": "page_id", "page_id": parent}
    return {"operation": "create_database", "requires_connected_tool": True, "request": request,
            "next_step": "After user-requested creation, fetch the returned data source and save personal config with its real collection:// ID. Do not save this request as config or auto-retry creation after an unknown outcome."}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--config", type=Path, default=config_path())
    commands = parser.add_subparsers(dest="command", required=True)
    resume = commands.add_parser("context", help="select a local handoff and load personal settings")
    resume.add_argument("--cwd", default=os.getcwd())
    resume.add_argument("--topic", default="")
    resume.add_argument("--handoff")
    setup = commands.add_parser("configure", help="validate and save a prepared personal configuration")
    setup.add_argument("--input", type=Path, required=True)
    notion = commands.add_parser("init-notion", help="prepare a creation request for the agent's connected Notion tool")
    notion.add_argument("--title", default="Pickup Tasks")
    notion.add_argument("--parent", help="optional parent page ID; omitted creates in Private")
    args = parser.parse_args()
    try:
        if args.command == "context":
            result = context(args.cwd, args.config.expanduser(), args.topic, args.handoff)
        elif args.command == "configure":
            settings = validate(json.loads(args.input.expanduser().read_text()))
            save_config(settings, args.config.expanduser())
            result = {"saved": str(args.config.expanduser()), "todos_enabled": settings["todos"] is not None}
        else:
            result = notion_request(args.title, args.parent)
        print(json.dumps(result, indent=2))
        return 0
    except (ValueError, OSError) as error:
        print(json.dumps({"error": str(error)}), file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main())
