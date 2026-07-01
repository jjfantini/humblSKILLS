package main

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/install"
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/internal/skillset"
)

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
	app.UI.Success("exported %d skill%s to %s", len(set.Skills), plural(len(set.Skills)), output)
	app.UI.Detail("commit it and run 'humblskills sync' on another machine to install the same set")
	return nil
}

// --- sync --------------------------------------------------------------------

func newSyncCmd(app *App) *cobra.Command {
	var f installFlags
	cmd := &cobra.Command{
		Use:   "sync [file-or-url]",
		Short: "Install every skill listed in a skillset file or URL",
		Long: "sync reads a skillset (default: ./humblskills.json) and installs " +
			"every skill it lists that isn't already present, pulling the current " +
			"registry version. The source can be a local path, a file:// URL, or an " +
			"http(s):// URL - so a team can host one canonical skillset and everyone " +
			"runs 'humblskills sync https://example.com/humblskills.json'. Already " +
			"up-to-date skills are skipped (use --force to reinstall). Platforms/scope " +
			"follow the same rules as install: explicit --platform/--scope/--global " +
			"flags win, otherwise your profile defaults apply.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := skillset.DefaultFilename
			if len(args) == 1 {
				path = args[0]
			}
			f.platformsSet = cmd.Flags().Changed("platform")
			f.scopeSet = cmd.Flags().Changed("scope")
			return runSync(app, path, f)
		},
	}
	cmd.Flags().StringSliceVar(&f.platforms, "platform", nil, "restrict install to these adapters (default: profile default, else all detected)")
	cmd.Flags().StringVar(&f.scope, "scope", "", "install scope: global|user|project|adapter-default (default: your profile's default scope)")
	cmd.Flags().BoolVar(&f.force, "force", false, "reinstall skills even if already up-to-date")
	cmd.Flags().BoolVar(&f.global, "global", false, "alias for --scope global")
	return cmd
}

func runSync(app *App, path string, f installFlags) error {
	set, err := skillset.LoadFrom(path)
	if err != nil {
		return err
	}
	if len(set.Skills) == 0 {
		app.UI.Info("skillset %s lists no skills — nothing to sync", path)
		return nil
	}

	adapterList, err := app.Adapters()
	if err != nil {
		return fmt.Errorf("load adapters: %w", err)
	}
	reg, _, err := registry.NewFetcher(app.Config.RegistryURL, app.Config.CacheDir).Load()
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

	engine := install.NewEngine(app.Config.CacheDir, app.Config.ManifestPath)
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

	if app.Config.JSON {
		return app.UI.JSON(aggregate)
	}
	printInstall(app, aggregate)
	if len(missing) > 0 {
		app.UI.Warn("%d skill%s in %s not found in the registry: %v",
			len(missing), plural(len(missing)), path, missing)
	}
	return nil
}
