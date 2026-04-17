package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/install"
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
)

func newUninstallCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall <skill>",
		Short: "Remove an installed skill from every target",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skill := args[0]

			m, err := manifest.Load(app.Config.ManifestPath)
			if err != nil {
				return fmt.Errorf("load manifest: %w", err)
			}
			entries := m.FindAll(skill)
			if len(entries) == 0 {
				app.UI.Warn("%s is not installed", skill)
				return nil
			}

			ok, err := app.Prompt.Confirm(
				fmt.Sprintf("Remove %s from %d target%s?", skill, len(entries), plural(len(entries))),
				true,
			)
			if err != nil {
				return err
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
		},
	}
}
