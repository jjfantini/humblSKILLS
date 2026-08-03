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

func TestUninstall_GlobalFanoutRemovesLinksAndCanonicalStore(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableClaudeCode(t, s)
	enableCodex(t, s)

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{
			Name: "foo", Version: "1.0.0",
			Files: testutil.SkillTree{"SKILL.md": sampleSkillMD},
		},
	})
	installRes := runCLIWithStdoutCapture(t,
		"install", "foo",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--manifest", s.ManifestPath,
		"--global",
		"--yes", "--json",
	)
	if installRes.RunErr != nil {
		t.Fatalf("install: %v\n%s", installRes.RunErr, installRes.Err)
	}

	canonical := filepath.Join(s.Home, ".humblskills", "skills", "foo")
	claudeLink := filepath.Join(s.Home, ".claude", "skills", "foo")
	codexLink := filepath.Join(s.Home, ".agents", "skills", "foo")

	res := runCLIWithStdoutCapture(t,
		"uninstall", "foo",
		"--manifest", s.ManifestPath,
		"--cache-dir", s.CacheDir,
		"--yes", "--json",
	)
	if res.RunErr != nil {
		t.Fatalf("uninstall: %v\n%s", res.RunErr, res.Err)
	}
	for _, path := range []string{canonical, claudeLink, codexLink} {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("%s should be removed, err=%v", path, err)
		}
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

// preserveSkillMD declares user-owned memory files, the thing every
// destructive path in this CLI has to stop and ask about.
const preserveSkillMD = `---
name: foo
description: Example skill with user-owned memory
metadata:
  version: 1.0.0
  preserve:
    - references/log.md
---

# foo
`

// installFooWithMemory installs `foo` globally and writes real content into its
// preserved file, returning (registry URL, canonical store path).
func installFooWithMemory(t *testing.T, s *testutil.Sandbox) (string, string) {
	t.Helper()
	enableClaudeCode(t, s)
	enableCodex(t, s)

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{{
		Name: "foo", Version: "1.0.0",
		Files: testutil.SkillTree{
			"SKILL.md":          preserveSkillMD,
			"references/log.md": "seed\n",
		},
	}})
	res := runCLIWithStdoutCapture(t,
		"install", "foo",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--manifest", s.ManifestPath,
		"--platform", "claude-code",
		"--global",
		"--yes", "--json",
	)
	if res.RunErr != nil {
		t.Fatalf("install failed: %v\n%s", res.RunErr, res.Err)
	}
	m, _ := manifest.Load(s.ManifestPath)
	if len(m.Installations) == 0 {
		t.Fatal("precondition: expected an install")
	}
	store := m.Installations[0].StorePath
	if store == "" {
		t.Fatal("precondition: expected a canonical store")
	}
	if err := os.WriteFile(filepath.Join(store, "references", "log.md"),
		[]byte("seed\n[SESSION] user work\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return regURL, store
}

// A non-interactive uninstall used to delete everything on an implied yes.
// Now it refuses, names the files it would have destroyed, and changes nothing.
func TestUninstall_NonInteractiveRequiresYes(t *testing.T) {
	s := testutil.NewSandbox(t)
	_, store := installFooWithMemory(t, s)

	res := runCLIWithStdoutCapture(t,
		"uninstall", "foo",
		"--manifest", s.ManifestPath,
		"--cache-dir", s.CacheDir,
		"--json",
	)
	if res.RunErr == nil {
		t.Fatal("uninstall without --yes must not proceed non-interactively")
	}
	msg := res.RunErr.Error()
	if !strings.Contains(msg, "--yes") {
		t.Errorf("error should name the flag to pass, got %q", msg)
	}
	if !strings.Contains(msg, "references/log.md") {
		t.Errorf("error should name the user-owned file at risk, got %q", msg)
	}
	if _, err := os.Stat(filepath.Join(store, "references", "log.md")); err != nil {
		t.Errorf("refused uninstall must not delete anything: %v", err)
	}
	m, _ := manifest.Load(s.ManifestPath)
	if len(m.Installations) == 0 {
		t.Error("refused uninstall must leave the manifest alone")
	}
}

// --force is the documented way to discard local content, so it must also be
// gated: same rule, same error, nothing touched.
func TestInstallForce_NonInteractiveRequiresYes(t *testing.T) {
	s := testutil.NewSandbox(t)
	regURL, store := installFooWithMemory(t, s)
	logPath := filepath.Join(store, "references", "log.md")
	before, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}

	res := runCLIWithStdoutCapture(t,
		"install", "foo", "--force",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--manifest", s.ManifestPath,
		"--platform", "claude-code",
		"--global",
		"--json",
	)
	if res.RunErr == nil {
		t.Fatal("install --force without --yes must not proceed non-interactively")
	}
	if !strings.Contains(res.RunErr.Error(), "references/log.md") {
		t.Errorf("error should name what --force discards, got %q", res.RunErr.Error())
	}
	after, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Error("refused --force must not overwrite preserved content")
	}
}

// Adding a platform to an installed skill is not destructive, so it must NOT
// require --yes and must not disturb preserved content.
func TestInstall_AddPlatformNeedsNoConfirmation(t *testing.T) {
	s := testutil.NewSandbox(t)
	regURL, store := installFooWithMemory(t, s)
	logPath := filepath.Join(store, "references", "log.md")
	before, _ := os.ReadFile(logPath)

	res := runCLIWithStdoutCapture(t,
		"install", "foo",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--manifest", s.ManifestPath,
		"--platform", "codex",
		"--global",
		"--json",
	)
	if res.RunErr != nil {
		t.Fatalf("adding a platform should just work: %v\n%s", res.RunErr, res.Err)
	}
	assertContains(t, res.Out, `"outcome": "linked"`)

	after, _ := os.ReadFile(logPath)
	if string(after) != string(before) {
		t.Errorf("platform add overwrote preserved content:\n got %q\nwant %q", after, before)
	}
	m, _ := manifest.Load(s.ManifestPath)
	if len(m.Installations) != 2 {
		t.Errorf("want an entry per platform, got %+v", m.Installations)
	}
}
