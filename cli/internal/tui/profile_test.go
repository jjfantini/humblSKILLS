package tui

import (
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/adapters"
	"github.com/jjfantini/humblSKILLS/cli/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

func newTestProfileModel(p profile.Profile) profileModel {
	return profileModel{
		theme:    ui.DefaultTheme(),
		adapters: []adapters.Adapter{{Name: "claude-code"}, {Name: "cursor"}},
		profile:  p,
		focus:    focusSettings,
		width:    100,
		height:   30,
	}
}

func TestProfileModel_CurrentSelectionIndex_UnsetResolvesToGlobal(t *testing.T) {
	m := newTestProfileModel(profile.Profile{})
	m.settingIdx = 1
	if got := m.currentSelectionIndex(); got != 0 {
		t.Errorf("unset scope should select index 0 (global), got %d", got)
	}
}

func TestProfileModel_CurrentSelectionIndex_ExplicitValues(t *testing.T) {
	cases := []struct {
		scope string
		want  int
	}{
		{profile.ScopeGlobal, 0},
		{profile.ScopeUser, 1},
		{profile.ScopeProject, 2},
		{profile.ScopeAdapterDefault, 3},
	}
	for _, c := range cases {
		m := newTestProfileModel(profile.Profile{DefaultScope: c.scope})
		m.settingIdx = 1
		if got := m.currentSelectionIndex(); got != c.want {
			t.Errorf("scope=%q: currentSelectionIndex() = %d, want %d", c.scope, got, c.want)
		}
	}
}

func TestProfileModel_ToggleCurrent_Scope_SetsExplicitValue(t *testing.T) {
	m := newTestProfileModel(profile.Profile{})
	m.settingIdx = 1
	m.valueIdx = 2 // project
	m = m.toggleCurrent()
	if m.profile.DefaultScope != profile.ScopeProject {
		t.Errorf("DefaultScope = %q, want %q", m.profile.DefaultScope, profile.ScopeProject)
	}
	if !m.changed {
		t.Error("expected changed=true")
	}
}

func TestProfileModel_ValueCount_ScopeHasFourOptions(t *testing.T) {
	m := newTestProfileModel(profile.Profile{})
	m.settingIdx = 1
	if got := m.valueCount(); got != 4 {
		t.Errorf("valueCount() = %d, want 4 (global/user/project/adapter-default)", got)
	}
}

func TestProfileModel_SettingBadge_Scope(t *testing.T) {
	cases := []struct {
		scope string
		want  string
	}{
		{"", "global humblskills"},
		{profile.ScopeGlobal, "global humblskills"},
		{profile.ScopeUser, "user"},
		{profile.ScopeProject, "project"},
		{profile.ScopeAdapterDefault, "adapter default"},
	}
	for _, c := range cases {
		m := newTestProfileModel(profile.Profile{DefaultScope: c.scope})
		if got := m.settingBadge("scope"); got != c.want {
			t.Errorf("scope=%q: settingBadge = %q, want %q", c.scope, got, c.want)
		}
	}
}

func TestProfileModel_SettingValueEmpty_ScopeNeverEmpty(t *testing.T) {
	m := newTestProfileModel(profile.Profile{})
	if m.settingValueEmpty("scope") {
		t.Error("scope should never report as empty — unset resolves to a concrete global default")
	}
}

func TestProfileModel_RenderScopeOptions_AdapterDefaultShowsNote(t *testing.T) {
	m := newTestProfileModel(profile.Profile{DefaultScope: profile.ScopeAdapterDefault})
	rows := m.renderScopeOptions("│", 60)
	joined := strings.Join(rows, "\n")
	if !strings.Contains(joined, "can't show a concrete location") {
		t.Errorf("expected adapter-default note in rows:\n%s", joined)
	}
}

func TestProfileModel_RenderScopeOptions_GlobalHasNoNote(t *testing.T) {
	m := newTestProfileModel(profile.Profile{DefaultScope: profile.ScopeGlobal})
	rows := m.renderScopeOptions("│", 60)
	joined := strings.Join(rows, "\n")
	if strings.Contains(joined, "can't show a concrete location") {
		t.Errorf("global scope should not show the adapter-default note:\n%s", joined)
	}
}
