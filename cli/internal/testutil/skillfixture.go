package testutil

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
)

// SkillFixture describes one skill to include in a BuildRegistry call.
type SkillFixture struct {
	Name        string
	Version     string
	Description string
	Tags        []string
	Platforms   []string
	Requires    []string
	Preserve    []string
	// Path inside the (fake) repo, e.g. "skills/foo". Defaults to
	// "skills/" + Name when empty.
	Path string
	// Files maps paths *inside* the skill directory to bodies. At
	// minimum supply a "SKILL.md".
	Files SkillTree
}

// BuildRegistry builds a registry.Registry pinned to owner/name@sha
// containing every SkillFixture (with correct DirSHA), and seeds the
// fetch tarball cache at cacheDir so install.Engine can resolve it
// offline.
//
// Returns the fully-populated Registry. Callers typically pass the
// result to install.Engine.Execute or write it to disk for
// file://-url registry tests.
func BuildRegistry(t testing.TB, cacheDir, owner, name, sha string, fixtures []SkillFixture) *registry.Registry {
	t.Helper()

	skills := make(map[string]SkillTree, len(fixtures))
	reg := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source: registry.Source{
			Repo: "github.com/" + owner + "/" + name,
			SHA:  sha,
		},
	}

	sorted := append([]SkillFixture(nil), fixtures...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	for _, f := range sorted {
		path := f.Path
		if path == "" {
			path = "skills/" + f.Name
		}
		if _, ok := f.Files["SKILL.md"]; !ok {
			t.Fatalf("fixture %q: Files must include SKILL.md", f.Name)
		}
		skills[path] = f.Files

		reg.Skills = append(reg.Skills, registry.Skill{
			Name:        f.Name,
			Version:     f.Version,
			Description: f.Description,
			Tags:        f.Tags,
			Platforms:   f.Platforms,
			Requires:    f.Requires,
			Preserve:    f.Preserve,
			Path:        path,
			DirSHA:      dirSHAForFiles(t, f.Files),
		})
	}

	SeedTarball(t, cacheDir, TarballSpec{
		Owner: owner, Name: name, SHA: sha,
		Skills: skills,
	})

	return reg
}

// dirSHAForFiles materialises files in a temp directory and returns
// registry.DirSHA(dir). Because DirSHA walks the tree deterministically,
// this must match what install.Engine computes after extraction.
func dirSHAForFiles(t testing.TB, files SkillTree) string {
	t.Helper()
	dir := t.TempDir()
	for rel, body := range files {
		abs := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(abs), err)
		}
		if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", abs, err)
		}
	}
	sha, err := registry.DirSHA(dir)
	if err != nil {
		t.Fatalf("dir_sha: %v", err)
	}
	return sha
}
