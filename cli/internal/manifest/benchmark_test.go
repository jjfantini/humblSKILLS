package manifest_test

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
)

// seedBench returns a manifest populated with 500 installations so
// Load/Save exercise realistic JSON volumes.
func seedBench() *manifest.Manifest {
	m := &manifest.Manifest{SchemaVersion: manifest.SchemaVersion}
	for i := 0; i < 500; i++ {
		m.Upsert(manifest.Installation{
			Skill: fmt.Sprintf("skill-%d", i), Version: "1.0.0",
			Platform:    "claude-code",
			Scope:       "user",
			Path:        fmt.Sprintf("/tmp/skill-%d", i),
			InstalledAt: time.Now().UTC(),
			SourceSHA:   "deadbeef",
			RegistryRef: fmt.Sprintf("ref-%d", i),
		})
	}
	return m
}

func BenchmarkManifestSave(b *testing.B) {
	m := seedBench()
	path := filepath.Join(b.TempDir(), "manifest.json")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := manifest.Save(path, m); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkManifestLoad(b *testing.B) {
	m := seedBench()
	path := filepath.Join(b.TempDir(), "manifest.json")
	if err := manifest.Save(path, m); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := manifest.Load(path); err != nil {
			b.Fatal(err)
		}
	}
}
