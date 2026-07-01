package adapters_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/adapters"
	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

func TestExpandPath_TildeAlone(t *testing.T) {
	// "~" by itself resolves to HOME.
	s := testutil.NewSandbox(t)
	got := adapters.ExpandPath("~")
	if got != s.Home {
		t.Errorf("ExpandPath(~) = %q, want %q", got, s.Home)
	}
}

func TestExpandPath_TildeSlash(t *testing.T) {
	s := testutil.NewSandbox(t)
	got := adapters.ExpandPath("~/skills")
	if got != filepath.Join(s.Home, "skills") {
		t.Errorf("ExpandPath(~/skills) = %q", got)
	}
}

func TestExpandPath_EnvSubstitution(t *testing.T) {
	t.Setenv("HUMBL_TEST_DIR", "/tmp/xyz")
	got := adapters.ExpandPath("$HUMBL_TEST_DIR/foo")
	if got != "/tmp/xyz/foo" {
		t.Errorf("ExpandPath = %q", got)
	}
	got = adapters.ExpandPath("${HUMBL_TEST_DIR}/bar")
	if got != "/tmp/xyz/bar" {
		t.Errorf("ExpandPath (braces) = %q", got)
	}
}

func TestExpandPath_PlainPath(t *testing.T) {
	got := adapters.ExpandPath("/absolute/path")
	if got != "/absolute/path" {
		t.Errorf("got %q", got)
	}
}

func TestExpandPath_EmptyReturnsEmpty(t *testing.T) {
	if got := adapters.ExpandPath(""); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestDetect_AllOfFailsEarly(t *testing.T) {
	// Build an adapter inline with all_of rules that must fail on the
	// first missing path, so describeRule is invoked on the failing
	// rule.
	ads := []adapters.Adapter{
		{
			Name: "t",
			Detect: adapters.DetectRules{
				AllOf: []adapters.DetectRule{
					{PathExists: "/this/does/not/exist"},
					{Env: "NEVER_SET_XYZZY"},
				},
			},
		},
	}
	got := adapters.Detect(ads)
	if len(got) != 1 {
		t.Fatalf("got %d results", len(got))
	}
	if got[0].Detected {
		t.Error("expected Detected=false")
	}
	if !strings.Contains(got[0].Reason, "not found") {
		t.Errorf("reason = %q", got[0].Reason)
	}
}

func TestDetect_AnyOfMatchesFirst(t *testing.T) {
	t.Setenv("HUMBL_ANYOF_TEST", "1")
	ads := []adapters.Adapter{
		{
			Name: "t",
			Detect: adapters.DetectRules{
				AnyOf: []adapters.DetectRule{
					{Env: "HUMBL_ANYOF_TEST"},
					{PathExists: "/nope"},
				},
			},
		},
	}
	got := adapters.Detect(ads)
	if !got[0].Detected {
		t.Errorf("expected match: %+v", got[0])
	}
	if !strings.Contains(got[0].Reason, "is set") {
		t.Errorf("reason = %q", got[0].Reason)
	}
}

func TestDetect_NoRulesReportsEmpty(t *testing.T) {
	ads := []adapters.Adapter{{Name: "bare"}}
	got := adapters.Detect(ads)
	if got[0].Detected {
		t.Error("empty rules must not detect")
	}
	if !strings.Contains(got[0].Reason, "no detect") {
		t.Errorf("reason = %q", got[0].Reason)
	}
}

func TestAdapter_Target_UnknownScope(t *testing.T) {
	a := adapters.Adapter{
		Name: "x",
		InstallTargets: map[string]string{
			"user": "/tmp/x-user",
		},
		DefaultScope: "user",
	}
	if _, err := a.Target("project"); err == nil {
		t.Error("expected error for unknown scope")
	}
}

func TestAdapter_Target_EmptyScopeFallsBackToDefault(t *testing.T) {
	a := adapters.Adapter{
		Name: "x",
		InstallTargets: map[string]string{
			"user": "/tmp/x",
		},
		DefaultScope: "user",
	}
	tg, err := a.Target("")
	if err != nil {
		t.Fatalf("Target: %v", err)
	}
	if tg.Scope != "user" {
		t.Errorf("Scope = %q", tg.Scope)
	}
}

func TestAdapter_Targets_ReturnsEveryScopeSorted(t *testing.T) {
	a := adapters.Adapter{
		Name: "x",
		InstallTargets: map[string]string{
			"user":    "/tmp/u",
			"project": "./proj",
		},
	}
	ts := a.Targets()
	if len(ts) != 2 {
		t.Fatalf("got %d", len(ts))
	}
	// Sorted lexically: project < user.
	if ts[0].Scope != "project" || ts[1].Scope != "user" {
		t.Errorf("scopes = %v", []string{ts[0].Scope, ts[1].Scope})
	}
	// Relative paths get resolved against cwd.
	if !filepath.IsAbs(ts[0].Path) {
		t.Errorf("project path not absolute: %q", ts[0].Path)
	}
}

func TestNameSet(t *testing.T) {
	ads := []adapters.Adapter{{Name: "a"}, {Name: "b"}}
	set := adapters.NameSet(ads)
	if _, ok := set["a"]; !ok {
		t.Error("a missing")
	}
	if _, ok := set["b"]; !ok {
		t.Error("b missing")
	}
	if _, ok := set["c"]; ok {
		t.Error("c unexpectedly present")
	}
}

func TestTarget_Writable_OnNewPath(t *testing.T) {
	// A brand-new nested path under a writable tempdir must be
	// considered writable (isWritable walks up to the first existing
	// parent and probes it).
	s := testutil.NewSandbox(t)
	a := adapters.Adapter{
		Name: "x",
		InstallTargets: map[string]string{
			"user": filepath.Join(s.Root, "new", "nested", "path"),
		},
		DefaultScope: "user",
	}
	tg, err := a.Target("user")
	if err != nil {
		t.Fatal(err)
	}
	if !tg.Writable {
		t.Error("expected Writable=true under a tempdir")
	}
}

func TestTarget_NotWritable_WhenParentIsFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission semantics differ on windows")
	}
	s := testutil.NewSandbox(t)
	blocker := filepath.Join(s.Root, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	a := adapters.Adapter{
		Name: "x",
		InstallTargets: map[string]string{
			"user": filepath.Join(blocker, "can-not-write"),
		},
		DefaultScope: "user",
	}
	tg, err := a.Target("user")
	if err != nil {
		t.Fatal(err)
	}
	if tg.Writable {
		t.Error("expected Writable=false when parent is a file")
	}
}

func TestLoad_ReturnsEmbeddedAdapters(t *testing.T) {
	// Load() reads *.yaml embedded at build time. Every shipped adapter
	// must parse, expose a non-empty Name, and a valid DefaultScope in
	// InstallTargets.
	ads, err := adapters.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(ads) == 0 {
		t.Fatal("no embedded adapters")
	}
	for _, a := range ads {
		if a.Name == "" {
			t.Errorf("adapter missing name: %+v", a)
		}
		if a.DefaultScope != "" {
			if _, ok := a.InstallTargets[a.DefaultScope]; !ok {
				t.Errorf("adapter %q: DefaultScope %q has no install target",
					a.Name, a.DefaultScope)
			}
		}
	}
}
