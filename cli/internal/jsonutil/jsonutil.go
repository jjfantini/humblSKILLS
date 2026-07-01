// Package jsonutil holds small JSON I/O helpers shared across the CLI. It
// replaces several private writeJSON copies that had drifted (some atomic,
// some not; some created parent dirs, some didn't).
package jsonutil

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// WriteFile marshals v as indented JSON (with a trailing newline) and writes
// it to path atomically: it creates the parent directory, writes to a temp
// file, then renames into place so a reader never sees a half-written file.
func WriteFile(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
