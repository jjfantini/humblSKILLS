#!/usr/bin/env python3
"""Offline checks for the optional handoff nudge hook."""

import importlib.util
import json
import os
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location("nudge", Path(__file__).resolve().parents[1] / "scripts" / "session-nudge.py")
nudge = importlib.util.module_from_spec(spec)
spec.loader.exec_module(nudge)


class NudgeTests(unittest.TestCase):
    def setUp(self):
        temp = tempfile.TemporaryDirectory()
        self.addCleanup(temp.cleanup)
        self.dir = Path(temp.name).resolve()
        self.root = self.dir / "work"
        self.root.mkdir()
        patcher = patch.dict(os.environ, {"XDG_STATE_HOME": str(self.dir / "state"), "HOME": str(self.dir / "home")})
        patcher.start()
        self.addCleanup(patcher.stop)
        self.lines = []

    def claude_edit(self, path):
        self.lines.append({"type": "assistant", "message": {"content": [{"type": "tool_use", "name": "Edit", "input": {"file_path": str(path)}}]}})

    def codex_patch(self, path):
        patch_text = f"await tools.apply_patch(`*** Begin Patch\n*** Update File: {path}\n@@\n-a\n+b\n*** End Patch`)"
        self.lines.append({"type": "response_item", "payload": {"type": "custom_tool_call", "name": "exec", "input": patch_text}})

    def run_hook(self, now, **extra):
        transcript = self.dir / "transcript.jsonl"
        transcript.write_text("".join(json.dumps(line) + "\n" for line in self.lines))
        hook = {"session_id": "s1", "transcript_path": str(transcript), "stop_hook_active": False, **extra}
        return nudge.decide(hook, self.root, now)

    def test_silent_without_work_edits(self):
        self.assertIsNone(self.run_hook(1000))
        for path in (self.root / ".claude" / "settings.json", self.dir / "elsewhere.md"):
            self.claude_edit(path)
        self.assertIsNone(self.run_hook(1000))

    def test_nudges_once_then_waits_for_time_and_new_edits(self):
        self.claude_edit(self.root / "a.md")
        self.assertIn("wrote no handoff", self.run_hook(1000))
        self.assertIsNone(self.run_hook(1000 + 60 * 60), "no new edits since the nudge")
        self.claude_edit(self.root / "b.md")
        self.assertIsNone(self.run_hook(1000 + 10 * 60), "new edit but too soon")
        self.assertIsNotNone(self.run_hook(1000 + 50 * 60))

    def test_handoff_resets_until_new_work(self):
        self.claude_edit(self.root / "a.md")
        self.claude_edit(self.root / "sub" / ".humblskills" / "handoffs" / "x-handoff-2026-10-14.md")
        self.assertIsNone(self.run_hook(1000))
        self.claude_edit(self.root / "c.md")
        self.assertIn("since its last handoff", self.run_hook(1000))

    def test_codex_patch_edits_count(self):
        self.codex_patch(self.root / "src" / "board.ts")
        self.assertIsNotNone(self.run_hook(1000))

    def test_shell_writes_count(self):
        self.lines.append({"type": "assistant", "message": {"content": [{"type": "tool_use", "name": "Bash", "input": {"command": "cat > notes.md <<'EOF'\nhi\nEOF"}}]}})
        self.assertIsNotNone(self.run_hook(1000))
        self.lines = [{"type": "response_item", "payload": {"type": "custom_tool_call", "name": "exec",
                       "input": 'text(await tools.exec_command({cmd:"printf \'hello\\\\n\' > notes.md 2>/dev/null","max_output_tokens":1000}));'}}]
        self.assertIsNotNone(nudge.decide({"session_id": "s2", "transcript_path": self.write()}, self.root, 1000))
        self.lines = [{"type": "assistant", "message": {"content": [{"type": "tool_use", "name": "Bash", "input": {"command": "ls > /dev/null 2>&1; echo hi >&2; cat x > /tmp/out.txt"}}]}}]
        self.assertIsNone(nudge.decide({"session_id": "s3", "transcript_path": self.write()}, self.root, 1000))

    def write(self):
        transcript = self.dir / "transcript.jsonl"
        transcript.write_text("".join(json.dumps(line) + "\n" for line in self.lines))
        return str(transcript)

    def test_loop_guard_and_missing_transcript(self):
        self.claude_edit(self.root / "a.md")
        self.assertIsNone(self.run_hook(1000, stop_hook_active=True))
        self.assertIsNone(nudge.decide({"transcript_path": str(self.dir / "missing.jsonl")}, self.root, 1000))

    def test_install_is_idempotent_and_preserves_other_hooks(self):
        project = self.dir / "proj"
        settings = project / ".claude" / "settings.local.json"
        settings.parent.mkdir(parents=True)
        other = {"hooks": [{"type": "command", "command": "echo other"}]}
        settings.write_text(json.dumps({"model": "x", "hooks": {"Stop": [other]}}))
        nudge.install("claude", str(project), 45)
        nudge.install("claude", str(project), 30)
        data = json.loads(settings.read_text())
        self.assertEqual(data["model"], "x")
        commands = [h["command"] for group in data["hooks"]["Stop"] for h in group["hooks"]]
        self.assertEqual(len(commands), 2)
        self.assertIn("--root", commands[1])
        self.assertIn("--minutes 30", commands[1])
        nudge.install("claude", str(project), 45, remove=True)
        self.assertEqual(json.loads(settings.read_text())["hooks"]["Stop"], [other])
        codex = nudge.install("codex", str(project), 45)
        self.assertEqual(codex, project / ".codex" / "hooks.json")
        with self.assertRaises(SystemExit):
            nudge.install("codex", None, 45)


if __name__ == "__main__":
    unittest.main()
