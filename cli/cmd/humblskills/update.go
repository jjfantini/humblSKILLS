package main

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/install"
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

type updateFlags struct {
	all   bool
	check bool
}

func newUpdateCmd(app *App) *cobra.Command {
	var f updateFlags
	cmd := &cobra.Command{
		Use:   "update [<skill>...]",
		Short: "Upgrade installed skills to the latest registry version",
		Long: "update with no args presents a picker of every skill that has " +
			"drifted from the registry. Names can be passed to narrow the set. " +
			"--all (or --yes) skips the picker and upgrades every drifted skill. " +
			"--check prints the diff and exits without changing anything.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(app, args, f)
		},
	}
	cmd.Flags().BoolVar(&f.all, "all", false, "update every drifted skill without prompting")
	cmd.Flags().BoolVar(&f.check, "check", false, "print what would change and exit")
	return cmd
}

func runUpdate(app *App, only []string, f updateFlags) error {
	m, err := manifest.Load(app.Config.ManifestPath)
	if err != nil {
		return fmt.Errorf("load manifest: %w", err)
	}
	if len(m.Installations) == 0 {
		app.UI.Info("no skills installed")
		return nil
	}

	reg, _, err := registry.NewFetcher(app.Config.RegistryURL, app.Config.CacheDir).Load()
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	plans := install.PlanUpdates(reg, m, only)
	sort.Slice(plans, func(i, j int) bool { return plans[i].Skill < plans[j].Skill })

	if f.check {
		return printUpdateCheck(app, plans)
	}

	if len(plans) == 0 {
		if len(only) > 0 {
			app.UI.Info("selected skills are already up-to-date")
		} else {
			app.UI.Info("all skills are up-to-date")
		}
		return nil
	}

	selected, err := chooseUpdates(app, plans, f.all)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		app.UI.Info("nothing selected")
		return nil
	}

	adapters, err := app.Adapters()
	if err != nil {
		return fmt.Errorf("load adapters: %w", err)
	}
	adapterKnown := map[string]struct{}{}
	for _, a := range adapters {
		adapterKnown[a.Name] = struct{}{}
	}

	engine := install.NewEngine(app.Config.CacheDir, app.Config.ManifestPath)
	useTUI := tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes)
	var aggregate install.Result

	run := func(sink install.EventSink) error {
		for _, plan := range selected {
			stepPlan, err := install.Plan(reg, plan.Skill)
			if err != nil {
				return err
			}

			// Group existing targets by scope; within each scope, the set of
			// platforms is exactly the adapters this skill is currently installed
			// onto. Skip platforms the binary no longer knows about (warn once).
			byScope := map[string][]string{}
			for _, t := range plan.Targets {
				if _, ok := adapterKnown[t.Platform]; !ok {
					app.UI.Warn("skipping unknown platform %q in manifest for %s", t.Platform, plan.Skill)
					continue
				}
				byScope[t.Scope] = append(byScope[t.Scope], t.Platform)
			}

			scopes := make([]string, 0, len(byScope))
			for s := range byScope {
				scopes = append(scopes, s)
			}
			sort.Strings(scopes)

			for _, scope := range scopes {
				plats := byScope[scope]
				sort.Strings(plats)
				res, err := engine.Execute(reg, stepPlan, install.ExecuteOpts{
					Adapters:  adapters,
					Platforms: plats,
					Scope:     scope,
					Force:     true,
					OnEvent:   sink,
				})
				if err != nil {
					return fmt.Errorf("%s: %w", plan.Skill, err)
				}
				aggregate.Results = append(aggregate.Results, res.Results...)
			}
		}
		return nil
	}

	if useTUI {
		if err := tui.ExecuteWithProgress(app.UI.Theme(), "update", run); err != nil {
			return err
		}
	} else {
		if err := run(nil); err != nil {
			return err
		}
	}

	if app.Config.JSON {
		return app.UI.JSON(aggregate)
	}
	printInstall(app, aggregate)
	return nil
}

func chooseUpdates(app *App, plans []install.UpdatePlan, all bool) ([]install.UpdatePlan, error) {
	if all || app.Config.Yes {
		return plans, nil
	}

	opts := make([]ui.MultiSelectOption, 0, len(plans))
	for _, p := range plans {
		opts = append(opts, ui.MultiSelectOption{
			Label:    fmt.Sprintf("%s  %s → %s", p.Skill, p.FromVersion, p.ToVersion),
			Value:    p.Skill,
			Selected: true,
		})
	}
	picked, err := app.Prompt.MultiSelect("Select skills to update", opts)
	if err != nil {
		return nil, err
	}
	pickedSet := map[string]struct{}{}
	for _, v := range picked {
		pickedSet[v] = struct{}{}
	}
	out := make([]install.UpdatePlan, 0, len(picked))
	for _, p := range plans {
		if _, ok := pickedSet[p.Skill]; ok {
			out = append(out, p)
		}
	}
	return out, nil
}

func printUpdateCheck(app *App, plans []install.UpdatePlan) error {
	if app.Config.JSON {
		return app.UI.JSON(struct {
			Updates []install.UpdatePlan `json:"updates"`
		}{plans})
	}
	if len(plans) == 0 {
		app.UI.Info("all skills are up-to-date")
		return nil
	}
	app.UI.Info("%d skill%s can be updated:", len(plans), plural(len(plans)))
	for _, p := range plans {
		app.UI.Detail("  %s  %s → %s  (%d target%s)",
			p.Skill, p.FromVersion, p.ToVersion, len(p.Targets), plural(len(p.Targets)))
	}
	return nil
}
