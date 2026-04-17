package main

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
)

func newSearchCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: "Search the registry by name, description, or tag",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, _, err := registry.NewFetcher(app.Config.RegistryURL, app.Config.CacheDir).Load()
			if err != nil {
				return err
			}
			query := ""
			if len(args) == 1 {
				query = strings.ToLower(args[0])
			}

			var hits []registry.Skill
			for _, s := range reg.Skills {
				if query == "" || matches(s, query) {
					hits = append(hits, s)
				}
			}
			sort.Slice(hits, func(i, j int) bool { return hits[i].Name < hits[j].Name })

			if app.Config.JSON {
				return app.UI.JSON(struct {
					Query   string           `json:"query,omitempty"`
					Results []registry.Skill `json:"results"`
				}{query, hits})
			}
			if len(hits) == 0 {
				app.UI.Warn("no skills matched %q", query)
				return nil
			}
			for _, s := range hits {
				app.UI.Info("%s@%s — %s", s.Name, s.Version, s.Description)
				if len(s.Tags) > 0 {
					app.UI.Detail("  tags: %s", strings.Join(s.Tags, ", "))
				}
				if len(s.Platforms) > 0 {
					app.UI.Detail("  platforms: %s", strings.Join(s.Platforms, ", "))
				}
			}
			return nil
		},
	}
}

func matches(s registry.Skill, q string) bool {
	if strings.Contains(strings.ToLower(s.Name), q) {
		return true
	}
	if strings.Contains(strings.ToLower(s.Description), q) {
		return true
	}
	for _, t := range s.Tags {
		if strings.Contains(strings.ToLower(t), q) {
			return true
		}
	}
	return false
}
