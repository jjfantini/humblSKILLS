package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/install"
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
)

func newUninstallCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall [skill]",
		Short: "Remove an installed skill from every target",
		Long: "uninstall <skill> removes a named skill. With no arg, it opens " +
			"an interactive picker listing every installed skill.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return runUninstall(app, args[0])
			}
			return runUninstallPicker(app, false)
		},
	}
}

// runUninstallPicker opens the shared two-pane browser over installed skills
// so the user can pick-and-remove without leaving the TUI.
func runUninstallPicker(app *App, fromDashboard bool) error {
	if !tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes) {
		return fmt.Errorf("skill name required — usage: humblskills uninstall <skill>")
	}

	m, err := manifest.Load(app.Config.ManifestPath)
	if err != nil {
		return fmt.Errorf("load manifest: %w", err)
	}
	if len(m.Installations) == 0 {
		app.UI.Info("no skills installed — nothing to uninstall")
		return nil
	}

	reg, _, _ := registry.NewFetcher(app.Config.RegistryURL, app.Config.CacheDir).Load()

	installedNames := uniqueSkillsFromManifest(m)
	skills := make([]registry.Skill, 0, len(installedNames))
	for _, name := range installedNames {
		if s := findRegistrySkill(reg, name); s != nil {
			skills = append(skills, *s)
			continue
		}
		inst := m.FindAll(name)
		if len(inst) == 0 {
			continue
		}
		skills = append(skills, registry.Skill{Name: name, Version: inst[0].Version})
	}
	items := buildSkillItems(skills, m)

	skill, action, err := runSkillBrowser(app, "Uninstall", items, modeInstalledOnly, "no skills installed", fromDashboard)
	if err != nil {
		return err
	}
	if action != "uninstall" || skill == "" {
		return nil
	}
	return runUninstall(app, skill)
}

// runUninstall performs the confirm → engine → print flow for a named skill.
// Extracted so list/uninstall TUIs can route straight into it.
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
