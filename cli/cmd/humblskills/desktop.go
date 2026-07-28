package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/install"
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/textutil"
)

// Claude Desktop and claude.ai can't read skills from the filesystem — they
// take an account-level zip upload (skill folder at the zip root). The
// claude-desktop ADAPTER is the primary path: select it like any platform and
// install/update keep ~/.humblskills/desktop stocked with current zips. This
// command is the manual complement: regenerate zips on demand without
// reinstalling anything.

const desktopUploadHint = "upload at claude.ai (or the Claude desktop app): Settings → Capabilities → Skills → upload the zip. Uploads don't auto-update — re-upload after 'humblskills update'."

// desktopZipInfo is one written zip, for JSON envelopes and printing.
type desktopZipInfo struct {
	Skill   string `json:"skill"`
	Version string `json:"version"`
	Zip     string `json:"zip"`
}

func newExportDesktopCmd(app *App) *cobra.Command {
	var outDir string
	cmd := &cobra.Command{
		Use:   "desktop [skill...]",
		Short: "Write claude.ai / Claude Desktop upload zips for installed skills",
		Long: "Claude Desktop and claude.ai don't read skills from the filesystem — " +
			"they take a zip upload (the skill folder at the zip root). " +
			"'export desktop' writes one upload-ready zip per installed skill " +
			"(or just the ones you name) from the canonical humblskills store. " +
			"Tip: select claude-desktop as a platform (install --platform " +
			"claude-desktop, or in your profile's default platforms) and " +
			"install/update keep the zips current automatically.",
		Args: cobra.ArbitraryArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			return runExportDesktop(app, args, outDir)
		},
	}
	cmd.Flags().StringVarP(&outDir, "output-dir", "o", "", "directory for the zips (default: ~/.humblskills/desktop)")
	return cmd
}

func runExportDesktop(app *App, names []string, outDir string) error {
	m, err := manifest.Load(app.Config.ManifestPath)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		names = uniqueSkillsFromManifest(m)
	}
	if len(names) == 0 {
		return fmt.Errorf("no skills installed — nothing to export")
	}
	if outDir == "" {
		if outDir, err = defaultDesktopDir(); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", outDir, err)
	}

	var infos []desktopZipInfo
	for _, name := range names {
		inst := installWithStorePath(m, name)
		if inst == nil {
			return fmt.Errorf("%q has no canonical store recorded — is it installed? (older installs: reinstall once with the current CLI)", name)
		}
		zipPath := filepath.Join(outDir, name+".zip")
		if err := install.WriteSkillZip(inst.StorePath, name, zipPath); err != nil {
			return fmt.Errorf("zip %s: %w", name, err)
		}
		infos = append(infos, desktopZipInfo{Skill: name, Version: inst.Version, Zip: zipPath})
	}

	if app.Config.JSON {
		return app.UI.JSON(struct {
			Zips []desktopZipInfo `json:"zips"`
		}{infos})
	}
	for _, z := range infos {
		app.UI.Success("desktop zip → %s", z.Zip)
	}
	app.UI.Info("%s", desktopUploadHint)
	app.UI.Info("wrote %d zip%s", len(infos), textutil.Plural(len(infos)))
	return nil
}

// installWithStorePath returns an installation of skill that records the
// canonical store path (any platform — the store is shared), or nil.
func installWithStorePath(m *manifest.Manifest, skill string) *manifest.Installation {
	for _, inst := range m.FindAll(skill) {
		if inst.StorePath != "" {
			return inst
		}
	}
	return nil
}

// defaultDesktopDir is where the zips land: ~/.humblskills/desktop — the same
// place the claude-desktop adapter targets.
func defaultDesktopDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home: %w", err)
	}
	return filepath.Join(home, ".humblskills", "desktop"), nil
}
