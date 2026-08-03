package main

import (
	"errors"
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
	// platforms reconciles each installed skill against the profile's
	// default_platforms, adding any it doesn't target yet.
	platforms bool
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
			"and reinstalls cleanly from the registry (equivalent to uninstall + install), " +
			"and asks for confirmation first. --platforms additionally adds any platform " +
			"in your profile's default_platforms that a skill doesn't target yet; that is " +
			"a symlink plus a manifest entry, with no refetch when the skill is current.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(app, args, f)
		},
	}
	cmd.Flags().BoolVar(&f.all, "all", false, "update every drifted skill without prompting")
	cmd.Flags().BoolVar(&f.check, "check", false, "print what would change and exit")
	cmd.Flags().BoolVar(&f.force, "force", false, "bypass local preserve edits; reinstall cleanly from registry (asks to confirm)")
	cmd.Flags().BoolVar(&f.platforms, "platforms", false, "also add missing platforms from your profile's default_platforms (symlink only)")
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

	// --platforms turns the profile's default_platforms into the target set every
	// installed skill should cover, so adding a platform to the profile and
	// running update is enough to backfill it everywhere.
	var wantPlatforms []string
	if f.platforms {
		p, err := profile.Load(app.Config.ProfilePath)
		if err != nil {
			return err
		}
		wantPlatforms = p.DefaultPlatforms
		if len(wantPlatforms) == 0 {
			app.UI.Warn("--platforms needs default_platforms in your profile — run 'humblskills profile' to set them")
		}
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
		for _, pl := range install.PlanUpdatesFor(rs.Reg, fm, only, wantPlatforms) {
			bySkill[pl.Skill] = regUpdatePlan{reg: rs.Reg, token: rs.Token, name: name}
			plans = append(plans, pl)
		}
	}
	sort.Slice(plans, func(i, j int) bool { return plans[i].Skill < plans[j].Skill })

	if f.check {
		return printUpdateCheck(app, plans)
	}

	if len(plans) == 0 {
		switch {
		case len(only) > 0:
			app.UI.Info("selected skills are already up-to-date")
		case f.platforms:
			app.UI.Info("all skills are up-to-date and cover every platform in your profile")
		default:
			app.UI.Info("all skills are up-to-date")
		}
		return nil
	}

	selected, forceOne, err := chooseUpdates(app, plans, f.all)
	if err != nil {
		return err
	}
	force := f.force || forceOne
	if len(selected) == 0 {
		app.UI.Info("nothing selected")
		return nil
	}

	// --force here means "ignore my local preserve edits", which is the one
	// update mode that destroys content. Confirm against the real file list.
	if force {
		names := make([]string, 0, len(selected))
		for _, pl := range selected {
			names = append(names, pl.Skill)
		}
		if err := confirmForce(app, "update --force", names); err != nil {
			if errors.Is(err, errCancelled) {
				app.UI.Info("cancelled")
				return nil
			}
			return err
		}
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

			// Backfilled platforms join an existing group rather than forming
			// their own, so one Execute call covers "refresh this skill" and
			// "also link it here". Two calls would fetch twice and, worse, the
			// second would see the store mid-refresh.
			if len(plan.AddPlatforms) > 0 {
				add := make([]string, 0, len(plan.AddPlatforms))
				for _, p := range plan.AddPlatforms {
					if _, ok := adapterKnown[p]; !ok {
						app.UI.Warn("skipping unknown platform %q for %s", p, plan.Skill)
						continue
					}
					add = append(add, p)
				}
				if len(groups) == 0 {
					if len(add) > 0 {
						app.UI.Warn("skipping %s: no known platform to anchor %v to", plan.Skill, add)
					}
				} else {
					byGroup[groups[0]] = append(byGroup[groups[0]], add...)
				}
			}

			for _, group := range groups {
				plats := byGroup[group]
				sort.Strings(plats)
				res, err := engine.Execute(rp.reg, stepPlan, install.ExecuteOpts{
					Adapters:     adapters,
					Platforms:    plats,
					Scope:        group.scope,
					Force:        force,
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
func chooseUpdates(app *App, plans []install.UpdatePlan, all bool) ([]install.UpdatePlan, bool, error) {
	if all || app.Config.Yes {
		return plans, false, nil
	}
	if !tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes) {
		return plans, false, nil
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
			{Key: "f", Label: "force one", Action: "force"},
		},
		EmptyMsg: "all skills are up-to-date",
	})
	if err != nil {
		return nil, false, err
	}

	switch res.Action {
	case "all":
		return plans, false, nil
	case "one", "force":
		it, ok := res.Item.(updatePlanItem)
		if !ok {
			return nil, false, nil
		}
		// "force one" is the TUI's counterpart to `update --force <skill>`;
		// runUpdate still routes it through the confirmation gate.
		return []install.UpdatePlan{it.p}, res.Action == "force", nil
	}
	return nil, false, nil
}

// updatePlanItem adapts install.UpdatePlan to tui.Item.
type updatePlanItem struct{ p install.UpdatePlan }

func (u updatePlanItem) Key() string { return u.p.Skill }
func (u updatePlanItem) FilterValue() string {
	return strings.ToLower(u.p.Skill)
}
func (u updatePlanItem) NaturalWidth(th *ui.Theme) int {
	if u.p.LinkOnly {
		what := "link " + strings.Join(u.p.AddPlatforms, ", ")
		badge := tui.Badge(th, tui.BadgeRO, "no content change")
		return 1 + 1 + lipgloss.Width(u.p.Skill) + 2 + lipgloss.Width(what) + 2 + lipgloss.Width(badge)
	}
	ver := u.p.FromVersion + " → " + u.p.ToVersion
	badge := tui.Badge(th, tui.BadgeRO, fmt.Sprintf("%d target%s", len(u.p.Targets), textutil.Plural(len(u.p.Targets))))
	// 1 (arrow) + 1 (space) + skill + 2 (gap) + version + 2 (gap) + badge.
	return 1 + 1 + lipgloss.Width(u.p.Skill) + 2 + lipgloss.Width(ver) + 2 + lipgloss.Width(badge)
}
func (u updatePlanItem) Row(th *ui.Theme, width int, selected bool) string {
	name := rowName(th, u.p.Skill, selected, true)
	// A link-only plan changes no content, so it must not borrow the visual
	// language of an upgrade: no ↑, no version transition.
	if u.p.LinkOnly {
		arrow := th.DotOK.Render("+")
		what := th.Version.Render("link " + strings.Join(u.p.AddPlatforms, ", "))
		badge := tui.Badge(th, tui.BadgeRO, "no content change")
		return rowWithTrailingBadge(arrow+" "+name+"  "+what, badge, width)
	}
	arrow := th.DotWarn.Render("↑")
	ver := th.Version.Render(u.p.FromVersion + " → " + u.p.ToVersion)
	label := fmt.Sprintf("%d target%s", len(u.p.Targets), textutil.Plural(len(u.p.Targets)))
	if len(u.p.AddPlatforms) > 0 {
		label += " +" + strings.Join(u.p.AddPlatforms, " +")
	}
	badge := tui.Badge(th, tui.BadgeRO, label)
	return rowWithTrailingBadge(arrow+" "+name+"  "+ver, badge, width)
}
func (u updatePlanItem) Detail(th *ui.Theme, width int) string {
	var sb strings.Builder
	sub := "v" + u.p.FromVersion + " → v" + u.p.ToVersion
	if u.p.LinkOnly {
		sub = "v" + u.p.FromVersion + " (already current)"
	}
	sb.WriteString(th.DetailTitle.Render(u.p.Skill) + "  " + th.DetailSub.Render(sub) + "\n\n")
	if len(u.p.AddPlatforms) > 0 {
		sb.WriteString(kvRow(th, "add platforms", th.Platform.Render(strings.Join(u.p.AddPlatforms, "  "))))
		sb.WriteString(kvRow(th, "changes files", th.KVValue.Render(map[bool]string{
			true: "no — symlink + manifest only", false: "yes — content refresh",
		}[u.p.LinkOnly])))
	}
	if u.p.RenamedFrom != "" {
		sb.WriteString(kvRow(th, "renamed from", th.KVValue.Render(u.p.RenamedFrom)))
	}
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
	app.UI.Info("%d skill%s to act on:", len(plans), textutil.Plural(len(plans)))
	// Println, not Detail: --check exists to show this list, and Detail is
	// verbose-only — so the list was invisible unless you also passed -v.
	th := app.UI.Theme()
	line := func(format string, args ...any) {
		app.UI.Println(th.Detail.Render(fmt.Sprintf(format, args...)))
	}
	for _, p := range plans {
		name := p.Skill
		if p.RenamedFrom != "" {
			// Say the old name out loud — the user knows the skill by that,
			// and a silent swap looks like an unrelated install.
			name = p.RenamedFrom + " → " + p.Skill
		}
		// Never print a version transition for a plan that changes no content:
		// "1.0.0 → 1.0.0" reads as a bug, and the real work is the link.
		if p.LinkOnly {
			line("  %s  link only  (+%s, no content change)",
				name, strings.Join(p.AddPlatforms, " +"))
			continue
		}
		extra := ""
		if len(p.AddPlatforms) > 0 {
			extra = ", +" + strings.Join(p.AddPlatforms, " +")
		}
		line("  %s  %s → %s  (%d target%s%s)",
			name, p.FromVersion, p.ToVersion, len(p.Targets), textutil.Plural(len(p.Targets)), extra)
	}
	return nil
}
