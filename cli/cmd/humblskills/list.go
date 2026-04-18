package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

func newListCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed skills",
		RunE: func(cmd *cobra.Command, _ []string) error {
			m, err := manifest.Load(app.Config.ManifestPath)
			if err != nil {
				return err
			}
			if app.Config.JSON {
				return app.UI.JSON(m)
			}
			if len(m.Installations) == 0 {
				app.UI.Info("no skills installed — try 'humblskills install <name>'")
				return nil
			}
			installs := append([]manifest.Installation(nil), m.Installations...)
			sort.Slice(installs, func(i, j int) bool {
				if installs[i].Skill != installs[j].Skill {
					return installs[i].Skill < installs[j].Skill
				}
				if installs[i].Platform != installs[j].Platform {
					return installs[i].Platform < installs[j].Platform
				}
				return installs[i].Scope < installs[j].Scope
			})

			if tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes) {
				return runListTUI(app, installs)
			}
			renderListTable(app, installs)
			return nil
		},
	}
}

// renderListTable prints installs as a bordered table using the shared theme.
// Used in non-TTY / piped / --json-less paths.
func renderListTable(app *App, installs []manifest.Installation) {
	th := app.UI.Theme()
	app.UI.Header("list")

	rows := make([][]string, 0, len(installs))
	for _, inst := range installs {
		rows = append(rows, []string{
			th.Name.Render(inst.Skill),
			th.Version.Render("v" + inst.Version),
			th.Platform.Render(inst.Platform),
			th.Label.Render(inst.Scope),
			th.Detail.Render(inst.Path),
		})
	}
	tbl := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(th.RuleLine).
		Headers("Skill", "Version", "Platform", "Scope", "Path").
		Rows(rows...).
		StyleFunc(func(row, _ int) lipgloss.Style {
			if row == table.HeaderRow {
				return th.Label.Padding(0, 1).Bold(true)
			}
			return lipgloss.NewStyle().Padding(0, 1)
		})
	fmt.Fprintln(app.UI.Out(), tbl.Render())
	fmt.Fprintln(app.UI.Out(), "  "+th.Crumb.Render(fmt.Sprintf(
		"%d install%s total", len(installs), plural(len(installs)))))
}

// runListTUI launches the shared browser in "installed" mode. Actions: update,
// uninstall, or quit. After the browser exits, the appropriate command is
// routed back through runUpdate / Uninstall so the user stays in one flow.
func runListTUI(app *App, installs []manifest.Installation) error {
	items := make([]tui.BrowseItem, 0, len(installs))
	for _, inst := range installs {
		items = append(items, installedBrowseItem{Installation: inst})
	}
	m, err := tui.Run(tui.NewBrowser(tui.BrowserConfig{
		Command:  "installed",
		Theme:    app.UI.Theme(),
		Items:    items,
		Actions:  []tui.BrowseAction{tui.ActionUpdate, tui.ActionUninstall},
		EmptyMsg: "no skills installed",
	}))
	if err != nil {
		return err
	}
	browser, ok := m.(tui.Browser)
	if !ok {
		return nil
	}
	res := browser.Selected()
	it, ok := res.Item.(installedBrowseItem)
	if !ok {
		return nil
	}
	switch res.Action {
	case tui.ActionUpdate:
		return runUpdate(app, []string{it.Skill}, updateFlags{})
	case tui.ActionUninstall:
		return runUninstall(app, it.Skill)
	}
	return nil
}

// installedBrowseItem adapts manifest.Installation to the BrowseItem contract.
type installedBrowseItem struct {
	manifest.Installation
}

func (i installedBrowseItem) FilterValue() string {
	return i.Skill + " " + i.Platform + " " + i.Scope
}
func (i installedBrowseItem) Title() string { return i.Skill + "  v" + i.Version }
func (i installedBrowseItem) Description() string {
	return i.Platform + "/" + i.Scope
}
func (i installedBrowseItem) Preview(theme *ui.Theme, width int) string {
	var sb strings.Builder
	sb.WriteString(theme.Name.Render(i.Skill))
	sb.WriteString("  ")
	sb.WriteString(theme.Version.Render("v" + i.Version))
	sb.WriteString("\n\n")
	sb.WriteString(theme.Label.Render("platform  ") + theme.Platform.Render(i.Platform) + "\n")
	sb.WriteString(theme.Label.Render("scope     ") + theme.Platform.Render(i.Scope) + "\n")
	sb.WriteString(theme.Label.Render("path      ") + theme.Detail.Render(i.Path) + "\n")
	if !i.InstalledAt.IsZero() {
		sb.WriteString(theme.Label.Render("installed ") +
			theme.Detail.Render(i.InstalledAt.Format("2006-01-02 15:04")) + "\n")
	}
	return sb.String()
}
