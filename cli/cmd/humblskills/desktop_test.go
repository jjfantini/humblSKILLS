package main

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/v2/internal/install"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/testutil"
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

	zipPath := filepath.Join(outDir, "foo.zip")
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

func TestInstall_ClaudeDesktopPlatform_WritesZip(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableClaudeCode(t, s)

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{
			Name: "foo", Version: "1.0.0",
			Platforms: []string{"claude-code", "claude-desktop"},
			Files:     testutil.SkillTree{"SKILL.md": sampleSkillMD},
		},
	})
	res := runCLIWithStdoutCapture(t,
		"install", "foo",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--manifest", s.ManifestPath,
		"--profile", s.ProfilePath,
		"--platform", "claude-code,claude-desktop",
		"--scope", "user",
		"--yes",
	)
	if res.RunErr != nil {
		t.Fatalf("install: %v\n%s", res.RunErr, res.Err)
	}

	m, err := manifest.Load(s.ManifestPath)
	if err != nil {
		t.Fatal(err)
	}
	desktop := m.FindOne("foo", "claude-desktop", "user")
	if desktop == nil {
		t.Fatalf("no claude-desktop manifest entry; manifest: %+v", m.Installations)
	}
	if desktop.InstallMode != install.InstallModeZip {
		t.Errorf("InstallMode = %q, want zip", desktop.InstallMode)
	}
	if !strings.HasSuffix(desktop.Path, "foo.zip") {
		t.Errorf("Path = %q, want …/foo.zip", desktop.Path)
	}
	names := zipNames(t, desktop.Path)
	if len(names) == 0 || !strings.HasPrefix(names[0], "foo/") {
		t.Errorf("zip layout wrong: %v", names)
	}
	assertContains(t, res.Out+res.Err, "upload at claude.ai")

	// Second install is idempotent for the zip target too.
	res = runCLIWithStdoutCapture(t,
		"install", "foo",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--manifest", s.ManifestPath,
		"--profile", s.ProfilePath,
		"--platform", "claude-desktop",
		"--scope", "user",
		"--json", "--yes",
	)
	if res.RunErr != nil {
		t.Fatalf("reinstall: %v\n%s", res.RunErr, res.Err)
	}
	assertContains(t, res.Out, "\"outcome\": \"skipped\"")

	// Uninstall removes the zip.
	if res := runCLIWithStdoutCapture(t,
		"uninstall", "foo",
		"--manifest", s.ManifestPath, "--profile", s.ProfilePath, "--yes",
	); res.RunErr != nil {
		t.Fatalf("uninstall: %v\n%s", res.RunErr, res.Err)
	}
	if _, err := os.Stat(desktop.Path); !os.IsNotExist(err) {
		t.Errorf("zip still on disk after uninstall: %v", err)
	}
}
