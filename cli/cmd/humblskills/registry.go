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
		Short: "Inspect and refresh the skill registry cache",
	}
	cmd.AddCommand(
		newRegistryRefreshCmd(app),
		newRegistryLoginCmd(app),
		newRegistryLogoutCmd(app),
	)
	return cmd
}

func newRegistryLoginCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Store a registry auth token (for a private registry) in the OS keychain",
		Long: "login saves the GitHub token used to read a private registry and download its skills. " +
			"The token is stored in the OS keychain (falling back to a 0600 file if the keychain is " +
			"unavailable) and used automatically on registry + skill fetches — no env var needed. " +
			"Provide it with the global --token flag, by piping it on stdin, or via the masked prompt.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRegistryLogin(app)
		},
	}
}

func newRegistryLogoutCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove the stored registry auth token from the keychain / secrets file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := secrets.DeleteRegistryToken(); err != nil {
				return err
			}
			if app.Config.JSON {
				return app.UI.JSON(map[string]string{"status": "removed"})
			}
			app.UI.Success("removed stored registry token")
			return nil
		},
	}
}

func runRegistryLogin(app *App) error {
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

	src, err := secrets.SetRegistryToken(token)
	if err != nil {
		return err
	}
	if app.Config.JSON {
		return app.UI.JSON(map[string]string{"stored": string(src)})
	}
	app.UI.Success("stored registry token in %s", src)
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
