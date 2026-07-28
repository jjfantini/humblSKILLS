package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/install"
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/internal/textutil"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

type updateFlags struct {
	all   bool
	check bool
	force bool
}

// regUpdatePlan records which registry (document + token + name) an update plan
// must be fetched from, so multi-registry updates hit the right source.
type regUpdatePlan struct {
	reg   *registry.Registry
	token string
	name  string
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
	useTUI := tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes)

	m, err := tui.RunWithLoadingIf(useTUI, app.UI.Theme(), "loading manifest…", func() (*manifest.Manifest, error) {
		m, err := manifest.Load(app.Config.ManifestPath)
		if err != nil {
			return nil, fmt.Errorf("load manifest: %w", err)
		}
		return m, nil
	})
	if err != nil {
		return err
	}
	if len(m.Installations) == 0 {
		app.UI.Info("no skills installed")
		return nil
	}

	loaded, err := tui.RunWithLoadingIf(useTUI, app.UI.Theme(), "loading registries…", func() ([]registrySkills, error) {
		return app.loadRegistries(), nil
	})
	if err != nil {
		return err
	}

	// Plan updates per ORIGIN registry so each installed skill is checked (and
	// later fetched) against the registry it came from, with that registry's
	// token. Legacy installs (no recorded origin) are attributed to whichever
	// registry currently lists the skill.
	ix := indexLoadedRegistries(loaded)
	regByName := make(map[string]registrySkills, len(loaded))
	for _, rs := range loaded {
		regByName[rs.Name] = rs
	}
	partitions := map[string][]manifest.Installation{}
	for _, inst := range m.Installations {
		origin := inst.RegistryName
		if origin == "" {
			origin = ix.registryOf(inst.Skill)
		}
		if origin == "" && len(loaded) > 0 {
			origin = loaded[0].Name
		}
		partitions[origin] = append(partitions[origin], inst)
	}

	bySkill := map[string]regUpdatePlan{}
	var plans []install.UpdatePlan
	for name, insts := range partitions {
		rs, ok := regByName[name]
		if !ok || rs.Reg == nil {
			continue // registry unavailable/unknown — skip its installs
		}
		fm := &manifest.Manifest{SchemaVersion: m.SchemaVersion, Installations: insts}
		for _, pl := range install.PlanUpdates(rs.Reg, fm, only) {
			bySkill[pl.Skill] = regUpdatePlan{reg: rs.Reg, token: rs.Token, name: name}
			plans = append(plans, pl)
		}
	}
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

	var aggregate install.Result

	run := func(sink install.EventSink) error {
		for _, plan := range selected {
			rp, ok := bySkill[plan.Skill]
			if !ok || rp.reg == nil {
				app.UI.Warn("skipping %s: source registry unresolved", plan.Skill)
				continue
			}
			engine := app.installEngineForToken(rp.token)
			stepPlan, err := install.Plan(rp.reg, plan.Skill)
			if err != nil {
				return err
			}

			type targetGroup struct {
				scope  string
				global bool
			}
			byGroup := map[targetGroup][]string{}
			for _, t := range plan.Targets {
				if _, ok := adapterKnown[t.Platform]; !ok {
					app.UI.Warn("skipping unknown platform %q in manifest for %s", t.Platform, plan.Skill)
					continue
				}
				group := targetGroup{
					scope:  t.Scope,
					global: t.InstallMode == install.InstallModeGlobal,
				}
				byGroup[group] = append(byGroup[group], t.Platform)
			}

			groups := make([]targetGroup, 0, len(byGroup))
			for g := range byGroup {
				groups = append(groups, g)
			}
			sort.Slice(groups, func(i, j int) bool {
				if groups[i].scope == groups[j].scope {
					return !groups[i].global && groups[j].global
				}
				return groups[i].scope < groups[j].scope
			})

			for _, group := range groups {
				plats := byGroup[group]
				sort.Strings(plats)
				res, err := engine.Execute(rp.reg, stepPlan, install.ExecuteOpts{
					Adapters:     adapters,
					Platforms:    plats,
					Scope:        group.scope,
					Force:        f.force,
					Global:       group.global,
					OnEvent:      sink,
					RegistryName: rp.name,
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
		p, err := profile.Load(app.Config.ProfilePath)
		if err != nil {
			return err
		}
		if err := tui.ExecuteWithProgress(app.UI.Theme(), "update", p.StatusAutoReturnDuration(), run); err != nil {
			return err
		}
		// Feedback already lives in the progress model's blocking done/summary
		// screen — see runInstall for why we don't also print to stdout here.
		return nil
	}
	if err := run(nil); err != nil {
		return err
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

	localMeta := func(items []tui.Item, _ int) string {
		return fmt.Sprintf("%d drifted", len(items))
	}
	meta := localMeta
	if app.Nav.Crumb != "" {
		status := app.Nav.Status
		theme := app.UI.Theme()
		meta = func(_ []tui.Item, _ int) string {
			return tui.RenderStatusMeta(theme, status)
		}
	}

	res, err := tui.RunListDetail(tui.Config{
		Theme:      app.UI.Theme(),
		Version:    resolveVersion().Version,
		Section:    app.headerSection("Update"),
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
	badge := tui.Badge(th, tui.BadgeRO, fmt.Sprintf("%d target%s", len(u.p.Targets), textutil.Plural(len(u.p.Targets))))
	// 1 (arrow) + 1 (space) + skill + 2 (gap) + version + 2 (gap) + badge.
	return 1 + 1 + lipgloss.Width(u.p.Skill) + 2 + lipgloss.Width(ver) + 2 + lipgloss.Width(badge)
}
func (u updatePlanItem) Row(th *ui.Theme, width int, selected bool) string {
	arrow := th.DotWarn.Render("↑")
	name := rowName(th, u.p.Skill, selected, true)
	ver := th.Version.Render(u.p.FromVersion + " → " + u.p.ToVersion)
	badge := tui.Badge(th, tui.BadgeRO, fmt.Sprintf("%d target%s", len(u.p.Targets), textutil.Plural(len(u.p.Targets))))
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
	app.UI.Info("%d skill%s can be updated:", len(plans), textutil.Plural(len(plans)))
	for _, p := range plans {
		app.UI.Detail("  %s  %s → %s  (%d target%s)",
			p.Skill, p.FromVersion, p.ToVersion, len(p.Targets), textutil.Plural(len(p.Targets)))
	}
	return nil
}
