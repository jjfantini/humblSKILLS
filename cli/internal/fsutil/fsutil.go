// Package fsutil holds small filesystem helpers shared across the CLI. It
// exists to replace several near-identical private copyTree implementations
// that had drifted apart (install staging, eval brain/harness/clitool).
package fsutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Options tunes CopyTree.
type Options struct {
	// RejectSymlinks makes CopyTree fail when it encounters a symlink rather
	// than copying the link target's contents. Install staging sets this so a
	// crafted skill tarball can't smuggle a symlink into the canonical store.
	RejectSymlinks bool
}

// CopyTree recursively copies the directory tree rooted at src into dst.
//
//   - File modes are preserved.
//   - Directories are created owner-traversable (source mode | 0700) so a
//     read-only source directory doesn't block the copy.
//   - Symlinks: when Options.RejectSymlinks is set, encountering one is an
//     error; otherwise the link target's contents are copied as a regular
//     file (symlinks are followed, never recreated).
//
// src must be an existing directory.
func CopyTree(src, dst string, opts Options) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("copytree: %s is not a directory", src)
	}
	return filepath.Walk(src, func(p string, fi os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if fi.IsDir() {
			return os.MkdirAll(target, fi.Mode()&0o777|0o700)
		}
		if fi.Mode()&os.ModeSymlink != 0 && opts.RejectSymlinks {
			return fmt.Errorf("copytree: refusing to copy symlink %s", p)
		}
		return copyFile(p, target)
	})
}

// copyFile copies the regular file at src to dst (following symlinks),
// creating parent directories and preserving the source file's permission
// bits.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	if fi, err := os.Stat(src); err == nil {
		_ = os.Chmod(dst, fi.Mode()&0o777)
	}
	return nil
}
