package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

func TestList_EmptyManifest_JSON(t *testing.T) {
	s := testutil.NewSandbox(t)

	res := runCLIWithStdoutCapture(t,
		"list",
		"--manifest", s.ManifestPath,
		"--json",
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v\n%s", res.RunErr, res.Err)
	}
	// Output should be JSON with zero installations.
	idx := strings.Index(res.Out, "{")
	if idx < 0 {
		t.Fatalf("no JSON in output:\n%s", res.Out)
	}
	var m struct {
		Installations []any `json:"installations"`
	}
	if err := json.Unmarshal([]byte(res.Out[idx:]), &m); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(m.Installations) != 0 {
		t.Errorf("installations = %d, want 0", len(m.Installations))
	}
}

func TestList_PopulatedManifest_JSON(t *testing.T) {
	s := testutil.NewSandbox(t)

	// Seed a manifest with two installations directly via the manifest
	// package, avoiding the need to run install for a list test.
	m := &manifest.Manifest{SchemaVersion: manifest.SchemaVersion}
	m.Upsert(manifest.Installation{
		Skill: "alpha", Version: "1.0.0", Platform: "claude-code",
		Scope: "user", Path: "/tmp/alpha",
		InstalledAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	m.Upsert(manifest.Installation{
		Skill: "beta", Version: "0.9.0", Platform: "cursor-agent",
		Scope: "project", Path: "/tmp/beta",
		InstalledAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
	})
	if err := manifest.Save(s.ManifestPath, m); err != nil {
		t.Fatal(err)
	}

	res := runCLIWithStdoutCapture(t,
		"list",
		"--manifest", s.ManifestPath,
		"--json",
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v\n%s", res.RunErr, res.Err)
	}

	idx := strings.Index(res.Out, "{")
	var got struct {
		Installations []struct {
			Skill    string `json:"skill"`
			Platform string `json:"platform"`
		} `json:"installations"`
	}
	if err := json.Unmarshal([]byte(res.Out[idx:]), &got); err != nil {
		t.Fatalf("parse: %v\n%s", err, res.Out)
	}
	if len(got.Installations) != 2 {
		t.Errorf("installations = %d, want 2", len(got.Installations))
	}
	names := map[string]string{}
	for _, i := range got.Installations {
		names[i.Skill] = i.Platform
	}
	if names["alpha"] != "claude-code" {
		t.Errorf("alpha platform = %q", names["alpha"])
	}
	if names["beta"] != "cursor-agent" {
		t.Errorf("beta platform = %q", names["beta"])
	}
}

func TestList_TextOutput_ShowsSourceColumn(t *testing.T) {
	s := testutil.NewSandbox(t)

	m := &manifest.Manifest{SchemaVersion: manifest.SchemaVersion}
	m.Upsert(manifest.Installation{
		Skill: "foo", Version: "1.0.0", Platform: "claude-code",
		Scope: "user", Path: "/tmp/.claude/skills/foo",
		StorePath:   "/tmp/.humblskills/skills/foo",
		InstalledAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	m.Upsert(manifest.Installation{
		Skill: "foo", Version: "1.0.0", Platform: "cursor",
		Scope: "user", Path: "/tmp/.cursor/skills/foo",
		StorePath:   "/tmp/.humblskills/skills/foo",
		InstalledAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err := manifest.Save(s.ManifestPath, m); err != nil {
		t.Fatal(err)
	}

	res := runCLIWithStdoutCapture(t,
		"list",
		"--manifest", s.ManifestPath,
		"--yes",
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v\n%s", res.RunErr, res.Err)
	}
	if !strings.Contains(res.Out, "Source") {
		t.Errorf("table should have a Source column header:\n%s", res.Out)
	}
	if !strings.Contains(res.Out, "/tmp/.humblskills/skills/foo") {
		t.Errorf("table should show the canonical store path:\n%s", res.Out)
	}
	if !strings.Contains(res.Out, "claude-code") || !strings.Contains(res.Out, "cursor") {
		t.Errorf("table should show every symlinked platform:\n%s", res.Out)
	}
}

func TestList_EmptyManifest_TextOutputHint(t *testing.T) {
	s := testutil.NewSandbox(t)

	// --yes keeps TUI off so we get the plain text output.
	res := runCLIWithStdoutCapture(t,
		"list",
		"--manifest", s.ManifestPath,
		"--yes",
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v\n%s", res.RunErr, res.Err)
	}
	if !strings.Contains(res.Out+res.Err, "no skills installed") {
		t.Errorf("expected helpful hint, got:\n%s\n---\n%s", res.Out, res.Err)
	}
}
