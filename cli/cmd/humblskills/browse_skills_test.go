package main

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

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

	items := buildSkillItems([]registry.Skill{{Name: "foo", Version: "1.0.0"}}, m, true)
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

	items := buildSkillItems([]registry.Skill{{Name: "foo", Version: "1.0.0"}}, m, true)
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

func TestSkillItem_Detail_LinesFitWidth(t *testing.T) {
	const width = 40
	longPath := "/Users/jenningsfantini/.humblskills/skills/deal-positioning-framework"
	it := skillItem{
		s: registry.Skill{
			Name:        "deal-positioning-framework",
			Version:     "0.1.0",
			Description: "Build an exec-ready HappyRobot Point of View (POV) that frames the customer's problem, our unique approach, and the commercial ask in one tight narrative for sellers.",
			Category:    "writing",
			Role:        "sdr",
			Tags:        []string{"sales", "pov", "positioning", "executive", "narrative"},
			Platforms:   []string{"claude-code", "cursor", "codex"},
			Registry:    "happyrobot",
		},
		installs: []manifest.Installation{
			{
				Skill: "deal-positioning-framework", Version: "0.1.0",
				Platform: "claude-code", Scope: "user",
				Path: longPath, StorePath: longPath,
			},
		},
	}
	detail := it.Detail(ui.DefaultTheme(), width)
	for i, line := range strings.Split(detail, "\n") {
		plain := ansi.Strip(line)
		if w := lipgloss.Width(plain); w > width {
			t.Errorf("line %d width %d > %d: %q", i, w, width, plain)
		}
	}
}

func TestBuildSkillItems_RolelessFirstThenByRoleWithinRegistry_Legacy(t *testing.T) {
	items := buildSkillItems([]registry.Skill{
		{Name: "zeta", Version: "1.0.0", Registry: "work", Role: "sdr"},
		{Name: "alpha", Version: "1.0.0", Registry: "work", Role: "fde"},
		{Name: "misc", Version: "1.0.0", Registry: "work"},
		{Name: "other", Version: "1.0.0", Registry: "personal"},
	}, nil, false)
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

func TestBuildSkillItems_SortsByCategoryWhenGrouped(t *testing.T) {
	items := buildSkillItems([]registry.Skill{
		{Name: "zeta", Version: "1.0.0", Registry: "work", Category: "writing", Role: "sdr"},
		{Name: "alpha", Version: "1.0.0", Registry: "work", Category: "design"},
		{Name: "misc", Version: "1.0.0", Registry: "work", Category: "writing"},
		{Name: "other", Version: "1.0.0", Registry: "personal", Category: "meta"},
	}, nil, true)
	got := make([]string, 0, len(items))
	for _, it := range items {
		got = append(got, it.s.Registry+"/"+skillCategory(it.s)+"/"+it.s.Role+"/"+it.s.Name)
	}
	want := []string{
		"personal/meta//other",
		"work/design//alpha",
		"work/writing//misc",
		"work/writing/sdr/zeta",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v, want %v", got, want)
		}
	}
}

func TestBuildSkillTree_CategoryThenRole(t *testing.T) {
	skills := buildSkillItems([]registry.Skill{
		{Name: "misc", Version: "1.0.0", Registry: "work", Category: "writing"},
		{Name: "alpha", Version: "1.0.0", Registry: "work", Category: "writing", Role: "fde"},
		{Name: "beta", Version: "1.0.0", Registry: "work", Category: "writing", Role: "fde"},
		{Name: "zeta", Version: "1.0.0", Registry: "work", Category: "design"},
		{Name: "other", Version: "1.0.0", Registry: "personal", Category: "meta"},
	}, nil, true)
	items := buildSkillTree(skills, true, true)

	got := make([]string, 0, len(items))
	for _, it := range items {
		switch v := it.(type) {
		case sectionHeaderItem:
			switch v.kind {
			case sectionRegistry:
				got = append(got, "reg:"+v.name)
			case sectionCategory:
				got = append(got, "cat:"+v.name)
			case sectionRole:
				got = append(got, "role:"+v.name)
			}
		case skillItem:
			got = append(got, v.s.Name)
		}
	}
	want := []string{
		"reg:personal", "cat:meta", "other",
		"reg:work", "cat:design", "zeta",
		"cat:writing", "misc", "role:fde", "alpha", "beta",
	}
	if len(got) != len(want) {
		t.Fatalf("items = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("items = %v, want %v", got, want)
		}
	}
}

func TestBuildSkillTree_LegacyRoleSubHeaders(t *testing.T) {
	skills := buildSkillItems([]registry.Skill{
		{Name: "misc", Version: "1.0.0", Registry: "work"},
		{Name: "alpha", Version: "1.0.0", Registry: "work", Role: "fde"},
		{Name: "beta", Version: "1.0.0", Registry: "work", Role: "fde"},
		{Name: "zeta", Version: "1.0.0", Registry: "work", Role: "sdr"},
		{Name: "other", Version: "1.0.0", Registry: "personal"},
	}, nil, false)
	items := buildSkillTree(skills, true, false)

	got := make([]string, 0, len(items))
	for _, it := range items {
		switch v := it.(type) {
		case sectionHeaderItem:
			switch v.kind {
			case sectionRegistry:
				got = append(got, "reg:"+v.name)
			case sectionRole:
				got = append(got, "role:"+v.name)
			default:
				got = append(got, "other:"+v.name)
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

func TestBuildSkillTree_LegacyRoleWithoutRegistryGrouping(t *testing.T) {
	skills := buildSkillItems([]registry.Skill{
		{Name: "misc", Version: "1.0.0"},
		{Name: "alpha", Version: "1.0.0", Role: "fde"},
	}, nil, false)
	items := buildSkillTree(skills, false, false)

	if len(items) != 3 {
		t.Fatalf("items = %d, want 3 (skill, role header, skill)", len(items))
	}
	h, ok := items[1].(sectionHeaderItem)
	if !ok || h.kind != sectionRole || h.name != "fde" {
		t.Fatalf("items[1] = %#v, want fde role header", items[1])
	}
}

func TestBuildSkillTree_NoRoleHeadersWhenNoRoles(t *testing.T) {
	skills := buildSkillItems([]registry.Skill{
		{Name: "a", Version: "1.0.0", Registry: "work", Category: "design"},
		{Name: "b", Version: "1.0.0", Registry: "personal", Category: "meta"},
	}, nil, true)
	items := buildSkillTree(skills, true, true)
	for _, it := range items {
		if h, ok := it.(sectionHeaderItem); ok && h.kind == sectionRole {
			t.Fatalf("unexpected role header %q", h.name)
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

func TestSectionHeader_CollapsibleAndParentKeys(t *testing.T) {
	skills := buildSkillItems([]registry.Skill{
		{Name: "alpha", Version: "1.0.0", Registry: "work", Category: "writing", Role: "sdr"},
	}, nil, true)
	items := buildSkillTree(skills, false, true)
	var cat, role sectionHeaderItem
	var skill skillItem
	for _, it := range items {
		switch v := it.(type) {
		case sectionHeaderItem:
			if v.kind == sectionCategory {
				cat = v
			}
			if v.kind == sectionRole {
				role = v
			}
		case skillItem:
			skill = v
		}
	}
	if !cat.IsCollapsible() || cat.IsHeader() {
		t.Fatalf("category should be collapsible, not a nav-skip header: %#v", cat)
	}
	if !role.IsCollapsible() {
		t.Fatal("role should be collapsible")
	}
	if len(role.ParentCollapseKeys()) != 1 || role.ParentCollapseKeys()[0] != cat.CollapseKey() {
		t.Fatalf("role parentKeys = %v, want [%s]", role.ParentCollapseKeys(), cat.CollapseKey())
	}
	if len(skill.ParentCollapseKeys()) != 2 {
		t.Fatalf("skill parentKeys = %v, want category+role", skill.ParentCollapseKeys())
	}
}
