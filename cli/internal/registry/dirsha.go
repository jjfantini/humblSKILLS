package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// DirSHA computes a deterministic sha256 over a skill directory tree. It
// canonicalizes the tree into sorted (rel_path, mode, content_sha) tuples and
// hashes that serialization. Symlinks are rejected — Phase 1 skills are plain
// text trees.
func DirSHA(root string) (string, error) {
	type entry struct {
		rel    string
		mode   fs.FileMode
		sumHex string
		isDir  bool
	}
	var entries []entry

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink not supported in skill tree: %s", rel)
		}
		if d.IsDir() {
			entries = append(entries, entry{rel: rel + "/", mode: info.Mode() & 0o777, isDir: true})
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			_ = f.Close()
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		entries = append(entries, entry{
			rel:    rel,
			mode:   info.Mode() & 0o777,
			sumHex: hex.EncodeToString(h.Sum(nil)),
		})
		return nil
	})
	if err != nil {
		return "", err
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].rel < entries[j].rel })

	h := sha256.New()
	for _, e := range entries {
		sum := e.sumHex
		if e.isDir {
			sum = "-"
		}
		fmt.Fprintf(h, "%s\x00%o\x00%s\n", e.rel, e.mode, sum)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
