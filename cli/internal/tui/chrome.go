package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/jjfantini/humblSKILLS/cli/v2/internal/ui"
)

// HeaderSpec is the bundle of text the header line renders. Version is the
// muted version tag ("v0.4.2"); Section is the capitalised crumb ("Adapters");
// Meta is the right-anchored muted line ("2 / 2 detected · scanned 18:04").
type HeaderSpec struct {
	Version string
	Section string
	Meta    string
}

// Header renders the design-01 header:
//
//	  humblskills v0.4.2  · Adapters                        2 / 2 detected
//	  ╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌
//
// width is the total line width; the dashed rule stretches edge-to-edge.
func Header(theme *ui.Theme, spec HeaderSpec, width int) string {
	if width < 40 {
		width = 40
	}
	left := theme.Brand.Render("humblskills")
	if spec.Version != "" {
		left += " " + theme.Version.Render(spec.Version)
	}
	if spec.Section != "" {
		left += "  " + theme.Crumb.Render("· "+spec.Section)
	}
	right := ""
	if spec.Meta != "" {
		right = theme.Meta.Render(spec.Meta)
	}
	line := padBetween(left, right, width-2)
	rule := DashedRule(theme, width-2)
	return "  " + line + "\n  " + rule
}

// KeyHint is one footer key → label pair.
type KeyHint struct {
	Keys  string
	Label string
}

// Footer renders the design-01 footer:
//
//	  ╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌
//	  [↑↓] select  [r] rescan  [q] quit               focused: claude-code
//
// right is the optional right-anchored context (e.g. "focused: <name>").
func Footer(theme *ui.Theme, hints []KeyHint, right string, width int) string {
	if width < 40 {
		width = 40
	}
	parts := make([]string, 0, len(hints))
	for _, h := range hints {
		parts = append(parts, theme.KbdKey.Render(h.Keys)+" "+theme.KbdLabel.Render(h.Label))
	}
	left := strings.Join(parts, "  ")
	// `right` is passed through unchanged so callers can compose mixed
	// styles (e.g. muted label + magenta value for `focused: <name>`).
	line := composeFooterLine(left, right, width-2)
	rule := DashedRule(theme, width-2)
	return "  " + rule + "\n  " + line
}

// composeFooterLine fits the hint block (left) and context (right) into avail
// display columns without overflowing — an overflowing footer wraps in the
// alt-screen and shoves the whole layout up a row. Priority: keep the
// keybindings. If left+right won't fit, drop the (secondary) right context;
// if the hints alone still overflow, truncate them with an ellipsis.
func composeFooterLine(left, right string, avail int) string {
	if avail < 1 {
		avail = 1
	}
	lw := lipgloss.Width(left)
	rw := lipgloss.Width(right)
	if rw > 0 && lw+1+rw <= avail {
		return padBetween(left, right, avail)
	}
	if lw <= avail {
		return left
	}
	return ansi.Truncate(left, avail, "…")
}

// DashedRule returns a dashed horizontal line (╌) coloured with the border
// token. Terminals that lack the dash codepoint fall back via lipgloss.
func DashedRule(theme *ui.Theme, width int) string {
	if width < 4 {
		width = 4
	}
	return theme.RuleLine.Render(strings.Repeat("╌", width))
}

// Frame composes header + body + footer into one renderable string, fitting
// the body to exactly bodyHeight rows so the footer sits at a predictable y.
func Frame(header, body, footer string, bodyHeight int) string {
	body = fitToHeight(body, bodyHeight)
	return header + "\n\n" + body + "\n" + footer
}

// padBetween places left and right on the same row separated by enough spaces
// to stretch the composed line to `width` display columns. Falls back to a
// single space when there isn't room.
func padBetween(left, right string, width int) string {
	lw := lipgloss.Width(left)
	rw := lipgloss.Width(right)
	gap := width - lw - rw
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

// fitToHeight forces s to exactly h rows: short bodies get blank rows appended
// so the footer lands at a fixed y, and over-tall bodies get their overflow
// dropped.
//
// The truncation half is load-bearing, not defensive. Every screen here renders
// into an alt-screen whose height is fixed, so a body even one row too tall
// makes the terminal scroll the whole frame up — the header disappears off the
// top and nothing can bring it back, because the frame is re-rendered from
// scratch each tick with the same overflow. Screens that own more content than
// fits must window it themselves (a viewport, or Model's left-pane window);
// this clamp only guarantees the chrome survives when one doesn't.
func fitToHeight(s string, h int) string {
	if h <= 0 {
		return s
	}
	lines := lipgloss.Height(s)
	if lines == h {
		return s
	}
	if lines < h {
		return s + strings.Repeat("\n", h-lines)
	}
	return strings.Join(strings.Split(s, "\n")[:h], "\n")
}

// fitToWidth forces s to exactly w display cells, padding short strings with
// spaces and truncating long ones with an ellipsis. ANSI-aware on both sides:
// lipgloss.Width ignores escape sequences when measuring, and ansi.Truncate
// keeps them intact so a clipped row can't leak its styling into the rest of
// the line.
func fitToWidth(s string, w int) string {
	if w < 1 {
		return ""
	}
	switch sw := lipgloss.Width(s); {
	case sw == w:
		return s
	case sw < w:
		return s + strings.Repeat(" ", w-sw)
	default:
		return ansi.Truncate(s, w, "…")
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
