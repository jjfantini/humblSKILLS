package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/adapters"
	"github.com/jjfantini/humblSKILLS/cli/internal/frontmatter"
	"github.com/jjfantini/humblSKILLS/cli/internal/install"
	"github.com/jjfantini/humblSKILLS/cli/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
)

type migrateFlags struct {
	global bool
	scope  string
	force  bool
}

type migrateResult struct {
	Migrated []string `json:"migrated"`
	Skipped  []string `json:"skipped"`
}

func newMigrateCmd(app *App) *cobra.Command {
	var f migrateFlags
	cmd := &cobra.Command{
		Use:   "migrate [platform]",
		Short: "Adopt existing platform skills into the humblskills canonical store",
		Long: "migrate claude-code scans existing Claude Code skills, adopts registry-known " +
			"skills into the humblskills canonical store, and replaces platform copies " +
			"with symlinks. Unknown personal skills are reported and skipped.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			platform := "claude-code"
			if len(args) == 1 {
				platform = args[0]
			}
			return runMigrate(app, platform, f)
		},
	}
	cmd.Flags().BoolVar(&f.global, "global", true, "install adopted skills into ~/.humblskills and fan out to all detected platforms")
	cmd.Flags().StringVar(&f.scope, "scope", "user", "source install scope to scan (user|project)")
	cmd.Flags().BoolVar(&f.force, "force", false, "reinstall even if an adopted skill is already up-to-date")
	return cmd
}

func runMigrate(app *App, sourcePlatform string, f migrateFlags) error {
	if sourcePlatform != "claude-code" {
		return fmt.Errorf("migration currently supports only claude-code")
	}

	adapterList, err := app.Adapters()
	if err != nil {
		return fmt.Errorf("load adapters: %w", err)
	}
	adapterIndex := map[string]adapters.Adapter{}
	for _, a := range adapterList {
		adapterIndex[a.Name] = a
	}
	sourceAdapter, ok := adapterIndex[sourcePlatform]
	if !ok {
		return fmt.Errorf("unknown platform %q", sourcePlatform)
	}
	sourceTarget, err := sourceAdapter.Target(f.scope)
	if err != nil {
		return err
	}

	reg, _, err := app.registryFetcher().Load()
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}
	registryNames := map[string]struct{}{}
	for _, skill := range reg.Skills {
		registryNames[skill.Name] = struct{}{}
	}

	candidates, skipped, err := migrationCandidates(sourceTarget.Path, registryNames)
	if err != nil {
		return err
	}
	sort.Strings(candidates)
	sort.Strings(skipped)

	selectedPlatforms, err := selectPlatforms(adapterList, nil, f.global, nil)
	if err != nil {
		return err
	}
	if len(selectedPlatforms) == 0 {
		return fmt.Errorf("no platforms selected - run 'humblskills doctor' to see what's detected")
	}

	engine := app.installEngine()
	useTUI := tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes)
	result := migrateResult{Skipped: skipped}
	var aggregate install.Result

	run := func(sink install.EventSink) error {
		for _, name := range candidates {
			plan, err := install.Plan(reg, name)
			if err != nil {
				return err
			}
			res, err := engine.Execute(reg, plan, install.ExecuteOpts{
				Adapters:  adapterList,
				Platforms: selectedPlatforms,
				Scope:     f.scope,
				Force:     f.force,
				Global:    f.global,
				OnEvent:   sink,
			})
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
			aggregate.Results = append(aggregate.Results, res.Results...)
			aggregate.Warnings = append(aggregate.Warnings, res.Warnings...)
			result.Migrated = append(result.Migrated, name)
		}
		return nil
	}

	if useTUI {
		p, err := profile.Load(app.Config.ProfilePath)
		if err != nil {
			return err
		}
		if err := tui.ExecuteWithProgress(app.UI.Theme(), "migrate", p.StatusAutoReturnDuration(), run); err != nil {
			return err
		}
	} else {
		if err := run(nil); err != nil {
			return err
		}
	}

	if app.Config.JSON {
		return app.UI.JSON(result)
	}
	for _, warning := range aggregate.Warnings {
		app.UI.Warn("%s", warning.Msg)
	}
	for _, name := range result.Migrated {
		app.UI.Success("migrated %s", name)
	}
	for _, name := range result.Skipped {
		app.UI.Warn("skipped unregistered skill %s", name)
	}
	if len(result.Migrated) == 0 && len(result.Skipped) == 0 {
		app.UI.Info("no existing skills found in %s", sourceTarget.Path)
	}
	return nil
}

func migrationCandidates(root string, registryNames map[string]struct{}) (matched, skipped []string, err error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("read %s: %w", root, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillMD := filepath.Join(root, entry.Name(), "SKILL.md")
		data, err := os.ReadFile(skillMD)
		if err != nil {
			continue
		}
		fm, _, err := frontmatter.Parse(data)
		if err != nil || fm.Name == "" {
			skipped = append(skipped, entry.Name())
			continue
		}
		if _, ok := registryNames[fm.Name]; ok {
			matched = append(matched, fm.Name)
			continue
		}
		skipped = append(skipped, fm.Name)
	}
	return matched, skipped, nil
}
