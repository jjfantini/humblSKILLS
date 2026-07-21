package main

import (
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/install"
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/internal/skillset"
	"github.com/jjfantini/humblSKILLS/cli/internal/textutil"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
)

// --- init -------------------------------------------------------------------

func newInitCmd(app *App) *cobra.Command {
	var fromInstalled bool
	var force bool
	cmd := &cobra.Command{
		Use:   "init [file]",
		Short: "Scaffold a new skillset file to share with your team",
		Long: "init writes a starter skillset file (default: ./humblskills.json) that " +
			"you commit to a repo so teammates can run 'humblskills sync' to land the " +
			"same set. By default it writes an empty skillset for you to fill in; pass " +
			"--from-installed to seed it from the skills you already have installed. " +
			"It refuses to overwrite an existing file unless you pass --force.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			path := skillset.DefaultFilename
			if len(args) == 1 {
				path = args[0]
			}
			return runInit(app, path, fromInstalled, force)
		},
	}
	cmd.Flags().BoolVar(&fromInstalled, "from-installed", false, "seed the skillset from your currently installed skills")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite an existing skillset file")
	return cmd
}

func runInit(app *App, path string, fromInstalled, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s already exists (use --force to overwrite, or edit it directly)", path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat %s: %w", path, err)
		}
	}

	set := skillset.New()
	if fromInstalled {
		m, err := manifest.Load(app.Config.ManifestPath)
		if err != nil {
			return err
		}
		for _, name := range uniqueSkillsFromManifest(m) {
			version := ""
			if insts := m.FindAll(name); len(insts) > 0 {
				version = insts[0].Version
			}
			set.Add(name, version)
		}
		set.Sort()
	}

	if app.Config.JSON {
		return app.UI.JSON(set)
	}
	if err := skillset.Save(path, set); err != nil {
		return err
	}
	app.UI.Success("created %s with %d skill%s", path, len(set.Skills), textutil.Plural(len(set.Skills)))
	if len(set.Skills) == 0 {
		app.UI.Detail("find skills with 'humblskills search <q>', add their names under \"skills\", then run 'humblskills sync'")
	} else {
		app.UI.Detail("commit it and run 'humblskills sync' on another machine to install the same set")
	}
	return nil
}

// --- export -----------------------------------------------------------------

func newExportCmd(app *App) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Write a shareable skillset file from your installed skills",
		Long: "export snapshots the unique skills in your install manifest into a " +
			"skillset file (default: ./humblskills.json). Commit it to a repo and " +
			"teammates run 'humblskills sync' to install the same set.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runExport(app, output)
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", skillset.DefaultFilename,
		"skillset file to write")
	return cmd
}

func runExport(app *App, output string) error {
	m, err := manifest.Load(app.Config.ManifestPath)
	if err != nil {
		return err
	}
	if len(m.Installations) == 0 {
		return fmt.Errorf("no skills installed — nothing to export")
	}

	set := skillset.New()
	for _, name := range uniqueSkillsFromManifest(m) {
		version := ""
		if insts := m.FindAll(name); len(insts) > 0 {
			version = insts[0].Version
		}
		set.Add(name, version)
	}
	set.Sort()

	if app.Config.JSON {
		return app.UI.JSON(set)
	}
	if err := skillset.Save(output, set); err != nil {
		return err
	}
	app.UI.Success("exported %d skill%s to %s", len(set.Skills), textutil.Plural(len(set.Skills)), output)
	app.UI.Detail("commit it and run 'humblskills sync' on another machine to install the same set")
	return nil
}

// --- sync --------------------------------------------------------------------

func newSyncCmd(app *App) *cobra.Command {
	var f installFlags
	var prune bool
	cmd := &cobra.Command{
		Use:   "sync [file-or-url]",
		Short: "Install every skill listed in a skillset file or URL",
		Long: "sync reads a skillset (default: ./humblskills.json) and installs " +
			"every skill it lists that isn't already present, pulling the current " +
			"registry version. The source can be a local path, a file:// URL, or an " +
			"http(s):// URL - so a team can host one canonical skillset and everyone " +
			"runs 'humblskills sync https://example.com/humblskills.json'. Already " +
			"up-to-date skills are skipped (use --force to reinstall). With --prune it " +
			"also uninstalls any skill you have that the skillset doesn't list, making " +
			"your local set match the file exactly. Platforms/scope follow the same " +
			"rules as install: explicit --platform/--scope/--global flags win, " +
			"otherwise your profile defaults apply.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := skillset.DefaultFilename
			if len(args) == 1 {
				path = args[0]
			}
			f.platformsSet = cmd.Flags().Changed("platform")
			f.scopeSet = cmd.Flags().Changed("scope")
			return runSync(app, path, f, prune)
		},
	}
	cmd.Flags().StringSliceVar(&f.platforms, "platform", nil, "restrict install to these adapters (default: profile default, else all detected)")
	cmd.Flags().StringVar(&f.scope, "scope", "", "install scope: global|user|project|adapter-default (default: your profile's default scope)")
	cmd.Flags().BoolVar(&f.force, "force", false, "reinstall skills even if already up-to-date")
	cmd.Flags().BoolVar(&f.global, "global", false, "alias for --scope global")
	cmd.Flags().BoolVar(&prune, "prune", false, "uninstall skills not listed in the skillset (make local set match the file)")
	return cmd
}

func runSync(app *App, path string, f installFlags, prune bool) error {
	set, err := skillset.LoadFrom(path)
	if err != nil {
		return err
	}
	if len(set.Skills) == 0 && !prune {
		app.UI.Info("skillset %s lists no skills — nothing to sync", path)
		return nil
	}

	adapterList, err := app.Adapters()
	if err != nil {
		return fmt.Errorf("load adapters: %w", err)
	}
	reg, _, err := app.registryFetcher().Load()
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}
	p, err := profile.Load(app.Config.ProfilePath)
	if err != nil {
		return err
	}

	scope, global, err := resolveInstallScope(f, p)
	if err != nil {
		return err
	}
	selected, err := selectPlatforms(adapterList, f.platforms, global, p.DefaultPlatforms)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		return fmt.Errorf("no platforms selected — run 'humblskills doctor' to see what's detected")
	}

	engine := app.installEngine()
	var aggregate install.Result
	var missing []string

	names := set.Names()
	sort.Strings(names)
	for _, name := range names {
		plan, planErr := install.Plan(reg, name)
		if planErr != nil {
			// A skill in the skillset that the registry doesn't know about is a
			// warning, not a hard failure — sync the rest.
			missing = append(missing, name)
			app.UI.Warn("skipping %q: %v", name, planErr)
			continue
		}
		res, execErr := engine.Execute(reg, plan, install.ExecuteOpts{
			Adapters:  adapterList,
			Platforms: selected,
			Scope:     scope,
			Force:     f.force,
			Global:    global,
		})
		if execErr != nil {
			return fmt.Errorf("%s: %w", name, execErr)
		}
		aggregate.Results = append(aggregate.Results, res.Results...)
		aggregate.Warnings = append(aggregate.Warnings, res.Warnings...)
	}

	var pruned []install.TargetResult
	if prune {
		pruned, err = pruneToSkillset(app, set)
		if err != nil {
			return err
		}
	}

	if app.Config.JSON {
		return app.UI.JSON(struct {
			install.Result
			Pruned []install.TargetResult `json:"pruned,omitempty"`
		}{aggregate, pruned})
	}
	printInstall(app, aggregate)
	for _, t := range pruned {
		app.UI.Success("pruned %s [%s/%s]", t.Skill, t.Platform, t.Scope)
	}
	if len(missing) > 0 {
		app.UI.Warn("%d skill%s in %s not found in the registry: %v",
			len(missing), textutil.Plural(len(missing)), path, missing)
	}
	return nil
}

// pruneToSkillset uninstalls every skill that's installed locally but absent
// from the skillset, so the local set ends up matching the file exactly. It
// reloads the manifest first because the preceding install pass may have added
// entries. Destructive, so it confirms (unless --yes / --json).
func pruneToSkillset(app *App, set *skillset.Set) ([]install.TargetResult, error) {
	m, err := manifest.Load(app.Config.ManifestPath)
	if err != nil {
		return nil, fmt.Errorf("load manifest: %w", err)
	}
	keep := map[string]bool{}
	for _, name := range set.Names() {
		keep[name] = true
	}
	var extra []string
	for _, name := range uniqueSkillsFromManifest(m) {
		if !keep[name] {
			extra = append(extra, name)
		}
	}
	if len(extra) == 0 {
		return nil, nil
	}
	sort.Strings(extra)

	if !app.Config.Yes && !app.Config.JSON {
		theme := app.UI.Theme()
		lines := make([]string, 0, len(extra))
		for _, name := range extra {
			lines = append(lines, theme.Name.Render(name))
		}
		ok, err := tui.ConfirmWithSummary(
			theme,
			"Prune skills not in the skillset",
			fmt.Sprintf("Uninstall %d skill%s not listed in the skillset?", len(extra), textutil.Plural(len(extra))),
			lines,
			false,
			app.Prompt.Interactive,
		)
		if err != nil {
			return nil, err
		}
		if !ok {
			app.UI.Info("prune cancelled — installed skills left untouched")
			return nil, nil
		}
	}

	engine := app.installEngine()
	var pruned []install.TargetResult
	for _, name := range extra {
		res, err := engine.Uninstall(name)
		if err != nil {
			return nil, fmt.Errorf("prune %s: %w", name, err)
		}
		pruned = append(pruned, res...)
	}
	return pruned, nil
}
