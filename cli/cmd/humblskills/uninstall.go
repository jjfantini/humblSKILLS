package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/install"
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
)

func newUninstallCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall <skill>",
		Short: "Remove an installed skill from every target",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUninstall(app, args[0])
		},
	}
}

// runUninstall performs the interactive uninstall flow. Extracted from the
// cobra RunE so the list TUI can route an 'x' action back through exactly the
// same confirm + engine + print path.
func runUninstall(app *App, skill string) error {
	m, err := manifest.Load(app.Config.ManifestPath)
	if err != nil {
		return fmt.Errorf("load manifest: %w", err)
	}
	entries := m.FindAll(skill)
	if len(entries) == 0 {
		app.UI.Warn("%s is not installed", skill)
		return nil
	}

	// Show exactly what's about to be removed before asking.
	theme := app.UI.Theme()
	lines := make([]string, 0, len(entries))
	for _, e := range entries {
		lines = append(lines, fmt.Sprintf(
			"%s  %s  %s",
			theme.Name.Render(e.Skill+"@"+e.Version),
			theme.Platform.Render("["+e.Platform+"/"+e.Scope+"]"),
			theme.Detail.Render(e.Path),
		))
	}
	ok := true
	if !app.Config.Yes && !app.Config.JSON {
		got, err := tui.ConfirmWithSummary(
			theme,
			fmt.Sprintf("Uninstall %s", skill),
			fmt.Sprintf("Remove %d target%s?", len(entries), plural(len(entries))),
			lines,
			true,
			app.Prompt.Interactive,
		)
		if err != nil {
			return err
		}
		ok = got
	}
	if !ok {
		app.UI.Info("cancelled")
		return nil
	}

	engine := install.NewEngine(app.Config.CacheDir, app.Config.ManifestPath)
	res, err := engine.Uninstall(skill)
	if err != nil {
		return err
	}
	if app.Config.JSON {
		return app.UI.JSON(struct {
			Results []install.TargetResult `json:"results"`
		}{res})
	}
	for _, t := range res {
		app.UI.Success("removed %s [%s/%s]", t.Skill, t.Platform, t.Scope)
	}
	return nil
}
