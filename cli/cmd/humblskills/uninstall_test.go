package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

// installFoo is a small helper for uninstall tests — seeds a registry,
// runs a successful install, and returns the regURL so callers can
// drive further CLI runs.
func installFoo(t *testing.T, s *testutil.Sandbox) string {
	t.Helper()
	enableClaudeCode(t, s)

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{
			Name: "foo", Version: "1.0.0",
			Platforms: []string{"claude-code"},
			Files:     testutil.SkillTree{"SKILL.md": sampleSkillMD},
		},
	})
	res := runCLIWithStdoutCapture(t,
		"install", "foo",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--manifest", s.ManifestPath,
		"--platform", "claude-code",
		"--scope", "user",
		"--yes", "--json",
	)
	if res.RunErr != nil {
		t.Fatalf("install failed: %v\n%s", res.RunErr, res.Err)
	}
	return regURL
}

func TestUninstall_RemovesFilesAndManifestEntry(t *testing.T) {
	s := testutil.NewSandbox(t)
	_ = installFoo(t, s)

	// Find install path to confirm it's actually deleted.
	m, _ := manifest.Load(s.ManifestPath)
	if len(m.Installations) == 0 {
		t.Fatal("precondition: expected one install")
	}
	installPath := m.Installations[0].Path

	res := runCLIWithStdoutCapture(t,
		"uninstall", "foo",
		"--manifest", s.ManifestPath,
		"--cache-dir", s.CacheDir,
		"--yes", "--json",
	)
	if res.RunErr != nil {
		t.Fatalf("uninstall: %v\n%s", res.RunErr, res.Err)
	}
	if _, err := os.Stat(filepath.Join(installPath, "SKILL.md")); err == nil {
		t.Error("SKILL.md should be removed after uninstall")
	}
	m2, _ := manifest.Load(s.ManifestPath)
	if len(m2.Installations) != 0 {
		t.Errorf("manifest still has entries: %+v", m2.Installations)
	}
}

func TestUninstall_UnknownSkillWarns(t *testing.T) {
	s := testutil.NewSandbox(t)

	// Seed an empty manifest.
	if err := os.MkdirAll(filepath.Dir(s.ManifestPath), 0o755); err != nil {
		t.Fatal(err)
	}
	_ = manifest.Save(s.ManifestPath, &manifest.Manifest{SchemaVersion: manifest.SchemaVersion})

	res := runCLIWithStdoutCapture(t,
		"uninstall", "ghost",
		"--manifest", s.ManifestPath,
		"--cache-dir", s.CacheDir,
		"--yes",
	)
	if res.RunErr != nil {
		t.Fatalf("uninstall should not error on unknown skill: %v", res.RunErr)
	}
	if !strings.Contains(res.Out+res.Err, "not installed") {
		t.Errorf("expected 'not installed' warning, got:\n%s", res.Out+res.Err)
	}
}

func TestUninstall_PreservesOtherSkills(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableClaudeCode(t, s)

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{
			Name: "foo", Version: "1.0.0", Platforms: []string{"claude-code"},
			Files: testutil.SkillTree{"SKILL.md": sampleSkillMD},
		},
		{
			Name: "bar", Version: "1.0.0", Platforms: []string{"claude-code"},
			Files: testutil.SkillTree{"SKILL.md": sampleSkillMD},
		},
	})
	for _, name := range []string{"foo", "bar"} {
		r := runCLIWithStdoutCapture(t,
			"install", name,
			"--registry", regURL,
			"--cache-dir", s.CacheDir,
			"--manifest", s.ManifestPath,
			"--platform", "claude-code", "--scope", "user",
			"--yes", "--json",
		)
		if r.RunErr != nil {
			t.Fatalf("install %s: %v", name, r.RunErr)
		}
	}

	r := runCLIWithStdoutCapture(t,
		"uninstall", "foo",
		"--manifest", s.ManifestPath,
		"--cache-dir", s.CacheDir,
		"--yes", "--json",
	)
	if r.RunErr != nil {
		t.Fatalf("uninstall: %v", r.RunErr)
	}

	m, _ := manifest.Load(s.ManifestPath)
	if len(m.Installations) != 1 || m.Installations[0].Skill != "bar" {
		t.Errorf("expected only bar remaining: %+v", m.Installations)
	}
}
