package install

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/platform"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
)

// seedTarball writes a fake GitHub-style tarball into cacheDir in the exact
// location the Fetcher expects, so Execute's Fetch call becomes a cache hit.
func seedTarball(t *testing.T, cacheDir, owner, name, sha, repoPrefix string, files map[string]string) {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	if err := tw.WriteHeader(&tar.Header{Name: repoPrefix + "/", Typeflag: tar.TypeDir, Mode: 0o755}); err != nil {
		t.Fatal(err)
	}
	for p, body := range files {
		if err := tw.WriteHeader(&tar.Header{
			Name: repoPrefix + "/" + p, Typeflag: tar.TypeReg,
			Mode: 0o644, Size: int64(len(body)),
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}

	dir := filepath.Join(cacheDir, "tars")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	fname := owner + "-" + name + "-" + sha + ".tar.gz"
	if err := os.WriteFile(filepath.Join(dir, fname), buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

// expectedDirSHA extracts a copy of the skill tree to a temp dir and computes
// its dir_sha so the test's registry entry matches exactly what the engine
// will compute on its own.
func expectedDirSHA(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for p, body := range files {
		full := filepath.Join(dir, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	sha, err := registry.DirSHA(dir)
	if err != nil {
		t.Fatal(err)
	}
	return sha
}

func TestEngine_InstallReplaceSkipForce(t *testing.T) {
	root := t.TempDir()
	cacheDir := filepath.Join(root, "cache")
	installRoot := filepath.Join(root, "home", ".claude", "skills")
	manifestPath := filepath.Join(root, "manifest.json")

	// Skill files (relative to repo root inside the tarball).
	skillFiles := map[string]string{
		"skills/foo/SKILL.md":      "# foo\n",
		"skills/foo/data/notes.md": "hello\n",
	}
	onlyFoo := map[string]string{
		"SKILL.md":      "# foo\n",
		"data/notes.md": "hello\n",
	}
	dirSHA := expectedDirSHA(t, onlyFoo)

	owner, name, sha := "example", "repo", "deadbeef000000"
	seedTarball(t, cacheDir, owner, name, sha, owner+"-"+name+"-abc1234", skillFiles)

	reg := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/example/repo", SHA: sha},
		Skills: []registry.Skill{{
			Name: "foo", Version: "0.1.0", Path: "skills/foo",
			Platforms: []string{"test"}, DirSHA: dirSHA,
		}},
	}

	adapter := platform.Adapter{
		Name:           "test",
		InstallTargets: map[string]string{"user": installRoot},
		DefaultScope:   "user",
	}

	engine := NewEngine(cacheDir, manifestPath)
	engine.Now = func() time.Time { return time.Unix(1700000000, 0).UTC() }

	plan, err := Plan(reg, "foo")
	if err != nil {
		t.Fatal(err)
	}
	res, err := engine.Execute(reg, plan, ExecuteOpts{
		Adapters:  []platform.Adapter{adapter},
		Platforms: []string{"test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Results) != 1 || res.Results[0].Outcome != OutcomeInstalled {
		t.Fatalf("first run: %+v", res.Results)
	}
	if _, err := os.Stat(filepath.Join(installRoot, "foo", "SKILL.md")); err != nil {
		t.Fatalf("skill not placed: %v", err)
	}

	// Second run without --force: skipped.
	res2, err := engine.Execute(reg, plan, ExecuteOpts{
		Adapters:  []platform.Adapter{adapter},
		Platforms: []string{"test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res2.Results[0].Outcome != OutcomeSkipped {
		t.Errorf("expected skipped, got %s", res2.Results[0].Outcome)
	}

	// Third run with --force: forced.
	res3, err := engine.Execute(reg, plan, ExecuteOpts{
		Adapters:  []platform.Adapter{adapter},
		Platforms: []string{"test"},
		Force:     true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res3.Results[0].Outcome != OutcomeForced {
		t.Errorf("expected forced, got %s", res3.Results[0].Outcome)
	}

	m, err := manifest.Load(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	inst := m.FindOne("foo", "test", "user")
	if inst == nil || inst.Version != "0.1.0" || inst.SourceSHA != sha {
		t.Errorf("manifest wrong: %+v", inst)
	}

	// Uninstall wipes the dir and the manifest entry.
	out, err := engine.Uninstall("foo")
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("uninstall results: %+v", out)
	}
	if _, err := os.Stat(filepath.Join(installRoot, "foo")); err == nil {
		t.Error("skill dir should be gone")
	}
	m, _ = manifest.Load(manifestPath)
	if m.Find("foo") != nil {
		t.Error("manifest should be empty")
	}
}

func TestEngine_DirSHAMismatchFails(t *testing.T) {
	root := t.TempDir()
	cacheDir := filepath.Join(root, "cache")
	installRoot := filepath.Join(root, "home")
	manifestPath := filepath.Join(root, "manifest.json")

	owner, name, sha := "example", "repo", "cafebabe000000"
	seedTarball(t, cacheDir, owner, name, sha, "example-repo-abc1234", map[string]string{
		"skills/foo/SKILL.md": "content\n",
	})

	reg := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/example/repo", SHA: sha},
		Skills: []registry.Skill{{
			Name: "foo", Version: "0.1.0", Path: "skills/foo",
			Platforms: []string{"test"},
			DirSHA:    "0000000000000000000000000000000000000000000000000000000000000000",
		}},
	}
	adapter := platform.Adapter{
		Name:           "test",
		InstallTargets: map[string]string{"user": installRoot},
		DefaultScope:   "user",
	}

	engine := NewEngine(cacheDir, manifestPath)
	plan, _ := Plan(reg, "foo")
	_, err := engine.Execute(reg, plan, ExecuteOpts{
		Adapters:  []platform.Adapter{adapter},
		Platforms: []string{"test"},
	})
	if err == nil {
		t.Fatal("expected dir_sha mismatch error")
	}
}
