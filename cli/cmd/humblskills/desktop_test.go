package main

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

func installFixtureSkill(t *testing.T, s *testutil.Sandbox, name string) {
	t.Helper()
	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{
			Name: name, Version: "1.0.0",
			Platforms: []string{"claude-code"},
			Files: testutil.SkillTree{
				"SKILL.md":            sampleSkillMD,
				"references/notes.md": "# notes\n",
			},
		},
	})
	res := runCLIWithStdoutCapture(t,
		"install", name,
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--manifest", s.ManifestPath,
		"--profile", s.ProfilePath,
		"--platform", "claude-code",
		"--scope", "user",
		"--yes",
	)
	if res.RunErr != nil {
		t.Fatalf("install: %v\n%s", res.RunErr, res.Err)
	}
}

func zipNames(t *testing.T, path string) []string {
	t.Helper()
	zr, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("open zip %s: %v", path, err)
	}
	defer zr.Close()
	var names []string
	for _, f := range zr.File {
		names = append(names, f.Name)
	}
	return names
}

func TestExportDesktop_WritesUploadReadyZip(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableClaudeCode(t, s)
	installFixtureSkill(t, s, "foo")

	outDir := filepath.Join(s.Root, "zips")
	res := runCLIWithStdoutCapture(t,
		"export", "desktop", "foo",
		"--manifest", s.ManifestPath,
		"--profile", s.ProfilePath,
		"-o", outDir,
		"--yes",
	)
	if res.RunErr != nil {
		t.Fatalf("export desktop: %v\n%s", res.RunErr, res.Err)
	}

	zipPath := filepath.Join(outDir, "foo-1.0.0.zip")
	names := zipNames(t, zipPath)
	want := map[string]bool{"foo/SKILL.md": false, "foo/references/notes.md": false}
	for _, n := range names {
		if _, ok := want[n]; ok {
			want[n] = true
		}
		if !strings.HasPrefix(n, "foo/") {
			t.Errorf("entry %q not under the skill folder at zip root", n)
		}
	}
	for n, seen := range want {
		if !seen {
			t.Errorf("zip missing %q; entries: %v", n, names)
		}
	}
	assertContains(t, res.Out+res.Err, "upload at claude.ai")
}

func TestExportDesktop_NoArgs_CoversAllInstalled(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableClaudeCode(t, s)
	installFixtureSkill(t, s, "foo")

	outDir := filepath.Join(s.Root, "zips")
	res := runCLIWithStdoutCapture(t,
		"export", "desktop",
		"--manifest", s.ManifestPath,
		"--profile", s.ProfilePath,
		"-o", outDir,
		"--json", "--yes",
	)
	if res.RunErr != nil {
		t.Fatalf("export desktop: %v\n%s", res.RunErr, res.Err)
	}
	idx := strings.Index(res.Out, "{")
	var out struct {
		Zips []struct {
			Skill string `json:"skill"`
			Zip   string `json:"zip"`
		} `json:"zips"`
	}
	if err := json.Unmarshal([]byte(res.Out[idx:]), &out); err != nil {
		t.Fatalf("parse: %v\n%s", err, res.Out)
	}
	if len(out.Zips) != 1 || out.Zips[0].Skill != "foo" {
		t.Fatalf("zips = %+v", out.Zips)
	}
	if _, err := os.Stat(out.Zips[0].Zip); err != nil {
		t.Errorf("zip not on disk: %v", err)
	}
}

func TestExportDesktop_NotInstalled_Errors(t *testing.T) {
	s := testutil.NewSandbox(t)
	res := runCLIWithStdoutCapture(t,
		"export", "desktop", "ghost",
		"--manifest", s.ManifestPath,
		"--profile", s.ProfilePath,
		"--yes",
	)
	if res.RunErr == nil {
		t.Fatal("expected error for a skill that isn't installed")
	}
}

func TestInstall_DesktopExportsSetting_WritesZip(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableClaudeCode(t, s)

	if err := os.MkdirAll(filepath.Dir(s.ProfilePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.ProfilePath, []byte(`{"schema_version":1,"desktop_exports":true}`), 0o644); err != nil {
		t.Fatal(err)
	}

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
		"--profile", s.ProfilePath,
		"--platform", "claude-code",
		"--scope", "user",
		"--json", "--yes",
	)
	if res.RunErr != nil {
		t.Fatalf("install: %v\n%s", res.RunErr, res.Err)
	}
	idx := strings.Index(res.Out, "{")
	var out struct {
		DesktopZips []struct {
			Zip string `json:"zip"`
		} `json:"desktop_zips"`
	}
	if err := json.Unmarshal([]byte(res.Out[idx:]), &out); err != nil {
		t.Fatalf("parse: %v\n%s", err, res.Out)
	}
	if len(out.DesktopZips) != 1 {
		t.Fatalf("desktop_zips = %+v, want 1 entry", out.DesktopZips)
	}
	if _, err := os.Stat(out.DesktopZips[0].Zip); err != nil {
		t.Errorf("auto-exported zip not on disk: %v", err)
	}
	names := zipNames(t, out.DesktopZips[0].Zip)
	if len(names) == 0 || !strings.HasPrefix(names[0], "foo/") {
		t.Errorf("zip layout wrong: %v", names)
	}
}

func TestInstall_DesktopExportsOff_NoZip(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableClaudeCode(t, s)
	installFixtureSkill(t, s, "foo")

	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".humblskills", "desktop")
	if entries, err := os.ReadDir(dir); err == nil && len(entries) > 0 {
		t.Errorf("desktop dir should be empty with the setting off; got %v", entries)
	}
}
