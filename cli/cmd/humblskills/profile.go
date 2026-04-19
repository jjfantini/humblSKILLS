package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/adapters"
	"github.com/jjfantini/humblSKILLS/cli/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
)

func newProfileCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "View or edit your humblskills profile (default platforms + scope)",
		Long: "profile edits the user profile that drives install defaults. " +
			"Run bare for an interactive TUI editor, or use subcommands " +
			"(`show`, `set`, `reset`, `path`) for scripting.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runProfileDefault(app)
		},
	}
	cmd.AddCommand(
		newProfileShowCmd(app),
		newProfileSetCmd(app),
		newProfileResetCmd(app),
		newProfilePathCmd(app),
	)
	return cmd
}

func newProfileShowCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Print the current profile",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runProfileShow(app)
		},
	}
}

func newProfileSetCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a profile value. Keys: platforms, scope.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProfileSet(app, args[0], args[1])
		},
	}
}

func newProfileResetCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Delete the profile file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runProfileReset(app)
		},
	}
}

func newProfilePathCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the resolved profile file path",
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintln(app.UI.Out(), app.Config.ProfilePath)
			return nil
		},
	}
}

func runProfileDefault(app *App) error {
	if tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes) {
		return runProfileEditor(app)
	}
	return runProfileShow(app)
}

func runProfileEditor(app *App) error {
	adapterList, err := app.Adapters()
	if err != nil {
		return fmt.Errorf("load adapters: %w", err)
	}
	p, err := profile.Load(app.Config.ProfilePath)
	if err != nil {
		return err
	}
	updated, saved, err := tui.RunProfileEditorWith(app.UI.Theme(), adapterList, p, tui.ProfileHeaderSpec{
		Section: app.headerSection("Profile"),
		Meta:    app.headerMeta(""),
	})
	if err != nil {
		return err
	}
	if !saved {
		app.UI.Info("profile unchanged")
		return nil
	}
	if err := validateProfileAgainstAdapters(updated, adapterList); err != nil {
		return err
	}
	if err := profile.Save(app.Config.ProfilePath, updated); err != nil {
		return err
	}
	app.UI.Success("saved profile → %s", app.Config.ProfilePath)
	return nil
}

func runProfileShow(app *App) error {
	p, err := profile.Load(app.Config.ProfilePath)
	if err != nil {
		return err
	}
	if app.Config.JSON {
		return app.UI.JSON(p)
	}
	th := app.UI.Theme()
	app.UI.Header("profile")
	app.UI.Section("Defaults")
	fmt.Fprintln(app.UI.Out(), "  "+th.KVKey.Render("platforms")+"    "+
		th.KVValue.Render(formatPlatforms(p.DefaultPlatforms)))
	fmt.Fprintln(app.UI.Out(), "  "+th.KVKey.Render("scope")+"        "+
		th.KVValue.Render(formatScope(p.DefaultScope)))
	fmt.Fprintln(app.UI.Out(), "  "+th.KVKey.Render("path")+"         "+
		th.KVValue.Render(app.Config.ProfilePath))
	return nil
}

func runProfileSet(app *App, key, value string) error {
	adapterList, err := app.Adapters()
	if err != nil {
		return fmt.Errorf("load adapters: %w", err)
	}
	p, err := profile.Load(app.Config.ProfilePath)
	if err != nil {
		return err
	}

	switch key {
	case "platforms":
		names := parseCSV(value)
		if len(names) == 0 {
			p.DefaultPlatforms = nil
		} else {
			known := adapters.NameSet(adapterList)
			for _, n := range names {
				if _, ok := known[n]; !ok {
					return fmt.Errorf("unknown platform %q — valid: %s", n, strings.Join(adapterNames(adapterList), ", "))
				}
			}
			p.DefaultPlatforms = names
		}
	case "scope":
		if !profile.IsValidScope(value) {
			return fmt.Errorf("invalid scope %q — valid: user, project, \"\"", value)
		}
		p.DefaultScope = value
	default:
		return fmt.Errorf("unknown key %q — valid keys: platforms, scope", key)
	}

	if err := profile.Save(app.Config.ProfilePath, p); err != nil {
		return err
	}
	if app.Config.JSON {
		return app.UI.JSON(p)
	}
	app.UI.Success("set %s → %s", key, value)
	return nil
}

func runProfileReset(app *App) error {
	if err := profile.Delete(app.Config.ProfilePath); err != nil {
		return err
	}
	if app.Config.JSON {
		return app.UI.JSON(map[string]string{"path": app.Config.ProfilePath, "status": "deleted"})
	}
	app.UI.Success("removed %s", app.Config.ProfilePath)
	return nil
}

// validateProfileAgainstAdapters drops any platform names that aren't in the
// current adapter set; warns but doesn't fail.
func validateProfileAgainstAdapters(p *profile.Profile, adapterList []adapters.Adapter) error {
	if len(p.DefaultPlatforms) == 0 {
		return nil
	}
	kept, dropped := profile.FilterKnownPlatforms(p.DefaultPlatforms, adapters.NameSet(adapterList))
	if len(dropped) > 0 {
		return fmt.Errorf("unknown platform(s): %s", strings.Join(dropped, ", "))
	}
	p.DefaultPlatforms = kept
	return nil
}

func parseCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func adapterNames(as []adapters.Adapter) []string {
	out := make([]string, 0, len(as))
	for _, a := range as {
		out = append(out, a.Name)
	}
	sort.Strings(out)
	return out
}

func formatPlatforms(p []string) string {
	if len(p) == 0 {
		return "(all detected)"
	}
	return strings.Join(p, ", ")
}

func formatScope(s string) string {
	if s == "" {
		return "(adapter default)"
	}
	return s
}
