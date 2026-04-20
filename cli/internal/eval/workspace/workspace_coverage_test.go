package workspace_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/workspace"
	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

func TestResolver_PrefersFlagOverEnvOverProfile(t *testing.T) {
	r := workspace.Resolver{
		FlagOverride:   "/flag",
		EnvOverride:    "/env",
		ProfileDefault: "/profile",
	}
	got, err := r.Root()
	if err != nil {
		t.Fatal(err)
	}
	if got != "/flag" {
		t.Errorf("got %q, want /flag", got)
	}
}

func TestResolver_FallsBackToEnvThenProfile(t *testing.T) {
	r := workspace.Resolver{EnvOverride: "/env", ProfileDefault: "/profile"}
	got, _ := r.Root()
	if got != "/env" {
		t.Errorf("got %q, want /env", got)
	}

	r = workspace.Resolver{ProfileDefault: "/profile"}
	got, _ = r.Root()
	if got != "/profile" {
		t.Errorf("got %q, want /profile", got)
	}
}

func TestResolver_FallsBackToDefaultRoot(t *testing.T) {
	s := testutil.NewSandbox(t)
	got, err := workspace.Resolver{}.Root()
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	if !strings.HasPrefix(got, s.XDGStateHome) {
		t.Errorf("default root %q not under XDG_STATE_HOME %q", got, s.XDGStateHome)
	}
}

func TestDefaultRoot_UsesXDGStateHome(t *testing.T) {
	s := testutil.NewSandbox(t)
	got, err := workspace.DefaultRoot()
	if err != nil {
		t.Fatalf("DefaultRoot: %v", err)
	}
	if !strings.HasPrefix(got, s.XDGStateHome) {
		t.Errorf("DefaultRoot = %q", got)
	}
}

func TestSkillDir_AndRegistryPath(t *testing.T) {
	if got := workspace.SkillDir("/r", "foo"); got != "/r/foo" {
		t.Errorf("SkillDir = %q", got)
	}
	if got := workspace.RegistryPath("/r", "foo"); got != "/r/foo/iterations.json" {
		t.Errorf("RegistryPath = %q", got)
	}
	if got := workspace.IterationDir("/r", "foo", 3); got != "/r/foo/iteration-3" {
		t.Errorf("IterationDir = %q", got)
	}
}

func TestLoadRegistry_MissingReturnsEmpty(t *testing.T) {
	root := t.TempDir()
	r, err := workspace.LoadRegistry(root, "foo")
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}
	if r == nil || r.SchemaVersion != workspace.SchemaVersion {
		t.Errorf("unexpected: %+v", r)
	}
}

func TestLoadRegistry_UnsupportedSchema(t *testing.T) {
	root := t.TempDir()
	path := workspace.RegistryPath(root, "foo")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"schema_version": 999}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.LoadRegistry(root, "foo"); err == nil {
		t.Fatal("expected unsupported schema_version error")
	}
}

func TestLoadRegistry_CorruptJSON(t *testing.T) {
	root := t.TempDir()
	path := workspace.RegistryPath(root, "foo")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := workspace.LoadRegistry(root, "foo"); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestSaveRegistry_NilReturnsError(t *testing.T) {
	if err := workspace.SaveRegistry(t.TempDir(), "x", nil); err == nil {
		t.Fatal("expected error for nil registry")
	}
}

func TestBeginIteration_AllocatesAndRecords(t *testing.T) {
	root := t.TempDir()
	n, dir, err := workspace.BeginIteration(root, "foo", "mock", []string{"smart_skill", "no_skill"}, []string{"s1"})
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	if n != 1 {
		t.Errorf("n = %d, want 1", n)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("iteration dir missing: %v", err)
	}
	reg, _ := workspace.LoadRegistry(root, "foo")
	if len(reg.Iterations) != 1 {
		t.Fatalf("iterations = %d", len(reg.Iterations))
	}
	if reg.Iterations[0].Status != workspace.StatusRunning {
		t.Errorf("status = %q", reg.Iterations[0].Status)
	}
}

func TestBeginIteration_MonotonicAcrossCalls(t *testing.T) {
	root := t.TempDir()
	n1, _, _ := workspace.BeginIteration(root, "foo", "mock", nil, nil)
	n2, _, _ := workspace.BeginIteration(root, "foo", "mock", nil, nil)
	if n2 != n1+1 {
		t.Errorf("expected n2=%d, got %d", n1+1, n2)
	}
}

func TestCompleteIteration_StampsAndRecordsStats(t *testing.T) {
	root := t.TempDir()
	n, _, _ := workspace.BeginIteration(root, "foo", "mock", nil, nil)
	err := workspace.CompleteIteration(root, "foo", n,
		map[string]float64{"smart_skill": 0.9},
		map[string]int{"total": 1000},
	)
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	reg, _ := workspace.LoadRegistry(root, "foo")
	it := reg.Iterations[0]
	if it.Status != workspace.StatusComplete {
		t.Errorf("status = %q", it.Status)
	}
	if it.CompletedAt == nil {
		t.Error("CompletedAt not set")
	}
	if it.HeadlinePassRt["smart_skill"] != 0.9 {
		t.Errorf("pass rate = %v", it.HeadlinePassRt)
	}
}

func TestCompleteIteration_UnknownN(t *testing.T) {
	root := t.TempDir()
	if err := workspace.CompleteIteration(root, "foo", 99, nil, nil); err == nil {
		t.Fatal("expected error for unknown iteration")
	}
}

func TestMarkIteration_UpdatesStatus(t *testing.T) {
	root := t.TempDir()
	n, _, _ := workspace.BeginIteration(root, "foo", "mock", nil, nil)
	if err := workspace.MarkIteration(root, "foo", n, workspace.StatusFailed); err != nil {
		t.Fatalf("Mark: %v", err)
	}
	reg, _ := workspace.LoadRegistry(root, "foo")
	if reg.Iterations[0].Status != workspace.StatusFailed {
		t.Errorf("status = %q", reg.Iterations[0].Status)
	}
}

func TestMarkIteration_UnknownN(t *testing.T) {
	if err := workspace.MarkIteration(t.TempDir(), "x", 1, workspace.StatusFailed); err == nil {
		t.Fatal("expected error for unknown iteration")
	}
}

func TestListSkills_IgnoresHiddenAndFiles(t *testing.T) {
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, "alpha"), 0o755)
	_ = os.MkdirAll(filepath.Join(root, "beta"), 0o755)
	_ = os.MkdirAll(filepath.Join(root, ".hidden"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "regular_file"), []byte("x"), 0o644)

	got, err := workspace.ListSkills(root)
	if err != nil {
		t.Fatalf("ListSkills: %v", err)
	}
	if len(got) != 2 || got[0] != "alpha" || got[1] != "beta" {
		t.Errorf("got %v", got)
	}
}

func TestListSkills_MissingRootIsEmpty(t *testing.T) {
	got, err := workspace.ListSkills(filepath.Join(t.TempDir(), "ghost"))
	if err != nil {
		t.Fatalf("ListSkills: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %v, want empty", got)
	}
}

func TestSizeBytes_SumsTreeAndHandlesMissing(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "a"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a", "x.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "y.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := workspace.SizeBytes(dir)
	if err != nil {
		t.Fatalf("SizeBytes: %v", err)
	}
	if got != 7 {
		t.Errorf("got %d, want 7", got)
	}

	// Missing path returns 0 without error.
	got, err = workspace.SizeBytes(filepath.Join(t.TempDir(), "nope"))
	if err != nil {
		t.Fatalf("missing path: %v", err)
	}
	if got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

func TestPrune_KeepLast(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 3; i++ {
		_, _, _ = workspace.BeginIteration(root, "foo", "mock", nil, nil)
	}
	res, err := workspace.Prune(root, "foo", workspace.PruneOpts{KeepLast: 1})
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if len(res.Removed) != 2 {
		t.Errorf("Removed = %v", res.Removed)
	}
	reg, _ := workspace.LoadRegistry(root, "foo")
	if len(reg.Iterations) != 1 {
		t.Errorf("remaining = %d", len(reg.Iterations))
	}
}

func TestPrune_OlderThan(t *testing.T) {
	root := t.TempDir()
	_, _, _ = workspace.BeginIteration(root, "foo", "mock", nil, nil)

	// Rewrite registry to backdate the single iteration far enough
	// that OlderThan matches.
	reg, _ := workspace.LoadRegistry(root, "foo")
	reg.Iterations[0].StartedAt = time.Now().Add(-48 * time.Hour)
	_ = workspace.SaveRegistry(root, "foo", reg)

	res, err := workspace.Prune(root, "foo", workspace.PruneOpts{OlderThan: 24 * time.Hour})
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if len(res.Removed) != 1 {
		t.Errorf("expected 1 removed, got %v", res.Removed)
	}
}

func TestPrune_AllDropsEverything(t *testing.T) {
	root := t.TempDir()
	_, _, _ = workspace.BeginIteration(root, "foo", "mock", nil, nil)
	_, _, _ = workspace.BeginIteration(root, "foo", "mock", nil, nil)

	res, err := workspace.Prune(root, "foo", workspace.PruneOpts{All: true})
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if len(res.Removed) != 2 {
		t.Errorf("Removed = %v", res.Removed)
	}
}

func TestPrune_DryRunKeepsFiles(t *testing.T) {
	root := t.TempDir()
	_, dir, _ := workspace.BeginIteration(root, "foo", "mock", nil, nil)

	res, err := workspace.Prune(root, "foo", workspace.PruneOpts{All: true, DryRun: true})
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if len(res.Removed) != 1 {
		t.Errorf("Removed report = %v", res.Removed)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Error("DryRun must not delete files")
	}
}

func TestHumanSize(t *testing.T) {
	cases := map[int64]string{
		0:                  "0 B",
		512:                "512 B",
		1024:               "1.0 KiB",
		1024 * 1024:        "1.0 MiB",
		int64(1024*1024*3): "3.0 MiB",
	}
	for in, want := range cases {
		if got := workspace.HumanSize(in); got != want {
			t.Errorf("HumanSize(%d) = %q, want %q", in, got, want)
		}
	}
}
