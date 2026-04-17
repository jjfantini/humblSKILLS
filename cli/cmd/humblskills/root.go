package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/platform"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/internal/ui"
)

// App is the shared context wired onto every subcommand via PersistentPreRunE.
type App struct {
	UI       *ui.Printer
	Config   Config
	Adapters func() ([]platform.Adapter, error)
}

// Config captures every flag/env-resolved setting used by subcommands.
type Config struct {
	RegistryURL  string
	CacheDir     string
	ManifestPath string
	AdaptersDir  string
	JSON         bool
	NoColor      bool
	Verbose      bool
	Quiet        bool
}

type globalFlags struct {
	registry    string
	cacheDir    string
	manifest    string
	adaptersDir string
	json        bool
	noColor     bool
	verbose     bool
	quiet       bool
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
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return configureApp(cmd, app, g)
		},
	}

	f := cmd.PersistentFlags()
	f.StringVar(&g.registry, "registry", "", "registry URL (or file:// path). Defaults to the hosted registry; env: HUMBLSKILLS_REGISTRY")
	f.StringVar(&g.cacheDir, "cache-dir", "", "cache directory (env: HUMBLSKILLS_CACHE_DIR; default: XDG_CACHE_HOME/humblskills)")
	f.StringVar(&g.manifest, "manifest", "", "install manifest path (env: HUMBLSKILLS_MANIFEST; default: XDG_STATE_HOME/humblskills/manifest.json)")
	f.StringVar(&g.adaptersDir, "adapters-dir", "", "override embedded adapters with a local directory (for development)")
	f.BoolVar(&g.json, "json", false, "emit machine-readable JSON")
	f.BoolVar(&g.noColor, "no-color", false, "disable ANSI colour output")
	f.BoolVarP(&g.verbose, "verbose", "v", false, "print extra detail")
	f.BoolVarP(&g.quiet, "quiet", "q", false, "suppress non-error output")

	cmd.AddCommand(
		newVersionCmd(app),
		newDoctorCmd(app),
		newRegistryCmd(app),
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

	cfg := Config{
		RegistryURL:  firstNonEmpty(g.registry, os.Getenv("HUMBLSKILLS_REGISTRY"), registry.DefaultURL),
		AdaptersDir:  g.adaptersDir,
		JSON:         g.json,
		NoColor:      g.noColor,
		Verbose:      g.verbose,
		Quiet:        g.quiet,
	}

	cacheDir, err := resolveCacheDir(firstNonEmpty(g.cacheDir, os.Getenv("HUMBLSKILLS_CACHE_DIR")))
	if err != nil {
		return err
	}
	cfg.CacheDir = cacheDir

	manifestPath, err := resolveManifestPath(firstNonEmpty(g.manifest, os.Getenv("HUMBLSKILLS_MANIFEST")))
	if err != nil {
		return err
	}
	cfg.ManifestPath = manifestPath

	app.Config = cfg
	app.Adapters = adapterLoader(cfg.AdaptersDir)

	return nil
}

func adapterLoader(overrideDir string) func() ([]platform.Adapter, error) {
	if overrideDir != "" {
		return func() ([]platform.Adapter, error) { return platform.LoadAll(overrideDir) }
	}
	return platform.LoadBuiltin
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

func firstNonEmpty(vs ...string) string {
	for _, v := range vs {
		if v != "" {
			return v
		}
	}
	return ""
}
