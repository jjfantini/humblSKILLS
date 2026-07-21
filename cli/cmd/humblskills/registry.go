package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/jjfantini/humblSKILLS/cli/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/internal/secrets"
	"github.com/jjfantini/humblSKILLS/cli/internal/textutil"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
)

type refreshResult struct {
	URL       string    `json:"url"`
	Source    string    `json:"source"`
	Skills    int       `json:"skills"`
	FetchedAt time.Time `json:"fetched_at"`
	CachePath string    `json:"cache_path,omitempty"`
}

func newRegistryCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "Manage skill registries (add/list/remove), refresh the cache, and auth",
	}
	cmd.AddCommand(
		newRegistryRefreshCmd(app),
		newRegistryLoginCmd(app),
		newRegistryLogoutCmd(app),
		newRegistryAddCmd(app),
		newRegistryListCmd(app),
		newRegistryRemoveCmd(app),
	)
	return cmd
}

func newRegistryAddCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "add <name> <url>",
		Short: "Add (or update) a named registry shown in aggregated views",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRegistryAdd(app, args[0], args[1])
		},
	}
}

func newRegistryListCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured registries",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRegistryList(app)
		},
	}
}

func newRegistryRemoveCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm"},
		Short:   "Remove a named registry (and its stored token)",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRegistryRemove(app, args[0])
		},
	}
}

func runRegistryAdd(app *App, name, url string) error {
	name = strings.TrimSpace(name)
	url = strings.TrimSpace(url)
	if name == "" {
		return fmt.Errorf("registry name must not be empty")
	}
	if !isPlausibleRegistry(url) {
		return fmt.Errorf("invalid registry URL %q — expected an http(s):// URL, a file:// URL, or a filesystem path", url)
	}
	p, err := profile.Load(app.Config.ProfilePath)
	if err != nil {
		return err
	}
	p.SetRegistry(name, url)
	if err := profile.Save(app.Config.ProfilePath, p); err != nil {
		return err
	}
	if app.Config.JSON {
		return app.UI.JSON(map[string]string{"name": name, "url": url})
	}
	app.UI.Success("added registry %q → %s", name, url)
	app.UI.Detail("store its token (if private) with: humblskills registry login --name %s", name)
	return nil
}

func runRegistryList(app *App) error {
	regs := app.resolvedRegistries()
	if app.Config.JSON {
		return app.UI.JSON(regs)
	}
	app.UI.Header("registries")
	p, _ := profile.Load(app.Config.ProfilePath)
	if p == nil || len(p.Registries) == 0 {
		app.UI.Info("no named registries configured — using the single default:")
	}
	th := app.UI.Theme()
	for _, r := range regs {
		_, src := secrets.GetRegistryTokenFor(r.Name)
		tok := "no token"
		if src != secrets.SourceAbsent {
			tok = "token: " + string(src)
		}
		fmt.Fprintln(app.UI.Out(), "  "+th.KVKey.Render(r.Name)+"  "+th.KVValue.Render(r.URL)+"  ("+tok+")")
	}
	return nil
}

func runRegistryRemove(app *App, name string) error {
	p, err := profile.Load(app.Config.ProfilePath)
	if err != nil {
		return err
	}
	if !p.RemoveRegistry(name) {
		return fmt.Errorf("no registry named %q", name)
	}
	if err := profile.Save(app.Config.ProfilePath, p); err != nil {
		return err
	}
	_ = secrets.DeleteRegistryTokenFor(name) // best-effort token cleanup
	if app.Config.JSON {
		return app.UI.JSON(map[string]string{"removed": name})
	}
	app.UI.Success("removed registry %q", name)
	return nil
}

func newRegistryLoginCmd(app *App) *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Store a registry auth token (for a private registry) in the OS keychain",
		Long: "login saves the GitHub token used to read a private registry and download its skills. " +
			"The token is stored in the OS keychain (falling back to a 0600 file if the keychain is " +
			"unavailable) and used automatically on registry + skill fetches — no env var needed. " +
			"Provide it with the global --token flag, by piping it on stdin, or via the masked prompt. " +
			"Use --name to attach the token to a specific named registry (see `registry add`).",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRegistryLogin(app, name)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "named registry to attach the token to (default: the single/default token)")
	return cmd
}

func newRegistryLogoutCmd(app *App) *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove the stored registry auth token from the keychain / secrets file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := secrets.DeleteRegistryTokenFor(name); err != nil {
				return err
			}
			if app.Config.JSON {
				return app.UI.JSON(map[string]string{"status": "removed", "name": name})
			}
			if name != "" {
				app.UI.Success("removed stored token for registry %q", name)
			} else {
				app.UI.Success("removed stored registry token")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "named registry whose token to remove (default: the single/default token)")
	return cmd
}

func runRegistryLogin(app *App, name string) error {
	// Priority: global --token flag > piped stdin > interactive masked prompt.
	token := strings.TrimSpace(app.Config.RegistryToken)
	if token == "" {
		if term.IsTerminal(int(os.Stdin.Fd())) {
			v, err := app.Prompt.Secret("Registry auth token (input hidden)")
			if err != nil {
				return err
			}
			token = strings.TrimSpace(v)
		} else {
			b, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("read token from stdin: %w", err)
			}
			token = strings.TrimSpace(string(b))
		}
	}
	if token == "" {
		return fmt.Errorf("no token provided (use --token, pipe it on stdin, or enter it at the prompt)")
	}

	src, err := secrets.SetRegistryTokenFor(name, token)
	if err != nil {
		return err
	}
	if app.Config.JSON {
		return app.UI.JSON(map[string]string{"stored": string(src), "name": name})
	}
	if name != "" {
		app.UI.Success("stored token for registry %q in %s", name, src)
	} else {
		app.UI.Success("stored registry token in %s", src)
	}
	if src == secrets.SourceFile {
		app.UI.Detail("OS keychain unavailable — stored in a 0600 secrets file instead")
	}
	return nil
}

func newRegistryRefreshCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "refresh",
		Short: "Force a refresh of the cached registry",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRegistryRefresh(app)
		},
	}
}

func runRegistryRefresh(app *App) error {
	f := app.registryFetcher()

	refresh := func() (refreshResult, error) {
		r, o, err := f.Refresh()
		if err != nil {
			return refreshResult{}, err
		}
		info := f.Inspect()
		return refreshResult{
			URL:       app.Config.RegistryURL,
			Source:    string(o),
			Skills:    len(r.Skills),
			FetchedAt: info.FetchedAt,
			CachePath: info.Path,
		}, nil
	}

	if tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes) {
		p, err := profile.Load(app.Config.ProfilePath)
		if err != nil {
			return err
		}
		_, err = tui.ExecuteWithStatus(app.UI.Theme(), "registry", "refreshing registry…", p.StatusAutoReturnDuration(),
			func() (tui.StatusResult, error) {
				res, err := refresh()
				if err != nil {
					return tui.StatusResult{}, err
				}
				return tui.StatusResult{
					Headline: fmt.Sprintf("registry refreshed: %d skill%s from %s", res.Skills, textutil.Plural(res.Skills), res.Source),
					Lines:    []string{"cache: " + res.CachePath},
				}, nil
			})
		return err
	}

	var res refreshResult
	err := tui.RunWithSpinner(app.UI.Theme(), "refreshing registry…", func() error {
		r, e := refresh()
		res = r
		return e
	})
	if err != nil {
		return err
	}
	if app.Config.JSON {
		return app.UI.JSON(res)
	}
	app.UI.Success("registry refreshed: %d skill%s from %s", res.Skills, textutil.Plural(res.Skills), res.Source)
	app.UI.Detail("  cache: %s", res.CachePath)
	return nil
}
