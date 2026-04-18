package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"golang.org/x/term"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// termWidth returns the width of stdout, falling back to fallback when stdout
// isn't a TTY (piped output, CI) so the dashed rule still has a sensible length.
func termWidth(fallback int) int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return fallback
	}
	return w
}

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
	width := termWidth(72)
	fmt.Println()
	fmt.Println("  " + theme.Brand.Render(title))
	fmt.Println("  " + DashedRule(theme, width-4))
	if len(lines) > 0 {
		for _, ln := range lines {
			fmt.Println("  " + ln)
		}
		fmt.Println("  " + DashedRule(theme, width-4))
	}
	fmt.Println()

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
