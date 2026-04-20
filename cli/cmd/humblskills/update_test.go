package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

// bumpRegistryVersion seeds a new registry where foo has advanced from
// 1.0.0 to 1.0.1 so the update command has drift to detect.
func bumpRegistryVersion(t *testing.T, s *testutil.Sandbox, newBody string) string {
	t.Helper()
	return seedTestRegistry(t, s, []testutil.SkillFixture{
		{
			Name: "foo", Version: "1.0.1", Platforms: []string{"claude-code"},
			Files: testutil.SkillTree{"SKILL.md": newBody},
		},
	})
}

func TestUpdate_NoInstallsInfoMessage(t *testing.T) {
	s := testutil.NewSandbox(t)

	res := runCLIWithStdoutCapture(t,
		"update",
		"--manifest", s.ManifestPath,
		"--cache-dir", s.CacheDir,
		"--registry", "file:///nonexistent/registry.json",
		"--yes",
	)
	if res.RunErr != nil {
		t.Fatalf("update on empty manifest must not error: %v", res.RunErr)
	}
	if !strings.Contains(res.Out+res.Err, "no skills installed") {
		t.Errorf("expected hint message, got:\n%s\n%s", res.Out, res.Err)
	}
}

func TestUpdate_AllUpToDate_InfoMessage(t *testing.T) {
	s := testutil.NewSandbox(t)
	regURL := installFoo(t, s)

	// Update against the SAME registry — nothing drifted.
	res := runCLIWithStdoutCapture(t,
		"update",
		"--manifest", s.ManifestPath,
		"--cache-dir", s.CacheDir,
		"--registry", regURL,
		"--yes",
	)
	if res.RunErr != nil {
		t.Fatalf("update: %v", res.RunErr)
	}
	if !strings.Contains(res.Out+res.Err, "up-to-date") {
		t.Errorf("expected up-to-date message, got:\n%s\n%s", res.Out, res.Err)
	}
}

func TestUpdate_Check_ReportsDrift(t *testing.T) {
	s := testutil.NewSandbox(t)
	_ = installFoo(t, s)

	newBody := strings.Replace(sampleSkillMD, "version: 1.0.0", "version: 1.0.1", 1)
	regURL := bumpRegistryVersion(t, s, newBody)

	res := runCLIWithStdoutCapture(t,
		"update", "--check",
		"--manifest", s.ManifestPath,
		"--cache-dir", s.CacheDir,
		"--registry", regURL,
		"--json",
	)
	if res.RunErr != nil {
		t.Fatalf("update --check: %v", res.RunErr)
	}
	idx := strings.Index(res.Out, "{")
	var out struct {
		Updates []struct {
			Skill       string `json:"skill"`
			FromVersion string `json:"from_version"`
			ToVersion   string `json:"to_version"`
		} `json:"updates"`
	}
	if err := json.Unmarshal([]byte(res.Out[idx:]), &out); err != nil {
		t.Fatalf("parse: %v\n%s", err, res.Out)
	}
	if len(out.Updates) != 1 || out.Updates[0].Skill != "foo" {
		t.Errorf("updates = %+v", out.Updates)
	}
	if out.Updates[0].FromVersion != "1.0.0" || out.Updates[0].ToVersion != "1.0.1" {
		t.Errorf("versions = %+v", out.Updates[0])
	}
}

func TestUpdate_AppliesDrift(t *testing.T) {
	s := testutil.NewSandbox(t)
	_ = installFoo(t, s)

	newBody := strings.Replace(sampleSkillMD, "version: 1.0.0", "version: 1.0.1", 1)
	regURL := bumpRegistryVersion(t, s, newBody)

	res := runCLIWithStdoutCapture(t,
		"update",
		"--manifest", s.ManifestPath,
		"--cache-dir", s.CacheDir,
		"--registry", regURL,
		"--yes", "--all", "--json",
	)
	if res.RunErr != nil {
		t.Fatalf("update: %v\n%s", res.RunErr, res.Err)
	}

	m, _ := manifest.Load(s.ManifestPath)
	if len(m.Installations) != 1 {
		t.Fatalf("installs = %d", len(m.Installations))
	}
	if m.Installations[0].Version != "1.0.1" {
		t.Errorf("version not bumped: %+v", m.Installations[0])
	}
}

func TestUpdate_Check_AllUpToDateNonJSON(t *testing.T) {
	s := testutil.NewSandbox(t)
	regURL := installFoo(t, s)
	// Check against same registry.
	res := runCLIWithStdoutCapture(t,
		"update", "--check",
		"--manifest", s.ManifestPath,
		"--cache-dir", s.CacheDir,
		"--registry", regURL,
		"--yes",
	)
	if res.RunErr != nil {
		t.Fatalf("update --check: %v", res.RunErr)
	}
	if !strings.Contains(res.Out+res.Err, "up-to-date") {
		t.Errorf("expected up-to-date, got:\n%s", res.Out+res.Err)
	}
}

func TestUpdate_OnlyNamedSkillUpToDate(t *testing.T) {
	s := testutil.NewSandbox(t)
	regURL := installFoo(t, s)

	res := runCLIWithStdoutCapture(t,
		"update", "foo",
		"--manifest", s.ManifestPath,
		"--cache-dir", s.CacheDir,
		"--registry", regURL,
		"--yes",
	)
	if res.RunErr != nil {
		t.Fatalf("update: %v", res.RunErr)
	}
	if !strings.Contains(res.Out+res.Err, "up-to-date") {
		t.Errorf("expected up-to-date message, got:\n%s", res.Out+res.Err)
	}
}
