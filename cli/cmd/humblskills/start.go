package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/adapters"
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
// Non-TTY paths fall through to the Cobra help text via printStartFallback,
// unless --fullscreen was asked for explicitly — then we surface an error
// instead of silently printing help.
func runStart(app *App) error {
	if !tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes) {
		if app.Config.Fullscreen {
			return fmt.Errorf("--fullscreen requires an interactive terminal (no TTY detected)")
		}
		return printStartFallback(app)
	}

	for {
		summary := buildDashboardSummary(app)
		status := tui.DashboardStatus{
			Healthy:   summary.drifted == 0,
			Platforms: summary.platforms,
			Skills:    summary.skills,
		}
		cfg := tui.DashboardConfig{
			Theme:   app.UI.Theme(),
			Version: resolveVersion().Version,
			Status:  status,
			Tiles:   tui.DefaultDashboardTiles(),
		}
		res, err := tui.RunDashboard(cfg)
		if err != nil {
			return err
		}
		if res.Quit || res.Command == "" {
			return nil
		}
		// Stash the breadcrumb + status on the App so every sub-screen renders
		// the same top header as the dashboard.
		app.Nav = NavContext{
			Crumb:  "Dashboard > " + crumbLabel(res.Command),
			Status: status,
		}
		if err := dispatchDashboardCommand(app, res.Command); err != nil {
			// Surface the error, then loop back so the user can keep working.
			app.UI.Error("%s: %v", res.Command, err)
		}
		app.Nav = NavContext{}
	}
}

// crumbLabel maps a dashboard command back to its display label for the
// breadcrumb. Falls back to Title-case of the command itself.
func crumbLabel(cmd string) string {
	switch cmd {
	case "install", "list", "update", "upgrade", "search", "uninstall", "profile", "eval", "doctor", "registry", "version":
		return strings.ToUpper(cmd[:1]) + cmd[1:]
	}
	return cmd
}

func dispatchDashboardCommand(app *App, cmd string) error {
	switch cmd {
	case "install":
		return runInstall(app, "", installFlags{}, true)
	case "list":
		return runList(app, true)
	case "update":
		return runUpdate(app, nil, updateFlags{})
	case "upgrade":
		return runUpgrade(app, upgradeFlags{})
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
	case "eval":
		return runEvalTUI(app)
	case "doctor":
		return runDoctor(app)
	case "registry":
		return runRegistryRefresh(app)
	case "version":
		return runVersion(app)
	}
	return fmt.Errorf("unknown dashboard command: %s", cmd)
}

// dashboardSummary is the data we pull once per dashboard re-entry to
// populate the top header status line (healthy · N platforms · M skills).
type dashboardSummary struct {
	drifted   int
	platforms int
	skills    int
}

func buildDashboardSummary(app *App) dashboardSummary {
	var s dashboardSummary

	if adapterList, err := app.Adapters(); err == nil {
		for _, r := range adapters.Detect(adapterList) {
			if r.Detected {
				s.platforms++
			}
		}
	}

	m, err := manifest.Load(app.Config.ManifestPath)
	if err != nil || m == nil {
		return s
	}
	seen := map[string]bool{}
	for _, inst := range m.Installations {
		if !seen[inst.Skill] {
			seen[inst.Skill] = true
			s.skills++
		}
	}

	reg, _, err := registry.NewFetcher(app.Config.RegistryURL, app.Config.CacheDir).Load()
	if err == nil && reg != nil {
		s.drifted = len(install.PlanUpdates(reg, m, nil))
	}
	return s
}

// headerSection returns the header Section string the caller should render.
// If Nav has a breadcrumb (we got here from the dashboard), use that; otherwise
// fall back to the caller-provided default (direct CLI invocation path).
func (a *App) headerSection(fallback string) string {
	if a.Nav.Crumb != "" {
		return a.Nav.Crumb
	}
	return fallback
}

// headerMeta returns the right-anchored header Meta string. If Nav is set
// (dashboard-launched sub-screen), render the shared status line; otherwise
// fall back to the caller-provided default (often "" or a command-specific
// summary).
func (a *App) headerMeta(fallback string) string {
	if a.Nav.Crumb != "" {
		return tui.RenderStatusMeta(a.UI.Theme(), a.Nav.Status)
	}
	return fallback
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
		{"upgrade", "upgrade the humblskills CLI itself"},
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
