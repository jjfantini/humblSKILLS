package main

import (
	"archive/zip"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/install"
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/internal/textutil"
)

// Claude Desktop and claude.ai can't read filesystem skills — they take an
// account-level zip upload (skill folder at the zip root, SKILL.md inside).
// This file is the bridge: `export desktop` writes upload-ready zips from the
// canonical store, and the profile's desktop_exports setting regenerates them
// automatically after install/update.

const desktopUploadHint = "upload at claude.ai (or the Claude desktop app): Settings → Capabilities → Skills → upload the zip. Uploads don't auto-update — re-export after 'humblskills update'."

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
			"Set 'humblskills profile set desktop_exports on' to regenerate the " +
			"zips automatically on every install/update.",
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
		zipPath := filepath.Join(outDir, name+"-"+inst.Version+".zip")
		if err := writeDesktopZip(inst.StorePath, name, zipPath); err != nil {
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

// defaultDesktopDir is where auto-exports and no-flag runs put the zips:
// ~/.humblskills/desktop, next to the canonical global store.
func defaultDesktopDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home: %w", err)
	}
	return filepath.Join(home, ".humblskills", "desktop"), nil
}

// writeDesktopZip zips the skill's store directory in claude.ai's required
// layout: every file under a single "<skillName>/…" folder at the zip root.
// Regular files only (symlinks skipped), sorted walk for deterministic output.
func writeDesktopZip(storePath, skillName, outPath string) error {
	var files []string
	err := filepath.WalkDir(storePath, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type().IsRegular() {
			files = append(files, p)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk %s: %w", storePath, err)
	}
	if len(files) == 0 {
		return fmt.Errorf("store %s has no files", storePath)
	}
	sort.Strings(files)

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	zw := zip.NewWriter(f)
	for _, p := range files {
		rel, err := filepath.Rel(storePath, p)
		if err != nil {
			zw.Close()
			f.Close()
			return err
		}
		w, err := zw.Create(skillName + "/" + filepath.ToSlash(rel))
		if err != nil {
			zw.Close()
			f.Close()
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			zw.Close()
			f.Close()
			return err
		}
		if _, err := w.Write(data); err != nil {
			zw.Close()
			f.Close()
			return err
		}
	}
	if err := zw.Close(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// maybeDesktopExport regenerates upload zips for the run's changed skills
// when the profile's desktop_exports setting is on. Returns the zip info for
// JSON envelopes; prints success lines + the upload hint on the text path.
// Never fails the install — zip problems are warnings.
func maybeDesktopExport(app *App, res install.Result) []desktopZipInfo {
	p, err := profile.Load(app.Config.ProfilePath)
	if err != nil || p == nil || !p.DesktopExports {
		return nil
	}
	outDir, err := defaultDesktopDir()
	if err != nil {
		app.UI.Warn("desktop zips: %v", err)
		return nil
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		app.UI.Warn("desktop zips: %v", err)
		return nil
	}

	seen := map[string]bool{}
	var infos []desktopZipInfo
	for _, t := range res.Results {
		if seen[t.Skill] || t.StorePath == "" || t.Outcome == install.OutcomeSkipped {
			continue
		}
		seen[t.Skill] = true
		zipPath := filepath.Join(outDir, t.Skill+"-"+t.Version+".zip")
		if err := writeDesktopZip(t.StorePath, t.Skill, zipPath); err != nil {
			app.UI.Warn("desktop zip for %s: %v", t.Skill, err)
			continue
		}
		infos = append(infos, desktopZipInfo{Skill: t.Skill, Version: t.Version, Zip: zipPath})
	}
	if len(infos) > 0 && !app.Config.JSON {
		for _, z := range infos {
			app.UI.Success("desktop zip → %s", z.Zip)
		}
		app.UI.Info("%s", desktopUploadHint)
	}
	return infos
}
