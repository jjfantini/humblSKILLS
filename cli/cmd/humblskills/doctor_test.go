package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// Doctor tests exercise the JSON output path (deterministic) rather
// than the TUI. --json plus an exit code of 0 means "nothing failed".

func extractDoctorJSON(t *testing.T, out string) doctorReport {
	t.Helper()
	idx := strings.Index(out, "{")
	if idx < 0 {
		t.Fatalf("no JSON in doctor output:\n%s", out)
	}
	var r doctorReport
	if err := json.Unmarshal([]byte(out[idx:]), &r); err != nil {
		t.Fatalf("parse doctor JSON: %v\n%s", err, out)
	}
	return r
}

func TestDoctor_ReportsAdapters(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableClaudeCode(t, s)

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{Name: "foo", Version: "1.0.0",
			Files: testutil.SkillTree{"SKILL.md": sampleSkillMD}},
	})
	testutil.UseFakeKeyring(t)

	res := runCLIWithStdoutCapture(t,
		"doctor",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--manifest", s.ManifestPath,
		"--profile", s.ProfilePath,
		"--json",
	)
	// doctor returns errDoctorFailed when any adapter isn't detected,
	// which is fine for this run — we only verify the report shape.
	got := extractDoctorJSON(t, res.Out)

	// claude-code enabled in sandbox should be Detected=true.
	found := false
	for _, a := range got.Adapters {
		if a.Name == "claude-code" {
			found = true
			if !a.Detected {
				t.Errorf("claude-code not detected: %+v", a)
			}
		}
	}
	if !found {
		t.Errorf("claude-code absent from doctor report")
	}
	if len(got.Registries) != 1 || got.Registries[0].Skills != 1 {
		t.Errorf("registries = %+v, want one with 1 skill", got.Registries)
	}
}

func TestAdapterItem_DetailHumanizesReasonAndExplainsBadges(t *testing.T) {
	th := ui.NewTheme(ui.DefaultPalette(), nil, true)
	it := adapterItem{a: adapterReport{
		Name:     "cursor",
		Detected: true,
		Reason:   "found ~/.cursor",
		Targets: []targetReport{
			{Scope: "user", Path: "/home/x/.cursor/skills", Writable: true},
			{Scope: "project", Path: "/ro/.cursor/skills", Writable: false},
		},
	}}
	d := it.Detail(th, 60)
	for _, want := range []string{"reason", "found ~/.cursor", "rw = writable", "read-only"} {
		if !strings.Contains(d, want) {
			t.Errorf("adapter detail missing %q:\n%s", want, d)
		}
	}
	if strings.Contains(d, "matched on") {
		t.Errorf("stale 'matched on' label still present:\n%s", d)
	}
}

func TestDoctor_DetectsCorruptManifest(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableClaudeCode(t, s)
	testutil.UseFakeKeyring(t)

	// Write a garbage manifest so doctor surfaces a load error.
	s.WriteFile(t, "xdg/state/humblskills/manifest.json", []byte("{not json"))

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{Name: "foo", Version: "1.0.0",
			Files: testutil.SkillTree{"SKILL.md": sampleSkillMD}},
	})

	res := runCLIWithStdoutCapture(t,
		"doctor",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--manifest", s.ManifestPath,
		"--profile", s.ProfilePath,
		"--json",
	)
	got := extractDoctorJSON(t, res.Out)
	found := false
	for _, iss := range got.Issues {
		if strings.Contains(iss, "manifest") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected manifest issue, got: %+v", got.Issues)
	}
}

func TestDoctor_RegistryUnreachableSurfacedAsIssue(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableClaudeCode(t, s)
	testutil.UseFakeKeyring(t)

	res := runCLIWithStdoutCapture(t,
		"doctor",
		// Point at a definitely-missing file URL.
		"--registry", "file:///nonexistent/registry.json",
		"--cache-dir", s.CacheDir,
		"--manifest", s.ManifestPath,
		"--profile", s.ProfilePath,
		"--json",
	)
	got := extractDoctorJSON(t, res.Out)
	if len(got.Registries) == 0 || got.Registries[0].Error == "" {
		t.Error("expected registry error on unreachable URL")
	}
}

func TestDoctor_JSONShapeSurvives(t *testing.T) {
	s := testutil.NewSandbox(t)
	enableClaudeCode(t, s)
	testutil.UseFakeKeyring(t)

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{Name: "foo", Version: "1.0.0",
			Files: testutil.SkillTree{"SKILL.md": sampleSkillMD}},
	})

	res := runCLIWithStdoutCapture(t,
		"doctor",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--manifest", s.ManifestPath,
		"--profile", s.ProfilePath,
		"--json",
	)
	got := extractDoctorJSON(t, res.Out)

	// Contract keys the report consumers (doctor --json, dashboards) rely on.
	if got.Eval.DefaultRunner == "" {
		// Empty is valid when no runner is detected, but Runners slice
		// must still be populated.
	}
	if got.Eval.Runners == nil {
		t.Error("Eval.Runners should be non-nil slice")
	}
	if got.Eval.APIKeys == nil {
		t.Error("Eval.APIKeys should be non-nil slice")
	}
}
