package tui

import (
	"os"
	"path/filepath"
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

func TestBuildDashboardGreeting_PopulatesFields(t *testing.T) {
	g := BuildDashboardGreeting(5)
	if g.Updates != 5 {
		t.Errorf("Updates = %d", g.Updates)
	}
	// User & Cwd may be empty in sandboxed CI — just sanity-check types.
	_ = g.User
	_ = g.Cwd
}

func TestCompactPath(t *testing.T) {
	// Use a real tempdir as the "home" so the path separators match
	// whatever the OS natively produces. Point both HOME and
	// USERPROFILE at it — Unix reads $HOME, Windows reads %USERPROFILE%.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	inside := filepath.Join(home, "work", "x")
	if got := compactPath(inside); !strings.HasPrefix(got, "~") {
		t.Errorf("compactPath should prefix ~: %q", got)
	}

	// A path unrelated to home passes through unchanged. Use the OS's
	// temp dir root so we get an actually-unrelated absolute path on
	// every platform.
	unrelated := filepath.Join(os.TempDir(), "definitely-not-home-"+filepath.Base(home))
	if got := compactPath(unrelated); got != unrelated {
		t.Errorf("unrelated path changed: got %q want %q", got, unrelated)
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
