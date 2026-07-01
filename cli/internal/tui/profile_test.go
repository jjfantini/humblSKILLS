package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

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

// --- enter vs space key semantics ------------------------------------------
//
// Regression coverage: multi-select values (platforms) must only toggle on
// space, matching every other multi-select surface in the TUI (e.g. the
// install platform modal). Enter previously also toggled here, which was
// inconsistent and easy to trigger by accident while navigating.

func TestProfileModel_MultiSelect_EnterDoesNotToggle_JustReturns(t *testing.T) {
	m := newTestProfileModel(profile.Profile{})
	m.settingIdx = 0 // platforms
	m.focus = focusValue
	m.valueIdx = 0 // claude-code

	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := out.(profileModel)

	if len(updated.profile.DefaultPlatforms) != 0 {
		t.Errorf("enter must not toggle a platform on; DefaultPlatforms = %v", updated.profile.DefaultPlatforms)
	}
	if updated.changed {
		t.Error("enter must not mark the profile as changed when it doesn't toggle anything")
	}
	if updated.focus != focusSettings {
		t.Errorf("enter should return focus to the settings pane, got %v", updated.focus)
	}
}

func TestProfileModel_MultiSelect_SpaceStillToggles_AndStaysInValuePane(t *testing.T) {
	m := newTestProfileModel(profile.Profile{})
	m.settingIdx = 0 // platforms
	m.focus = focusValue
	m.valueIdx = 0 // claude-code

	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	updated := out.(profileModel)

	if len(updated.profile.DefaultPlatforms) != 1 || updated.profile.DefaultPlatforms[0] != "claude-code" {
		t.Errorf("space should toggle claude-code on; DefaultPlatforms = %v", updated.profile.DefaultPlatforms)
	}
	if !updated.changed {
		t.Error("expected changed=true after toggling")
	}
	if updated.focus != focusValue {
		t.Errorf("space should not leave the value pane, got focus=%v", updated.focus)
	}
}

// --- status auto-return setting --------------------------------------------

func TestProfileModel_AutoReturn_CurrentSelectionIndex_UnsetResolvesToDefault(t *testing.T) {
	m := newTestProfileModel(profile.Profile{})
	m.settingIdx = 2
	if got := m.currentSelectionIndex(); got != 0 {
		t.Errorf("unset auto-return should select index 0 (default), got %d", got)
	}
}

func TestProfileModel_AutoReturn_CurrentSelectionIndex_ExplicitValues(t *testing.T) {
	ptr := func(n int) *int { return &n }
	cases := []struct {
		seconds *int
		want    int
	}{
		{ptr(10), 1},
		{ptr(15), 2},
		{ptr(30), 3},
		{ptr(0), 4},
	}
	for _, c := range cases {
		m := newTestProfileModel(profile.Profile{StatusAutoReturnSeconds: c.seconds})
		m.settingIdx = 2
		if got := m.currentSelectionIndex(); got != c.want {
			t.Errorf("seconds=%v: currentSelectionIndex() = %d, want %d", *c.seconds, got, c.want)
		}
	}
}

func TestProfileModel_AutoReturn_ValueCount(t *testing.T) {
	m := newTestProfileModel(profile.Profile{})
	m.settingIdx = 2
	if got := m.valueCount(); got != 5 {
		t.Errorf("valueCount() = %d, want 5 (5s/10s/15s/30s/disabled)", got)
	}
}

func TestProfileModel_AutoReturn_ToggleCurrent_SetsExplicitValue(t *testing.T) {
	m := newTestProfileModel(profile.Profile{})
	m.settingIdx = 2
	m.valueIdx = 4 // disabled
	m = m.toggleCurrent()
	if m.profile.StatusAutoReturnSeconds == nil || *m.profile.StatusAutoReturnSeconds != 0 {
		t.Errorf("StatusAutoReturnSeconds = %v, want 0 (disabled)", m.profile.StatusAutoReturnSeconds)
	}
	if !m.changed {
		t.Error("expected changed=true")
	}
}

func TestProfileModel_AutoReturn_SettingBadge(t *testing.T) {
	ptr := func(n int) *int { return &n }
	cases := []struct {
		seconds *int
		want    string
	}{
		{nil, "5s (default)"},
		{ptr(10), "10s"},
		{ptr(0), "disabled"},
	}
	for _, c := range cases {
		m := newTestProfileModel(profile.Profile{StatusAutoReturnSeconds: c.seconds})
		if got := m.settingBadge("status_auto_return"); got != c.want {
			t.Errorf("seconds=%v: settingBadge = %q, want %q", c.seconds, got, c.want)
		}
	}
}

func TestProfileModel_Radio_EnterTogglesAndReturns(t *testing.T) {
	m := newTestProfileModel(profile.Profile{})
	m.settingIdx = 1 // scope (radio)
	m.focus = focusValue
	m.valueIdx = 2 // project

	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := out.(profileModel)

	if updated.profile.DefaultScope != profile.ScopeProject {
		t.Errorf("enter should commit the highlighted radio option; DefaultScope = %q", updated.profile.DefaultScope)
	}
	if updated.focus != focusSettings {
		t.Errorf("enter should return focus to the settings pane, got %v", updated.focus)
	}
}
