package main

import (
	"errors"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/v2/internal/adapters"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/install"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/textutil"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/tui"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/ui"
)

type installFlags struct {
	platforms    []string
	platformsSet bool
	scope        string
	scopeSet     bool
	force        bool
	global       bool
	from         string // registry name to disambiguate a skill present in several
}

func newInstallCmd(app *App) *cobra.Command {
	var f installFlags
	cmd := &cobra.Command{
		Use:   "install [skill...]",
		Short: "Install one or more skills (and their deps) onto every detected platform",
		Long: "install <skill>... installs the named skills, sharing one dependency " +
			"resolution and one platform prompt across all of them. With no arg, it " +
			"opens an interactive, filterable picker listing every skill in the " +
			"registry, where space picks several to install at once.",
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			f.platformsSet = cmd.Flags().Changed("platform")
			f.scopeSet = cmd.Flags().Changed("scope")
			return runInstall(app, args, f, false)
		},
	}
	cmd.Flags().StringSliceVar(&f.platforms, "platform", nil, "restrict install to these adapters (default: profile default, else all detected)")
	cmd.Flags().StringVar(&f.scope, "scope", "", "install scope: global|user|project|adapter-default (default: your profile's default scope, itself global unless changed)")
	cmd.Flags().BoolVar(&f.force, "force", false, "reinstall even if already up-to-date")
	cmd.Flags().BoolVar(&f.global, "global", false, "alias for --scope global: install once into ~/.humblskills and symlink to the selected platforms")
	cmd.Flags().StringVar(&f.from, "from", "", "registry to install from when a skill name exists in more than one")
	// Completion stays available for every positional now that the command takes
	// several, minus the names already typed on the line.
	cmd.ValidArgsFunction = func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		taken := make(map[string]bool, len(args))
		for _, a := range args {
			taken[a] = true
		}
		all := app.completeSkillNames(toComplete)
		out := make([]string, 0, len(all))
		for _, name := range all {
			if !taken[name] {
				out = append(out, name)
			}
		}
		return out, cobra.ShellCompDirectiveNoFileComp
	}
	_ = cmd.RegisterFlagCompletionFunc("from", func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return app.completeRegistryNames(toComplete), cobra.ShellCompDirectiveNoFileComp
	})
	return cmd
}

// runInstall installs every named skill. Passing no names opens the picker,
// where the user can pick several.
//
// The batch is not N independent installs: the whole set shares one dependency
// resolution, one platform/scope prompt, one --force confirmation and one
// progress screen. Skills are grouped by owning registry because a registry
// carries its own document and token, so each group is a separate engine run —
// but all of them happen inside the one progress screen.
func runInstall(app *App, skills []string, f installFlags, fromDashboard bool) error {
	useTUI := tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes)

	type preload struct {
		adapters []adapters.Adapter
		loaded   []registrySkills
		merged   *registry.Registry
	}
	pre, err := tui.RunWithLoadingIf(useTUI, app.UI.Theme(), "loading adapters + registry…", func() (preload, error) {
		adapterList, err := app.Adapters()
		if err != nil {
			return preload{}, fmt.Errorf("load adapters: %w", err)
		}
		loaded := app.loadRegistries()
		merged := mergedRegistry(loaded)
		if len(merged.Skills) == 0 {
			for _, rs := range loaded {
				if rs.Err != nil {
					return preload{}, fmt.Errorf("load registry %q: %w", rs.Name, rs.Err)
				}
			}
		}
		return preload{adapters: adapterList, loaded: loaded, merged: merged}, nil
	})
	if err != nil {
		return err
	}
	adapterList, loaded, reg := pre.adapters, pre.loaded, pre.merged

	skills = dedupeNames(skills)
	if len(skills) == 0 {
		skills, err = pickSkills(app, reg, fromDashboard)
		if err != nil {
			if fromDashboard && err.Error() == "no skill selected" {
				return nil
			}
			return err
		}
	}
	if len(skills) == 0 {
		return fmt.Errorf("no skill selected")
	}

	// Resolve which registry holds each chosen skill so we plan and fetch against
	// that registry's document (Source) and token, then group the batch by
	// registry. Resolution happens for every name up front, before anything is
	// written: an unknown or ambiguous name in a batch of five should fail with
	// nothing installed rather than leaving a partial result behind.
	groups, err := groupSkillsByRegistry(app, loaded, skills, f.from, useTUI)
	if err != nil {
		return err
	}

	p, err := profile.Load(app.Config.ProfilePath)
	if err != nil {
		return err
	}

	// Any explicit flag (--platform, --scope, --global) opts out of the
	// interactive modal — scripted/explicit invocations should never be
	// silently overridden by a prompt the caller can't see.
	explicitFlags := f.platformsSet || f.scopeSet || f.global
	platforms := f.platforms
	scope := f.scope
	global := f.global
	force := f.force
	useTUIForModal := !explicitFlags && useTUI
	if useTUIForModal {
		// One modal for the whole batch — the platform/scope answer is a property
		// of this install run, not of each skill in it.
		res, ok, err := promptInstallTargets(app, adapterList, batchLabel(skills))
		if err != nil {
			return err
		}
		if !ok {
			if fromDashboard {
				return nil
			}
			return fmt.Errorf("install cancelled")
		}
		platforms = res.Platforms
		scope = res.Scope
		global = res.Global
		// The modal offers the same force toggle the flag does, so the TUI can
		// reach every install the CLI can. --force on the command line still
		// wins if it was passed.
		force = force || res.Force
	} else {
		scope, global, err = resolveInstallScope(f, p)
		if err != nil {
			return err
		}
	}

	selected, err := selectPlatforms(adapterList, platforms, global, p.DefaultPlatforms)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		return fmt.Errorf("no platforms selected — run 'humblskills doctor' to see what's detected")
	}

	// One plan per registry group. PlanAll shares dependency resolution across
	// every root in the group, so a dep two picked skills have in common is
	// fetched and placed once.
	plans := make([][]install.Step, len(groups))
	for i, g := range groups {
		plan, err := install.PlanAll(g.target.Reg, g.skills)
		if err != nil {
			return err
		}
		plans[i] = plan
	}

	// --force is what bypasses the preserve list, so it's the one install flag
	// that can destroy user data. Name the files first and make the user agree —
	// once for the batch, listing every skill it will overwrite.
	if force {
		var names []string
		for _, plan := range plans {
			for _, s := range plan {
				names = append(names, s.Skill.Name)
			}
		}
		if err := confirmForce(app, "install --force", names); err != nil {
			if errors.Is(err, errCancelled) {
				app.UI.Info("cancelled")
				return nil
			}
			return err
		}
	}

	if !useTUI {
		app.UI.Detail("plan:")
		for i, g := range groups {
			if len(groups) > 1 {
				app.UI.Detail("  from %s:", g.target.Name)
			}
			for _, s := range plans[i] {
				tag := "root"
				if s.IsDep {
					tag = "dep"
				}
				app.UI.Detail("  %s %s@%s", tag, s.Skill.Name, s.Skill.Version)
			}
		}
	}

	var res install.Result
	run := func(sink install.EventSink) error {
		// Sequential, and it aborts on the first failing group: a later group
		// depends on nothing the earlier ones did, but continuing past a real
		// error (bad token, network gone) would just pile up the same failure.
		for i, g := range groups {
			r, err := app.installEngineForToken(g.target.Token).Execute(g.target.Reg, plans[i], install.ExecuteOpts{
				Adapters:     adapterList,
				Platforms:    selected,
				Scope:        scope,
				Force:        force,
				Global:       global,
				OnEvent:      sink,
				RegistryName: g.target.Name,
			})
			res.Results = append(res.Results, r.Results...)
			res.Warnings = append(res.Warnings, r.Warnings...)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if useTUI {
		if err := tui.ExecuteWithProgress(app.UI.Theme(), "install", p.StatusAutoReturnDuration(), run); err != nil {
			return err
		}
		// Feedback already lives in the progress model's blocking done/summary
		// screen — printing to the normal buffer here would just get hidden
		// the instant the dashboard loop re-enters the alt-screen.
		return nil
	}
	if err := run(nil); err != nil {
		return err
	}

	if app.Config.JSON {
		return app.UI.JSON(res)
	}

	printInstall(app, res)
	return nil
}

// resolveInstallScope figures out the effective (scope, global) pair when
// the install isn't going through the interactive platform modal (the modal
// resolves its own scope/global from the profile internally). Precedence:
// explicit --global / --scope flags, then the profile's resolved default,
// then the historical "adapter default" catch-all.
func resolveInstallScope(f installFlags, p *profile.Profile) (scope string, global bool, err error) {
	if f.global {
		if f.scopeSet && f.scope == profile.ScopeProject {
			return "", false, fmt.Errorf("--global installs to user-scope platform targets; use --scope project without --global")
		}
		return "", true, nil
	}
	if f.scopeSet {
		switch f.scope {
		case profile.ScopeGlobal:
			return "", true, nil
		case profile.ScopeAdapterDefault, "":
			return "", false, nil
		case profile.ScopeUser, profile.ScopeProject:
			return f.scope, false, nil
		default:
			return "", false, fmt.Errorf("unknown scope %q — valid: global, user, project, adapter-default", f.scope)
		}
	}
	switch p.ResolvedScope() {
	case profile.ScopeGlobal:
		return "", true, nil
	case profile.ScopeUser, profile.ScopeProject:
		return p.DefaultScope, false, nil
	default: // adapter-default
		return "", false, nil
	}
}

// selectPlatforms returns the adapter names to install onto. If the user
// passed --platform, it's the intersection of that list with the declared
// adapters — an explicit request always wins, global or not. Otherwise it
// falls back to the profile's saved platforms, or (failing that) the same
// default cascade the TUI uses: global scope symlinks every detected
// platform; non-global scopes prefer claude-code over cursor when both are
// detected, since Cursor can read ~/.claude/skills natively (issue #84).
func selectPlatforms(adapterList []adapters.Adapter, requested []string, global bool, profileDefaults []string) ([]string, error) {
	known := adapters.NameSet(adapterList)
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

	detected := map[string]bool{}
	for _, d := range adapters.Detect(adapterList) {
		detected[d.Name] = d.Detected
	}
	out := adapters.PreferredDefaults(adapterList, detected, profileDefaults, global)
	sort.Strings(out)
	return out, nil
}

func printInstall(app *App, r install.Result) {
	for _, w := range r.Warnings {
		where := ""
		if w.Skill != "" {
			where = w.Skill
			if w.Platform != "" {
				where += " [" + w.Platform + "/" + w.Scope + "]"
			}
			where += ": "
		}
		app.UI.Warn("%s%s", where, w.Msg)
	}
	if len(r.Results) == 0 {
		app.UI.Warn("nothing to do - skill(s) declared no matching platforms")
		return
	}
	var installed, replaced, skipped, forced, linked []install.TargetResult
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
		case install.OutcomeLinked:
			linked = append(linked, t)
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
	// "linked" is deliberately worded so nobody reads it as a content change:
	// the skill was already current, this run only added a platform.
	for _, t := range linked {
		app.UI.Success("linked %s → %s [%s/%s] (no content change)", t.Skill, t.Path, t.Platform, t.Scope)
	}
	for _, t := range skipped {
		app.UI.Detail("already up-to-date: %s [%s/%s]", t.Skill, t.Platform, t.Scope)
	}
	if len(installed)+len(replaced)+len(forced)+len(linked) == 0 {
		app.UI.Info("%d target%s already up-to-date (use --force to reinstall)", len(skipped), textutil.Plural(len(skipped)))
	}
	for _, t := range r.Results {
		if t.Platform == "claude-desktop" && t.Outcome != install.OutcomeSkipped {
			app.UI.Info("%s", desktopUploadHint)
			break
		}
	}
}

// promptInstallTargets opens a huh modal asking the user which platforms to
// install `skill` into (defaults come from profile), returning the confirmed
// platforms, scope and force toggle. If the user picks "edit profile" inside the modal, the
// profile editor opens and the modal re-prompts with the updated defaults.
// Returns ok=false if the user cancelled.
func promptInstallTargets(app *App, adapterList []adapters.Adapter, skill string) (tui.InstallModalResult, bool, error) {
	detected := map[string]bool{}
	for _, r := range adapters.Detect(adapterList) {
		detected[r.Name] = r.Detected
	}
	for {
		p, err := profile.Load(app.Config.ProfilePath)
		if err != nil {
			return tui.InstallModalResult{}, false, err
		}
		res, err := tui.RunInstallPlatformModal(app.UI.Theme(), skill, adapterList, detected, p)
		if err != nil {
			return tui.InstallModalResult{}, false, err
		}
		if res.EditProfile {
			if err := runProfileEditor(app); err != nil {
				return tui.InstallModalResult{}, false, err
			}
			continue
		}
		if !res.Confirmed {
			return tui.InstallModalResult{}, false, nil
		}
		return res, true, nil
	}
}

// pickSkills opens the shared two-pane skill browser over the registry and
// returns the chosen skills' names — several of them when the user ticked rows
// with space. Matches the search surface 1:1 so the user can't tell them apart —
// a zero-arg install IS a searchable picker.
func pickSkills(app *App, reg *registry.Registry, fromDashboard bool) ([]string, error) {
	if len(reg.Skills) == 0 {
		return nil, fmt.Errorf("registry is empty")
	}
	if !tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes) {
		return nil, fmt.Errorf("skill name required — usage: humblskills install <skill>...")
	}

	skills := append([]registry.Skill(nil), reg.Skills...)
	sort.Slice(skills, func(i, j int) bool { return skills[i].Name < skills[j].Name })

	// ShouldUseTUI already returned true above, so this always runs behind
	// its own alt-screen loading spinner rather than on the exposed
	// terminal buffer right before the picker takes over.
	m, _ := tui.RunWithLoading(app.UI.Theme(), "loading manifest…", func() (*manifest.Manifest, error) {
		m, err := manifest.Load(app.Config.ManifestPath)
		if err != nil {
			return &manifest.Manifest{}, nil
		}
		return m, nil
	})
	items := buildSkillItems(skills, m, app.resolvedGroupByCategory())

	picked, action, err := runSkillBrowser(app, "Install", items, modeSearch, "registry is empty", fromDashboard)
	if err != nil {
		return nil, err
	}
	if action != "install" || len(picked) == 0 {
		return nil, fmt.Errorf("no skill selected")
	}
	return picked, nil
}

// dedupeNames drops blanks and repeats while preserving the caller's order, so
// `install a b a` plans `a` once and the plan reads in the order asked for.
func dedupeNames(names []string) []string {
	out := make([]string, 0, len(names))
	seen := make(map[string]bool, len(names))
	for _, n := range names {
		if n == "" || seen[n] {
			continue
		}
		seen[n] = true
		out = append(out, n)
	}
	return out
}

// batchLabel names the install for the platform modal's title: the skill itself
// when there's one, otherwise a count, since listing eight names would blow out
// the modal header.
func batchLabel(skills []string) string {
	if len(skills) == 1 {
		return skills[0]
	}
	return fmt.Sprintf("%d skills", len(skills))
}

// installGroup is one registry's share of a batch install: the registry to plan
// and fetch against, plus the requested root skills that live in it.
type installGroup struct {
	target registrySkills
	skills []string
}

// groupSkillsByRegistry resolves every name to its owning registry and buckets
// the batch accordingly, preserving first-appearance order for both the groups
// and the skills inside them so a run is reproducible.
//
// Ambiguity (a name in several registries) is resolved exactly as it was for a
// single skill: --from wins, else prompt in a TUI, else an error telling the
// user to pass --from. --from applies to the whole batch, and is only consulted
// for names that are actually ambiguous — so it can't drag an unambiguous skill
// to the wrong registry.
func groupSkillsByRegistry(app *App, loaded []registrySkills, skills []string, from string, useTUI bool) ([]installGroup, error) {
	var groups []installGroup
	index := map[string]int{} // registry name → position in groups
	for _, skill := range skills {
		matches := allRegistriesForSkill(loaded, skill)
		var target registrySkills
		switch {
		case len(matches) == 0:
			return nil, fmt.Errorf("skill %q not found in any configured registry", skill)
		case len(matches) == 1:
			target = matches[0]
		default:
			picked := ""
			if from != "" {
				picked = from
			} else if useTUI {
				opts := make([]ui.SelectOption, 0, len(matches))
				for _, m := range matches {
					opts = append(opts, ui.SelectOption{Label: m.Name + "  —  " + m.URL, Value: m.Name})
				}
				sel, err := app.Prompt.Select(fmt.Sprintf("%q is in %d registries — install from?", skill, len(matches)), "", opts)
				if err != nil {
					return nil, err
				}
				picked = sel
			} else {
				return nil, fmt.Errorf("skill %q exists in multiple registries (%s) — choose one with --from <registry>", skill, registryNames(matches))
			}
			found := false
			for _, m := range matches {
				if m.Name == picked {
					target = m
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("skill %q is not in registry %q (available in: %s)", skill, picked, registryNames(matches))
			}
		}

		if i, ok := index[target.Name]; ok {
			groups[i].skills = append(groups[i].skills, skill)
			continue
		}
		index[target.Name] = len(groups)
		groups = append(groups, installGroup{target: target, skills: []string{skill}})
	}
	return groups, nil
}
