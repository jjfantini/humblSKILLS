package install

import (
	"archive/zip"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// WriteSkillZip zips a skill's store directory in claude.ai's required upload
// layout: every file under a single "<skillName>/…" folder at the zip root.
// Regular files only (symlinks skipped), sorted walk for deterministic
// output. Used by the claude-desktop adapter's zip transform and the
// `export desktop` command.
func WriteSkillZip(storePath, skillName, outPath string) error {
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

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	zw := zip.NewWriter(f)
	write := func() error {
		for _, p := range files {
			rel, err := filepath.Rel(storePath, p)
			if err != nil {
				return err
			}
			w, err := zw.Create(skillName + "/" + filepath.ToSlash(rel))
			if err != nil {
				return err
			}
			data, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			if _, err := w.Write(data); err != nil {
				return err
			}
		}
		return zw.Close()
	}
	if err := write(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
