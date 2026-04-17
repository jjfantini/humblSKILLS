package main

import (
	"sort"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
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
			for _, inst := range installs {
				app.UI.Info("%s@%s  [%s/%s]  %s", inst.Skill, inst.Version, inst.Platform, inst.Scope, inst.Path)
			}
			return nil
		},
	}
}
