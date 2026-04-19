package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
	force bool
}

func newUpdateCmd(app *App) *cobra.Command {
	var f updateFlags
	cmd := &cobra.Command{
		Use:   "update [<skill>...]",
		Short: "Upgrade installed skills to the latest registry version",
		Long: "update with no args opens an interactive picker of every skill that " +
			"has drifted from the registry. Names can be passed to narrow the set. " +
			"--all (or --yes) skips the picker and upgrades every drifted skill. " +
			"--check prints the diff and exits without changing anything. " +
			"By default, the preserve list on each installed SKILL.md is honored so " +
			"your local customizations survive. --force ignores local preserve edits " +
			"and reinstalls cleanly from the registry (equivalent to uninstall + install).",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(app, args, f)
		},
	}
	cmd.Flags().BoolVar(&f.all, "all", false, "update every drifted skill without prompting")
	cmd.Flags().BoolVar(&f.check, "check", false, "print what would change and exit")
	cmd.Flags().BoolVar(&f.force, "force", false, "bypass local preserve edits; reinstall cleanly from registry")
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
					Force:     f.force,
					OnEvent:   sink,
				})
				if err != nil {
					return fmt.Errorf("%s: %w", plan.Skill, err)
				}
				aggregate.Results = append(aggregate.Results, res.Results...)
				aggregate.Warnings = append(aggregate.Warnings, res.Warnings...)
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

// chooseUpdates is the pre-execute picker. With --all / --yes it returns every
// plan. On an interactive TTY it opens the two-pane listdetail so the user can
// inspect each drifted skill before applying. Non-interactive (pipe, --json)
// returns every plan so scripts that don't pass --all still work — matching
// the pre-refactor behaviour.
func chooseUpdates(app *App, plans []install.UpdatePlan, all bool) ([]install.UpdatePlan, error) {
	if all || app.Config.Yes {
		return plans, nil
	}
	if !tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes) {
		return plans, nil
	}

	items := make([]tui.Item, 0, len(plans))
	for _, p := range plans {
		items = append(items, updatePlanItem{p: p})
	}

	meta := func(items []tui.Item, _ int) string {
		return fmt.Sprintf("%d drifted", len(items))
	}

	res, err := tui.RunListDetail(tui.Config{
		Theme:      app.UI.Theme(),
		Version:    resolveVersion().Version,
		Section:    "Update",
		Meta:       meta,
		Items:      items,
		LeftTitle:  "DRIFTED",
		RightTitle: "DETAIL",
		Actions: []tui.ActionSpec{
			{Key: "u", Label: "apply all", Action: "all"},
			{Key: "enter", Label: "apply one", Action: "one"},
		},
		EmptyMsg: "all skills are up-to-date",
	})
	if err != nil {
		return nil, err
	}

	switch res.Action {
	case "all":
		return plans, nil
	case "one":
		it, ok := res.Item.(updatePlanItem)
		if !ok {
			return nil, nil
		}
		return []install.UpdatePlan{it.p}, nil
	}
	return nil, nil
}

// updatePlanItem adapts install.UpdatePlan to tui.Item.
type updatePlanItem struct{ p install.UpdatePlan }

func (u updatePlanItem) Key() string { return u.p.Skill }
func (u updatePlanItem) FilterValue() string {
	return strings.ToLower(u.p.Skill)
}
func (u updatePlanItem) NaturalWidth(th *ui.Theme) int {
	ver := u.p.FromVersion + " → " + u.p.ToVersion
	badge := tui.Badge(th, tui.BadgeRO, fmt.Sprintf("%d target%s", len(u.p.Targets), plural(len(u.p.Targets))))
	// 1 (arrow) + 1 (space) + skill + 2 (gap) + version + 2 (gap) + badge.
	return 1 + 1 + lipgloss.Width(u.p.Skill) + 2 + lipgloss.Width(ver) + 2 + lipgloss.Width(badge)
}
func (u updatePlanItem) Row(th *ui.Theme, width int, selected bool) string {
	arrow := th.DotWarn.Render("↑")
	name := rowName(th, u.p.Skill, selected, true)
	ver := th.Version.Render(u.p.FromVersion + " → " + u.p.ToVersion)
	badge := tui.Badge(th, tui.BadgeRO, fmt.Sprintf("%d target%s", len(u.p.Targets), plural(len(u.p.Targets))))
	return rowWithTrailingBadge(arrow+" "+name+"  "+ver, badge, width)
}
func (u updatePlanItem) Detail(th *ui.Theme, width int) string {
	var sb strings.Builder
	sb.WriteString(th.DetailTitle.Render(u.p.Skill) + "  " +
		th.DetailSub.Render("v"+u.p.FromVersion+" → v"+u.p.ToVersion) + "\n\n")
	sb.WriteString(kvRow(th, "from", th.KVValue.Render("v"+u.p.FromVersion)))
	sb.WriteString(kvRow(th, "to", th.KVValue.Render("v"+u.p.ToVersion)))
	sb.WriteString(kvRow(th, "targets", th.KVValue.Render(fmt.Sprintf("%d", len(u.p.Targets)))))

	if len(u.p.Targets) > 0 {
		sb.WriteString("\n" + th.SectionTitle.Render("TARGETS") + "\n")
		for i, t := range u.p.Targets {
			if i > 0 {
				sb.WriteString(tui.DashedRule(th, width) + "\n")
			}
			scope := th.PathLabel.Render(padRight(t.Scope, 7))
			plat := th.Platform.Render(t.Platform)
			path := th.PathValue.Render(t.Path)
			sb.WriteString("  " + scope + "  " + plat + "  " + path + "\n")
		}
	}
	return sb.String()
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
