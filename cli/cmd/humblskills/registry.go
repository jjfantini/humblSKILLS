package main

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
)

type refreshResult struct {
	URL       string    `json:"url"`
	Source    string    `json:"source"`
	Skills    int       `json:"skills"`
	FetchedAt time.Time `json:"fetched_at"`
	CachePath string    `json:"cache_path,omitempty"`
}

func newRegistryCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "Inspect and refresh the skill registry cache",
	}
	cmd.AddCommand(newRegistryRefreshCmd(app))
	return cmd
}

func newRegistryRefreshCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "refresh",
		Short: "Force a refresh of the cached registry",
		RunE: func(cmd *cobra.Command, _ []string) error {
			f := registry.NewFetcher(app.Config.RegistryURL, app.Config.CacheDir)
			reg, origin, err := f.Refresh()
			if err != nil {
				return err
			}
			info := f.Inspect()
			res := refreshResult{
				URL:       app.Config.RegistryURL,
				Source:    string(origin),
				Skills:    len(reg.Skills),
				FetchedAt: info.FetchedAt,
				CachePath: info.Path,
			}
			if app.Config.JSON {
				return app.UI.JSON(res)
			}
			app.UI.Success("registry refreshed: %d skill%s from %s", res.Skills, plural(res.Skills), res.Source)
			app.UI.Detail("  cache: %s", res.CachePath)
			return nil
		},
	}
}
