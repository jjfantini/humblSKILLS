package main

import (
	"fmt"
	"sort"
	"strconv"
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
		Use: "set <key> <value>",
		Short: "Set a profile value. Keys: platforms (csv), scope (global|user|project|adapter-default), " +
			"registry (URL/file:// path, or \"\" to clear), status_auto_return_seconds (seconds, or default|off), " +
			"tui_router (on|off — experimental single-program TUI, no flash between panes).",
		Args: cobra.ExactArgs(2),
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
	useTUI := tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes)

	type preload struct {
		adapters []adapters.Adapter
		profile  *profile.Profile
	}
	pre, err := tui.RunWithLoadingIf(useTUI, app.UI.Theme(), "loading profile…", func() (preload, error) {
		adapterList, err := app.Adapters()
		if err != nil {
			return preload{}, fmt.Errorf("load adapters: %w", err)
		}
		p, err := profile.Load(app.Config.ProfilePath)
		if err != nil {
			return preload{}, err
		}
		return preload{adapters: adapterList, profile: p}, nil
	})
	if err != nil {
		return err
	}
	adapterList, p := pre.adapters, pre.profile

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
	fmt.Fprintln(app.UI.Out(), "  "+th.KVKey.Render("registry")+"     "+
		th.KVValue.Render(formatRegistry(p.Registry)))
	fmt.Fprintln(app.UI.Out(), "  "+th.KVKey.Render("auto-return")+"  "+
		th.KVValue.Render(formatAutoReturn(p.StatusAutoReturnSeconds)))
	fmt.Fprintln(app.UI.Out(), "  "+th.KVKey.Render("tui-router")+"   "+
		th.KVValue.Render(formatTUIRouter(p.TUIRouter)))
	fmt.Fprintln(app.UI.Out(), "  "+th.KVKey.Render("path")+"         "+
		th.KVValue.Render(app.Config.ProfilePath))

	if len(p.Registries) > 0 {
		app.UI.Section("Registries")
		for _, r := range p.Registries {
			fmt.Fprintln(app.UI.Out(), "  "+th.KVKey.Render(r.Name)+"  "+
				th.KVValue.Render(r.URL)+"  "+th.Detail.Render("("+registryTokenLabel(r.Name)+")"))
		}
	}
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
			return fmt.Errorf("invalid scope %q — valid: global, user, project, adapter-default, \"\"", value)
		}
		p.DefaultScope = value
	case "registry":
		reg := normalizeRegistryURL(strings.TrimSpace(value))
		if reg != "" && !isPlausibleRegistry(reg) {
			return fmt.Errorf("invalid registry %q — expected owner/repo, a github.com URL, an http(s):// URL, a file:// path, or a filesystem path", value)
		}
		p.Registry = reg
	case "status_auto_return_seconds":
		seconds, err := parseStatusAutoReturnSeconds(value)
		if err != nil {
			return err
		}
		p.StatusAutoReturnSeconds = seconds
	case "tui_router":
		on, err := parseTUIRouter(value)
		if err != nil {
			return err
		}
		p.TUIRouter = on
	default:
		return fmt.Errorf("unknown key %q — valid keys: platforms, scope, registry, status_auto_return_seconds, tui_router", key)
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

// parseStatusAutoReturnSeconds parses the `profile set status_auto_return_seconds`
// value: "default"/"" resets to unset (built-in default), "off"/"disabled"
// disables the timer, anything else must be a non-negative integer.
func parseStatusAutoReturnSeconds(value string) (*int, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "default", "":
		return nil, nil
	case "off", "disabled":
		n := 0
		return &n, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 0 {
		return nil, fmt.Errorf("invalid status_auto_return_seconds %q — expected a non-negative integer, \"off\"/\"disabled\", or \"default\"", value)
	}
	return &n, nil
}

// formatAutoReturn renders StatusAutoReturnSeconds for `profile show`.
func formatAutoReturn(seconds *int) string {
	switch {
	case seconds == nil:
		return fmt.Sprintf("%ds (default)", profile.DefaultStatusAutoReturnSeconds)
	case *seconds <= 0:
		return "disabled — wait for enter/q"
	default:
		return fmt.Sprintf("%ds", *seconds)
	}
}

// parseTUIRouter parses the `profile set tui_router` value.
func parseTUIRouter(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "on", "true", "1":
		return true, nil
	case "off", "false", "0", "default", "":
		return false, nil
	}
	return false, fmt.Errorf("invalid tui_router %q — expected on or off", value)
}

// formatTUIRouter renders Profile.TUIRouter for `profile show`.
func formatTUIRouter(on bool) string {
	if on {
		return "on — experimental single-program TUI"
	}
	return "off (default)"
}

// formatRegistry renders Profile.Registry for `profile show`.
func formatRegistry(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(hosted default)"
	}
	return s
}

// isPlausibleRegistry does a light sanity check on a registry value: an
// http(s):// or file:// URL, or something path-like. It intentionally stays
// lenient — the fetcher does the real interpretation.
func isPlausibleRegistry(s string) bool {
	return strings.HasPrefix(s, "http://") ||
		strings.HasPrefix(s, "https://") ||
		strings.HasPrefix(s, "file://") ||
		strings.Contains(s, "/") ||
		strings.HasSuffix(s, ".json")
}

func formatScope(s string) string {
	p := profile.Profile{DefaultScope: s}
	switch resolved := p.ResolvedScope(); resolved {
	case profile.ScopeGlobal:
		if s == "" {
			return "global humblskills (default)"
		}
		return "global humblskills"
	case profile.ScopeAdapterDefault:
		return "adapter default"
	default:
		return resolved
	}
}
