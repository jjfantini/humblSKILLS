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

func TestBuildSkillItems_RolelessFirstThenByRoleWithinRegistry(t *testing.T) {
	items := buildSkillItems([]registry.Skill{
		{Name: "zeta", Version: "1.0.0", Registry: "work", Role: "sdr"},
		{Name: "alpha", Version: "1.0.0", Registry: "work", Role: "fde"},
		{Name: "misc", Version: "1.0.0", Registry: "work"},
		{Name: "other", Version: "1.0.0", Registry: "personal"},
	}, nil)
	got := make([]string, 0, len(items))
	for _, it := range items {
		got = append(got, it.s.Registry+"/"+it.s.Role+"/"+it.s.Name)
	}
	want := []string{"personal//other", "work//misc", "work/fde/alpha", "work/sdr/zeta"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v, want %v", got, want)
		}
	}
}

func TestInterleaveHeaders_RoleSubHeaders(t *testing.T) {
	skills := buildSkillItems([]registry.Skill{
		{Name: "misc", Version: "1.0.0", Registry: "work"},
		{Name: "alpha", Version: "1.0.0", Registry: "work", Role: "fde"},
		{Name: "beta", Version: "1.0.0", Registry: "work", Role: "fde"},
		{Name: "zeta", Version: "1.0.0", Registry: "work", Role: "sdr"},
		{Name: "other", Version: "1.0.0", Registry: "personal"},
	}, nil)
	items := interleaveRegistryHeaders(skills, true)

	got := make([]string, 0, len(items))
	for _, it := range items {
		switch v := it.(type) {
		case registryHeaderItem:
			if v.sub {
				got = append(got, "role:"+v.name)
			} else {
				got = append(got, "reg:"+v.name)
			}
		case skillItem:
			got = append(got, v.s.Name)
		}
	}
	want := []string{"reg:personal", "other", "reg:work", "misc", "role:fde", "alpha", "beta", "role:sdr", "zeta"}
	if len(got) != len(want) {
		t.Fatalf("items = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("items = %v, want %v", got, want)
		}
	}
}

func TestInterleaveHeaders_RoleSubHeadersWithoutRegistryGrouping(t *testing.T) {
	// Single registry (grouped=false): role sub-headers still appear.
	skills := buildSkillItems([]registry.Skill{
		{Name: "misc", Version: "1.0.0"},
		{Name: "alpha", Version: "1.0.0", Role: "fde"},
	}, nil)
	items := interleaveRegistryHeaders(skills, false)

	if len(items) != 3 {
		t.Fatalf("items = %d, want 3 (skill, sub-header, skill)", len(items))
	}
	h, ok := items[1].(registryHeaderItem)
	if !ok || !h.sub || h.name != "fde" {
		t.Fatalf("items[1] = %#v, want fde sub-header", items[1])
	}
}

func TestInterleaveHeaders_NoRoleSubHeadersWhenNoRoles(t *testing.T) {
	skills := buildSkillItems([]registry.Skill{
		{Name: "a", Version: "1.0.0", Registry: "work"},
		{Name: "b", Version: "1.0.0", Registry: "personal"},
	}, nil)
	items := interleaveRegistryHeaders(skills, true)
	for _, it := range items {
		if h, ok := it.(registryHeaderItem); ok && h.sub {
			t.Fatalf("unexpected role sub-header %q", h.name)
		}
	}
}

func TestSkillItem_RoleInDetailAndFilterValue(t *testing.T) {
	it := skillItem{s: registry.Skill{Name: "alpha", Version: "1.0.0", Role: "fde"}}
	th := ui.DefaultTheme()
	if !strings.Contains(it.Detail(th, 80), "fde") {
		t.Error("Detail missing role")
	}
	if !strings.Contains(it.FilterValue(), "fde") {
		t.Error("FilterValue missing role")
	}
}
