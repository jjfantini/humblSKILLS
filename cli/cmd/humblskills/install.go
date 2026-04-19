package main

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/adapters"
	"github.com/jjfantini/humblSKILLS/cli/internal/install"
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
)

type installFlags struct {
	platforms    []string
	platformsSet bool
	scope        string
	force        bool
}

func newInstallCmd(app *App) *cobra.Command {
	var f installFlags
	cmd := &cobra.Command{
		Use:   "install [skill]",
		Short: "Install a skill (and its deps) onto every detected platform",
		Long: "install <skill> installs a named skill. With no arg, it opens " +
			"an interactive, filterable picker listing every skill in the registry.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skill := ""
			if len(args) == 1 {
				skill = args[0]
			}
			f.platformsSet = cmd.Flags().Changed("platform")
			return runInstall(app, skill, f, false)
		},
	}
	cmd.Flags().StringSliceVar(&f.platforms, "platform", nil, "restrict install to these adapters (default: all detected)")
	cmd.Flags().StringVar(&f.scope, "scope", "", "install scope (user|project; default: adapter's default)")
	cmd.Flags().BoolVar(&f.force, "force", false, "reinstall even if already up-to-date")
	return cmd
}

func runInstall(app *App, skill string, f installFlags, fromDashboard bool) error {
	adapterList, err := app.Adapters()
	if err != nil {
		return fmt.Errorf("load adapters: %w", err)
	}

	reg, _, err := registry.NewFetcher(app.Config.RegistryURL, app.Config.CacheDir).Load()
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	if skill == "" {
		skill, err = pickSkill(app, reg, fromDashboard)
		if err != nil {
			if fromDashboard && err.Error() == "no skill selected" {
				return nil
			}
			return err
		}
	}

	platforms := f.platforms
	scope := f.scope
	useTUIForModal := !f.platformsSet && tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes)
	if useTUIForModal {
		plats, scp, ok, err := promptInstallTargets(app, adapterList, skill)
		if err != nil {
			return err
		}
		if !ok {
			if fromDashboard {
				return nil
			}
			return fmt.Errorf("install cancelled")
		}
		platforms = plats
		if scp != "" {
			scope = scp
		}
	}

	selected, err := selectPlatforms(adapterList, platforms)
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

	engine := install.NewEngine(app.Config.CacheDir, app.Config.ManifestPath)
	useTUI := tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes)

	if !useTUI {
		app.UI.Detail("plan:")
		for _, s := range plan {
			tag := "root"
			if s.IsDep {
				tag = "dep"
			}
			app.UI.Detail("  %s %s@%s", tag, s.Skill.Name, s.Skill.Version)
		}
	}

	var res install.Result
	run := func(sink install.EventSink) error {
		r, err := engine.Execute(reg, plan, install.ExecuteOpts{
			Adapters:  adapterList,
			Platforms: selected,
			Scope:     scope,
			Force:     f.force,
			OnEvent:   sink,
		})
		res = r
		return err
	}

	if useTUI {
		if err := tui.ExecuteWithProgress(app.UI.Theme(), "install", run); err != nil {
			return err
		}
	} else {
		if err := run(nil); err != nil {
			return err
		}
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
func selectPlatforms(adapterList []adapters.Adapter, requested []string) ([]string, error) {
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

	detected := adapters.Detect(adapterList)
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

// promptInstallTargets opens a huh modal asking the user which platforms to
// install `skill` into (defaults come from profile), returning the confirmed
// platforms + scope. If the user picks "edit profile" inside the modal, the
// profile editor opens and the modal re-prompts with the updated defaults.
// Returns ok=false if the user cancelled.
func promptInstallTargets(app *App, adapterList []adapters.Adapter, skill string) ([]string, string, bool, error) {
	detected := map[string]bool{}
	for _, r := range adapters.Detect(adapterList) {
		detected[r.Name] = r.Detected
	}
	for {
		p, err := profile.Load(app.Config.ProfilePath)
		if err != nil {
			return nil, "", false, err
		}
		res, err := tui.RunInstallPlatformModal(app.UI.Theme(), skill, adapterList, detected, p)
		if err != nil {
			return nil, "", false, err
		}
		if res.EditProfile {
			if err := runProfileEditor(app); err != nil {
				return nil, "", false, err
			}
			continue
		}
		if !res.Confirmed {
			return nil, "", false, nil
		}
		return res.Platforms, res.Scope, true, nil
	}
}

// pickSkill opens the shared two-pane skill browser over the registry and
// returns the chosen skill's name. Matches the search surface 1:1 so the user
// can't tell them apart — a zero-arg install IS a searchable picker.
func pickSkill(app *App, reg *registry.Registry, fromDashboard bool) (string, error) {
	if len(reg.Skills) == 0 {
		return "", fmt.Errorf("registry is empty")
	}
	if !tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes) {
		return "", fmt.Errorf("skill name required — usage: humblskills install <skill>")
	}

	skills := append([]registry.Skill(nil), reg.Skills...)
	sort.Slice(skills, func(i, j int) bool { return skills[i].Name < skills[j].Name })

	m, err := manifest.Load(app.Config.ManifestPath)
	if err != nil {
		m = &manifest.Manifest{}
	}
	items := buildSkillItems(skills, m)

	skill, action, err := runSkillBrowser(app, "Install", items, modeSearch, "registry is empty", fromDashboard)
	if err != nil {
		return "", err
	}
	if action != "install" || skill == "" {
		return "", fmt.Errorf("no skill selected")
	}
	return skill, nil
}
