package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/adapters"
	"github.com/jjfantini/humblSKILLS/cli/internal/env"
	"github.com/jjfantini/humblSKILLS/cli/internal/install"
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/internal/secrets"
	"github.com/jjfantini/humblSKILLS/cli/internal/textutil"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// App is the shared context wired onto every subcommand via PersistentPreRunE.
type App struct {
	UI       *ui.Printer
	Prompt   *ui.Prompter
	Config   Config
	Adapters func() ([]adapters.Adapter, error)

	// Nav is populated when a sub-command is launched from the dashboard loop.
	// Sub-screens mirror this into their HeaderSpec so the top header stays
	// consistent ("Dashboard > Install") with the shared status line.
	Nav NavContext
}

// NavContext carries the breadcrumb + status line that sub-screens render in
// their shared header when the user is navigating from the dashboard.
type NavContext struct {
	Crumb  string // "Dashboard > Install" (empty when entered via a direct CLI invocation)
	Status tui.DashboardStatus
}

// Config captures every flag/env-resolved setting used by subcommands.
type Config struct {
	RegistryURL string
	// RegistryToken holds ONLY the --token flag value. The effective token is
	// resolved lazily by (*App).registryToken(): flag > HUMBLSKILLS_TOKEN env >
	// OS keyring > secrets file. Resolving lazily keeps commands that never
	// fetch (list, profile, doctor) from touching the keychain.
	RegistryToken string
	CacheDir      string
	ManifestPath  string
	ProfilePath   string
	JSON          bool
	NoColor       bool
	Verbose       bool
	Quiet         bool
	Yes           bool
	Fullscreen    bool
}

type globalFlags struct {
	registry   string
	token      string
	cacheDir   string
	manifest   string
	profile    string
	json       bool
	noColor    bool
	verbose    bool
	quiet      bool
	yes        bool
	fullscreen bool
}

func newRootCmd() *cobra.Command {
	var g globalFlags
	app := &App{}

	cmd := &cobra.Command{
		Use:           "humblskills",
		Short:         "Install agentskills.io-format skills into your agent platform",
		Long:          "humblskills installs curated skills (agentskills.io format) into Claude Code, Cursor, and friends — single binary, no server, no telemetry.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return configureApp(cmd, app, g)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Bare `humblskills` on an interactive TTY launches the
			// dashboard. Non-TTY falls through to the help-style fallback so
			// pipes, agents, and --json callers still get something useful.
			return runStart(app)
		},
	}

	f := cmd.PersistentFlags()
	f.StringVar(&g.registry, "registry", "", "registry URL (or file:// path). Defaults to the hosted registry; env: HUMBLSKILLS_REGISTRY")
	f.StringVar(&g.token, "token", "", "auth token for a private registry, sent as a Bearer token on registry + skill fetches (env: HUMBLSKILLS_TOKEN)")
	f.StringVar(&g.cacheDir, "cache-dir", "", "cache directory (env: HUMBLSKILLS_CACHE_DIR; default: XDG_CACHE_HOME/humblskills)")
	f.StringVar(&g.manifest, "manifest", "", "install manifest path (env: HUMBLSKILLS_MANIFEST; default: XDG_STATE_HOME/humblskills/manifest.json)")
	f.StringVar(&g.profile, "profile", "", "profile config path (env: HUMBLSKILLS_PROFILE; default: ~/.humblskills/profile.json)")
	f.BoolVar(&g.json, "json", false, "emit machine-readable JSON")
	f.BoolVar(&g.noColor, "no-color", false, "disable ANSI colour output")
	f.BoolVarP(&g.verbose, "verbose", "v", false, "print extra detail")
	f.BoolVarP(&g.quiet, "quiet", "q", false, "suppress non-error output")
	f.BoolVarP(&g.yes, "yes", "y", false, "skip confirmation prompts (auto-accept)")
	f.BoolVar(&g.fullscreen, "fullscreen", false, "open the interactive dashboard in full-screen TUI mode")

	cmd.AddCommand(
		newStartCmd(app),
		newVersionCmd(app),
		newDoctorCmd(app),
		newRegistryCmd(app),
		newInstallCmd(app),
		newMigrateCmd(app),
		newUninstallCmd(app),
		newUpdateCmd(app),
		newUpgradeCmd(app),
		newListCmd(app),
		newSearchCmd(app),
		newInitCmd(app),
		newExportCmd(app),
		newSyncCmd(app),
		newProfileCmd(app),
		newEvalCmd(app),
	)

	return cmd
}

func configureApp(_ *cobra.Command, app *App, g globalFlags) error {
	if g.quiet && g.verbose {
		return fmt.Errorf("--quiet and --verbose are mutually exclusive")
	}

	level := ui.LevelNormal
	switch {
	case g.quiet:
		level = ui.LevelQuiet
	case g.verbose:
		level = ui.LevelVerbose
	}

	app.UI = ui.New(ui.Options{
		Level:   level,
		NoColor: g.noColor,
		JSON:    g.json,
	})
	// JSON mode is inherently non-interactive; callers get machine output.
	app.Prompt = ui.NewPrompter(g.yes || g.json)

	// Load .env from repo root if present BEFORE reading env vars below,
	// so users who drop API keys in .env get them picked up on every
	// invocation without export-ing. Env always wins over file.
	if res, err := env.LoadDotEnv(""); err != nil {
		app.UI.Warn(".env: %v", err)
	} else if res.Path != "" && g.verbose {
		app.UI.Detail("loaded %d key(s) from %s (kept %d that were already set)",
			len(res.Loaded), res.Path, len(res.Kept))
	}

	profilePath, err := resolveProfilePath(textutil.FirstNonEmpty(g.profile, os.Getenv("HUMBLSKILLS_PROFILE")))
	if err != nil {
		return err
	}

	// Load the profile up front (best-effort) so a persisted registry URL can
	// act as a fallback below. A missing profile yields an empty one, so a read
	// error here is non-fatal — commands that need the profile load it and
	// surface real errors themselves.
	var profileRegistry string
	if p, perr := profile.Load(profilePath); perr == nil && p != nil {
		profileRegistry = p.Registry
	}

	cfg := Config{
		RegistryURL:   textutil.FirstNonEmpty(g.registry, os.Getenv("HUMBLSKILLS_REGISTRY"), profileRegistry, registry.DefaultURL),
		RegistryToken: g.token, // flag only; env/keyring/file resolved in registryToken()
		ProfilePath:   profilePath,
		JSON:          g.json,
		NoColor:       g.noColor,
		Verbose:       g.verbose,
		Quiet:         g.quiet,
		Yes:           g.yes,
		Fullscreen:    g.fullscreen,
	}

	cacheDir, err := resolveCacheDir(textutil.FirstNonEmpty(g.cacheDir, os.Getenv("HUMBLSKILLS_CACHE_DIR")))
	if err != nil {
		return err
	}
	cfg.CacheDir = cacheDir

	manifestPath, err := resolveManifestPath(textutil.FirstNonEmpty(g.manifest, os.Getenv("HUMBLSKILLS_MANIFEST")))
	if err != nil {
		return err
	}
	cfg.ManifestPath = manifestPath

	app.Config = cfg
	app.Adapters = adapters.Load

	return nil
}

// registryToken resolves the effective registry auth token, in precedence
// order: the --token flag > HUMBLSKILLS_TOKEN env > OS keyring > secrets file.
// It is called only when a fetcher/engine is actually built, so commands that
// never fetch don't touch the keychain.
func (a *App) registryToken() string {
	if a.Config.RegistryToken != "" {
		return a.Config.RegistryToken
	}
	tok, _ := secrets.GetRegistryToken()
	return tok
}

// registryFetcher builds a registry Fetcher using the resolved registry URL,
// cache dir, and (optional) auth token. Centralising construction here keeps the
// token wired into every command that reads the registry.
func (a *App) registryFetcher() *registry.Fetcher {
	f := registry.NewFetcher(a.Config.RegistryURL, a.Config.CacheDir)
	f.Token = a.registryToken()
	return f
}

// installEngine builds an install Engine with the (optional) auth token wired
// into its tarball fetcher, so skill content can be pulled from a private
// registry's backing repo.
func (a *App) installEngine() *install.Engine {
	e := install.NewEngine(a.Config.CacheDir, a.Config.ManifestPath)
	e.Fetcher.Token = a.registryToken()
	return e
}

func resolveCacheDir(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	if xdg.CacheHome != "" {
		return filepath.Join(xdg.CacheHome, "humblskills"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve cache dir: %w", err)
	}
	return filepath.Join(home, ".cache", "humblskills"), nil
}

func resolveManifestPath(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	return manifest.DefaultPath()
}

func resolveProfilePath(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	return profile.DefaultPath()
}
