package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

func asciiTheme() *ui.Theme {
	return ui.NewTheme(ui.DefaultPalette(), nil, true)
}

// A footer that overflows the terminal width wraps in the alt-screen and shoves
// the layout up a row, so no rendered line may exceed the requested width.
func TestFooter_DoesNotOverflowWidth(t *testing.T) {
	th := asciiTheme()
	hints := []KeyHint{
		{Keys: "↑↓", Label: "select"},
		{Keys: "/", Label: "filter"},
		{Keys: "⇞⇟", Label: "scroll"},
		{Keys: "i/enter", Label: "install"},
		{Keys: "u", Label: "update"},
		{Keys: "x", Label: "uninstall"},
		{Keys: "q", Label: "quit"},
	}
	for _, w := range []int{40, 55, 80, 120} {
		f := Footer(th, hints, "focused: some-really-long-skill-name", w)
		for _, ln := range strings.Split(f, "\n") {
			if got := lipgloss.Width(ln); got > w {
				t.Errorf("width=%d: footer line %q width=%d exceeds %d", w, ln, got, w)
			}
		}
	}
}

func TestComposeFooterLine_TruncatesLongHints(t *testing.T) {
	left := strings.Repeat("x", 100)
	got := composeFooterLine(left, "", 20)
	if w := lipgloss.Width(got); w > 20 {
		t.Errorf("width=%d exceeds 20: %q", w, got)
	}
	if !strings.HasSuffix(got, "…") {
		t.Errorf("expected ellipsis suffix, got %q", got)
	}
}

func TestComposeFooterLine_DropsRightWhenTight(t *testing.T) {
	left := strings.Repeat("x", 18)
	right := "yyyy"
	// 18 + 1 + 4 = 23 > 20, so the right context must be dropped entirely.
	got := composeFooterLine(left, right, 20)
	if strings.Contains(got, "y") {
		t.Errorf("right context should be dropped when it doesn't fit: %q", got)
	}
	if w := lipgloss.Width(got); w > 20 {
		t.Errorf("width=%d exceeds 20: %q", w, got)
	}
}

func TestComposeFooterLine_KeepsRightWhenItFits(t *testing.T) {
	got := composeFooterLine("abc", "xyz", 40)
	if !strings.Contains(got, "abc") || !strings.Contains(got, "xyz") {
		t.Errorf("expected both left and right to fit: %q", got)
	}
}
