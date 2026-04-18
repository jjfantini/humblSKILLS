package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// BadgeKind selects one of the reverse-video pill styles from the design
// handoff. Each kind is a semantic role: detected vs. missing for adapters,
// rw/ro for writable-target indicators, plus ghost for neutral chips.
type BadgeKind int

const (
	BadgeDetected BadgeKind = iota
	BadgeMissing
	BadgeRW
	BadgeRO
	BadgeGhost
)

// Badge renders `text` as a padded, reverse-video pill using the caller's
// theme. Every surface calls this helper so the design stays consistent.
func Badge(theme *ui.Theme, kind BadgeKind, text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	var style lipgloss.Style
	switch kind {
	case BadgeDetected:
		style = theme.BadgeDetected
	case BadgeMissing:
		style = theme.BadgeMissing
	case BadgeRW:
		style = theme.BadgeRW
	case BadgeRO:
		style = theme.BadgeRO
	default:
		style = theme.BadgeGhost
	}
	return style.Render(text)
}
