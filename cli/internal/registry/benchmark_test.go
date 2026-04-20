package registry_test

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
)

// buildBenchTree materialises a representative skill tree for DirSHA
// benchmarking. N files across M subdirs approximates a medium skill
// (SKILL.md + references/ + scripts/).
func buildBenchTree(b *testing.B, root string, files int) {
	b.Helper()
	for i := 0; i < files; i++ {
		sub := filepath.Join(root, "dir"+strconv.Itoa(i%5))
		if err := os.MkdirAll(sub, 0o755); err != nil {
			b.Fatal(err)
		}
		body := "file " + strconv.Itoa(i) + " content with some padding for realism."
		if err := os.WriteFile(filepath.Join(sub, "f"+strconv.Itoa(i)+".md"), []byte(body), 0o644); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDirSHA_SmallTree(b *testing.B) {
	dir := b.TempDir()
	buildBenchTree(b, dir, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := registry.DirSHA(dir); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDirSHA_MediumTree(b *testing.B) {
	dir := b.TempDir()
	buildBenchTree(b, dir, 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := registry.DirSHA(dir); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDirSHA_LargeTree(b *testing.B) {
	dir := b.TempDir()
	buildBenchTree(b, dir, 500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := registry.DirSHA(dir); err != nil {
			b.Fatal(err)
		}
	}
}
