package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

func TestDefaultDashboardTiles_Shape(t *testing.T) {
	tiles := DefaultDashboardTiles()
	if len(tiles) < 9 {
		t.Fatalf("want at least 9 tiles, got %d", len(tiles))
	}
	hotkeys := map[string]bool{}
	commands := map[string]bool{}
	for _, tl := range tiles {
		if tl.Command == "" {
			t.Errorf("tile missing command: %+v", tl)
		}
		if tl.Label == "" {
			t.Errorf("tile missing label: %+v", tl)
		}
		if tl.Hotkey != "" && hotkeys[tl.Hotkey] {
			t.Errorf("duplicate hotkey %q", tl.Hotkey)
		}
		hotkeys[tl.Hotkey] = true
		if commands[tl.Command] {
			t.Errorf("duplicate command %q", tl.Command)
		}
		commands[tl.Command] = true
	}
	// Every expected command shipped.
	for _, want := range []string{"install", "list", "update", "search", "uninstall", "profile", "doctor", "registry", "version", "eval"} {
		if !commands[want] {
			t.Errorf("missing dashboard command %q", want)
		}
	}
}

func TestDashboardStatus_Fields(t *testing.T) {
	// Round-trip through RenderStatusMeta just to ensure it doesn't panic
	// and renders something visible.
	th := ui.DefaultTheme()
	got := RenderStatusMeta(th, DashboardStatus{Healthy: true, Platforms: 2, Skills: 5})
	if !strings.Contains(got, "2") || !strings.Contains(got, "5") {
		t.Errorf("status missing counts: %q", got)
	}
}

func TestDashboardModel_QuitKeys(t *testing.T) {
	m := dashboardModel{
		cfg: DashboardConfig{
			Theme: ui.DefaultTheme(),
			Tiles: DefaultDashboardTiles(),
		},
	}
	m.rebuildVisible()

	cases := []string{"q", "esc", "ctrl+c"}
	for _, k := range cases {
		out, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
		if k == "ctrl+c" {
			out, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
		} else if k == "esc" {
			out, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		}
		dm, ok := out.(dashboardModel)
		if !ok {
			t.Fatalf("key %q: Update did not return dashboardModel", k)
		}
		if !dm.result.Quit {
			t.Errorf("key %q: expected Quit=true, got %+v", k, dm.result)
		}
		if cmd == nil {
			t.Errorf("key %q: expected tea.Quit cmd", k)
		}
	}
}

func TestDashboardModel_VimKeysNavigateViaSharedKeymap(t *testing.T) {
	m := dashboardModel{cfg: DashboardConfig{Theme: ui.DefaultTheme(), Tiles: DefaultDashboardTiles()}}
	m.width = 120 // 3 columns
	m.rebuildVisible()

	// j moves down a full row (cols tiles) via keys.Down ("down"/"j").
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	dm := out.(dashboardModel)
	if dm.cursor == 0 {
		t.Errorf("j should move the cursor down, got %d", dm.cursor)
	}

	// l moves right one tile (keys.Right = right/l), and must NOT launch the
	// "list" command whose hotkey is also "l" — movement shadows it.
	out, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	dm = out.(dashboardModel)
	if dm.result.Command != "" {
		t.Errorf("l should move right, not launch a command: %+v", dm.result)
	}
	if dm.cursor != 1 {
		t.Errorf("l should move cursor right to 1, got %d", dm.cursor)
	}

	// A hotkey that doesn't collide with hjkl still launches.
	out, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	dm = out.(dashboardModel)
	if dm.result.Command != "doctor" {
		t.Errorf("d should launch doctor, got %+v", dm.result)
	}
}

func TestDashboardModel_HotkeyLaunchesCommand(t *testing.T) {
	m := dashboardModel{
		cfg: DashboardConfig{
			Theme: ui.DefaultTheme(),
			Tiles: DefaultDashboardTiles(),
		},
	}
	m.rebuildVisible()

	// "i" is install's hotkey.
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	dm, ok := out.(dashboardModel)
	if !ok {
		t.Fatal("Update did not return dashboardModel")
	}
	if dm.result.Command != "install" {
		t.Errorf("Command = %q, want install", dm.result.Command)
	}
}

func TestDashboardModel_SearchTogglesAndFilters(t *testing.T) {
	m := dashboardModel{
		cfg: DashboardConfig{
			Theme: ui.DefaultTheme(),
			Tiles: DefaultDashboardTiles(),
		},
	}
	m.rebuildVisible()

	// Press / to open search.
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	dm := out.(dashboardModel)
	if !dm.searchOn {
		t.Fatal("searchOn should be true after /")
	}
	// Type "ins".
	for _, r := range "ins" {
		out, _ = dm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		dm = out.(dashboardModel)
	}
	if dm.query != "ins" {
		t.Errorf("query = %q", dm.query)
	}
	// "install" tile should be visible; "list" might not be.
	sawInstall := false
	for _, idx := range dm.visible {
		if dm.cfg.Tiles[idx].Command == "install" {
			sawInstall = true
		}
	}
	if !sawInstall {
		t.Errorf("install not in filtered visible set: %v", dm.visible)
	}

	// Esc clears query and closes search.
	out, _ = dm.Update(tea.KeyMsg{Type: tea.KeyEsc})
	dm = out.(dashboardModel)
	if dm.searchOn {
		t.Error("searchOn should be false after esc")
	}
	if dm.query != "" {
		t.Errorf("query should be cleared: %q", dm.query)
	}
}

func TestDashboardModel_BackspaceShrinksQuery(t *testing.T) {
	m := dashboardModel{cfg: DashboardConfig{Theme: ui.DefaultTheme(), Tiles: DefaultDashboardTiles()}}
	m.rebuildVisible()
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	dm := out.(dashboardModel)
	for _, r := range "abc" {
		out, _ = dm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		dm = out.(dashboardModel)
	}
	out, _ = dm.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	dm = out.(dashboardModel)
	if dm.query != "ab" {
		t.Errorf("query = %q", dm.query)
	}
}

func TestDashboardModel_EnterLaunchesCursor(t *testing.T) {
	m := dashboardModel{cfg: DashboardConfig{Theme: ui.DefaultTheme(), Tiles: DefaultDashboardTiles()}}
	m.rebuildVisible()
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	dm := out.(dashboardModel)
	if dm.result.Command == "" {
		t.Errorf("enter should emit a command: %+v", dm.result)
	}
}

func TestTruncateDisplay(t *testing.T) {
	cases := map[string]int{
		"hello":                      10,
		"this is a very long string": 10,
	}
	for in, width := range cases {
		got := truncateDisplay(in, width)
		// Display width must not exceed the target.
		// Don't overspecify the ellipsis format — just that it fits.
		if len([]rune(got)) > width+3 {
			t.Errorf("truncateDisplay(%q, %d) = %q too wide", in, width, got)
		}
	}
}

func TestIndentBlock_AddsLeadingSpaces(t *testing.T) {
	got := indentBlock("a\nb", 4)
	for _, line := range strings.Split(got, "\n") {
		if !strings.HasPrefix(line, "    ") {
			t.Errorf("line not indented: %q", line)
		}
	}
}

func TestVersionScreenModel_QuitOnKey(t *testing.T) {
	m := versionScreenModel{
		cfg:    VersionScreenConfig{Theme: ui.DefaultTheme(), Info: VersionInfo{Version: "v1", Commit: "abc"}},
		width:  80,
		height: 24,
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Error("expected Quit cmd on q")
	}
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Error("expected Quit cmd on esc")
	}
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("expected Quit cmd on enter")
	}
}

func TestVersionScreenModel_View_RendersInfo(t *testing.T) {
	m := versionScreenModel{
		cfg:    VersionScreenConfig{Theme: ui.DefaultTheme(), Info: VersionInfo{Version: "v1.2.3", Commit: "abc123", Dirty: true}},
		width:  100,
		height: 24,
	}
	view := m.View()
	for _, want := range []string{"humblskills", "v1.2.3", "abc123", "dirty"} {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q:\n%s", want, view)
		}
	}
}

func TestVersionScreenModel_View_EmptyBeforeSize(t *testing.T) {
	m := versionScreenModel{cfg: VersionScreenConfig{Theme: ui.DefaultTheme()}}
	if got := m.View(); got != "" {
		t.Errorf("view should be empty before size set, got:\n%s", got)
	}
}

func TestPadRight(t *testing.T) {
	if padRight("abc", 6) != "abc   " {
		t.Errorf("got %q", padRight("abc", 6))
	}
	if padRight("abcdef", 3) != "abcdef" {
		t.Errorf("got %q", padRight("abcdef", 3))
	}
}
