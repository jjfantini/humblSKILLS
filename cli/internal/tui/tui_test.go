// Tests for the shared TUI helpers that don't require driving a full
// bubbletea program. Rendering and keymap assertions live here; full
// interactive model tests live alongside each screen's _test.go.
package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

func TestShouldUseTUI_FlagsOverrideTTY(t *testing.T) {
	cases := []struct {
		name                  string
		json, quiet, yes, want bool
	}{
		{"json disables", true, false, false, false},
		{"quiet disables", false, true, false, false},
		{"yes disables", false, false, true, false},
		// Default case depends on TTY; we only assert the disable paths.
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ShouldUseTUI(tc.json, tc.quiet, tc.yes); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDefaultKeys_RegistersEveryBinding(t *testing.T) {
	k := DefaultKeys()
	// Every binding should have at least one key.
	for _, b := range []key.Binding{
		k.Up, k.Down, k.Left, k.Right, k.Filter, k.Enter, k.Back, k.Help, k.Quit,
	} {
		if len(b.Keys()) == 0 {
			t.Errorf("binding missing keys: help=%q", b.Help())
		}
	}
	// Arrow + vim bindings: Up should include both "up" and "k".
	has := func(bind key.Binding, k string) bool {
		for _, kk := range bind.Keys() {
			if kk == k {
				return true
			}
		}
		return false
	}
	if !has(k.Up, "up") || !has(k.Up, "k") {
		t.Errorf("Up missing arrow or vim binding")
	}
	if !has(k.Down, "down") || !has(k.Down, "j") {
		t.Errorf("Down missing arrow or vim binding")
	}
	if !has(k.Quit, "ctrl+c") {
		t.Errorf("Quit missing ctrl+c")
	}
}

func TestBadge_EmptyTextReturnsEmpty(t *testing.T) {
	th := ui.DefaultTheme()
	if got := Badge(th, BadgeDetected, "  "); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestBadge_AllKindsRender(t *testing.T) {
	th := ui.DefaultTheme()
	for _, kind := range []BadgeKind{BadgeDetected, BadgeMissing, BadgeRW, BadgeRO, BadgeGhost} {
		got := Badge(th, kind, "text")
		if !strings.Contains(got, "text") {
			t.Errorf("kind %d lost text: %q", kind, got)
		}
	}
}

func TestHeader_RendersBrandAndMetaAndRule(t *testing.T) {
	th := ui.DefaultTheme()
	got := Header(th, HeaderSpec{Version: "v1.2.3", Section: "Install", Meta: "3 / 3 detected"}, 80)
	for _, want := range []string{"humblskills", "v1.2.3", "Install", "3 / 3 detected", "╌"} {
		if !strings.Contains(got, want) {
			t.Errorf("Header missing %q:\n%s", want, got)
		}
	}
}

func TestHeader_ClampsShortWidth(t *testing.T) {
	th := ui.DefaultTheme()
	got := Header(th, HeaderSpec{Version: "v1", Section: "x"}, 10)
	if got == "" {
		t.Error("Header returned empty for narrow width")
	}
}

func TestFooter_KeyHintsAndRule(t *testing.T) {
	th := ui.DefaultTheme()
	hints := []KeyHint{
		{Keys: "↑↓", Label: "select"},
		{Keys: "q", Label: "quit"},
	}
	got := Footer(th, hints, "focused: x", 80)
	for _, want := range []string{"↑↓", "select", "q", "quit", "focused: x", "╌"} {
		if !strings.Contains(got, want) {
			t.Errorf("Footer missing %q:\n%s", want, got)
		}
	}
}

func TestFooter_EmptyHints(t *testing.T) {
	th := ui.DefaultTheme()
	got := Footer(th, nil, "", 80)
	if !strings.Contains(got, "╌") {
		t.Errorf("Footer missing rule:\n%s", got)
	}
}

func TestDashedRule_MinWidth(t *testing.T) {
	th := ui.DefaultTheme()
	// Even at width=1 the helper clamps to a minimum usable length.
	if got := DashedRule(th, 1); got == "" {
		t.Error("DashedRule returned empty")
	}
}

func TestFrame_ComposesHeaderBodyFooter(t *testing.T) {
	got := Frame("HEADER", "BODY", "FOOTER", 4)
	// Body lifted to 4 rows then footer on the fifth.
	if !strings.Contains(got, "HEADER") || !strings.Contains(got, "FOOTER") {
		t.Errorf("missing chrome:\n%s", got)
	}
	if !strings.Contains(got, "BODY") {
		t.Errorf("missing body:\n%s", got)
	}
}

func TestFrame_ShortCircuitsWithZeroHeight(t *testing.T) {
	got := Frame("H", "B", "F", 0)
	if !strings.Contains(got, "B") {
		t.Errorf("body dropped at height 0:\n%s", got)
	}
}

func TestPadBetween_RoomForBoth(t *testing.T) {
	got := padBetween("LEFT", "RIGHT", 20)
	if len(got) < 20 {
		t.Errorf("got width %d", len(got))
	}
	if !strings.HasPrefix(got, "LEFT") || !strings.HasSuffix(got, "RIGHT") {
		t.Errorf("got %q", got)
	}
}

func TestPadBetween_NoRoom(t *testing.T) {
	// Width too small → single-space fallback.
	got := padBetween("LEFT", "RIGHT", 2)
	if !strings.Contains(got, "LEFT RIGHT") {
		t.Errorf("fallback failed: %q", got)
	}
}

func TestPadToHeight_AddsLines(t *testing.T) {
	got := padToHeight("line1\nline2", 5)
	if strings.Count(got, "\n") < 4 {
		t.Errorf("expected 5 rows, got: %q", got)
	}
}

func TestPadToHeight_NoOpWhenTallEnough(t *testing.T) {
	got := padToHeight("a\nb\nc", 2)
	if got != "a\nb\nc" {
		t.Errorf("got %q", got)
	}
}

func TestMaxHelper(t *testing.T) {
	if max(1, 2) != 2 {
		t.Error()
	}
	if max(5, 3) != 5 {
		t.Error()
	}
}
