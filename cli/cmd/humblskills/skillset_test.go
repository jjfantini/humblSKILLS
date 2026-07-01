package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/skillset"
	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

func TestExport_WritesSkillset(t *testing.T) {
	s := testutil.NewSandbox(t)
	_ = installFoo(t, s) // foo 1.0.0 installed

	out := filepath.Join(s.Root, "humblskills.json")
	res := runCLIWithStdoutCapture(t,
		"export",
		"--manifest", s.ManifestPath,
		"--output", out,
		"--yes",
	)
	if res.RunErr != nil {
		t.Fatalf("export: %v\n%s", res.RunErr, res.Err)
	}
	set, err := skillset.Load(out)
	if err != nil {
		t.Fatalf("load exported skillset: %v", err)
	}
	if len(set.Skills) != 1 || set.Skills[0].Name != "foo" {
		t.Fatalf("unexpected skillset: %+v", set.Skills)
	}
	if set.Skills[0].Version != "1.0.0" {
		t.Errorf("foo version = %q, want 1.0.0", set.Skills[0].Version)
	}
}

func TestExport_EmptyManifest_Errors(t *testing.T) {
	s := testutil.NewSandbox(t)
	res := runCLIWithStdoutCapture(t,
		"export",
		"--manifest", s.ManifestPath,
		"--output", filepath.Join(s.Root, "out.json"),
		"--yes",
	)
	if res.RunErr == nil {
		t.Fatal("expected error exporting an empty manifest")
	}
}

func TestSync_InstallsSkillsFromFile(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableClaudeCode(t, s)

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{
			Name: "foo", Version: "1.0.0",
			Platforms: []string{"claude-code"},
			Files:     testutil.SkillTree{"SKILL.md": sampleSkillMD},
		},
	})

	setPath := filepath.Join(s.Root, "humblskills.json")
	if err := os.WriteFile(setPath, []byte(`{"schema_version":1,"skills":[{"name":"foo"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	res := runCLIWithStdoutCapture(t,
		"sync", setPath,
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--manifest", s.ManifestPath,
		"--platform", "claude-code",
		"--scope", "user",
		"--yes",
	)
	if res.RunErr != nil {
		t.Fatalf("sync: %v\n%s", res.RunErr, res.Err)
	}

	m, err := manifest.Load(s.ManifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.FindAll("foo")) == 0 {
		t.Errorf("expected foo to be installed after sync; manifest: %+v", m.Installations)
	}
}

func TestSync_UnknownSkill_WarnsNonFatal(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableClaudeCode(t, s)

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{
			Name: "foo", Version: "1.0.0",
			Platforms: []string{"claude-code"},
			Files:     testutil.SkillTree{"SKILL.md": sampleSkillMD},
		},
	})

	setPath := filepath.Join(s.Root, "humblskills.json")
	if err := os.WriteFile(setPath, []byte(`{"schema_version":1,"skills":[{"name":"foo"},{"name":"ghost"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	res := runCLIWithStdoutCapture(t,
		"sync", setPath,
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--manifest", s.ManifestPath,
		"--platform", "claude-code",
		"--scope", "user",
		"--yes",
	)
	// Unknown skills are warnings, not hard errors — the rest still syncs.
	if res.RunErr != nil {
		t.Fatalf("sync should not fail on unknown skill: %v\n%s", res.RunErr, res.Err)
	}
	if !strings.Contains(res.Out+res.Err, "ghost") {
		t.Errorf("expected a warning naming the missing skill:\n%s\n%s", res.Out, res.Err)
	}
	m, _ := manifest.Load(s.ManifestPath)
	if len(m.FindAll("foo")) == 0 {
		t.Errorf("foo should still install despite ghost being missing")
	}
}

func TestSync_MissingFile_Errors(t *testing.T) {
	s := testutil.NewSandbox(t)
	res := runCLIWithStdoutCapture(t,
		"sync", filepath.Join(s.Root, "nope.json"),
		"--manifest", s.ManifestPath,
		"--cache-dir", s.CacheDir,
		"--registry", "file:///nonexistent/registry.json",
		"--yes",
	)
	if res.RunErr == nil {
		t.Fatal("expected error when skillset file is missing")
	}
}
