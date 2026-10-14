#!/usr/bin/env python3
"""Offline checks for selection, configuration, and the Notion setup request."""

import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch

import pickup


class PickupTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name).resolve()
        self.cwd = self.root / "projects" / "billing"
        self.cwd.mkdir(parents=True)
        self.config = self.root / "settings" / "smart-pickup.json"

    def document(self, directory, name, timestamp):
        folder = directory / ".humblskills" / "handoffs"
        folder.mkdir(parents=True, exist_ok=True)
        path = folder / name
        path.write_text("# Handoff\n")
        os.utime(path, (timestamp, timestamp))
        return path

    def settings(self, **updates):
        settings = {"version": 1, "project_roots": [str(self.root / "projects")], "todos": None}
        settings.update(updates)
        pickup.save_config(settings, self.config)
        return settings

    def test_newest_local_ignores_parent_and_nested(self):
        self.document(self.cwd.parent, "parent.md", 900)
        self.document(self.cwd / "nested", "nested.md", 800)
        self.document(self.cwd, "older.md", 100)
        selected = self.document(self.cwd, "newer.md", 200)
        result = pickup.context(self.cwd, self.config)
        self.assertEqual(result["selected_handoff"], str(selected))
        self.assertTrue(result["needs_configuration"])
        self.assertEqual(result["candidates"], [])

    def test_topic_is_literal_case_insensitive(self):
        self.document(self.cwd, "Invoices.md", 100)
        selected = self.document(self.cwd, "billing.md", 200)
        self.document(self.cwd, "newest.md", 300)
        self.assertEqual(pickup.context(self.cwd, self.config, "BILLING")["selected_handoff"], str(selected))
        self.assertIsNone(pickup.context(self.cwd, self.config, "*")["selected_handoff"])

    def test_equal_mtime_is_deterministic(self):
        self.document(self.cwd, "a.md", 100)
        selected = self.document(self.cwd, "z.md", 100)
        self.assertEqual(pickup.context(self.cwd, self.config)["selected_handoff"], str(selected))

    def test_explicit_temp_document_without_search(self):
        path = self.root / "temporary.md"
        path.write_text("# Explicit handoff\n")
        result = pickup.context(self.cwd, self.config, explicit=str(path))
        self.assertEqual(result["selection"], "explicit")
        self.assertEqual(result["selected_handoff"], str(path))
        self.assertTrue(path.exists())

    def test_relative_explicit_document(self):
        path = self.cwd / "session.md"
        path.write_text("# Handoff\n")
        self.assertEqual(pickup.context(self.cwd, self.config, explicit="session.md")["selected_handoff"], str(path))

    def test_invalid_explicit_document_fails(self):
        with self.assertRaises(ValueError):
            pickup.context(self.cwd, self.config, explicit="missing.md")

    def test_fallback_requires_choice_and_is_depth_bounded(self):
        self.settings()
        self.document(self.cwd.parent / "shipping", "older.md", 100)
        selected = self.document(self.cwd.parent / "shipping", "latest.md", 200)
        allowed = self.document(self.cwd.parent / "a" / "b" / "c" / "d", "depth-four.md", 300)
        self.document(self.cwd.parent / "a" / "b" / "c" / "d" / "e", "depth-five.md", 400)
        result = pickup.context(self.cwd, self.config)
        self.assertIsNone(result["selected_handoff"])
        self.assertEqual(result["selection"], "choose_workstream")
        self.assertEqual(set(result["candidates"]), {str(selected), str(allowed)})

    def test_empty_config_does_not_search_machine(self):
        self.document(self.cwd.parent / "shipping", "sibling.md", 100)
        result = pickup.context(self.cwd, self.config)
        self.assertEqual(result["candidates"], [])

    def test_scope_matches_exact_deepest_segment(self):
        todos = {"provider": "file", "source": "~/tasks.md", "fields": {"project": "Project"}, "scope_values": {"projects": "all", "billing": "billing-team"}}
        self.settings(todos=todos)
        result = pickup.context(self.cwd, self.config)
        self.assertEqual(result["todo_scope"], {"field": "Project", "value": "billing-team"})
        self.assertFalse(result["todo_scope_unmapped"])
        other = self.root / "billing-extra"
        other.mkdir()
        result = pickup.context(other, self.config)
        self.assertIsNone(result["todo_scope"])
        self.assertTrue(result["todo_scope_unmapped"])

    def test_invalid_update_preserves_existing_config(self):
        self.settings()
        before = self.config.read_bytes()
        for invalid in ({"version": True, "todos": None}, {"version": 1}, {"version": 1, "todos": None, "token": "value"}, {"version": 1, "todos": None, "timezone": "not-a-zone"}):
            with self.assertRaises(ValueError):
                pickup.save_config(invalid, self.config)
            self.assertEqual(before, self.config.read_bytes())

    def test_mapping_validation(self):
        for todos in ({"provider": "notion", "source": "collection://------------------------------------"}, {"provider": "file", "source": "tasks.md", "scope_values": {"billing": "Billing"}}, {"provider": "file", "source": "tasks.md", "closed_statuses": ["Done"]}, {"provider": "file", "source": "tasks.md", "command": "anything"}):
            with self.assertRaises(ValueError):
                pickup.validate({"version": 1, "todos": todos})

    def test_configure_cli_and_permissions(self):
        source = self.root / "prepared.json"
        source.write_text(json.dumps({"version": 1, "todos": None}))
        result = subprocess.run([sys.executable, str(Path(pickup.__file__)), "--config", str(self.config), "configure", "--input", str(source)], capture_output=True, text=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertFalse(json.loads(result.stdout)["todos_enabled"])
        self.assertEqual(self.config.stat().st_mode & 0o777, 0o600)
        result = subprocess.run([sys.executable, str(Path(pickup.__file__)), "--config", str(self.config), "context", "--cwd", str(self.cwd)], capture_output=True, text=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertFalse(json.loads(result.stdout)["needs_configuration"])

    def test_xdg_config_location(self):
        with patch.dict(os.environ, {"XDG_CONFIG_HOME": str(self.root / "xdg")}):
            self.assertEqual(pickup.config_path(), self.root / "xdg" / "humblskills" / "smart-pickup.json")

    def test_notion_request_has_no_side_effects(self):
        before = list(self.root.iterdir())
        result = pickup.notion_request("Planning Tasks")
        self.assertEqual(result["operation"], "create_database")
        self.assertTrue(result["requires_connected_tool"])
        self.assertEqual(result["request"]["title"], "Planning Tasks")
        self.assertNotIn("parent", result["request"])
        for field in ("Task name", "Status", "Priority", "Project", "Due date"):
            self.assertIn('"' + field + '"', result["request"]["schema"])
        self.assertEqual(list(self.root.iterdir()), before)

    def test_notion_parent_validation(self):
        parent = "00000000-0000-0000-0000-000000000001"
        self.assertEqual(pickup.notion_request("Tasks", parent)["request"]["parent"]["page_id"], parent)
        with self.assertRaises(ValueError):
            pickup.notion_request("Tasks", "not-a-page-id")

    def handoff_snapshot(self, cwd, *arguments):
        preflight = Path(__file__).resolve().parents[2] / "smart-handoff" / "scripts" / "preflight.sh"
        if not preflight.is_file():
            self.skipTest("tandem checks require smart-handoff alongside smart-pickup")
        result = subprocess.run(["bash", str(preflight), "--slug", "billing", *arguments], cwd=cwd, capture_output=True, text=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        return dict(line.split("=", 1) for line in result.stdout.splitlines() if "=" in line and line.split("=", 1)[0].isupper())

    def test_handoff_pickup_roundtrip_in_nested_git_workstream(self):
        subprocess.run(["git", "init", "-q", str(self.cwd.parent)], check=True, capture_output=True)
        snapshot = self.handoff_snapshot(self.cwd)
        path = Path(snapshot["PERSIST_FILE"])
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text("# Billing - Session Handoff\n\n## Next Steps\n\n1. Verify the retry fix.\n")
        os.utime(path, (100, 100))
        self.document(self.cwd.parent, "newer-root-handoff.md", 900)
        self.document(self.cwd.parent / "shipping", "newer-shipping-handoff.md", 1000)
        result = pickup.context(self.cwd, self.config)
        self.assertEqual(result["selected_handoff"], str(path))
        self.assertEqual(path.parent, self.cwd / ".humblskills" / "handoffs")
        self.assertEqual(result["candidates"], [])
        self.assertEqual(snapshot["REPO_ROOT"], str(self.cwd.parent))

    def test_handoff_snapshot_from_another_directory(self):
        snapshot = self.handoff_snapshot(self.root, "--cwd", str(self.cwd))
        self.assertEqual(snapshot["PERSIST_DIR"], str(self.cwd / ".humblskills" / "handoffs"))
        self.assertEqual(snapshot["SESSION_START_DIR"], str(self.cwd))

    def test_handoff_ignore_check_uses_workstream_directory(self):
        subprocess.run(["git", "init", "-q", str(self.cwd.parent)], check=True, capture_output=True)
        (self.cwd.parent / ".gitignore").write_text("billing/.humblskills/\n")
        snapshot = self.handoff_snapshot(self.cwd)
        self.assertEqual(snapshot["PERSIST_DIR_GITIGNORED"], "yes")

    def test_handoff_pickup_roundtrip_without_git(self):
        snapshot = self.handoff_snapshot(self.cwd)
        path = Path(snapshot["PERSIST_FILE"])
        path.parent.mkdir(parents=True)
        path.write_text("# Billing handoff\n")
        self.assertEqual(pickup.context(self.cwd, self.config)["selected_handoff"], str(path))
        self.assertEqual(snapshot["REPO_ROOT"], "none")


if __name__ == "__main__":
    unittest.main()
