package main

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
)

func newListCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed skills",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runList(app, false)
		},
	}
}

// runList is the package-level entry point. `fromDashboard` softens ESC/quit
// semantics so the launcher loop in start.go can re-enter the dashboard
// instead of exiting the process.
func runList(app *App, fromDashboard bool) error {
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
		return runListTUI(app, m, fromDashboard)
	}
	renderListTable(app, installs)
	return nil
}

// renderListTable prints installs as a bordered table using the shared theme.
// Used in non-TTY / piped paths.
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

// runListTUI opens the shared two-pane browser in installed-only mode so list
// looks identical to search, differing only in which skills are shown and
// which actions are wired up.
func runListTUI(app *App, m *manifest.Manifest, fromDashboard bool) error {
	// Pull the registry to know whether each installed skill has drifted. If
	// the registry is unreachable, fall back to the manifest versions.
	reg, _, _ := registry.NewFetcher(app.Config.RegistryURL, app.Config.CacheDir).Load()

	// Build a "virtual" registry view of only the skills we have installed,
	// preferring the registry version when available so `outdated` can render.
	installedSkills := uniqueSkillsFromManifest(m)
	skills := make([]registry.Skill, 0, len(installedSkills))
	for _, name := range installedSkills {
		if s := findRegistrySkill(reg, name); s != nil {
			skills = append(skills, *s)
			continue
		}
		// Registry missing this skill — synthesise from the manifest entry so
		// it still shows in the browser.
		inst := m.FindAll(name)
		if len(inst) == 0 {
			continue
		}
		skills = append(skills, registry.Skill{
			Name:    name,
			Version: inst[0].Version,
		})
	}

	items := buildSkillItems(skills, m)

	skill, action, err := runSkillBrowser(app, "Installed", items, modeInstalledOnly, "no skills installed", fromDashboard)
	if err != nil {
		return err
	}
	switch action {
	case "update":
		return runUpdate(app, []string{skill}, updateFlags{})
	case "uninstall":
		return runUninstall(app, skill)
	}
	return nil
}

func uniqueSkillsFromManifest(m *manifest.Manifest) []string {
	seen := map[string]bool{}
	out := make([]string, 0)
	for _, inst := range m.Installations {
		if !seen[inst.Skill] {
			seen[inst.Skill] = true
			out = append(out, inst.Skill)
		}
	}
	sort.Strings(out)
	return out
}

func findRegistrySkill(reg *registry.Registry, name string) *registry.Skill {
	if reg == nil {
		return nil
	}
	for i := range reg.Skills {
		if reg.Skills[i].Name == name {
			return &reg.Skills[i]
		}
	}
	return nil
}
