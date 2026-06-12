package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jjfantini/humblSKILLS/cli/internal/adapters"
	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

func newTestModal(selected, detected map[string]bool) installModalModel {
	if selected == nil {
		selected = map[string]bool{}
	}
	if detected == nil {
		detected = map[string]bool{}
	}
	return installModalModel{
		theme: ui.DefaultTheme(),
		skill: "foo",
		adapters: []adapters.Adapter{
			{Name: "claude-code"},
			{Name: "cursor"},
		},
		detected: detected,
		selected: selected,
		scopes: []scopeOpt{
			{label: "Global fanout", value: "global"},
			{label: "adapter default", value: ""},
			{label: "user", value: "user"},
			{label: "project", value: "project"},
		},
		scopeIdx: 1,
		actions: []actionOpt{
			{label: "install", value: "install"},
			{label: "cancel", value: "cancel"},
		},
		actionIdx: 0,
		group:     groupPlatforms,
		cursor:    0,
		width:     120,
		height:    30,
	}
}

func TestInstallModal_SpaceTogglesInPlatforms(t *testing.T) {
	m := newTestModal(
		map[string]bool{"claude-code": true},
		map[string]bool{"claude-code": true, "cursor": true},
	)
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown}) // move to cursor row
	m = out.(installModalModel)
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.cursor)
	}
	out, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = out.(installModalModel)
	if !m.selected["cursor"] {
		t.Error("space should toggle cursor on")
	}
	if m.group != groupPlatforms {
		t.Errorf("space must not advance group; got %v", m.group)
	}
}

func TestInstallModal_EnterAdvancesFromPlatforms_NoToggle(t *testing.T) {
	m := newTestModal(
		map[string]bool{"claude-code": true},
		map[string]bool{"claude-code": true, "cursor": true},
	)
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = out.(installModalModel)
	if m.group != groupScope {
		t.Errorf("enter should advance to scope; got group=%v", m.group)
	}
	if !m.selected["claude-code"] {
		t.Error("enter must not toggle selection off")
	}
	if m.selected["cursor"] {
		t.Error("enter must not toggle selection on")
	}
}

func TestInstallModal_EnterAdvancesFromScope(t *testing.T) {
	m := newTestModal(nil, nil)
	m.group = groupScope
	m.cursor = 2
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = out.(installModalModel)
	if m.group != groupAction {
		t.Errorf("enter should advance to action; got %v", m.group)
	}
	if m.scopeIdx != 2 {
		t.Errorf("scopeIdx = %d, want 2", m.scopeIdx)
	}
}

func TestInstallModal_GlobalFanoutCommitsGlobalResult(t *testing.T) {
	m := newTestModal(
		map[string]bool{"claude-code": true},
		map[string]bool{"claude-code": true, "cursor": true},
	)
	m.group = groupAction
	m.scopeIdx = 0
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = out.(installModalModel)
	if !m.result.Confirmed {
		t.Fatal("result should be confirmed")
	}
	if !m.result.Global {
		t.Fatalf("global fanout scope should set Global=true: %+v", m.result)
	}
}

func TestInstallModal_TabKeptAsSilentAdvance(t *testing.T) {
	m := newTestModal(nil, nil)
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = out.(installModalModel)
	if m.group != groupScope {
		t.Errorf("tab should still advance; got %v", m.group)
	}
}

func TestInstallModal_Hints_PlatformsGroup(t *testing.T) {
	m := newTestModal(nil, nil)
	m.group = groupPlatforms
	hints := m.hints()
	keys := hintKeys(hints)
	if !contains(keys, "space") {
		t.Errorf("platforms hints missing 'space': %v", keys)
	}
	if !contains(keys, "↵") {
		t.Errorf("platforms hints missing '↵': %v", keys)
	}
	if contains(keys, "tab") {
		t.Errorf("platforms hints should not advertise tab: %v", keys)
	}
}

func TestInstallModal_Hints_ScopeGroup(t *testing.T) {
	m := newTestModal(nil, nil)
	m.group = groupScope
	keys := hintKeys(m.hints())
	if contains(keys, "space") {
		t.Errorf("scope hints should not mention space: %v", keys)
	}
	if contains(keys, "tab") {
		t.Errorf("scope hints should not mention tab: %v", keys)
	}
	if !contains(keys, "↵") {
		t.Errorf("scope hints missing '↵': %v", keys)
	}
}

func TestInstallModal_InfoPane_BothSelected_Warning(t *testing.T) {
	m := newTestModal(
		map[string]bool{"claude-code": true, "cursor": true},
		map[string]bool{"claude-code": true, "cursor": true},
	)
	heading, body := m.infoContent(40)
	if !strings.Contains(heading, "Duplicate") {
		t.Errorf("heading should warn about duplicates; got %q", heading)
	}
	if !strings.Contains(body, "drift") {
		t.Errorf("body should mention drift; got %q", body)
	}
}

func TestInstallModal_InfoPane_ClaudeOnly_Tip(t *testing.T) {
	m := newTestModal(
		map[string]bool{"claude-code": true},
		map[string]bool{"claude-code": true, "cursor": true},
	)
	heading, body := m.infoContent(40)
	if !strings.Contains(heading, "Tip") {
		t.Errorf("heading should be a tip; got %q", heading)
	}
	if !strings.Contains(body, "Cursor") {
		t.Errorf("body should mention Cursor; got %q", body)
	}
}

func TestInstallModal_InfoPane_CursorOnly_Note(t *testing.T) {
	m := newTestModal(
		map[string]bool{"cursor": true},
		map[string]bool{"claude-code": true, "cursor": true},
	)
	heading, body := m.infoContent(40)
	if !strings.Contains(heading, "Note") {
		t.Errorf("heading should be a note; got %q", heading)
	}
	if !strings.Contains(body, "claude-code is not selected") {
		t.Errorf("body should explain claude-code is not selected; got %q", body)
	}
}

func TestInstallModal_InfoPane_CursorOnly_ClaudeNotDetected(t *testing.T) {
	m := newTestModal(
		map[string]bool{"cursor": true},
		map[string]bool{"cursor": true},
	)
	_, body := m.infoContent(40)
	if !strings.Contains(body, "not detected") {
		t.Errorf("body should note claude-code not detected; got %q", body)
	}
}

func TestInstallModal_InfoPane_NoneSelected_EmptyState(t *testing.T) {
	m := newTestModal(nil, map[string]bool{"claude-code": true})
	heading, body := m.infoContent(40)
	if heading != "" {
		t.Errorf("empty state should have no heading; got %q", heading)
	}
	if !strings.Contains(body, "Select") {
		t.Errorf("empty state body should prompt selection; got %q", body)
	}
}

func TestInstallModal_InfoPane_GlobalFanout(t *testing.T) {
	m := newTestModal(
		map[string]bool{"claude-code": true, "cursor": true},
		map[string]bool{"claude-code": true, "cursor": true},
	)
	m.scopeIdx = 0
	heading, body := m.infoContent(60)
	if !strings.Contains(heading, "Global fanout") {
		t.Errorf("heading should describe global fanout; got %q", heading)
	}
	if !strings.Contains(body, ".humblskills") {
		t.Errorf("body should mention canonical .humblskills store; got %q", body)
	}
}

func TestInstallModal_NarrowTerminal_NoDivider(t *testing.T) {
	m := newTestModal(
		map[string]bool{"claude-code": true, "cursor": true},
		map[string]bool{"claude-code": true, "cursor": true},
	)
	m.width = 70
	body := m.renderBody()
	if strings.Contains(body, "│") {
		t.Errorf("narrow terminal should omit divider; got:\n%s", body)
	}
}

func TestInstallModal_WideTerminal_HasDivider(t *testing.T) {
	m := newTestModal(
		map[string]bool{"claude-code": true, "cursor": true},
		map[string]bool{"claude-code": true, "cursor": true},
	)
	m.width = 120
	body := m.renderBody()
	if !strings.Contains(body, "│") {
		t.Errorf("wide terminal should render divider; got:\n%s", body)
	}
}

func TestInstallModal_PaneWidths_Clamp(t *testing.T) {
	m := newTestModal(nil, nil)
	// Tiny total: right pane clamped to 24, left clamped to 40.
	leftW, rightW := m.paneWidths(50)
	if rightW != 24 {
		t.Errorf("rightW = %d, want 24 (clamped min)", rightW)
	}
	if leftW != 40 {
		t.Errorf("leftW = %d, want 40 (clamped min)", leftW)
	}

	// Huge total: right pane clamped to 48.
	_, rightW = m.paneWidths(300)
	if rightW != 48 {
		t.Errorf("rightW = %d, want 48 (clamped max)", rightW)
	}
}

// --- helpers ---

func hintKeys(hs []KeyHint) []string {
	out := make([]string, 0, len(hs))
	for _, h := range hs {
		out = append(out, h.Keys)
	}
	return out
}

func contains(xs []string, x string) bool {
	for _, s := range xs {
		if s == x {
			return true
		}
	}
	return false
}
