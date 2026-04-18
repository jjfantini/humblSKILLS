package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// ConfirmWithSummary renders a framed summary panel listing the items being
// acted on, then asks the user to confirm via the shared huh theme. Use it for
// destructive actions (uninstall, force-reinstall) where the user should see
// exactly what's about to change before they press enter.
//
// Returns (answer, error). When the process isn't interactive, the panel is
// still printed (so CI logs show intent) but the confirm falls back to dflt.
func ConfirmWithSummary(
	theme *ui.Theme,
	title, prompt string,
	lines []string,
	dflt bool,
	interactive bool,
) (bool, error) {
	fmt.Println()
	fmt.Println("  " + theme.Brand.Render(title))
	fmt.Println()
	if len(lines) > 0 {
		body := strings.Join(lines, "\n")
		fmt.Println(theme.Panel.Render(body))
		fmt.Println()
	}

	if !interactive {
		return dflt, nil
	}

	v := dflt
	err := huh.NewConfirm().
		Title(prompt).
		Affirmative("Yes").
		Negative("No").
		Value(&v).
		WithTheme(ui.HuhTheme(theme)).
		Run()
	if err != nil {
		return dflt, err
	}
	return v, nil
}
