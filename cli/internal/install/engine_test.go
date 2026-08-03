package install

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jjfantini/humblSKILLS/cli/internal/adapters"
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
)

// seedTarball writes a fake GitHub-style tarball into cacheDir in the exact
// location the Fetcher expects, so Execute's Fetch call becomes a cache hit.
// seedTarball writes the fixture tarball and returns its path, so a test can
// delete it to prove a later run does not fetch.
func seedTarball(t *testing.T, cacheDir, owner, name, sha, repoPrefix string, files map[string]string) string {
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
	path := filepath.Join(dir, fname)
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
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

func withCwd(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prev)
	})
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

	engine := newTestEngine(t, root, cacheDir, manifestPath)
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

// TestEngine_SkipOnStaleSourceSHA is the engine-side counterpart to
// TestPlanUpdates_StaleSourceSHAIsNotDrift: when the humblSKILLS repo
// SHA advances but the skill's Version and DirSHA are unchanged, a
// re-run must Skip rather than Replace. Previously Source.SHA was part
// of the up-to-date check, so every install was marked as replaced (and
// shown as drifted in the dashboard) after each CLI release.
func TestEngine_SkipOnStaleSourceSHA(t *testing.T) {
	root := t.TempDir()
	cacheDir := filepath.Join(root, "cache")
	installRoot := filepath.Join(root, "home", ".claude", "skills")
	manifestPath := filepath.Join(root, "manifest.json")

	skillFiles := map[string]string{"skills/foo/SKILL.md": "# foo\n"}
	onlyFoo := map[string]string{"SKILL.md": "# foo\n"}
	dirSHA := expectedDirSHA(t, onlyFoo)

	owner, name := "example", "repo"
	sha1 := "sha1aaaaaaaaaaaa"
	sha2 := "sha2bbbbbbbbbbbb"
	// Both tarballs ship the exact same skill tree — only the repo SHA
	// differs, simulating an unrelated commit (doc, other-skill, CI).
	seedTarball(t, cacheDir, owner, name, sha1, owner+"-"+name+"-abc1234", skillFiles)
	seedTarball(t, cacheDir, owner, name, sha2, owner+"-"+name+"-def5678", skillFiles)

	mkReg := func(sha string) *registry.Registry {
		return &registry.Registry{
			SchemaVersion: registry.SchemaVersion,
			Source:        registry.Source{Repo: "github.com/example/repo", SHA: sha},
			Skills: []registry.Skill{{
				Name: "foo", Version: "0.1.0", Path: "skills/foo",
				Platforms: []string{"test"}, DirSHA: dirSHA,
			}},
		}
	}

	adapter := adapters.Adapter{
		Name:           "test",
		InstallTargets: map[string]string{"user": installRoot},
		DefaultScope:   "user",
	}
	engine := newTestEngine(t, root, cacheDir, manifestPath)
	engine.Now = func() time.Time { return time.Unix(1700000000, 0).UTC() }

	// Install under sha1.
	plan1, _ := Plan(mkReg(sha1), "foo")
	if _, err := engine.Execute(mkReg(sha1), plan1, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"},
	}); err != nil {
		t.Fatal(err)
	}

	// Re-run under sha2 — same version + dir_sha, new repo SHA. Must skip.
	plan2, _ := Plan(mkReg(sha2), "foo")
	res, err := engine.Execute(mkReg(sha2), plan2, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Results) != 1 || res.Results[0].Outcome != OutcomeSkipped {
		t.Fatalf("expected skipped after source_sha-only change, got %+v", res.Results)
	}
}

func TestEngine_ProjectScopeMovesOldInstall(t *testing.T) {
	root := t.TempDir()
	withCwd(t, filepath.Join(root, "new-project"))
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

	engine := newTestEngine(t, root, cacheDir, manifestPath)
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
		"skills/foo/SKILL.md":     "# foo\n",
		"skills/foo/wiki/seed.md": "seed-v1\n",
		"skills/foo/log.md":       "initial\n",
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

	engine := newTestEngine(t, root, cacheDir, manifestPath)
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
	engine := newTestEngine(t, root, cacheDir, manifestPath)
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
	engine := newTestEngine(t, root, cacheDir, manifestPath)
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

// TestInstall_ForceBypassesPreserve pins the new semantics: --force gives the
// caller a clean upstream install. Preserve is NOT applied, so any user edits
// in preserved paths are overwritten. This is the documented escape hatch for
// users who want to drop their local customizations without running uninstall.
func TestInstall_ForceBypassesPreserve(t *testing.T) {
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
	engine := newTestEngine(t, root, cacheDir, manifestPath)
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
	if string(got) != "initial\n" {
		t.Errorf("force should have restored upstream bytes; got %q", got)
	}
}

// skillMD is a tiny helper for writing test SKILL.md bodies with valid
// frontmatter. The tests exercise the locally-edited preserve behavior, so
// the files they ship/install need real YAML up top for Parse to succeed.
func skillMD(name, version string, preserve []string) string {
	sb := strings.Builder{}
	sb.WriteString("---\n")
	sb.WriteString("name: " + name + "\n")
	sb.WriteString("description: test skill\n")
	sb.WriteString("version: " + version + "\n")
	if len(preserve) > 0 {
		sb.WriteString("preserve:\n")
		for _, p := range preserve {
			sb.WriteString("  - " + p + "\n")
		}
	} else {
		sb.WriteString("preserve: []\n")
	}
	sb.WriteString("---\n\n# " + name + "\n")
	return sb.String()
}

// TestUpdate_UsesLocalPreserveList: the user added a new entry to their
// installed SKILL.md's preserve list. On update that user-only entry must
// survive even though the registry doesn't list it.
func TestUpdate_UsesLocalPreserveList(t *testing.T) {
	root := t.TempDir()
	cacheDir := filepath.Join(root, "cache")
	installRoot := filepath.Join(root, "home", ".claude", "skills")
	manifestPath := filepath.Join(root, "manifest.json")

	v1Body := skillMD("foo", "0.1.0", []string{"log.md"})
	v1Files := map[string]string{
		"skills/foo/SKILL.md": v1Body,
		"skills/foo/log.md":   "initial\n",
	}
	v1Flat := map[string]string{"SKILL.md": v1Body, "log.md": "initial\n"}
	v1SHA := expectedDirSHA(t, v1Flat)
	src1 := "shauser1aaaaa"
	seedTarball(t, cacheDir, "ex", "r", src1, "ex-r-abc", v1Files)

	v2Body := skillMD("foo", "0.2.0", []string{"log.md"})
	v2Files := map[string]string{
		"skills/foo/SKILL.md": v2Body,
		"skills/foo/log.md":   "shipped-v2\n",
	}
	v2Flat := map[string]string{"SKILL.md": v2Body, "log.md": "shipped-v2\n"}
	v2SHA := expectedDirSHA(t, v2Flat)
	src2 := "shauser2bbbbb"
	seedTarball(t, cacheDir, "ex", "r", src2, "ex-r-def", v2Files)

	adapter := adapters.Adapter{
		Name:           "test",
		InstallTargets: map[string]string{"user": installRoot},
		DefaultScope:   "user",
	}
	engine := newTestEngine(t, root, cacheDir, manifestPath)
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

	// User extends their preserve list locally and drops a file.
	userBody := skillMD("foo", "0.1.0", []string{"log.md", "notes.md"})
	if err := os.WriteFile(filepath.Join(installRoot, "foo", "SKILL.md"), []byte(userBody), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(installRoot, "foo", "notes.md"), []byte("user-notes\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(installRoot, "foo", "log.md"), []byte("user-log\n"), 0o644); err != nil {
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
	r, err := engine.Execute(reg2, plan2, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Warnings) != 0 {
		t.Errorf("unexpected warnings: %+v", r.Warnings)
	}

	if got, _ := os.ReadFile(filepath.Join(installRoot, "foo", "log.md")); string(got) != "user-log\n" {
		t.Errorf("log.md: want user-log, got %q", got)
	}
	if got, err := os.ReadFile(filepath.Join(installRoot, "foo", "notes.md")); err != nil || string(got) != "user-notes\n" {
		t.Errorf("notes.md: want user-notes, got %q err=%v", got, err)
	}

	// The rewritten SKILL.md must keep the user's preserve list AND ship
	// upstream's version/body bump. This is the merge-preserve-key contract.
	finalSKILL, err := os.ReadFile(filepath.Join(installRoot, "foo", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(finalSKILL), "- notes.md") {
		t.Errorf("SKILL.md should carry the user's notes.md preserve entry; got:\n%s", finalSKILL)
	}
	if !strings.Contains(string(finalSKILL), "version: 0.2.0") {
		t.Errorf("SKILL.md should have v2 version from upstream; got:\n%s", finalSKILL)
	}
}

// TestUpdate_MergesPreserveKeyKeepsUpstreamBody: verify the merge-preserve
// contract - every non-preserve frontmatter key and the markdown body flow
// from upstream on update, while the user's preserve list rides along.
func TestUpdate_MergesPreserveKeyKeepsUpstreamBody(t *testing.T) {
	root := t.TempDir()
	cacheDir := filepath.Join(root, "cache")
	installRoot := filepath.Join(root, "home", ".claude", "skills")
	manifestPath := filepath.Join(root, "manifest.json")

	v1Body := skillMD("foo", "0.1.0", []string{"log.md"})
	v1Files := map[string]string{
		"skills/foo/SKILL.md": v1Body,
		"skills/foo/log.md":   "initial\n",
	}
	v1Flat := map[string]string{"SKILL.md": v1Body, "log.md": "initial\n"}
	v1SHA := expectedDirSHA(t, v1Flat)
	src1 := "shamrg1aaaaa"
	seedTarball(t, cacheDir, "ex", "r", src1, "ex-r-abc", v1Files)

	// v2 SKILL.md has a richer body so we can prove it flows through.
	v2Body := "---\nname: foo\ndescription: v2 rewrite\nversion: 0.2.0\npreserve:\n  - log.md\n---\n\n# foo v2\n\nBrand new body with a warning.\n"
	v2Files := map[string]string{
		"skills/foo/SKILL.md": v2Body,
		"skills/foo/log.md":   "shipped-v2\n",
	}
	v2Flat := map[string]string{"SKILL.md": v2Body, "log.md": "shipped-v2\n"}
	v2SHA := expectedDirSHA(t, v2Flat)
	src2 := "shamrg2bbbbb"
	seedTarball(t, cacheDir, "ex", "r", src2, "ex-r-def", v2Files)

	adapter := adapters.Adapter{
		Name:           "test",
		InstallTargets: map[string]string{"user": installRoot},
		DefaultScope:   "user",
	}
	engine := newTestEngine(t, root, cacheDir, manifestPath)
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

	// User extends preserve with a custom entry.
	userBody := skillMD("foo", "0.1.0", []string{"log.md", "decisions.md"})
	if err := os.WriteFile(filepath.Join(installRoot, "foo", "SKILL.md"), []byte(userBody), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(installRoot, "foo", "decisions.md"), []byte("d1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	reg2 := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/ex/r", SHA: src2},
		Skills: []registry.Skill{{
			Name: "foo", Version: "0.2.0", Path: "skills/foo",
			Platforms: []string{"test"}, DirSHA: v2SHA,
			Preserve: []string{"log.md"}, // upstream never listed decisions.md
		}},
	}
	plan2, _ := Plan(reg2, "foo")
	if _, err := engine.Execute(reg2, plan2, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"},
	}); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(installRoot, "foo", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	// Upstream bits: description bumped, version bumped, body copied.
	if !strings.Contains(s, "description: v2 rewrite") {
		t.Errorf("SKILL.md missing v2 description:\n%s", s)
	}
	if !strings.Contains(s, "version: 0.2.0") {
		t.Errorf("SKILL.md missing v2 version:\n%s", s)
	}
	if !strings.Contains(s, "Brand new body with a warning.") {
		t.Errorf("SKILL.md body not refreshed from upstream:\n%s", s)
	}
	// User bit: their custom preserve entry survived.
	if !strings.Contains(s, "- decisions.md") {
		t.Errorf("SKILL.md lost user's decisions.md preserve entry:\n%s", s)
	}
	if !strings.Contains(s, "- log.md") {
		t.Errorf("SKILL.md lost log.md preserve entry:\n%s", s)
	}
	// And the user-only file should still exist too.
	if _, err := os.Stat(filepath.Join(installRoot, "foo", "decisions.md")); err != nil {
		t.Errorf("decisions.md user file missing after update: %v", err)
	}
}

// TestUpdate_PreserveSurvivesMultipleUpdates: the rewritten SKILL.md must
// carry the user's preserve list so the NEXT update continues to honour it
// without the user having to re-edit every time.
func TestUpdate_PreserveSurvivesMultipleUpdates(t *testing.T) {
	root := t.TempDir()
	cacheDir := filepath.Join(root, "cache")
	installRoot := filepath.Join(root, "home", ".claude", "skills")
	manifestPath := filepath.Join(root, "manifest.json")

	build := func(version string) (string, string, map[string]string, string) {
		body := skillMD("foo", version, []string{"log.md"})
		files := map[string]string{
			"skills/foo/SKILL.md": body,
			"skills/foo/log.md":   "shipped-" + version + "\n",
		}
		flat := map[string]string{"SKILL.md": body, "log.md": "shipped-" + version + "\n"}
		return body, expectedDirSHA(t, flat), files, version
	}
	_, v1SHA, v1Files, _ := build("0.1.0")
	_, v2SHA, v2Files, _ := build("0.2.0")
	_, v3SHA, v3Files, _ := build("0.3.0")
	src1, src2, src3 := "shamul1aaaaa", "shamul2bbbbb", "shamul3ccccc"
	seedTarball(t, cacheDir, "ex", "r", src1, "ex-r-v1", v1Files)
	seedTarball(t, cacheDir, "ex", "r", src2, "ex-r-v2", v2Files)
	seedTarball(t, cacheDir, "ex", "r", src3, "ex-r-v3", v3Files)

	adapter := adapters.Adapter{
		Name:           "test",
		InstallTargets: map[string]string{"user": installRoot},
		DefaultScope:   "user",
	}
	engine := newTestEngine(t, root, cacheDir, manifestPath)
	engine.Now = func() time.Time { return time.Unix(1700000000, 0).UTC() }

	reg := func(version, sha, dirSHA string) *registry.Registry {
		return &registry.Registry{
			SchemaVersion: registry.SchemaVersion,
			Source:        registry.Source{Repo: "github.com/ex/r", SHA: sha},
			Skills: []registry.Skill{{
				Name: "foo", Version: version, Path: "skills/foo",
				Platforms: []string{"test"}, DirSHA: dirSHA,
				Preserve: []string{"log.md"},
			}},
		}
	}

	plan1, _ := Plan(reg("0.1.0", src1, v1SHA), "foo")
	if _, err := engine.Execute(reg("0.1.0", src1, v1SHA), plan1, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"},
	}); err != nil {
		t.Fatal(err)
	}

	// User extends preserve with references/ dir.
	userBody := skillMD("foo", "0.1.0", []string{"log.md", "references/"})
	if err := os.WriteFile(filepath.Join(installRoot, "foo", "SKILL.md"), []byte(userBody), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(installRoot, "foo", "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(installRoot, "foo", "references", "note.md"), []byte("keep-me\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Update v1 -> v2.
	plan2, _ := Plan(reg("0.2.0", src2, v2SHA), "foo")
	if _, err := engine.Execute(reg("0.2.0", src2, v2SHA), plan2, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"},
	}); err != nil {
		t.Fatal(err)
	}

	// Update v2 -> v3 WITHOUT re-editing SKILL.md. The user's preserve
	// list should still ride through because the previous update wrote
	// it back into the on-disk SKILL.md.
	plan3, _ := Plan(reg("0.3.0", src3, v3SHA), "foo")
	if _, err := engine.Execute(reg("0.3.0", src3, v3SHA), plan3, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"},
	}); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(installRoot, "foo", "references", "note.md"))
	if err != nil || string(got) != "keep-me\n" {
		t.Errorf("references/note.md lost across two updates: got=%q err=%v", got, err)
	}
	skillMDAfter, _ := os.ReadFile(filepath.Join(installRoot, "foo", "SKILL.md"))
	if !strings.Contains(string(skillMDAfter), "- references/") {
		t.Errorf("SKILL.md lost user's references/ preserve entry after two updates:\n%s", skillMDAfter)
	}
	if !strings.Contains(string(skillMDAfter), "version: 0.3.0") {
		t.Errorf("SKILL.md should have v3 version; got:\n%s", skillMDAfter)
	}
}

// TestUpdate_LocalPreserveRemovedEntryWipes: user cleared their local
// preserve list, so an entry the registry still preserves now gets the
// upstream bytes instead of the user's.
func TestUpdate_LocalPreserveRemovedEntryWipes(t *testing.T) {
	root := t.TempDir()
	cacheDir := filepath.Join(root, "cache")
	installRoot := filepath.Join(root, "home", ".claude", "skills")
	manifestPath := filepath.Join(root, "manifest.json")

	v1Body := skillMD("foo", "0.1.0", []string{"log.md"})
	v1Files := map[string]string{
		"skills/foo/SKILL.md": v1Body,
		"skills/foo/log.md":   "initial\n",
	}
	v1Flat := map[string]string{"SKILL.md": v1Body, "log.md": "initial\n"}
	v1SHA := expectedDirSHA(t, v1Flat)
	src1 := "sharmv1aaaaaa"
	seedTarball(t, cacheDir, "ex", "r", src1, "ex-r-abc", v1Files)

	v2Body := skillMD("foo", "0.2.0", []string{"log.md"})
	v2Files := map[string]string{
		"skills/foo/SKILL.md": v2Body,
		"skills/foo/log.md":   "shipped-v2\n",
	}
	v2Flat := map[string]string{"SKILL.md": v2Body, "log.md": "shipped-v2\n"}
	v2SHA := expectedDirSHA(t, v2Flat)
	src2 := "sharmv2bbbbbb"
	seedTarball(t, cacheDir, "ex", "r", src2, "ex-r-def", v2Files)

	adapter := adapters.Adapter{
		Name:           "test",
		InstallTargets: map[string]string{"user": installRoot},
		DefaultScope:   "user",
	}
	engine := newTestEngine(t, root, cacheDir, manifestPath)
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

	// User opts out of preserving anything AND edits log.md.
	userBody := skillMD("foo", "0.1.0", nil)
	if err := os.WriteFile(filepath.Join(installRoot, "foo", "SKILL.md"), []byte(userBody), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(installRoot, "foo", "log.md"), []byte("user-log\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	reg2 := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/ex/r", SHA: src2},
		Skills: []registry.Skill{{
			Name: "foo", Version: "0.2.0", Path: "skills/foo",
			Platforms: []string{"test"}, DirSHA: v2SHA,
			Preserve: []string{"log.md"}, // registry still preserves, but user opted out.
		}},
	}
	plan2, _ := Plan(reg2, "foo")
	if _, err := engine.Execute(reg2, plan2, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"},
	}); err != nil {
		t.Fatal(err)
	}

	if got, _ := os.ReadFile(filepath.Join(installRoot, "foo", "log.md")); string(got) != "shipped-v2\n" {
		t.Errorf("log.md: user opted out of preserve, want shipped-v2, got %q", got)
	}
}

// TestUpdate_LocalSkillMdUnparseable_FallsBackToRegistry: a mangled SKILL.md
// must not cause user-data loss. The engine falls back to the registry's
// preserve list and emits a warning.
func TestUpdate_LocalSkillMdUnparseable_FallsBackToRegistry(t *testing.T) {
	root := t.TempDir()
	cacheDir := filepath.Join(root, "cache")
	installRoot := filepath.Join(root, "home", ".claude", "skills")
	manifestPath := filepath.Join(root, "manifest.json")

	v1Body := skillMD("foo", "0.1.0", []string{"log.md"})
	v1Files := map[string]string{
		"skills/foo/SKILL.md": v1Body,
		"skills/foo/log.md":   "initial\n",
	}
	v1Flat := map[string]string{"SKILL.md": v1Body, "log.md": "initial\n"}
	v1SHA := expectedDirSHA(t, v1Flat)
	src1 := "shaunp1aaaaa"
	seedTarball(t, cacheDir, "ex", "r", src1, "ex-r-abc", v1Files)

	v2Body := skillMD("foo", "0.2.0", []string{"log.md"})
	v2Files := map[string]string{
		"skills/foo/SKILL.md": v2Body,
		"skills/foo/log.md":   "shipped-v2\n",
	}
	v2Flat := map[string]string{"SKILL.md": v2Body, "log.md": "shipped-v2\n"}
	v2SHA := expectedDirSHA(t, v2Flat)
	src2 := "shaunp2bbbbb"
	seedTarball(t, cacheDir, "ex", "r", src2, "ex-r-def", v2Files)

	adapter := adapters.Adapter{
		Name:           "test",
		InstallTargets: map[string]string{"user": installRoot},
		DefaultScope:   "user",
	}
	engine := newTestEngine(t, root, cacheDir, manifestPath)
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

	// Corrupt the on-disk SKILL.md.
	if err := os.WriteFile(filepath.Join(installRoot, "foo", "SKILL.md"), []byte(":::not yaml:::\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(installRoot, "foo", "log.md"), []byte("user-log\n"), 0o644); err != nil {
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
	r, err := engine.Execute(reg2, plan2, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if got, _ := os.ReadFile(filepath.Join(installRoot, "foo", "log.md")); string(got) != "user-log\n" {
		t.Errorf("log.md: fallback to registry preserve should keep user bytes, got %q", got)
	}
	if len(r.Warnings) == 0 {
		t.Errorf("expected a warning about unparseable SKILL.md")
	}
}

// TestUpdate_LocalPreserveInvalid_FallsBackToRegistry: the local SKILL.md
// parses but carries an invalid preserve list (e.g. a `..` traversal).
// Engine rejects it, warns, and falls back to the registry list.
func TestUpdate_LocalPreserveInvalid_FallsBackToRegistry(t *testing.T) {
	root := t.TempDir()
	cacheDir := filepath.Join(root, "cache")
	installRoot := filepath.Join(root, "home", ".claude", "skills")
	manifestPath := filepath.Join(root, "manifest.json")

	v1Body := skillMD("foo", "0.1.0", []string{"log.md"})
	v1Files := map[string]string{
		"skills/foo/SKILL.md": v1Body,
		"skills/foo/log.md":   "initial\n",
	}
	v1Flat := map[string]string{"SKILL.md": v1Body, "log.md": "initial\n"}
	v1SHA := expectedDirSHA(t, v1Flat)
	src1 := "shainv1aaaaa"
	seedTarball(t, cacheDir, "ex", "r", src1, "ex-r-abc", v1Files)

	v2Body := skillMD("foo", "0.2.0", []string{"log.md"})
	v2Files := map[string]string{
		"skills/foo/SKILL.md": v2Body,
		"skills/foo/log.md":   "shipped-v2\n",
	}
	v2Flat := map[string]string{"SKILL.md": v2Body, "log.md": "shipped-v2\n"}
	v2SHA := expectedDirSHA(t, v2Flat)
	src2 := "shainv2bbbbb"
	seedTarball(t, cacheDir, "ex", "r", src2, "ex-r-def", v2Files)

	adapter := adapters.Adapter{
		Name:           "test",
		InstallTargets: map[string]string{"user": installRoot},
		DefaultScope:   "user",
	}
	engine := newTestEngine(t, root, cacheDir, manifestPath)
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

	userBody := skillMD("foo", "0.1.0", []string{"../evil"})
	if err := os.WriteFile(filepath.Join(installRoot, "foo", "SKILL.md"), []byte(userBody), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(installRoot, "foo", "log.md"), []byte("user-log\n"), 0o644); err != nil {
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
	r, err := engine.Execute(reg2, plan2, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if got, _ := os.ReadFile(filepath.Join(installRoot, "foo", "log.md")); string(got) != "user-log\n" {
		t.Errorf("log.md: fallback should keep user bytes, got %q", got)
	}
	if len(r.Warnings) == 0 {
		t.Errorf("expected a warning about invalid preserve list")
	}
}

func TestInstall_PreserveScopeMove(t *testing.T) {
	root := t.TempDir()
	withCwd(t, filepath.Join(root, "new"))
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
	engine := newTestEngine(t, root, cacheDir, manifestPath)
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
	engine := newTestEngine(t, root, cacheDir, manifestPath)
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
	engine := newTestEngine(t, root, cacheDir, manifestPath)
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

	engine := newTestEngine(t, root, cacheDir, manifestPath)
	plan, _ := Plan(reg, "foo")
	_, err := engine.Execute(reg, plan, ExecuteOpts{
		Adapters:  []adapters.Adapter{adapter},
		Platforms: []string{"test"},
	})
	if err == nil {
		t.Fatal("expected dir_sha mismatch error")
	}
}

// TestRename_CarriesPreservedDataToNewName is the regression test for the
// data-loss hole a rename opens: `preserve` is looked up by skill name, so a
// renamed skill matches no existing installation and every preserved path is
// written fresh from the tarball -- silently discarding the user's log,
// decisions, and patterns. `previous_names` closes it.
func TestRename_CarriesPreservedDataToNewName(t *testing.T) {
	root := t.TempDir()
	cacheDir := filepath.Join(root, "cache")
	installRoot := filepath.Join(root, "home", ".claude", "skills")
	manifestPath := filepath.Join(root, "manifest.json")

	oldBody := skillMD("use-foo", "0.1.0", []string{"log.md"})
	oldFiles := map[string]string{
		"skills/use-foo/SKILL.md": oldBody,
		"skills/use-foo/log.md":   "shipped\n",
	}
	oldSHA := expectedDirSHA(t, map[string]string{"SKILL.md": oldBody, "log.md": "shipped\n"})
	src1 := "sharename1aaa"
	seedTarball(t, cacheDir, "ex", "r", src1, "ex-r-abc", oldFiles)

	newBody := skillMD("foo", "0.2.0", []string{"log.md"})
	newFiles := map[string]string{
		"skills/foo/SKILL.md": newBody,
		"skills/foo/log.md":   "shipped-v2\n",
	}
	newSHA := expectedDirSHA(t, map[string]string{"SKILL.md": newBody, "log.md": "shipped-v2\n"})
	src2 := "sharename2bbb"
	seedTarball(t, cacheDir, "ex", "r", src2, "ex-r-def", newFiles)

	adapter := adapters.Adapter{
		Name:           "test",
		InstallTargets: map[string]string{"user": installRoot},
		DefaultScope:   "user",
	}
	engine := newTestEngine(t, root, cacheDir, manifestPath)
	engine.Now = func() time.Time { return time.Unix(1700000000, 0).UTC() }

	reg1 := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/ex/r", SHA: src1},
		Skills: []registry.Skill{{
			Name: "use-foo", Version: "0.1.0", Path: "skills/use-foo",
			Platforms: []string{"test"}, DirSHA: oldSHA, Preserve: []string{"log.md"},
		}},
	}
	plan1, _ := Plan(reg1, "use-foo")
	if _, err := engine.Execute(reg1, plan1, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"},
	}); err != nil {
		t.Fatal(err)
	}

	// The user accumulates brain data under the OLD name.
	userLog := "shipped\n[SESSION] real work the user cannot get back\n"
	if err := os.WriteFile(filepath.Join(installRoot, "use-foo", "log.md"), []byte(userLog), 0o644); err != nil {
		t.Fatal(err)
	}

	// Upstream renames use-foo -> foo.
	reg2 := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/ex/r", SHA: src2},
		Skills: []registry.Skill{{
			Name: "foo", Version: "0.2.0", Path: "skills/foo",
			Platforms: []string{"test"}, DirSHA: newSHA, Preserve: []string{"log.md"},
			PreviousNames: []string{"use-foo"},
		}},
	}
	plan2, _ := Plan(reg2, "foo")
	r, err := engine.Execute(reg2, plan2, ExecuteOpts{
		Adapters: []adapters.Adapter{adapter}, Platforms: []string{"test"},
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(installRoot, "foo", "log.md"))
	if err != nil {
		t.Fatalf("renamed skill not installed: %v", err)
	}
	if string(got) != userLog {
		t.Errorf("user brain data lost across rename\n got: %q\nwant: %q", got, userLog)
	}

	// The superseded install is retired, so the skill is not double-registered.
	if _, err := os.Stat(filepath.Join(installRoot, "use-foo")); !os.IsNotExist(err) {
		t.Error("old install directory should be removed after a successful rename")
	}
	m, err := manifest.Load(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if m.Find("use-foo") != nil {
		t.Error("manifest should no longer track the old name")
	}
	if m.Find("foo") == nil {
		t.Error("manifest should track the new name")
	}

	// The carry-over is reported, not silent.
	var announced bool
	for _, w := range r.Warnings {
		if strings.Contains(w.Msg, "renamed from use-foo") {
			announced = true
		}
	}
	if !announced {
		t.Errorf("rename carry-over should be surfaced to the user, got %+v", r.Warnings)
	}
}

// TestRename_WithoutPreviousNamesLeavesOldAlone guards the blast radius: a
// skill that merely shares a prefix must never trigger retirement.
func TestRename_WithoutPreviousNamesLeavesOldAlone(t *testing.T) {
	m := &manifest.Manifest{Installations: []manifest.Installation{
		{Skill: "use-foo", Platform: "test", Scope: "user"},
	}}
	if got := m.FindPrevious(nil, "test", "user"); got != nil {
		t.Error("no previous_names must match nothing")
	}
	if got := m.FindPrevious([]string{"", "unrelated"}, "test", "user"); got != nil {
		t.Error("unrelated names must not match")
	}
	if got := m.FindPrevious([]string{"use-foo"}, "test", "user"); got == nil {
		t.Error("declared previous name should match")
	}
	if got := m.FindPrevious([]string{"use-foo"}, "other", "user"); got != nil {
		t.Error("must not cross platform boundaries")
	}
}

// TestInstall_AddPlatform_KeepsPreservedContentWithoutFetching reproduces the
// incident this whole path exists for: a skill installed --global for one
// platform, then installed for a SECOND platform. The canonical store is
// shared, so the second install must not overwrite it — the user's
// metadata.preserve files are their accumulated memory, and --force was never
// passed.
//
// The cached tarball is deleted before the second run: any fetch attempt fails
// the test, which is what locks in "adding a platform costs no network".
func TestInstall_AddPlatform_KeepsPreservedContentWithoutFetching(t *testing.T) {
	root := t.TempDir()
	cacheDir := filepath.Join(root, "cache")
	claudeRoot := filepath.Join(root, "home", ".claude", "skills")
	agentsRoot := filepath.Join(root, "home", ".agents", "skills")
	manifestPath := filepath.Join(root, "manifest.json")

	body := skillMD("foo", "0.1.0", []string{"references/log.md"})
	files := map[string]string{
		"skills/foo/SKILL.md":          body,
		"skills/foo/references/log.md": "seed\n",
	}
	dirSHA := expectedDirSHA(t, map[string]string{
		"SKILL.md": body, "references/log.md": "seed\n",
	})
	sha := "shaaddplat111"
	tarName := seedTarball(t, cacheDir, "ex", "r", sha, "ex-r-abc", files)

	claude := adapters.Adapter{
		Name: "claude-code", InstallTargets: map[string]string{"user": claudeRoot}, DefaultScope: "user",
	}
	codex := adapters.Adapter{
		Name: "codex", InstallTargets: map[string]string{"user": agentsRoot}, DefaultScope: "user",
	}
	all := []adapters.Adapter{claude, codex}

	engine := newTestEngine(t, root, cacheDir, manifestPath)
	engine.Now = func() time.Time { return time.Unix(1700000000, 0).UTC() }

	reg := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/ex/r", SHA: sha},
		Skills: []registry.Skill{{
			Name: "foo", Version: "0.1.0", Path: "skills/foo",
			DirSHA: dirSHA, Preserve: []string{"references/log.md"},
		}},
	}
	plan, err := Plan(reg, "foo")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Execute(reg, plan, ExecuteOpts{
		Adapters: all, Platforms: []string{"claude-code"}, Global: true,
	}); err != nil {
		t.Fatal(err)
	}

	storePath, err := CanonicalSkillPath("foo", "user", true)
	if err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(storePath, "references", "log.md")

	// Real accumulated use: content no upstream revision ever shipped.
	userLog := "seed\n[SESSION] decisions the user cannot get back\n"
	if err := os.WriteFile(logPath, []byte(userLog), 0o644); err != nil {
		t.Fatal(err)
	}

	// Nothing has drifted, so nothing may be downloaded. Removing the cache
	// entry turns any fetch into a hard failure instead of a silent refetch.
	if err := os.Remove(tarName); err != nil {
		t.Fatal(err)
	}

	res, err := engine.Execute(reg, plan, ExecuteOpts{
		Adapters: all, Platforms: []string{"codex"}, Global: true,
	})
	if err != nil {
		t.Fatalf("adding a platform must not fetch: %v", err)
	}

	got, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("preserved file gone: %v", err)
	}
	if string(got) != userLog {
		t.Errorf("preserved file was overwritten by the platform add:\n got %q\nwant %q", got, userLog)
	}

	if len(res.Results) != 1 || res.Results[0].Outcome != OutcomeLinked {
		t.Fatalf("want one linked target, got %+v", res.Results)
	}
	link, err := os.Readlink(filepath.Join(agentsRoot, "foo"))
	if err != nil {
		t.Fatalf("codex target is not a symlink: %v", err)
	}
	if link != storePath {
		t.Errorf("codex symlink -> %q, want the shared store %q", link, storePath)
	}

	// The first platform keeps working: same store, still one canonical copy.
	if l, err := os.Readlink(filepath.Join(claudeRoot, "foo")); err != nil || l != storePath {
		t.Errorf("claude-code symlink broke: %q %v", l, err)
	}
	m, err := manifest.Load(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Installations) != 2 {
		t.Errorf("want 2 manifest entries (one per platform), got %d", len(m.Installations))
	}
}

// TestInstall_AddPlatformWithDrift_UpdatesStoreAndPreserves covers the awkward
// case: the registry moved on AND the user is adding a platform in the same
// run. The store must be refreshed to the new revision with preserved content
// carried over, not skipped and not clobbered.
func TestInstall_AddPlatformWithDrift_UpdatesStoreAndPreserves(t *testing.T) {
	root := t.TempDir()
	cacheDir := filepath.Join(root, "cache")
	claudeRoot := filepath.Join(root, "home", ".claude", "skills")
	agentsRoot := filepath.Join(root, "home", ".agents", "skills")
	manifestPath := filepath.Join(root, "manifest.json")

	v1 := skillMD("foo", "0.1.0", []string{"references/log.md"})
	seedTarball(t, cacheDir, "ex", "r", "shadrift1aaa", "ex-r-abc", map[string]string{
		"skills/foo/SKILL.md":          v1,
		"skills/foo/references/log.md": "seed\n",
	})
	sha1 := expectedDirSHA(t, map[string]string{"SKILL.md": v1, "references/log.md": "seed\n"})

	v2 := skillMD("foo", "0.2.0", []string{"references/log.md"})
	seedTarball(t, cacheDir, "ex", "r", "shadrift2bbb", "ex-r-def", map[string]string{
		"skills/foo/SKILL.md":          v2,
		"skills/foo/references/log.md": "seed-v2\n",
		"skills/foo/NEW.md":            "upstream addition\n",
	})
	sha2 := expectedDirSHA(t, map[string]string{
		"SKILL.md": v2, "references/log.md": "seed-v2\n", "NEW.md": "upstream addition\n",
	})

	claude := adapters.Adapter{
		Name: "claude-code", InstallTargets: map[string]string{"user": claudeRoot}, DefaultScope: "user",
	}
	codex := adapters.Adapter{
		Name: "codex", InstallTargets: map[string]string{"user": agentsRoot}, DefaultScope: "user",
	}
	all := []adapters.Adapter{claude, codex}

	engine := newTestEngine(t, root, cacheDir, manifestPath)
	engine.Now = func() time.Time { return time.Unix(1700000000, 0).UTC() }

	mkReg := func(sha, version, dirSHA string) *registry.Registry {
		return &registry.Registry{
			SchemaVersion: registry.SchemaVersion,
			Source:        registry.Source{Repo: "github.com/ex/r", SHA: sha},
			Skills: []registry.Skill{{
				Name: "foo", Version: version, Path: "skills/foo",
				DirSHA: dirSHA, Preserve: []string{"references/log.md"},
			}},
		}
	}

	reg1 := mkReg("shadrift1aaa", "0.1.0", sha1)
	plan1, _ := Plan(reg1, "foo")
	if _, err := engine.Execute(reg1, plan1, ExecuteOpts{
		Adapters: all, Platforms: []string{"claude-code"}, Global: true,
	}); err != nil {
		t.Fatal(err)
	}

	storePath, err := CanonicalSkillPath("foo", "user", true)
	if err != nil {
		t.Fatal(err)
	}
	userLog := "seed\n[SESSION] hard-won context\n"
	if err := os.WriteFile(filepath.Join(storePath, "references", "log.md"), []byte(userLog), 0o644); err != nil {
		t.Fatal(err)
	}

	reg2 := mkReg("shadrift2bbb", "0.2.0", sha2)
	plan2, _ := Plan(reg2, "foo")
	res, err := engine.Execute(reg2, plan2, ExecuteOpts{
		Adapters: all, Platforms: []string{"codex"}, Global: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	got, _ := os.ReadFile(filepath.Join(storePath, "references", "log.md"))
	if string(got) != userLog {
		t.Errorf("preserved file lost across the drifted platform add:\n got %q\nwant %q", got, userLog)
	}
	if _, err := os.Stat(filepath.Join(storePath, "NEW.md")); err != nil {
		t.Errorf("upstream addition missing — store was not refreshed: %v", err)
	}
	if len(res.Results) != 1 || res.Results[0].Outcome != OutcomeInstalled {
		t.Fatalf("want one installed target, got %+v", res.Results)
	}

	// The platform that wasn't named in this run shares the store, so its
	// manifest row must report the revision that's actually on disk.
	m, err := manifest.Load(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, inst := range m.Installations {
		if inst.Version != "0.2.0" || inst.RegistryRef != sha2 {
			t.Errorf("stale manifest row for %s: v%s ref %s", inst.Platform, inst.Version, inst.RegistryRef)
		}
	}
}

func TestPreservedAtRisk(t *testing.T) {
	store := t.TempDir()
	body := skillMD("foo", "0.1.0", []string{"references/log.md", "references/wiki/", "gone.md"})
	if err := os.WriteFile(filepath.Join(store, "SKILL.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(store, "references", "wiki"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(store, "references", "log.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := PreservedAtRisk(store, nil)
	want := []string{"references/log.md", "references/wiki/"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}

	// No local SKILL.md: fall back to the registry's list, still filtered by
	// what exists on disk.
	bare := t.TempDir()
	if err := os.WriteFile(filepath.Join(bare, "keep.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got = PreservedAtRisk(bare, []string{"keep.md", "absent.md"})
	if len(got) != 1 || got[0] != "keep.md" {
		t.Errorf("registry fallback: got %v", got)
	}
}
