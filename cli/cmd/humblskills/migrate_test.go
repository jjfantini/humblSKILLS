package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

const migratableSkillMD = `---
name: foo
description: Example skill for migration tests
metadata:
  preserve:
    - references/
---

# foo
`

func TestMigrate_ClaudeCodeRegistrySkillToGlobalFanout(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableClaudeCode(t, s)
	enableCodex(t, s)

	claudeFoo := filepath.Join(s.Home, ".claude", "skills", "foo")
	if err := os.MkdirAll(filepath.Join(claudeFoo, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeFoo, "SKILL.md"), []byte(migratableSkillMD), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeFoo, "references", "log.md"), []byte("local brain\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	claudeUnknown := filepath.Join(s.Home, ".claude", "skills", "personal")
	if err := os.MkdirAll(claudeUnknown, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeUnknown, "SKILL.md"), []byte(strings.Replace(migratableSkillMD, "name: foo", "name: personal", 1)), 0o644); err != nil {
		t.Fatal(err)
	}

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{
			Name:        "foo",
			Version:     "1.0.0",
			Description: "test skill",
			Platforms:   []string{"claude-code"},
			Preserve:    []string{"references/"},
			Files: testutil.SkillTree{
				"SKILL.md": migratableSkillMD,
			},
		},
	})

	res := runCLIWithStdoutCapture(t,
		"migrate", "claude-code",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--manifest", s.ManifestPath,
		"--global",
		"--yes", "--json",
	)
	if res.RunErr != nil {
		t.Fatalf("migrate: %v\nerr: %s\nout: %s", res.RunErr, res.Err, res.Out)
	}
	if !strings.Contains(res.Out, "foo") {
		t.Fatalf("migrate output should include foo: %s", res.Out)
	}
	if !strings.Contains(res.Out, "personal") {
		t.Fatalf("migrate output should report skipped personal skill: %s", res.Out)
	}

	canonical := filepath.Join(s.Home, ".humblskills", "skills", "foo")
	got, err := os.ReadFile(filepath.Join(canonical, "references", "log.md"))
	if err != nil {
		t.Fatalf("preserved brain file missing: %v", err)
	}
	if string(got) != "local brain\n" {
		t.Fatalf("preserved brain file = %q", got)
	}
	if !targetIsSymlinkTo(t, filepath.Join(s.Home, ".claude", "skills", "foo"), canonical) {
		t.Fatal("claude target should be a symlink to canonical store")
	}
	if !targetIsSymlinkTo(t, filepath.Join(s.Home, ".agents", "skills", "foo"), canonical) {
		t.Fatal("codex target should be a symlink to canonical store")
	}

	m, err := manifest.Load(s.ManifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Installations) != 2 {
		t.Fatalf("manifest installs = %d, want 2: %+v", len(m.Installations), m.Installations)
	}
}

func targetIsSymlinkTo(t *testing.T, path, want string) bool {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("lstat %s: %v", path, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return false
	}
	got, err := os.Readlink(path)
	if err != nil {
		t.Fatalf("readlink %s: %v", path, err)
	}
	return got == want
}
