package install

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jjfantini/humblSKILLS/cli/internal/adapters"
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
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

	adapter := adapters.Adapter{
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
		Adapters:  []adapters.Adapter{adapter},
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
		Adapters:  []adapters.Adapter{adapter},
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
		Adapters:  []adapters.Adapter{adapter},
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

func TestEngine_ProjectScopeMovesOldInstall(t *testing.T) {
	root := t.TempDir()
	cacheDir := filepath.Join(root, "cache")
	oldRoot := filepath.Join(root, "old-project", ".claude", "skills")
	newRoot := filepath.Join(root, "new-project", ".claude", "skills")
	manifestPath := filepath.Join(root, "manifest.json")

	skillFiles := map[string]string{
		"skills/foo/SKILL.md": "# foo\n",
	}
	onlyFoo := map[string]string{"SKILL.md": "# foo\n"}
	dirSHA := expectedDirSHA(t, onlyFoo)

	owner, name, sha := "example", "repo", "abcd00000000"
	seedTarball(t, cacheDir, owner, name, sha, owner+"-"+name+"-abc1234", skillFiles)

	reg := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/example/repo", SHA: sha},
		Skills: []registry.Skill{{
			Name: "foo", Version: "0.1.0", Path: "skills/foo",
			Platforms: []string{"test"}, DirSHA: dirSHA,
		}},
	}

	// Pre-seed manifest and on-disk content at an old project location.
	oldPath := filepath.Join(oldRoot, "foo")
	if err := os.MkdirAll(oldPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldPath, "SKILL.md"), []byte("# foo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := &manifest.Manifest{SchemaVersion: manifest.SchemaVersion}
	m.Upsert(manifest.Installation{
		Skill: "foo", Version: "0.1.0", Platform: "test", Scope: "project",
		Path: oldPath, InstalledAt: time.Unix(1700000000, 0).UTC(),
		SourceSHA: sha, RegistryRef: dirSHA,
	})
	if err := manifest.Save(manifestPath, m); err != nil {
		t.Fatal(err)
	}

	// Adapter now resolves project-scope to a DIFFERENT path (simulates
	// running from a new CWD).
	adapter := adapters.Adapter{
		Name:           "test",
		InstallTargets: map[string]string{"project": newRoot},
		DefaultScope:   "project",
	}

	engine := NewEngine(cacheDir, manifestPath)
	engine.Now = func() time.Time { return time.Unix(1700000000, 0).UTC() }
	plan, err := Plan(reg, "foo")
	if err != nil {
		t.Fatal(err)
	}
	res, err := engine.Execute(reg, plan, ExecuteOpts{
		Adapters:  []adapters.Adapter{adapter},
		Platforms: []string{"test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Results) != 1 {
		t.Fatalf("expected 1 result, got %+v", res.Results)
	}
	// Old path should be gone.
	if _, err := os.Stat(oldPath); err == nil {
		t.Error("old install path should have been removed")
	}
	// New path should contain the skill.
	if _, err := os.Stat(filepath.Join(newRoot, "foo", "SKILL.md")); err != nil {
		t.Errorf("new install path not written: %v", err)
	}
	// Manifest entry should point at the new path.
	m2, err := manifest.Load(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	inst := m2.FindOne("foo", "test", "project")
	if inst == nil || inst.Path != filepath.Join(newRoot, "foo") {
		t.Errorf("manifest path not migrated: %+v", inst)
	}
}

func TestInstall_PreserveFreshInstall_SeedsFromStaging(t *testing.T) {
	root := t.TempDir()
	cacheDir := filepath.Join(root, "cache")
	installRoot := filepath.Join(root, "home", ".claude", "skills")
	manifestPath := filepath.Join(root, "manifest.json")

	skillFiles := map[string]string{
		"skills/foo/SKILL.md":       "# foo\n",
		"skills/foo/wiki/seed.md":   "seed-v1\n",
		"skills/foo/log.md":         "initial\n",
	}
	onlyFoo := map[string]string{
		"SKILL.md":     "# foo\n",
		"wiki/seed.md": "seed-v1\n",
		"log.md":       "initial\n",
	}
	dirSHA := expectedDirSHA(t, onlyFoo)
	owner, name, sha := "ex", "r", "sha1fresh000000"
	seedTarball(t, cacheDir, owner, name, sha, owner+"-"+name+"-abc", skillFiles)

	reg := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/ex/r", SHA: sha},
		Skills: []registry.Skill{{
			Name: "foo", Version: "0.1.0", Path: "skills/foo",
			Platforms: []string{"test"}, DirSHA: dirSHA,
			Preserve: []string{"log.md", "wiki/"},
		}},
	}
	adapter := adapters.Adapter{
		Name:           "test",
		InstallTargets: map[string]string{"user": installRoot},
		DefaultScope:   "user",
	}

	engine := NewEngine(cacheDir, manifestPath)
	engine.Now = func() time.Time { return time.Unix(1700000000, 0).UTC() }
	plan, _ := Plan(reg, "foo")
	if _, err := engine.Execute(reg, plan, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"},
	}); err != nil {
		t.Fatal(err)
	}

	for _, rel := range []string{"SKILL.md", "wiki/seed.md", "log.md"} {
		if _, err := os.Stat(filepath.Join(installRoot, "foo", rel)); err != nil {
			t.Errorf("%s missing after fresh install: %v", rel, err)
		}
	}
}

func TestInstall_PreserveFile_UserWinsOnReplace(t *testing.T) {
	root := t.TempDir()
	cacheDir := filepath.Join(root, "cache")
	installRoot := filepath.Join(root, "home", ".claude", "skills")
	manifestPath := filepath.Join(root, "manifest.json")

	// v1 ships log.md with "initial".
	v1Files := map[string]string{
		"skills/foo/SKILL.md": "# foo\n",
		"skills/foo/log.md":   "initial\n",
	}
	v1Flat := map[string]string{"SKILL.md": "# foo\n", "log.md": "initial\n"}
	v1SHA := expectedDirSHA(t, v1Flat)
	owner, name := "ex", "r"
	src1 := "sha1aaaaaaaaaa"
	seedTarball(t, cacheDir, owner, name, src1, owner+"-"+name+"-abc", v1Files)

	// v2 ships log.md with "shipped-v2".
	v2Files := map[string]string{
		"skills/foo/SKILL.md": "# foo v2\n",
		"skills/foo/log.md":   "shipped-v2\n",
	}
	v2Flat := map[string]string{"SKILL.md": "# foo v2\n", "log.md": "shipped-v2\n"}
	v2SHA := expectedDirSHA(t, v2Flat)
	src2 := "sha2bbbbbbbbbb"
	seedTarball(t, cacheDir, owner, name, src2, owner+"-"+name+"-def", v2Files)

	adapter := adapters.Adapter{
		Name:           "test",
		InstallTargets: map[string]string{"user": installRoot},
		DefaultScope:   "user",
	}
	engine := NewEngine(cacheDir, manifestPath)
	engine.Now = func() time.Time { return time.Unix(1700000000, 0).UTC() }

	// Install v1.
	reg1 := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/ex/r", SHA: src1},
		Skills: []registry.Skill{{
			Name: "foo", Version: "0.1.0", Path: "skills/foo",
			Platforms: []string{"test"}, DirSHA: v1SHA,
			Preserve: []string{"log.md"},
		}},
	}
	plan1, _ := Plan(reg1, "foo")
	if _, err := engine.Execute(reg1, plan1, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"},
	}); err != nil {
		t.Fatal(err)
	}

	// User edits log.md.
	logPath := filepath.Join(installRoot, "foo", "log.md")
	if err := os.WriteFile(logPath, []byte("user-edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Install v2.
	reg2 := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/ex/r", SHA: src2},
		Skills: []registry.Skill{{
			Name: "foo", Version: "0.2.0", Path: "skills/foo",
			Platforms: []string{"test"}, DirSHA: v2SHA,
			Preserve: []string{"log.md"},
		}},
	}
	plan2, _ := Plan(reg2, "foo")
	if _, err := engine.Execute(reg2, plan2, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"},
	}); err != nil {
		t.Fatal(err)
	}

	// log.md should have user bytes; SKILL.md should be v2 bytes.
	got, _ := os.ReadFile(logPath)
	if string(got) != "user-edit\n" {
		t.Errorf("log.md: want user-edit, got %q", got)
	}
	gotSkill, _ := os.ReadFile(filepath.Join(installRoot, "foo", "SKILL.md"))
	if string(gotSkill) != "# foo v2\n" {
		t.Errorf("SKILL.md: want v2, got %q", gotSkill)
	}
}

func TestInstall_PreserveDir_DeepMerge(t *testing.T) {
	root := t.TempDir()
	cacheDir := filepath.Join(root, "cache")
	installRoot := filepath.Join(root, "home", ".claude", "skills")
	manifestPath := filepath.Join(root, "manifest.json")

	v1Files := map[string]string{
		"skills/foo/SKILL.md":     "# foo\n",
		"skills/foo/wiki/seed.md": "seed-v1\n",
	}
	v1Flat := map[string]string{"SKILL.md": "# foo\n", "wiki/seed.md": "seed-v1\n"}
	v1SHA := expectedDirSHA(t, v1Flat)
	owner, name := "ex", "r"
	src1 := "shadir1aaaaaaa"
	seedTarball(t, cacheDir, owner, name, src1, owner+"-"+name+"-abc", v1Files)

	v2Files := map[string]string{
		"skills/foo/SKILL.md":     "# foo v2\n",
		"skills/foo/wiki/seed.md": "seed-v2\n",
	}
	v2Flat := map[string]string{"SKILL.md": "# foo v2\n", "wiki/seed.md": "seed-v2\n"}
	v2SHA := expectedDirSHA(t, v2Flat)
	src2 := "shadir2bbbbbbb"
	seedTarball(t, cacheDir, owner, name, src2, owner+"-"+name+"-def", v2Files)

	adapter := adapters.Adapter{
		Name:           "test",
		InstallTargets: map[string]string{"user": installRoot},
		DefaultScope:   "user",
	}
	engine := NewEngine(cacheDir, manifestPath)
	engine.Now = func() time.Time { return time.Unix(1700000000, 0).UTC() }

	reg1 := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/ex/r", SHA: src1},
		Skills: []registry.Skill{{
			Name: "foo", Version: "0.1.0", Path: "skills/foo",
			Platforms: []string{"test"}, DirSHA: v1SHA,
			Preserve: []string{"wiki/"},
		}},
	}
	plan1, _ := Plan(reg1, "foo")
	if _, err := engine.Execute(reg1, plan1, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"},
	}); err != nil {
		t.Fatal(err)
	}

	// User adds a new file and edits the seeded one.
	wikiDir := filepath.Join(installRoot, "foo", "wiki")
	if err := os.WriteFile(filepath.Join(wikiDir, "mynote.md"), []byte("my-note\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wikiDir, "seed.md"), []byte("user-edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	reg2 := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/ex/r", SHA: src2},
		Skills: []registry.Skill{{
			Name: "foo", Version: "0.2.0", Path: "skills/foo",
			Platforms: []string{"test"}, DirSHA: v2SHA,
			Preserve: []string{"wiki/"},
		}},
	}
	plan2, _ := Plan(reg2, "foo")
	if _, err := engine.Execute(reg2, plan2, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"},
	}); err != nil {
		t.Fatal(err)
	}

	// seed.md should now be staging's v2 bytes (staging wins on conflict).
	got, _ := os.ReadFile(filepath.Join(wikiDir, "seed.md"))
	if string(got) != "seed-v2\n" {
		t.Errorf("seed.md: want seed-v2, got %q", got)
	}
	// mynote.md should survive.
	got2, err := os.ReadFile(filepath.Join(wikiDir, "mynote.md"))
	if err != nil || string(got2) != "my-note\n" {
		t.Errorf("mynote.md: want my-note, got %q err=%v", got2, err)
	}
}

func TestInstall_PreserveWithForce(t *testing.T) {
	root := t.TempDir()
	cacheDir := filepath.Join(root, "cache")
	installRoot := filepath.Join(root, "home", ".claude", "skills")
	manifestPath := filepath.Join(root, "manifest.json")

	files := map[string]string{
		"skills/foo/SKILL.md": "# foo\n",
		"skills/foo/log.md":   "initial\n",
	}
	flat := map[string]string{"SKILL.md": "# foo\n", "log.md": "initial\n"}
	dirSHA := expectedDirSHA(t, flat)
	owner, name, sha := "ex", "r", "shaforce000000"
	seedTarball(t, cacheDir, owner, name, sha, owner+"-"+name+"-abc", files)

	reg := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/ex/r", SHA: sha},
		Skills: []registry.Skill{{
			Name: "foo", Version: "0.1.0", Path: "skills/foo",
			Platforms: []string{"test"}, DirSHA: dirSHA,
			Preserve: []string{"log.md"},
		}},
	}
	adapter := adapters.Adapter{
		Name:           "test",
		InstallTargets: map[string]string{"user": installRoot},
		DefaultScope:   "user",
	}
	engine := NewEngine(cacheDir, manifestPath)
	engine.Now = func() time.Time { return time.Unix(1700000000, 0).UTC() }

	plan, _ := Plan(reg, "foo")
	if _, err := engine.Execute(reg, plan, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"},
	}); err != nil {
		t.Fatal(err)
	}

	logPath := filepath.Join(installRoot, "foo", "log.md")
	if err := os.WriteFile(logPath, []byte("user-edit\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := engine.Execute(reg, plan, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"}, Force: true,
	}); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(logPath)
	if string(got) != "user-edit\n" {
		t.Errorf("forced install wiped preserved file: got %q", got)
	}
}

func TestInstall_PreserveScopeMove(t *testing.T) {
	root := t.TempDir()
	cacheDir := filepath.Join(root, "cache")
	oldRoot := filepath.Join(root, "old", ".claude", "skills")
	newRoot := filepath.Join(root, "new", ".claude", "skills")
	manifestPath := filepath.Join(root, "manifest.json")

	files := map[string]string{
		"skills/foo/SKILL.md":     "# foo\n",
		"skills/foo/wiki/seed.md": "seed\n",
	}
	flat := map[string]string{"SKILL.md": "# foo\n", "wiki/seed.md": "seed\n"}
	dirSHA := expectedDirSHA(t, flat)
	owner, name, sha := "ex", "r", "shamove000000"
	seedTarball(t, cacheDir, owner, name, sha, owner+"-"+name+"-abc", files)

	// Pre-seed old install with user content in preserve dir.
	oldFoo := filepath.Join(oldRoot, "foo")
	if err := os.MkdirAll(filepath.Join(oldFoo, "wiki"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldFoo, "SKILL.md"), []byte("# foo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldFoo, "wiki", "seed.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldFoo, "wiki", "mynote.md"), []byte("my-note\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := &manifest.Manifest{SchemaVersion: manifest.SchemaVersion}
	m.Upsert(manifest.Installation{
		Skill: "foo", Version: "0.1.0", Platform: "test", Scope: "project",
		Path: oldFoo, InstalledAt: time.Unix(1700000000, 0).UTC(),
		SourceSHA: sha, RegistryRef: dirSHA,
	})
	if err := manifest.Save(manifestPath, m); err != nil {
		t.Fatal(err)
	}

	reg := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/ex/r", SHA: sha},
		Skills: []registry.Skill{{
			Name: "foo", Version: "0.1.0", Path: "skills/foo",
			Platforms: []string{"test"}, DirSHA: dirSHA,
			Preserve: []string{"wiki/"},
		}},
	}
	adapter := adapters.Adapter{
		Name:           "test",
		InstallTargets: map[string]string{"project": newRoot},
		DefaultScope:   "project",
	}
	engine := NewEngine(cacheDir, manifestPath)
	engine.Now = func() time.Time { return time.Unix(1700000000, 0).UTC() }
	plan, _ := Plan(reg, "foo")
	if _, err := engine.Execute(reg, plan, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"},
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(oldFoo); err == nil {
		t.Error("old path should be gone")
	}
	// User content migrated to new location.
	got, err := os.ReadFile(filepath.Join(newRoot, "foo", "wiki", "mynote.md"))
	if err != nil || string(got) != "my-note\n" {
		t.Errorf("migrated user note missing: err=%v got=%q", err, got)
	}
}

func TestInstall_PreserveMissingOnDisk_UsesStaging(t *testing.T) {
	root := t.TempDir()
	cacheDir := filepath.Join(root, "cache")
	installRoot := filepath.Join(root, "home", ".claude", "skills")
	manifestPath := filepath.Join(root, "manifest.json")

	v1Files := map[string]string{
		"skills/foo/SKILL.md": "# foo\n",
		"skills/foo/log.md":   "initial\n",
	}
	v1Flat := map[string]string{"SKILL.md": "# foo\n", "log.md": "initial\n"}
	v1SHA := expectedDirSHA(t, v1Flat)
	owner, name := "ex", "r"
	src1 := "shamiss1aaaaa"
	seedTarball(t, cacheDir, owner, name, src1, owner+"-"+name+"-abc", v1Files)

	v2Files := map[string]string{
		"skills/foo/SKILL.md": "# foo v2\n",
		"skills/foo/log.md":   "shipped-v2\n",
	}
	v2Flat := map[string]string{"SKILL.md": "# foo v2\n", "log.md": "shipped-v2\n"}
	v2SHA := expectedDirSHA(t, v2Flat)
	src2 := "shamiss2bbbbb"
	seedTarball(t, cacheDir, owner, name, src2, owner+"-"+name+"-def", v2Files)

	adapter := adapters.Adapter{
		Name:           "test",
		InstallTargets: map[string]string{"user": installRoot},
		DefaultScope:   "user",
	}
	engine := NewEngine(cacheDir, manifestPath)
	engine.Now = func() time.Time { return time.Unix(1700000000, 0).UTC() }

	reg1 := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/ex/r", SHA: src1},
		Skills: []registry.Skill{{
			Name: "foo", Version: "0.1.0", Path: "skills/foo",
			Platforms: []string{"test"}, DirSHA: v1SHA,
			Preserve: []string{"log.md"},
		}},
	}
	plan1, _ := Plan(reg1, "foo")
	if _, err := engine.Execute(reg1, plan1, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"},
	}); err != nil {
		t.Fatal(err)
	}

	// Remove the preserved file from disk.
	if err := os.Remove(filepath.Join(installRoot, "foo", "log.md")); err != nil {
		t.Fatal(err)
	}

	reg2 := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/ex/r", SHA: src2},
		Skills: []registry.Skill{{
			Name: "foo", Version: "0.2.0", Path: "skills/foo",
			Platforms: []string{"test"}, DirSHA: v2SHA,
			Preserve: []string{"log.md"},
		}},
	}
	plan2, _ := Plan(reg2, "foo")
	if _, err := engine.Execute(reg2, plan2, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"},
	}); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(installRoot, "foo", "log.md"))
	if string(got) != "shipped-v2\n" {
		t.Errorf("missing preserve should have landed staging seed: got %q", got)
	}
}

func TestInstall_PreserveTypeMismatch_Errors(t *testing.T) {
	root := t.TempDir()
	cacheDir := filepath.Join(root, "cache")
	installRoot := filepath.Join(root, "home", ".claude", "skills")
	manifestPath := filepath.Join(root, "manifest.json")

	// v1 doesn't ship note.md at all; user will create it later as a dir.
	v1Files := map[string]string{"skills/foo/SKILL.md": "# foo\n"}
	v1Flat := map[string]string{"SKILL.md": "# foo\n"}
	v1SHA := expectedDirSHA(t, v1Flat)
	owner, name := "ex", "r"
	src1 := "shatyp1aaaaaa"
	seedTarball(t, cacheDir, owner, name, src1, owner+"-"+name+"-abc", v1Files)

	v2Files := map[string]string{
		"skills/foo/SKILL.md": "# foo v2\n",
		"skills/foo/note.md":  "file-v2\n",
	}
	v2Flat := map[string]string{"SKILL.md": "# foo v2\n", "note.md": "file-v2\n"}
	v2SHA := expectedDirSHA(t, v2Flat)
	src2 := "shatyp2bbbbbb"
	seedTarball(t, cacheDir, owner, name, src2, owner+"-"+name+"-def", v2Files)

	adapter := adapters.Adapter{
		Name:           "test",
		InstallTargets: map[string]string{"user": installRoot},
		DefaultScope:   "user",
	}
	engine := NewEngine(cacheDir, manifestPath)
	engine.Now = func() time.Time { return time.Unix(1700000000, 0).UTC() }

	reg1 := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/ex/r", SHA: src1},
		Skills: []registry.Skill{{
			Name: "foo", Version: "0.1.0", Path: "skills/foo",
			Platforms: []string{"test"}, DirSHA: v1SHA,
			Preserve: []string{"note.md"},
		}},
	}
	plan1, _ := Plan(reg1, "foo")
	if _, err := engine.Execute(reg1, plan1, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"},
	}); err != nil {
		t.Fatal(err)
	}

	// User creates note.md as a DIRECTORY.
	if err := os.Mkdir(filepath.Join(installRoot, "foo", "note.md"), 0o755); err != nil {
		t.Fatal(err)
	}

	reg2 := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/ex/r", SHA: src2},
		Skills: []registry.Skill{{
			Name: "foo", Version: "0.2.0", Path: "skills/foo",
			Platforms: []string{"test"}, DirSHA: v2SHA,
			Preserve: []string{"note.md"}, // declared as file
		}},
	}
	plan2, _ := Plan(reg2, "foo")
	_, err := engine.Execute(reg2, plan2, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"},
	})
	if err == nil {
		t.Fatal("expected type-mismatch error")
	}
	// User dir must still exist — no destructive cleanup.
	if _, serr := os.Stat(filepath.Join(installRoot, "foo", "note.md")); serr != nil {
		t.Errorf("user dir should survive failed update: %v", serr)
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
	adapter := adapters.Adapter{
		Name:           "test",
		InstallTargets: map[string]string{"user": installRoot},
		DefaultScope:   "user",
	}

	engine := NewEngine(cacheDir, manifestPath)
	plan, _ := Plan(reg, "foo")
	_, err := engine.Execute(reg, plan, ExecuteOpts{
		Adapters:  []adapters.Adapter{adapter},
		Platforms: []string{"test"},
	})
	if err == nil {
		t.Fatal("expected dir_sha mismatch error")
	}
}
