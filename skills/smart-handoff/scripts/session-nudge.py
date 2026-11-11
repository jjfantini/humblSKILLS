#!/usr/bin/env python3
"""Optional Stop hook: remind the agent to offer a handoff when a session has unsaved work.

Claude Code and Codex run Stop hooks after every reply. This hook stays silent unless the
session edited files under the workstream root since its last handoff. After one nudge it
waits REPEAT_MINUTES and for new edits before nudging that session again.

  python3 session-nudge.py [--root DIR]            # hook mode: reads hook JSON on stdin
  python3 session-nudge.py install --agent claude|codex [--project DIR] [--minutes N]
  python3 session-nudge.py uninstall --agent claude|codex [--project DIR]
"""

import argparse
import json
import os
from pathlib import Path
import re
import shlex
import sys
import time

REPEAT_MINUTES = 45
MARKER = "session-nudge.py"
SKIP = re.compile(r"/(\.git|node_modules|\.next|__pycache__|\.venv|\.claude|\.codex|\.humblskills)/")
HANDOFF = re.compile(r"/\.humblskills/handoffs/[^/]+\.md$")
CODEX_PATCH = re.compile(r"\*\*\* (?:Add|Update) File: ([^\s\"'`\\]+)")
CLAUDE_EDITS = {"Edit", "Write", "MultiEdit", "NotebookEdit"}
CODEX_COMMAND = re.compile(r'cmd:\s*"((?:[^"\\]|\\.)*)"')
# ponytail: shell writes are recognized only through > >> and tee targets; scripts
# that write files themselves (python, sed -i) are not seen.
REDIRECT = re.compile(r"(?:^|[\s;&|(])(?:\d?>>?|tee(?:\s+-a)?)\s*(['\"]?)([^\s;&|<>'\"()]+)\1")


def shell_targets(command, base):
    for _quote, target in REDIRECT.findall(command):
        if not target.startswith(("&", "/dev/")):
            yield os.path.join(base, os.path.expanduser(target))


def edited_paths(transcript, base):
    """Files this session wrote, in order, from a Claude Code or Codex transcript."""
    for line in transcript.open(errors="replace"):
        try:
            entry = json.loads(line)
        except ValueError:
            continue
        if entry.get("type") == "assistant":
            for item in entry.get("message", {}).get("content") or []:
                if not (isinstance(item, dict) and item.get("type") == "tool_use"):
                    continue
                arguments = item.get("input") or {}
                if item.get("name") in CLAUDE_EDITS and (arguments.get("file_path") or arguments.get("notebook_path")):
                    yield arguments.get("file_path") or arguments.get("notebook_path")
                elif item.get("name") == "Bash":
                    yield from shell_targets(str(arguments.get("command") or ""), base)
        elif entry.get("type") == "response_item":
            payload = entry.get("payload") or {}
            if payload.get("type") in ("custom_tool_call", "function_call"):
                code = str(payload.get("input") or payload.get("arguments") or "")
                yield from CODEX_PATCH.findall(code)
                for command in CODEX_COMMAND.findall(code):
                    yield from shell_targets(json.loads(f'"{command}"'), base)


def state_file(session):
    base = Path(os.environ.get("XDG_STATE_HOME", Path.home() / ".local" / "state"))
    # ponytail: one tiny file per nudged session, never pruned; prune by age if it ever matters.
    return base / "humblskills" / "handoff-nudge" / f"{re.sub(r'[^A-Za-z0-9_.-]', '_', session)}.json"


def decide(hook, root, now, minutes=REPEAT_MINUTES):
    """Return the nudge reason, or None to stay silent. Records the nudge when it returns one."""
    if hook.get("stop_hook_active"):
        return None
    transcript = Path(hook.get("transcript_path") or "")
    if not transcript.is_file():
        return None
    prefix = str(root).rstrip("/") + "/"
    work = handed_off = unsaved = 0
    for path in map(os.path.realpath, edited_paths(transcript, str(root))):
        if HANDOFF.search(path):
            handed_off, unsaved = 1, 0
        elif path.startswith(prefix) and not SKIP.search(path):
            work += 1
            unsaved += 1
    if not unsaved:
        return None
    state = state_file(hook.get("session_id") or str(transcript))
    try:
        last = json.loads(state.read_text())
    except (OSError, ValueError):
        last = None
    if last and (work <= last.get("edits", 0) or now - last.get("at", 0) < minutes * 60):
        return None
    state.parent.mkdir(parents=True, exist_ok=True)
    state.write_text(json.dumps({"at": now, "edits": work}))
    done = "since its last handoff" if handed_off else "but wrote no handoff"
    return (f"Handoff check: this session edited files under {root} {done}. "
            f"In one or two sentences, offer to write a smart-handoff to {root}/.humblskills/handoffs/, "
            "saying what closed as well as what opened. Accept a no without pushing.")


def settings_path(agent, project):
    if agent == "claude":
        return (Path(project) / ".claude" / "settings.local.json") if project else Path.home() / ".claude" / "settings.json"
    if not project:
        raise SystemExit("Codex installs are per project: pass --project DIR")
    return Path(project) / ".codex" / "hooks.json"


def install(agent, project, minutes, remove=False):
    path = settings_path(agent, project)
    settings = json.loads(path.read_text()) if path.exists() else {}
    stops = settings.setdefault("hooks", {}).setdefault("Stop", [])
    stops[:] = [group for group in stops if not any(MARKER in h.get("command", "") for h in group.get("hooks", []))]
    if not remove:
        command = f"python3 {shlex.quote(str(Path(__file__).resolve()))}"
        if project:
            command += f" --root {shlex.quote(str(Path(project).resolve()))}"
        if minutes != REPEAT_MINUTES:
            command += f" --minutes {minutes}"
        stops.append({"hooks": [{"type": "command", "command": command, "timeout": 15,
                                 "statusMessage": "Checking for unsaved handoff work..."}]})
    if not stops:
        del settings["hooks"]["Stop"]
    path.parent.mkdir(parents=True, exist_ok=True)
    temporary = path.with_name(f".{path.name}.tmp")
    temporary.write_text(json.dumps(settings, indent=2) + "\n")
    os.replace(temporary, path)
    return path


def main():
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("action", nargs="?", choices=["install", "uninstall"])
    parser.add_argument("--agent", choices=["claude", "codex"])
    parser.add_argument("--project", help="install for one project instead of every session")
    parser.add_argument("--root", help="workstream root (hook mode); defaults to the session's project directory")
    parser.add_argument("--minutes", type=int, default=REPEAT_MINUTES)
    args = parser.parse_args()
    if args.action:
        if not args.agent:
            parser.error("--agent is required")
        print(install(args.agent, args.project, args.minutes, remove=args.action == "uninstall"))
        return 0
    try:
        hook = json.load(sys.stdin)
    except ValueError:
        return 0
    root = Path(args.root or os.environ.get("CLAUDE_PROJECT_DIR") or hook.get("cwd") or os.getcwd()).resolve()
    reason = decide(hook, root, int(os.environ.get("NUDGE_NOW") or time.time()), args.minutes)
    if reason:
        print(json.dumps({"decision": "block", "reason": reason}))
    return 0


if __name__ == "__main__":
    sys.exit(main())
