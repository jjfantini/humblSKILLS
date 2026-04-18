package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
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

			// With no query and an interactive terminal, drop into the TUI
			// browser so the user can filter + preview + install without
			// leaving the CLI.
			useTUI := query == "" && tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes)
			if useTUI {
				return runSearchTUI(app, hits)
			}

			app.UI.Println(renderSearchResults(app.UI.Theme(), hits, query))
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

// runSearchTUI launches the bubbletea browser over hits. When the user exits
// with the install action, the selected skill is handed straight to
// runInstall so one TUI flows into the next.
func runSearchTUI(app *App, hits []registry.Skill) error {
	items := make([]tui.BrowseItem, 0, len(hits))
	for _, s := range hits {
		items = append(items, skillBrowseItem{Skill: s})
	}
	m, err := tui.Run(tui.NewBrowser(tui.BrowserConfig{
		Command:  "search",
		Theme:    app.UI.Theme(),
		Items:    items,
		Actions:  []tui.BrowseAction{tui.ActionInstall},
		EmptyMsg: "no skills in registry",
	}))
	if err != nil {
		return err
	}
	browser, ok := m.(tui.Browser)
	if !ok {
		return nil
	}
	res := browser.Selected()
	if res.Action != tui.ActionInstall {
		return nil
	}
	it, ok := res.Item.(skillBrowseItem)
	if !ok {
		return nil
	}
	app.UI.Info("installing %s…", app.UI.Theme().Name.Render(it.Name))
	return runInstall(app, it.Name, installFlags{})
}

// skillBrowseItem adapts registry.Skill to the tui.BrowseItem contract.
type skillBrowseItem struct {
	registry.Skill
}

func (s skillBrowseItem) FilterValue() string {
	return s.Name + " " + strings.Join(s.Tags, " ") + " " + s.Skill.Description
}
func (s skillBrowseItem) Title() string { return s.Name + "  v" + s.Version }
func (s skillBrowseItem) Description() string {
	return s.Skill.Description
}
func (s skillBrowseItem) Preview(theme *ui.Theme, width int) string {
	var sb strings.Builder
	sb.WriteString(theme.Name.Render(s.Name))
	sb.WriteString("  ")
	sb.WriteString(theme.Version.Render("v" + s.Version))
	sb.WriteString("\n\n")
	if s.Skill.Description != "" {
		sb.WriteString(theme.Desc.Width(width).Render(s.Skill.Description))
		sb.WriteString("\n\n")
	}
	if len(s.Tags) > 0 {
		chips := make([]string, 0, len(s.Tags))
		for _, t := range s.Tags {
			chips = append(chips, theme.Tag.Render("#"+t))
		}
		sb.WriteString(theme.Label.Render("tags    ") + strings.Join(chips, "  ") + "\n")
	}
	if len(s.Platforms) > 0 {
		plats := make([]string, 0, len(s.Platforms))
		for _, p := range s.Platforms {
			plats = append(plats, theme.Platform.Render(p))
		}
		sb.WriteString(theme.Label.Render("target  ") + strings.Join(plats, "  ") + "\n")
	}
	if len(s.Requires) > 0 {
		sb.WriteString(theme.Label.Render("deps    ") +
			theme.Detail.Render(strings.Join(s.Requires, ", ")) + "\n")
	}
	return sb.String()
}

// renderSearchResults returns a multi-line, themed string listing every hit.
// Layout is terminal-width-aware; every colour token is sourced from the
// shared Theme so styling stays in sync with every other surface.
func renderSearchResults(theme *ui.Theme, hits []registry.Skill, query string) string {
	width := termWidth()
	if width > 96 {
		width = 96
	}
	if width < 48 {
		width = 48
	}
	inner := width - 4

	var sb strings.Builder
	header := "  " + theme.Brand.Render("humblskills") + theme.Crumb.Render("  ›  search")
	if query != "" {
		header += theme.Crumb.Render("  ›  ") + theme.Hit.Render(fmt.Sprintf("%q", query))
	}
	sb.WriteString("\n")
	sb.WriteString(header)
	sb.WriteString("\n  ")
	sb.WriteString(theme.RuleLine.Render(strings.Repeat("─", inner)))
	sb.WriteString("\n")

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
	sb.WriteString(theme.Count.Render(summary))
	sb.WriteString("\n\n")

	for i, s := range hits {
		left := theme.Bullet.Render("▸ ") + highlightName(s.Name, query, theme.Name, theme.Hit)
		right := theme.Version.Render("v" + s.Version)
		pad := inner - lipgloss.Width(left) - lipgloss.Width(right)
		if pad < 1 {
			pad = 1
		}
		sb.WriteString("  " + left + strings.Repeat(" ", pad) + right + "\n")

		if s.Description != "" {
			wrapped := theme.Desc.Width(inner - 2).Render(s.Description)
			for _, line := range strings.Split(wrapped, "\n") {
				sb.WriteString("    " + line + "\n")
			}
		}
		if len(s.Tags) > 0 {
			chips := make([]string, 0, len(s.Tags))
			for _, t := range s.Tags {
				chips = append(chips, theme.Tag.Render("#"+t))
			}
			sb.WriteString("    " + theme.Label.Render("tags    ") + strings.Join(chips, "  ") + "\n")
		}
		if len(s.Platforms) > 0 {
			plats := make([]string, 0, len(s.Platforms))
			for _, pl := range s.Platforms {
				plats = append(plats, theme.Platform.Render(pl))
			}
			sb.WriteString("    " + theme.Label.Render("target  ") + strings.Join(plats, "  ") + "\n")
		}

		if i < len(hits)-1 {
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n  ")
	sb.WriteString(theme.RuleLine.Render(strings.Repeat("─", inner)))
	sb.WriteString("\n  ")
	sb.WriteString(theme.Crumb.Render("install with  "))
	sb.WriteString(theme.Name.Render("humblskills install <name>"))
	sb.WriteString("\n")

	return sb.String()
}

// highlightName wraps every case-insensitive query match inside hit style,
// leaving non-matching segments in base. Empty query → untouched name.
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

func termWidth() int {
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	return 80
}
