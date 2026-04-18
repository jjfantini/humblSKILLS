package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
)

func newSearchCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: "Search the registry by name, description, or tag",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, _, err := registry.NewFetcher(app.Config.RegistryURL, app.Config.CacheDir).Load()
			if err != nil {
				return err
			}
			query := ""
			if len(args) == 1 {
				query = strings.ToLower(args[0])
			}

			var hits []registry.Skill
			for _, s := range reg.Skills {
				if query == "" || matches(s, query) {
					hits = append(hits, s)
				}
			}
			sort.Slice(hits, func(i, j int) bool { return hits[i].Name < hits[j].Name })

			if app.Config.JSON {
				return app.UI.JSON(struct {
					Query   string           `json:"query,omitempty"`
					Results []registry.Skill `json:"results"`
				}{query, hits})
			}
			if len(hits) == 0 {
				app.UI.Warn("no skills matched %q", query)
				return nil
			}
			app.UI.Println(renderSearchResults(hits, query, app.Config.NoColor))
			return nil
		},
	}
}

func matches(s registry.Skill, q string) bool {
	if strings.Contains(strings.ToLower(s.Name), q) {
		return true
	}
	if strings.Contains(strings.ToLower(s.Description), q) {
		return true
	}
	for _, t := range s.Tags {
		if strings.Contains(strings.ToLower(t), q) {
			return true
		}
	}
	return false
}

// renderSearchResults returns a multi-line, styled string listing every
// skill. Layout is terminal-width-aware and degrades to plain ASCII when
// noColor is set.
func renderSearchResults(hits []registry.Skill, query string, noColor bool) string {
	r := lipgloss.DefaultRenderer()
	if noColor {
		r = lipgloss.NewRenderer(os.Stdout)
		r.SetColorProfile(termenv.Ascii)
	}

	width := termWidth()
	if width > 96 {
		width = 96
	}
	if width < 48 {
		width = 48
	}
	inner := width - 4 // gutter on each side

	var (
		brand     = r.NewStyle().Bold(true).Foreground(lipgloss.Color("#A78BFA"))
		crumb     = r.NewStyle().Foreground(lipgloss.Color("244"))
		count     = r.NewStyle().Foreground(lipgloss.Color("244")).Italic(true)
		rule      = r.NewStyle().Foreground(lipgloss.Color("238"))
		bullet    = r.NewStyle().Foreground(lipgloss.Color("#A78BFA"))
		nameStyle = r.NewStyle().Bold(true).Foreground(lipgloss.Color("#5EEAD4"))
		version   = r.NewStyle().Foreground(lipgloss.Color("244"))
		descStyle = r.NewStyle().Foreground(lipgloss.Color("252"))
		tag       = r.NewStyle().Foreground(lipgloss.Color("#93C5FD"))
		platform  = r.NewStyle().Foreground(lipgloss.Color("#F0ABFC"))
		hit       = r.NewStyle().Bold(true).Foreground(lipgloss.Color("#FDE68A"))
		label     = r.NewStyle().Foreground(lipgloss.Color("244"))
	)

	var sb strings.Builder

	// Header: "  humblskills › search  "foo""
	header := "  " + brand.Render("humblskills") + crumb.Render("  ›  search")
	if query != "" {
		header += crumb.Render("  ›  ") + hit.Render(fmt.Sprintf("%q", query))
	}
	sb.WriteString("\n")
	sb.WriteString(header)
	sb.WriteString("\n  ")
	sb.WriteString(rule.Render(strings.Repeat("─", inner)))
	sb.WriteString("\n")

	// Count line
	noun := "skill"
	if len(hits) != 1 {
		noun = "skills"
	}
	var summary string
	if query != "" {
		summary = fmt.Sprintf("%d %s matching your query", len(hits), noun)
	} else {
		summary = fmt.Sprintf("%d %s in registry", len(hits), noun)
	}
	sb.WriteString("  ")
	sb.WriteString(count.Render(summary))
	sb.WriteString("\n\n")

	for i, s := range hits {
		// Title line: ▸ name-with-highlight                              v0.1.0
		left := bullet.Render("▸ ") + highlightName(s.Name, query, nameStyle, hit)
		right := version.Render("v" + s.Version)
		pad := inner - lipgloss.Width(left) - lipgloss.Width(right)
		if pad < 1 {
			pad = 1
		}
		sb.WriteString("  " + left + strings.Repeat(" ", pad) + right + "\n")

		// Description, soft-wrapped to inner-2 and indented.
		if s.Description != "" {
			wrapped := descStyle.Width(inner - 2).Render(s.Description)
			for _, line := range strings.Split(wrapped, "\n") {
				sb.WriteString("    " + line + "\n")
			}
		}

		// Metadata: tags + platforms on one or two lines.
		if len(s.Tags) > 0 {
			chips := make([]string, 0, len(s.Tags))
			for _, t := range s.Tags {
				chips = append(chips, tag.Render("#"+t))
			}
			sb.WriteString("    " + label.Render("tags    ") + strings.Join(chips, "  ") + "\n")
		}
		if len(s.Platforms) > 0 {
			plats := make([]string, 0, len(s.Platforms))
			for _, pl := range s.Platforms {
				plats = append(plats, platform.Render(pl))
			}
			sb.WriteString("    " + label.Render("target  ") + strings.Join(plats, "  ") + "\n")
		}

		if i < len(hits)-1 {
			sb.WriteString("\n")
		}
	}

	// Hint footer.
	sb.WriteString("\n  ")
	sb.WriteString(rule.Render(strings.Repeat("─", inner)))
	sb.WriteString("\n  ")
	sb.WriteString(crumb.Render("install with  "))
	sb.WriteString(nameStyle.Render("humblskills install <name>"))
	sb.WriteString("\n")

	return sb.String()
}

// highlightName returns the skill name with case-insensitive query matches
// wrapped in the hit style. The non-matching parts are rendered with base.
// If query is empty, the name is returned under base untouched.
func highlightName(name, query string, base, hit lipgloss.Style) string {
	if query == "" {
		return base.Render(name)
	}
	lower := strings.ToLower(name)
	idx := strings.Index(lower, query)
	if idx < 0 {
		return base.Render(name)
	}
	var sb strings.Builder
	cursor := 0
	for idx >= 0 {
		if idx > cursor {
			sb.WriteString(base.Render(name[cursor:idx]))
		}
		end := idx + len(query)
		sb.WriteString(hit.Render(name[idx:end]))
		cursor = end
		rest := lower[cursor:]
		next := strings.Index(rest, query)
		if next < 0 {
			break
		}
		idx = cursor + next
	}
	if cursor < len(name) {
		sb.WriteString(base.Render(name[cursor:]))
	}
	return sb.String()
}

// termWidth returns a sensible content width, falling back to 80 when stdout
// isn't a TTY (piped output, CI, tests).
func termWidth() int {
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	return 80
}
