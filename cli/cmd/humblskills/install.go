package main

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/install"
	"github.com/jjfantini/humblSKILLS/cli/internal/platform"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
)

type installFlags struct {
	platforms []string
	scope     string
	force     bool
}

func newInstallCmd(app *App) *cobra.Command {
	var f installFlags
	cmd := &cobra.Command{
		Use:   "install <skill>",
		Short: "Install a skill (and its deps) onto every detected platform",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstall(app, args[0], f)
		},
	}
	cmd.Flags().StringSliceVar(&f.platforms, "platform", nil, "restrict install to these adapters (default: all detected)")
	cmd.Flags().StringVar(&f.scope, "scope", "", "install scope (user|project; default: adapter's default)")
	cmd.Flags().BoolVar(&f.force, "force", false, "reinstall even if already up-to-date")
	return cmd
}

func runInstall(app *App, skill string, f installFlags) error {
	adapters, err := app.Adapters()
	if err != nil {
		return fmt.Errorf("load adapters: %w", err)
	}

	reg, _, err := registry.NewFetcher(app.Config.RegistryURL, app.Config.CacheDir).Load()
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	selected, err := selectPlatforms(adapters, f.platforms)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		return fmt.Errorf("no platforms selected — run 'humblskills doctor' to see what's detected")
	}

	plan, err := install.Plan(reg, skill)
	if err != nil {
		return err
	}

	app.UI.Detail("plan:")
	for _, s := range plan {
		tag := "root"
		if s.IsDep {
			tag = "dep"
		}
		app.UI.Detail("  %s %s@%s", tag, s.Skill.Name, s.Skill.Version)
	}

	engine := install.NewEngine(app.Config.CacheDir, app.Config.ManifestPath)
	res, err := engine.Execute(reg, plan, install.ExecuteOpts{
		Adapters:  adapters,
		Platforms: selected,
		Scope:     f.scope,
		Force:     f.force,
	})
	if err != nil {
		return err
	}

	if app.Config.JSON {
		return app.UI.JSON(res)
	}

	printInstall(app, res)
	return nil
}

// selectPlatforms returns the adapter names to install onto. If the user
// passed --platform, it's the intersection of that list with the declared
// adapters; otherwise it's every detected adapter.
func selectPlatforms(adapters []platform.Adapter, requested []string) ([]string, error) {
	known := platform.NameSet(adapters)
	if len(requested) > 0 {
		out := make([]string, 0, len(requested))
		for _, r := range requested {
			if _, ok := known[r]; !ok {
				return nil, fmt.Errorf("unknown platform %q", r)
			}
			out = append(out, r)
		}
		return out, nil
	}

	detected := platform.Detect(adapters)
	out := make([]string, 0, len(detected))
	for _, d := range detected {
		if d.Detected {
			out = append(out, d.Name)
		}
	}
	sort.Strings(out)
	return out, nil
}

func printInstall(app *App, r install.Result) {
	if len(r.Results) == 0 {
		app.UI.Warn("nothing to do — skill(s) declared no matching platforms")
		return
	}
	var installed, replaced, skipped, forced []install.TargetResult
	for _, t := range r.Results {
		switch t.Outcome {
		case install.OutcomeInstalled:
			installed = append(installed, t)
		case install.OutcomeReplaced:
			replaced = append(replaced, t)
		case install.OutcomeSkipped:
			skipped = append(skipped, t)
		case install.OutcomeForced:
			forced = append(forced, t)
		}
	}
	for _, t := range installed {
		app.UI.Success("installed %s → %s [%s/%s]", t.Skill, t.Path, t.Platform, t.Scope)
	}
	for _, t := range replaced {
		app.UI.Success("replaced %s → %s [%s/%s]", t.Skill, t.Path, t.Platform, t.Scope)
	}
	for _, t := range forced {
		app.UI.Success("reinstalled %s → %s [%s/%s]", t.Skill, t.Path, t.Platform, t.Scope)
	}
	for _, t := range skipped {
		app.UI.Detail("already up-to-date: %s [%s/%s]", t.Skill, t.Platform, t.Scope)
	}
	if len(installed)+len(replaced)+len(forced) == 0 {
		app.UI.Info("%d target%s already up-to-date (use --force to reinstall)", len(skipped), plural(len(skipped)))
	}
}
