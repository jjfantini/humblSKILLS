package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// Header renders the standard breadcrumb that sits atop every TUI model:
//
//	  humblskills › <command> › <detail>
//	  ──────────────────────────────────
//
// detail may be empty. width bounds the rule line so the frame keeps its shape
// on narrow terminals.
func Header(theme *ui.Theme, command, detail string, width int) string {
	if width < 20 {
		width = 20
	}
	line := "  " + theme.Brand.Render("humblskills") +
		theme.Crumb.Render("  ›  "+command)
	if detail != "" {
		line += theme.Crumb.Render("  ›  ") + theme.Hit.Render(detail)
	}
	rule := theme.RuleLine.Render(strings.Repeat("─", max(width-4, 4)))
	return line + "\n  " + rule
}

// KeyHint is one "label → keys" pair shown in the footer keybar.
type KeyHint struct {
	Keys  string
	Label string
}

// Footer renders the shared keybar. Example:
//
//	  ↑/↓ navigate · / filter · enter select · q quit
func Footer(theme *ui.Theme, hints []KeyHint) string {
	var parts []string
	for _, h := range hints {
		parts = append(parts, theme.Name.Render(h.Keys)+" "+theme.Crumb.Render(h.Label))
	}
	return "  " + strings.Join(parts, theme.Crumb.Render(" · "))
}

// Frame composes header + body + footer into a single renderable string,
// padding the body to bodyHeight rows so the footer sits at a predictable
// place regardless of content length.
func Frame(header, body, footer string, bodyHeight int) string {
	body = padToHeight(body, bodyHeight)
	return header + "\n\n" + body + "\n\n" + footer
}

func padToHeight(s string, h int) string {
	if h <= 0 {
		return s
	}
	lines := lipgloss.Height(s)
	if lines >= h {
		return s
	}
	return s + strings.Repeat("\n", h-lines)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
