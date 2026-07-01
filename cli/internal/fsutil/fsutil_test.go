package fsutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCopyTree_NestedAndModes(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "out")

	mustWrite(t, filepath.Join(src, "a.txt"), "alpha", 0o644)
	mustWrite(t, filepath.Join(src, "nested", "b.sh"), "#!/bin/sh\n", 0o755)

	if err := CopyTree(src, dst, Options{}); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, filepath.Join(dst, "a.txt")); got != "alpha" {
		t.Errorf("a.txt = %q", got)
	}
	if got := readFile(t, filepath.Join(dst, "nested", "b.sh")); got != "#!/bin/sh\n" {
		t.Errorf("b.sh = %q", got)
	}
	// Executable bit preserved (skip on Windows where perms are emulated).
	if runtime.GOOS != "windows" {
		fi, err := os.Stat(filepath.Join(dst, "nested", "b.sh"))
		if err != nil {
			t.Fatal(err)
		}
		if fi.Mode()&0o111 == 0 {
			t.Errorf("expected b.sh to stay executable, got %v", fi.Mode())
		}
	}
}

func TestCopyTree_NotADirectory(t *testing.T) {
	src := t.TempDir()
	file := filepath.Join(src, "f")
	mustWrite(t, file, "x", 0o644)
	if err := CopyTree(file, filepath.Join(t.TempDir(), "out"), Options{}); err == nil {
		t.Error("expected error copying a non-directory src")
	}
}

func TestCopyTree_SymlinkReject(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics differ on Windows")
	}
	src := t.TempDir()
	mustWrite(t, filepath.Join(src, "real.txt"), "data", 0o644)
	if err := os.Symlink(filepath.Join(src, "real.txt"), filepath.Join(src, "link.txt")); err != nil {
		t.Fatal(err)
	}

	// RejectSymlinks: error.
	if err := CopyTree(src, filepath.Join(t.TempDir(), "reject"), Options{RejectSymlinks: true}); err == nil {
		t.Error("expected error when RejectSymlinks is set and a symlink exists")
	}

	// Follow (default): the link target's contents are copied as a regular file.
	dst := filepath.Join(t.TempDir(), "follow")
	if err := CopyTree(src, dst, Options{}); err != nil {
		t.Fatalf("follow copy: %v", err)
	}
	if got := readFile(t, filepath.Join(dst, "link.txt")); got != "data" {
		t.Errorf("followed symlink content = %q, want data", got)
	}
	fi, err := os.Lstat(filepath.Join(dst, "link.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		t.Error("copied entry should be a regular file, not a symlink")
	}
}

func mustWrite(t *testing.T, path, body string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
