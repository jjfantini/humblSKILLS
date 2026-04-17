package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirSHA_Deterministic(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "SKILL.md", "hello")
	writeFile(t, dir, "assets/a.txt", "a")
	writeFile(t, dir, "assets/b.txt", "b")

	first, err := DirSHA(dir)
	if err != nil {
		t.Fatal(err)
	}
	second, err := DirSHA(dir)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("not deterministic: %s vs %s", first, second)
	}
}

func TestDirSHA_ContentChangeFlipsHash(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "SKILL.md", "hello")
	a, err := DirSHA(dir)
	if err != nil {
		t.Fatal(err)
	}

	writeFile(t, dir, "SKILL.md", "hello world")
	b, err := DirSHA(dir)
	if err != nil {
		t.Fatal(err)
	}

	if a == b {
		t.Errorf("hash did not change after content edit")
	}
}

func TestDirSHA_NewFileFlipsHash(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "SKILL.md", "hello")
	a, _ := DirSHA(dir)

	writeFile(t, dir, "extra.txt", "")
	b, _ := DirSHA(dir)

	if a == b {
		t.Errorf("hash did not change after file added")
	}
}

func TestDirSHA_RejectsSymlink(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "SKILL.md", "hello")
	if err := os.Symlink("SKILL.md", filepath.Join(dir, "link")); err != nil {
		t.Skip("symlinks not supported on this platform")
	}
	if _, err := DirSHA(dir); err == nil {
		t.Error("expected symlink error")
	}
}

func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	full := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
