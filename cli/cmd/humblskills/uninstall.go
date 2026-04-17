package main

import (
	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/install"
)

func newUninstallCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall <skill>",
		Short: "Remove an installed skill from every target",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := install.NewEngine(app.Config.CacheDir, app.Config.ManifestPath)
			res, err := engine.Uninstall(args[0])
			if err != nil {
				return err
			}
			if app.Config.JSON {
				return app.UI.JSON(struct {
					Results []install.TargetResult `json:"results"`
				}{res})
			}
			if len(res) == 0 {
				app.UI.Warn("%s is not installed", args[0])
				return nil
			}
			for _, t := range res {
				app.UI.Success("removed %s [%s/%s]", t.Skill, t.Platform, t.Scope)
			}
			return nil
		},
	}
}
