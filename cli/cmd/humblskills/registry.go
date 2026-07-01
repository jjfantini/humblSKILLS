package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
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
			return runRegistryRefresh(app)
		},
	}
}

func runRegistryRefresh(app *App) error {
	f := registry.NewFetcher(app.Config.RegistryURL, app.Config.CacheDir)

	refresh := func() (refreshResult, error) {
		r, o, err := f.Refresh()
		if err != nil {
			return refreshResult{}, err
		}
		info := f.Inspect()
		return refreshResult{
			URL:       app.Config.RegistryURL,
			Source:    string(o),
			Skills:    len(r.Skills),
			FetchedAt: info.FetchedAt,
			CachePath: info.Path,
		}, nil
	}

	if tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes) {
		p, err := profile.Load(app.Config.ProfilePath)
		if err != nil {
			return err
		}
		_, err = tui.ExecuteWithStatus(app.UI.Theme(), "registry", "refreshing registry…", p.StatusAutoReturnDuration(),
			func() (tui.StatusResult, error) {
				res, err := refresh()
				if err != nil {
					return tui.StatusResult{}, err
				}
				return tui.StatusResult{
					Headline: fmt.Sprintf("registry refreshed: %d skill%s from %s", res.Skills, plural(res.Skills), res.Source),
					Lines:    []string{"cache: " + res.CachePath},
				}, nil
			})
		return err
	}

	var res refreshResult
	err := tui.RunWithSpinner(app.UI.Theme(), "refreshing registry…", func() error {
		r, e := refresh()
		res = r
		return e
	})
	if err != nil {
		return err
	}
	if app.Config.JSON {
		return app.UI.JSON(res)
	}
	app.UI.Success("registry refreshed: %d skill%s from %s", res.Skills, plural(res.Skills), res.Source)
	app.UI.Detail("  cache: %s", res.CachePath)
	return nil
}
