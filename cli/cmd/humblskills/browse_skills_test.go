package main

import (
	"strings"
	"testing"
	"time"

	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

func TestBuildSkillItems_CarriesEveryInstallationForASkill(t *testing.T) {
	m := &manifest.Manifest{SchemaVersion: manifest.SchemaVersion}
	m.Upsert(manifest.Installation{
		Skill: "foo", Version: "1.0.0", Platform: "claude-code", Scope: "user",
		Path: "/home/u/.claude/skills/foo", StorePath: "/home/u/.humblskills/skills/foo",
	})
	m.Upsert(manifest.Installation{
		Skill: "foo", Version: "1.0.0", Platform: "cursor", Scope: "user",
		Path: "/home/u/.cursor/skills/foo", StorePath: "/home/u/.humblskills/skills/foo",
	})
	m.Upsert(manifest.Installation{
		Skill: "foo", Version: "1.0.0", Platform: "codex", Scope: "user",
		Path: "/home/u/.agents/skills/foo", StorePath: "/home/u/.humblskills/skills/foo",
	})

	items := buildSkillItems([]registry.Skill{{Name: "foo", Version: "1.0.0"}}, m)
	if len(items) != 1 {
		t.Fatalf("items = %d, want 1", len(items))
	}
	it := items[0]
	if len(it.installs) != 3 {
		t.Fatalf("installs = %d, want 3: %+v", len(it.installs), it.installs)
	}
	if it.outdated {
		t.Error("should not be outdated when every install matches the registry version")
	}
	platforms := map[string]bool{}
	for _, inst := range it.installs {
		platforms[inst.Platform] = true
	}
	for _, want := range []string{"claude-code", "cursor", "codex"} {
		if !platforms[want] {
			t.Errorf("missing platform %q in installs: %+v", want, it.installs)
		}
	}
}

func TestBuildSkillItems_OutdatedWhenAnyInstallDrifts(t *testing.T) {
	m := &manifest.Manifest{SchemaVersion: manifest.SchemaVersion}
	m.Upsert(manifest.Installation{
		Skill: "foo", Version: "1.0.0", Platform: "claude-code", Scope: "user",
		Path: "/home/u/.claude/skills/foo", StorePath: "/home/u/.humblskills/skills/foo",
	})
	m.Upsert(manifest.Installation{
		Skill: "foo", Version: "0.9.0", Platform: "cursor", Scope: "user",
		Path: "/home/u/.cursor/skills/foo", StorePath: "/home/u/.humblskills/skills/foo",
	})

	items := buildSkillItems([]registry.Skill{{Name: "foo", Version: "1.0.0"}}, m)
	if !items[0].outdated {
		t.Error("expected outdated=true when one install lags the registry version")
	}
}

func TestSkillItem_Detail_ShowsStorePathAndEveryPlatform(t *testing.T) {
	it := skillItem{
		s: registry.Skill{Name: "foo", Version: "1.0.0"},
		installs: []manifest.Installation{
			{
				Skill: "foo", Version: "1.0.0", Platform: "claude-code", Scope: "user",
				Path: "/home/u/.claude/skills/foo", StorePath: "/home/u/.humblskills/skills/foo",
				InstalledAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				Skill: "foo", Version: "1.0.0", Platform: "cursor", Scope: "user",
				Path: "/home/u/.cursor/skills/foo", StorePath: "/home/u/.humblskills/skills/foo",
				InstalledAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	detail := it.Detail(ui.DefaultTheme(), 80)
	for _, want := range []string{
		"/home/u/.humblskills/skills/foo", // canonical store
		"claude-code", "/home/u/.claude/skills/foo",
		"cursor", "/home/u/.cursor/skills/foo",
	} {
		if !strings.Contains(detail, want) {
			t.Errorf("detail missing %q:\n%s", want, detail)
		}
	}
}

func TestSkillItem_Detail_NotInstalled_NoInstalledSection(t *testing.T) {
	it := skillItem{s: registry.Skill{Name: "foo", Version: "1.0.0"}}
	detail := it.Detail(ui.DefaultTheme(), 80)
	if strings.Contains(detail, "INSTALLED") {
		t.Errorf("uninstalled skill should not render an INSTALLED section:\n%s", detail)
	}
}
