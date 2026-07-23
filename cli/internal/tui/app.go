// Package tui hosts every bubbletea model + shared TUI chrome used by the
// humblskills CLI. Each command's "interactive" mode (browse, wizard, confirm)
// lives here so the visual frame — header, keybar, palette — stays consistent
// across commands.
//
// Callers are expected to check ShouldUseTUI before invoking anything here;
// when that returns false the command should fall back to its static
// (pipe-friendly, --json-safe) path.
package tui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

// ShouldUseTUI reports whether the caller can safely take over the screen with
// an alt-screen bubbletea model. It's false whenever the user opted out
// explicitly (--json, --quiet, --yes) or the process isn't attached to a TTY
// on both stdin and stdout.
func ShouldUseTUI(jsonMode, quiet, yes bool) bool {
	if jsonMode || quiet || yes {
		return false
	}
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

// Run starts a bubbletea program using the shared set of program options
// (alt-screen + mouse cell motion). It blocks until the model finishes and
// returns the last rendered model so callers can extract results.
func Run(m tea.Model) (tea.Model, error) {
	// Opt-in single-program router: run every screen on one long-lived program
	// so the alt-screen isn't torn down between panes (no flash). Off by default.
	if RouterEnabled() {
		return runOnSession(m)
	}
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	return p.Run()
}
