package testutil_test

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/adrg/xdg"

	"github.com/jjfantini/humblSKILLS/cli/internal/fetch"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/internal/secrets"
	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

func TestSandbox_IsolatesEnvAndXDG(t *testing.T) {
	prevHome := os.Getenv("HOME")

	s := testutil.NewSandbox(t)

	if got := os.Getenv("HOME"); got != s.Home {
		t.Errorf("HOME = %q, want %q", got, s.Home)
	}
	if got := os.Getenv("XDG_CONFIG_HOME"); got != s.XDGConfigHome {
		t.Errorf("XDG_CONFIG_HOME = %q, want %q", got, s.XDGConfigHome)
	}
	if !strings.HasPrefix(xdg.ConfigHome, s.XDGConfigHome) {
		t.Errorf("xdg.ConfigHome = %q, want prefix %q", xdg.ConfigHome, s.XDGConfigHome)
	}
	// HUMBLSKILLS_* should be cleared, not inherited.
	if got := os.Getenv("HUMBLSKILLS_REGISTRY"); got != "" {
		t.Errorf("HUMBLSKILLS_REGISTRY leaked: %q", got)
	}
	// ANTHROPIC_API_KEY must be neutralised so secret tests are deterministic.
	if got := os.Getenv("ANTHROPIC_API_KEY"); got != "" {
		t.Errorf("ANTHROPIC_API_KEY leaked: %q", got)
	}

	// After the test returns, env must be restored. We can't observe
	// t.Cleanup directly from within the same test, but we can assert
	// that the previous value we captured still lives somewhere and
	// t.Cleanup will run later. Sanity check: prev captured non-empty.
	if prevHome == "" {
		t.Log("developer HOME was empty — cleanup check skipped")
	}
}

func TestRegistryServer_ServesRegistryAndTracksRequests(t *testing.T) {
	testutil.NewSandbox(t)

	reg := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "x/y", SHA: "deadbeef"},
	}
	srv := testutil.NewRegistryServer(t, reg)

	// 200 first
	resp, err := http.Get(srv.URL())
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("status = %d", resp.StatusCode)
	}
	if !strings.Contains(string(body), `"sha": "deadbeef"`) {
		t.Errorf("body missing sha: %s", body)
	}

	// switch to 500 and confirm the new status flows
	srv.SetStatus(500)
	resp2, _ := http.Get(srv.URL())
	resp2.Body.Close()
	if resp2.StatusCode != 500 {
		t.Errorf("status after SetStatus(500) = %d", resp2.StatusCode)
	}

	if got := srv.Requests(); got != 2 {
		t.Errorf("requests = %d, want 2", got)
	}
}

func TestBuildRegistry_SeedsTarballAndEngineCanExtract(t *testing.T) {
	s := testutil.NewSandbox(t)

	reg := testutil.BuildRegistry(t, s.CacheDir, "example", "repo", "0123456789abcdef", []testutil.SkillFixture{
		{
			Name:    "foo",
			Version: "1.0.0",
			Files: testutil.SkillTree{
				"SKILL.md":       "# foo\n",
				"data/notes.md":  "hello\n",
			},
		},
	})

	if len(reg.Skills) != 1 {
		t.Fatalf("skills = %d", len(reg.Skills))
	}
	skill := reg.Skills[0]
	if skill.DirSHA == "" {
		t.Error("DirSHA empty")
	}

	// Extract via the production Fetch path and confirm the seeded
	// tarball is usable and DirSHA recomputes to the same value.
	f := fetch.NewFetcher(s.CacheDir)
	tarPath, err := f.Fetch(reg.Source.Repo, reg.Source.SHA)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	dest := filepath.Join(t.TempDir(), "out")
	if err := fetch.Extract(tarPath, skill.Path, dest); err != nil {
		t.Fatalf("extract: %v", err)
	}
	got, err := registry.DirSHA(dest)
	if err != nil {
		t.Fatalf("dir_sha: %v", err)
	}
	if got != skill.DirSHA {
		t.Errorf("DirSHA mismatch after round-trip: seeded=%s got=%s", skill.DirSHA, got)
	}
}

func TestFakeKeyring_IntegratesWithSecretsLayeredStore(t *testing.T) {
	s := testutil.NewSandbox(t)
	testutil.UseFakeKeyring(t)

	store, err := secrets.NewStore(s.SecretsPath)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	src, err := store.Set("anthropic", "sk-test-abc")
	if err != nil {
		t.Fatalf("Set: %v", err)
	}
	if src != secrets.SourceKeyring {
		t.Errorf("Set source = %q, want keyring", src)
	}
	got, src, err := store.Get("anthropic")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "sk-test-abc" || src != secrets.SourceKeyring {
		t.Errorf("Get = (%q, %q), want (sk-test-abc, keyring)", got, src)
	}

	// After Delete, no source should hold the secret.
	if err := store.Delete("anthropic"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, src, _ = store.Get("anthropic")
	if src != secrets.SourceAbsent {
		t.Errorf("post-Delete source = %q, want absent", src)
	}
}

func TestUnavailableKeyring_FallsBackToFile(t *testing.T) {
	s := testutil.NewSandbox(t)
	testutil.UseUnavailableKeyring(t, errBoom{})

	store, err := secrets.NewStore(s.SecretsPath)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	src, err := store.Set("anthropic", "sk-file")
	if err != nil {
		t.Fatalf("Set: %v", err)
	}
	if src != secrets.SourceFile {
		t.Errorf("Set source = %q, want file", src)
	}
	info, err := os.Stat(s.SecretsPath)
	if err != nil {
		t.Fatalf("stat secrets: %v", err)
	}
	// 0o600 enforced on write — but Windows doesn't honor the Unix
	// permission bits passed to OpenFile, so this assertion is POSIX-only.
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Errorf("secrets perm = %o, want 0600", info.Mode().Perm())
	}
}

type errBoom struct{}

func (errBoom) Error() string { return "boom" }

func TestSnapshotAndGoldenDir_RoundTrip(t *testing.T) {
	src := t.TempDir()
	if err := os.MkdirAll(filepath.Join(src, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "a.txt"), []byte("alpha"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "sub", "b.txt"), []byte("beta"), 0o644); err != nil {
		t.Fatal(err)
	}

	snap := testutil.SnapshotDir(t, src)
	if snap["a.txt"] != "alpha" {
		t.Errorf("snap[a.txt] = %q", snap["a.txt"])
	}
	if snap["sub/b.txt"] != "beta" {
		t.Errorf("snap[sub/b.txt] = %q", snap["sub/b.txt"])
	}
	if _, ok := snap["sub/"]; !ok {
		t.Errorf("snap missing sub/ directory marker: %v", snap)
	}
}
