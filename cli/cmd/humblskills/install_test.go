package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

// sampleSkillMD is the minimum valid SKILL.md body the install tests use.
const sampleSkillMD = `---
name: foo
description: Example skill for install tests
version: 1.0.0
---

# foo

Body.
`

func enableClaudeCode(t *testing.T, s *testutil.Sandbox) {
	t.Helper()
	// Create ~/.claude so the claude-code adapter's path_exists rule
	// matches under the sandboxed HOME.
	if err := os.MkdirAll(filepath.Join(s.Home, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestInstall_HappyPath_SingleSkill(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableClaudeCode(t, s)

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{
			Name: "foo", Version: "1.0.0",
			Description: "test skill",
			Platforms:   []string{"claude-code"},
			Files: testutil.SkillTree{
				"SKILL.md": sampleSkillMD,
			},
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
		t.Fatalf("run: %v\nerr: %s", res.RunErr, res.Err)
	}

	// Manifest should now have one installation under claude-code/user.
	m, err := manifest.Load(s.ManifestPath)
	if err != nil {
		t.Fatalf("manifest: %v", err)
	}
	if len(m.Installations) != 1 {
		t.Fatalf("installs = %d, body=%s", len(m.Installations), res.Out)
	}
	got := m.Installations[0]
	if got.Skill != "foo" || got.Platform != "claude-code" || got.Scope != "user" {
		t.Errorf("unexpected install: %+v", got)
	}
	if _, err := os.Stat(filepath.Join(got.Path, "SKILL.md")); err != nil {
		t.Errorf("installed SKILL.md missing: %v", err)
	}
}

func TestInstall_IdempotentWhenUpToDate(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableClaudeCode(t, s)

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{
			Name: "foo", Version: "1.0.0",
			Platforms: []string{"claude-code"},
			Files:     testutil.SkillTree{"SKILL.md": sampleSkillMD},
		},
	})

	runOnce := func() execResult {
		return runCLIWithStdoutCapture(t,
			"install", "foo",
			"--registry", regURL,
			"--cache-dir", s.CacheDir,
			"--manifest", s.ManifestPath,
			"--platform", "claude-code",
			"--scope", "user",
			"--yes", "--json",
		)
	}
	first := runOnce()
	if first.RunErr != nil {
		t.Fatalf("first run: %v\n%s", first.RunErr, first.Err)
	}

	second := runOnce()
	if second.RunErr != nil {
		t.Fatalf("second run: %v\n%s", second.RunErr, second.Err)
	}

	// Every target on the second run must be "skipped".
	extractJSON := func(s string) string {
		idx := strings.Index(s, "{")
		return s[idx:]
	}
	var payload struct {
		Results []struct {
			Outcome string `json:"outcome"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(extractJSON(second.Out)), &payload); err != nil {
		t.Fatalf("parse json: %v\n%s", err, second.Out)
	}
	if len(payload.Results) == 0 {
		t.Fatal("second run reported no results")
	}
	for _, r := range payload.Results {
		if r.Outcome != "skipped" {
			t.Errorf("second run outcome = %q, want skipped", r.Outcome)
		}
	}
}

func TestInstall_UnknownSkillErrors(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableClaudeCode(t, s)

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{
			Name: "foo", Version: "1.0.0",
			Platforms: []string{"claude-code"},
			Files:     testutil.SkillTree{"SKILL.md": sampleSkillMD},
		},
	})
	res := runCLIWithStdoutCapture(t,
		"install", "ghost",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--manifest", s.ManifestPath,
		"--platform", "claude-code",
		"--yes", "--json",
	)
	if res.RunErr == nil {
		t.Fatal("expected error for unknown skill")
	}
}

func TestInstall_UnknownPlatformErrors(t *testing.T) {
	s := testutil.NewSandbox(t)
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
		"--platform", "ghost-platform",
		"--yes", "--json",
	)
	if res.RunErr == nil {
		t.Fatal("expected error for unknown platform")
	}
	if !strings.Contains(res.RunErr.Error(), "unknown platform") {
		t.Errorf("err: %v", res.RunErr)
	}
}

func TestInstall_ForceReinstallsUpToDateTarget(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableClaudeCode(t, s)

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{
			Name: "foo", Version: "1.0.0",
			Platforms: []string{"claude-code"},
			Files:     testutil.SkillTree{"SKILL.md": sampleSkillMD},
		},
	})

	args := []string{
		"install", "foo",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--manifest", s.ManifestPath,
		"--platform", "claude-code",
		"--scope", "user",
		"--yes", "--json",
	}
	if r := runCLIWithStdoutCapture(t, args...); r.RunErr != nil {
		t.Fatalf("first run: %v\n%s", r.RunErr, r.Err)
	}

	// Second run with --force should report "forced" outcomes.
	forceArgs := append(args, "--force")
	res := runCLIWithStdoutCapture(t, forceArgs...)
	if res.RunErr != nil {
		t.Fatalf("force run: %v\n%s", res.RunErr, res.Err)
	}
	if !strings.Contains(res.Out, "forced") {
		t.Errorf("expected forced outcome in output:\n%s", res.Out)
	}
}

func TestInstall_SelectPlatforms_DetectedOnly(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableClaudeCode(t, s)
	// cursor NOT enabled (no ~/.cursor) — it must not appear in selected set.

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{
			Name: "foo", Version: "1.0.0",
			Platforms: []string{"claude-code", "cursor"},
			Files:     testutil.SkillTree{"SKILL.md": sampleSkillMD},
		},
	})

	// Omit --platform → auto-detect must pick only claude-code.
	res := runCLIWithStdoutCapture(t,
		"install", "foo",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--manifest", s.ManifestPath,
		"--scope", "user",
		"--yes", "--json",
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v\nerr: %s", res.RunErr, res.Err)
	}
	m, err := manifest.Load(s.ManifestPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, inst := range m.Installations {
		if inst.Platform == "cursor" {
			t.Errorf("cursor installed despite not being detected: %+v", inst)
		}
	}
	// claude-code must be present.
	found := false
	for _, inst := range m.Installations {
		if inst.Platform == "claude-code" {
			found = true
		}
	}
	if !found {
		t.Error("claude-code missing from manifest")
	}
}

func TestInstall_DefaultsPreferClaudeCodeWhenBothDetected(t *testing.T) {
	// Issue #84: when both Claude Code and Cursor are detected and the user
	// doesn't pass --platform, only claude-code should be installed — Cursor
	// can read ~/.claude/skills natively.
	s := testutil.NewSandbox(t)
	enableClaudeCode(t, s)
	if err := os.MkdirAll(filepath.Join(s.Home, ".cursor"), 0o755); err != nil {
		t.Fatal(err)
	}

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{
			Name: "foo", Version: "1.0.0",
			Platforms: []string{"claude-code", "cursor"},
			Files:     testutil.SkillTree{"SKILL.md": sampleSkillMD},
		},
	})
	res := runCLIWithStdoutCapture(t,
		"install", "foo",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--manifest", s.ManifestPath,
		"--scope", "user",
		"--yes", "--json",
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v\nerr: %s", res.RunErr, res.Err)
	}
	m, err := manifest.Load(s.ManifestPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, inst := range m.Installations {
		if inst.Platform == "cursor" {
			t.Errorf("cursor should not be installed by default when claude-code is also detected: %+v", inst)
		}
	}
	found := false
	for _, inst := range m.Installations {
		if inst.Platform == "claude-code" {
			found = true
		}
	}
	if !found {
		t.Error("claude-code missing from manifest")
	}
}

func TestInstall_ExplicitPlatformBothStillWorks(t *testing.T) {
	// Users who really want both IDEs can still ask for them via --platform.
	s := testutil.NewSandbox(t)
	enableClaudeCode(t, s)
	if err := os.MkdirAll(filepath.Join(s.Home, ".cursor"), 0o755); err != nil {
		t.Fatal(err)
	}

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{
			Name: "foo", Version: "1.0.0",
			Platforms: []string{"claude-code", "cursor"},
			Files:     testutil.SkillTree{"SKILL.md": sampleSkillMD},
		},
	})
	res := runCLIWithStdoutCapture(t,
		"install", "foo",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--manifest", s.ManifestPath,
		"--platform", "claude-code,cursor",
		"--scope", "user",
		"--yes", "--json",
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v\nerr: %s", res.RunErr, res.Err)
	}
	m, err := manifest.Load(s.ManifestPath)
	if err != nil {
		t.Fatal(err)
	}
	platforms := map[string]bool{}
	for _, inst := range m.Installations {
		platforms[inst.Platform] = true
	}
	if !platforms["claude-code"] || !platforms["cursor"] {
		t.Errorf("explicit --platform both should install to both; got %v", platforms)
	}
}

func TestInstall_NoPlatformsDetectedErrors(t *testing.T) {
	s := testutil.NewSandbox(t)
	// Neither ~/.claude nor ~/.cursor exist. Auto-detect yields nothing.

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
		"--scope", "user",
		"--yes", "--json",
	)
	if res.RunErr == nil {
		t.Fatal("expected error when no platforms detected")
	}
}

func TestSelectPlatforms_RequestedSubset(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableClaudeCode(t, s)
	// Cursor also enabled.
	_ = os.MkdirAll(filepath.Join(s.Home, ".cursor"), 0o755)

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{
			Name: "foo", Version: "1.0.0",
			Platforms: []string{"claude-code", "cursor"},
			Files:     testutil.SkillTree{"SKILL.md": sampleSkillMD},
		},
	})

	// Requested only cursor → claude-code must not be installed.
	res := runCLIWithStdoutCapture(t,
		"install", "foo",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--manifest", s.ManifestPath,
		"--platform", "cursor",
		"--scope", "user",
		"--yes", "--json",
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v\n%s", res.RunErr, res.Err)
	}
	m, _ := manifest.Load(s.ManifestPath)
	for _, inst := range m.Installations {
		if inst.Platform == "claude-code" {
			t.Errorf("claude-code installed despite --platform cursor: %+v", inst)
		}
	}
}
