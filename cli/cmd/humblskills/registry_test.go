package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/v2/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/testutil"
)

func TestRegistryRefresh_FileURLReportsFileOrigin(t *testing.T) {
	s := testutil.NewSandbox(t)

	regURL := seedTestRegistry(t, s, []testutil.SkillFixture{
		{Name: "foo", Version: "1.0.0",
			Files: testutil.SkillTree{"SKILL.md": sampleSkillMD}},
	})

	res := runCLIWithStdoutCapture(t,
		"registry", "refresh",
		"--registry", regURL,
		"--cache-dir", s.CacheDir,
		"--json",
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v\n%s", res.RunErr, res.Err)
	}
	idx := strings.Index(res.Out, "{")
	if idx < 0 {
		t.Fatalf("no JSON in output:\n%s", res.Out)
	}
	var out struct {
		URL    string `json:"url"`
		Source string `json:"source"`
		Skills int    `json:"skills"`
	}
	if err := json.Unmarshal([]byte(res.Out[idx:]), &out); err != nil {
		t.Fatalf("parse: %v\n%s", err, res.Out)
	}
	if out.Source != "file" {
		t.Errorf("source = %q, want file", out.Source)
	}
	if out.Skills != 1 {
		t.Errorf("skills = %d", out.Skills)
	}
}

// Naming the first registry must not silently drop the registry that was already
// in effect: resolvedRegistries() returns ONLY the named set once it is non-empty,
// so `registry add work …` on a fresh profile used to hide every public skill.
func TestRegistryAdd_SeedsPreviouslyEffectiveRegistry(t *testing.T) {
	t.Run("fresh profile keeps the hosted default as public", func(t *testing.T) {
		s := testutil.NewSandbox(t)

		res := runCLIWithStdoutCapture(t,
			"registry", "add", "work", "https://example.com/work/registry.json",
			"--profile", s.ProfilePath,
			"--cache-dir", s.CacheDir,
			"--json",
		)
		if res.RunErr != nil {
			t.Fatalf("run: %v\n%s", res.RunErr, res.Err)
		}

		p, err := profile.Load(s.ProfilePath)
		if err != nil {
			t.Fatalf("load profile: %v", err)
		}
		got := map[string]string{}
		for _, r := range p.Registries {
			got[r.Name] = r.URL
		}
		if len(got) != 2 {
			t.Fatalf("registries = %v, want work + public", got)
		}
		if got["work"] != "https://example.com/work/registry.json" {
			t.Errorf("work = %q", got["work"])
		}
		if got["public"] != registry.DefaultURL {
			t.Errorf("public = %q, want %q", got["public"], registry.DefaultURL)
		}
		if idx := strings.Index(res.Out, "{"); idx >= 0 {
			var out struct{ Seeded string }
			if err := json.Unmarshal([]byte(res.Out[idx:]), &out); err == nil && out.Seeded != "public" {
				t.Errorf("json seeded = %q, want public", out.Seeded)
			}
		}
	})

	t.Run("a custom profile registry is kept under default", func(t *testing.T) {
		s := testutil.NewSandbox(t)
		if res := runCLI(t, "profile", "set", "registry", "https://example.com/mine/registry.json",
			"--profile", s.ProfilePath, "--cache-dir", s.CacheDir); res.RunErr != nil {
			t.Fatalf("profile set: %v\n%s", res.RunErr, res.Err)
		}

		if res := runCLI(t, "registry", "add", "work", "https://example.com/work/registry.json",
			"--profile", s.ProfilePath, "--cache-dir", s.CacheDir); res.RunErr != nil {
			t.Fatalf("registry add: %v\n%s", res.RunErr, res.Err)
		}

		p, err := profile.Load(s.ProfilePath)
		if err != nil {
			t.Fatalf("load profile: %v", err)
		}
		got := map[string]string{}
		for _, r := range p.Registries {
			got[r.Name] = r.URL
		}
		if got["default"] != "https://example.com/mine/registry.json" {
			t.Errorf("default = %q, want the profile's own registry URL (got %v)", got["default"], got)
		}
	})

	t.Run("second add seeds nothing further", func(t *testing.T) {
		s := testutil.NewSandbox(t)
		for _, name := range []string{"work", "extra"} {
			if res := runCLI(t, "registry", "add", name, "https://example.com/"+name+"/registry.json",
				"--profile", s.ProfilePath, "--cache-dir", s.CacheDir); res.RunErr != nil {
				t.Fatalf("add %s: %v\n%s", name, res.RunErr, res.Err)
			}
		}
		p, err := profile.Load(s.ProfilePath)
		if err != nil {
			t.Fatalf("load profile: %v", err)
		}
		// work + extra + one seeded public, and nothing more.
		if len(p.Registries) != 3 {
			t.Errorf("registries = %d, want 3 (work, extra, public): %+v", len(p.Registries), p.Registries)
		}
	})

	t.Run("naming public itself seeds nothing", func(t *testing.T) {
		s := testutil.NewSandbox(t)
		if res := runCLI(t, "registry", "add", "public", registry.DefaultURL,
			"--profile", s.ProfilePath, "--cache-dir", s.CacheDir); res.RunErr != nil {
			t.Fatalf("add: %v\n%s", res.RunErr, res.Err)
		}
		p, err := profile.Load(s.ProfilePath)
		if err != nil {
			t.Fatalf("load profile: %v", err)
		}
		if len(p.Registries) != 1 {
			t.Errorf("registries = %d, want 1: %+v", len(p.Registries), p.Registries)
		}
	})
}
