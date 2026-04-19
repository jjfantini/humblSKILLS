package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/install"
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
)

func newStartCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Open the humblskills dashboard (default when run on a TTY)",
		Long: "start opens the interactive dashboard — a tile grid with fuzzy " +
			"search that routes into every humblskills command. When stdout " +
			"isn't a terminal, prints the same command table as `--help`.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runStart(app)
		},
	}
}

// runStart is the launcher loop. It re-enters the dashboard after every
// sub-command returns, so ESC in any sub-screen bounces back to the grid.
// Non-TTY paths fall through to the Cobra help text via printStartFallback.
func runStart(app *App) error {
	if !tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes) {
		return printStartFallback(app)
	}

	for {
		cfg := tui.DashboardConfig{
			Theme:    app.UI.Theme(),
			Version:  resolveVersion().Version,
			Greeting: tui.BuildDashboardGreeting(countDrifted(app)),
			Tiles:    tui.DefaultDashboardTiles(),
		}
		res, err := tui.RunDashboard(cfg)
		if err != nil {
			return err
		}
		if res.Quit || res.Command == "" {
			return nil
		}
		if err := dispatchDashboardCommand(app, res.Command); err != nil {
			// Surface the error, then loop back so the user can keep working.
			app.UI.Error("%s: %v", res.Command, err)
		}
	}
}

func dispatchDashboardCommand(app *App, cmd string) error {
	switch cmd {
	case "install":
		return runInstall(app, "", installFlags{}, true)
	case "list":
		return runList(app, true)
	case "update":
		return runUpdate(app, nil, updateFlags{})
	case "search":
		reg, _, err := registry.NewFetcher(app.Config.RegistryURL, app.Config.CacheDir).Load()
		if err != nil {
			return err
		}
		hits := append([]registry.Skill(nil), reg.Skills...)
		return runSearchTUI(app, hits, true)
	case "uninstall":
		return runUninstallPicker(app, true)
	case "profile":
		return runProfileEditor(app)
	case "doctor":
		return runDoctor(app)
	case "registry":
		return runRegistryRefresh(app)
	case "version":
		return runVersion(app)
	}
	return fmt.Errorf("unknown dashboard command: %s", cmd)
}

// countDrifted counts how many installed skills have a newer registry version.
// Returns 0 on any error — the banner treats it as "up-to-date".
func countDrifted(app *App) int {
	m, err := manifest.Load(app.Config.ManifestPath)
	if err != nil || m == nil {
		return 0
	}
	reg, _, err := registry.NewFetcher(app.Config.RegistryURL, app.Config.CacheDir).Load()
	if err != nil || reg == nil {
		return 0
	}
	return len(install.PlanUpdates(reg, m, nil))
}

func printStartFallback(app *App) error {
	th := app.UI.Theme()
	out := app.UI.Out()
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  "+th.Brand.Render("humblskills")+"  "+th.Crumb.Render("— skill installer for Claude Code, Cursor, and friends"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  "+th.SectionTitle.Render("COMMANDS"))
	cmds := []struct{ name, desc string }{
		{"install", "add a skill to every detected platform"},
		{"list", "show installed skills"},
		{"update", "pull newer registry versions"},
		{"search", "browse the registry"},
		{"uninstall", "remove a skill"},
		{"profile", "edit install defaults"},
		{"doctor", "inspect adapter health"},
		{"registry refresh", "refresh the registry cache"},
		{"version", "print version + commit"},
	}
	for _, c := range cmds {
		fmt.Fprintf(out, "  %s  %s\n", th.Name.Render(padRight(c.name, 18)), th.Detail.Render(c.desc))
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  "+th.Crumb.Render("run 'humblskills <cmd> --help' for flags"))
	fmt.Fprintln(out)
	return nil
}
