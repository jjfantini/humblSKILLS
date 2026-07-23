package profile_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jjfantini/humblSKILLS/cli/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

func TestLoad_MissingFileReturnsEmptyProfile(t *testing.T) {
	s := testutil.NewSandbox(t)

	p, err := profile.Load(s.ProfilePath)
	if err != nil {
		t.Fatalf("Load on missing file returned error: %v", err)
	}
	if p == nil {
		t.Fatal("Load returned nil profile")
	}
	if p.SchemaVersion != profile.SchemaVersion {
		t.Errorf("SchemaVersion = %d, want %d", p.SchemaVersion, profile.SchemaVersion)
	}
	if len(p.DefaultPlatforms) != 0 || p.DefaultScope != "" || p.Eval != nil {
		t.Errorf("fresh profile has non-zero fields: %+v", p)
	}
}

func TestSave_WritesAtomicallyAndRoundTrips(t *testing.T) {
	s := testutil.NewSandbox(t)

	in := &profile.Profile{
		DefaultPlatforms: []string{"claude-code", "cursor"},
		DefaultScope:     "project",
		TUIRouter:        boolPtr(false),
		Eval: &profile.EvalProfile{
			Runner:                 "anthropic-api",
			ExecutorModel:          "claude-sonnet-4-6",
			GraderModel:            "claude-opus-4-7",
			RunsPerConfiguration:   3,
			Parallel:               2,
			DefaultWorkspace:       "/tmp/ws",
			IncludeBlindComparator: true,
		},
	}
	if err := profile.Save(s.ProfilePath, in); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// The .tmp file must not linger after a successful rename.
	if _, err := os.Stat(s.ProfilePath + ".tmp"); err == nil {
		t.Error("tmp file leaked after successful Save")
	}

	out, err := profile.Load(s.ProfilePath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if out.DefaultScope != "project" {
		t.Errorf("DefaultScope = %q", out.DefaultScope)
	}
	if got := strings.Join(out.DefaultPlatforms, ","); got != "claude-code,cursor" {
		t.Errorf("DefaultPlatforms = %q", got)
	}
	if out.Eval == nil || out.Eval.Runner != "anthropic-api" {
		t.Errorf("Eval lost in round-trip: %+v", out.Eval)
	}
	// Explicit false must survive the trip — it's the opt-out from the
	// default-on router, so omitempty dropping it would silently re-enable.
	if out.TUIRouter == nil || *out.TUIRouter {
		t.Errorf("TUIRouter = %v, want explicit false", out.TUIRouter)
	}
	if !out.Eval.IncludeBlindComparator {
		t.Error("IncludeBlindComparator not round-tripped")
	}
}

func TestSave_StampsSchemaVersionWhenZero(t *testing.T) {
	s := testutil.NewSandbox(t)

	// Caller didn't set SchemaVersion — Save must stamp the current one.
	in := &profile.Profile{DefaultScope: "user"}
	if err := profile.Save(s.ProfilePath, in); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(s.ProfilePath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var raw struct {
		SchemaVersion int `json:"schema_version"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if raw.SchemaVersion != profile.SchemaVersion {
		t.Errorf("on-disk schema_version = %d, want %d", raw.SchemaVersion, profile.SchemaVersion)
	}
}

func TestSave_NilProfileErrors(t *testing.T) {
	s := testutil.NewSandbox(t)
	if err := profile.Save(s.ProfilePath, nil); err == nil {
		t.Fatal("expected error for nil profile")
	}
}

func TestLoad_RejectsUnknownSchemaVersion(t *testing.T) {
	s := testutil.NewSandbox(t)
	body := `{"schema_version": 999, "default_scope": "user"}` + "\n"
	if err := os.MkdirAll(filepath.Dir(s.ProfilePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.ProfilePath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := profile.Load(s.ProfilePath); err == nil {
		t.Fatal("expected error for unsupported schema_version")
	}
}

func TestLoad_RejectsCorruptJSON(t *testing.T) {
	s := testutil.NewSandbox(t)
	if err := os.MkdirAll(filepath.Dir(s.ProfilePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.ProfilePath, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := profile.Load(s.ProfilePath); err == nil {
		t.Fatal("expected parse error for corrupt profile")
	}
}

func TestLoad_UpgradesLegacyZeroSchemaVersion(t *testing.T) {
	s := testutil.NewSandbox(t)
	// Hand-written / legacy profile with no schema_version at all.
	body := `{"default_scope": "user"}` + "\n"
	if err := os.MkdirAll(filepath.Dir(s.ProfilePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.ProfilePath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := profile.Load(s.ProfilePath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if p.SchemaVersion != profile.SchemaVersion {
		t.Errorf("SchemaVersion not upgraded: got %d", p.SchemaVersion)
	}
	if p.DefaultScope != "user" {
		t.Errorf("DefaultScope = %q", p.DefaultScope)
	}
}

func TestDelete_RemovesFile_AndToleratesMissing(t *testing.T) {
	s := testutil.NewSandbox(t)
	if err := profile.Save(s.ProfilePath, &profile.Profile{DefaultScope: "user"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := profile.Delete(s.ProfilePath); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := os.Stat(s.ProfilePath); err == nil {
		t.Error("profile file still exists after Delete")
	}
	// Deleting a missing file must be a no-op, not an error.
	if err := profile.Delete(s.ProfilePath); err != nil {
		t.Errorf("Delete on missing file: %v", err)
	}
}

func TestDefaultPath_ResolvesUnderHumblskillsHome(t *testing.T) {
	s := testutil.NewSandbox(t)
	got, err := profile.DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}
	want := filepath.Join(s.Home, ".humblskills", "profile.json")
	if got != want {
		t.Errorf("DefaultPath = %q, want %q", got, want)
	}
}

func TestDefaultPath_MigratesLegacyXDGConfigProfile(t *testing.T) {
	s := testutil.NewSandbox(t)

	legacy := filepath.Join(s.XDGConfigHome, "humblskills", "config.json")
	body := `{"schema_version": 1, "default_scope": "project"}` + "\n"
	if err := os.MkdirAll(filepath.Dir(legacy), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacy, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := profile.DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}
	want := filepath.Join(s.Home, ".humblskills", "profile.json")
	if got != want {
		t.Errorf("DefaultPath = %q, want %q", got, want)
	}

	// The legacy file must be gone - migration moves, it doesn't copy.
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Errorf("legacy profile still exists at %q after migration: err=%v", legacy, err)
	}

	p, err := profile.Load(got)
	if err != nil {
		t.Fatalf("Load migrated profile: %v", err)
	}
	if p.DefaultScope != "project" {
		t.Errorf("migrated profile DefaultScope = %q, want %q", p.DefaultScope, "project")
	}

	// A second resolution must be a no-op (no legacy file left to migrate,
	// and DefaultPath is a pure resolver - it doesn't require Save to have
	// run first).
	got2, err := profile.DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath (second call): %v", err)
	}
	if got2 != want {
		t.Errorf("DefaultPath (second call) = %q, want %q", got2, want)
	}
}

func TestDefaultPath_NoMigrationWhenNewProfileAlreadyExists(t *testing.T) {
	s := testutil.NewSandbox(t)

	// A legacy file exists, but so does a profile at the new location -
	// the new one must win and the legacy file must be left untouched.
	legacy := filepath.Join(s.XDGConfigHome, "humblskills", "config.json")
	if err := os.MkdirAll(filepath.Dir(legacy), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacy, []byte(`{"schema_version": 1, "default_scope": "project"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	newPath := filepath.Join(s.Home, ".humblskills", "profile.json")
	if err := profile.Save(newPath, &profile.Profile{DefaultScope: "user"}); err != nil {
		t.Fatal(err)
	}

	got, err := profile.DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}
	if got != newPath {
		t.Errorf("DefaultPath = %q, want %q", got, newPath)
	}
	if _, err := os.Stat(legacy); err != nil {
		t.Errorf("legacy profile should be left alone when a new profile already exists: %v", err)
	}
	p, err := profile.Load(got)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if p.DefaultScope != "user" {
		t.Errorf("DefaultScope = %q, want %q (legacy must not have overwritten the new file)", p.DefaultScope, "user")
	}
}

func TestFilterKnownPlatforms(t *testing.T) {
	known := map[string]struct{}{"claude-code": {}, "cursor": {}}
	kept, dropped := profile.FilterKnownPlatforms(
		[]string{"claude-code", "vscode", "cursor", "emacs"},
		known,
	)
	if strings.Join(kept, ",") != "claude-code,cursor" {
		t.Errorf("kept = %v", kept)
	}
	if strings.Join(dropped, ",") != "vscode,emacs" {
		t.Errorf("dropped = %v", dropped)
	}
}

func TestSave_FailsWhenParentUnwritable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod 0500 does not prevent write on windows")
	}
	s := testutil.NewSandbox(t)
	// Create a readonly parent directory and target a file inside it.
	parent := filepath.Join(s.Root, "ro")
	if err := os.MkdirAll(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(parent, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(parent, 0o755) })

	target := filepath.Join(parent, "nested", "config.json")
	if err := profile.Save(target, &profile.Profile{DefaultScope: "user"}); err == nil {
		t.Fatal("expected Save to fail when parent is unwritable")
	}
}

func TestDelete_ReturnsErrorWhenPathIsDirectory(t *testing.T) {
	s := testutil.NewSandbox(t)
	// A directory at the profile path makes os.Remove fail with a
	// non-ErrNotExist error, which Delete should surface.
	if err := os.MkdirAll(s.ProfilePath, 0o755); err != nil {
		t.Fatal(err)
	}
	// Put a sentinel inside so os.Remove fails even on permissive OSes
	// (darwin/linux allow removing an empty dir with os.Remove).
	if err := os.WriteFile(filepath.Join(s.ProfilePath, "guard"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := profile.Delete(s.ProfilePath); err == nil {
		t.Fatal("expected error removing non-empty directory at profile path")
	}
}

func TestIsValidScope(t *testing.T) {
	cases := map[string]bool{
		"":                true,
		"user":            true,
		"project":         true,
		"global":          true,
		"adapter-default": true,
		"USER":            false,
		" user":           false,
	}
	for in, want := range cases {
		if got := profile.IsValidScope(in); got != want {
			t.Errorf("IsValidScope(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestStatusAutoReturnDuration(t *testing.T) {
	ptr := func(n int) *int { return &n }
	cases := []struct {
		name string
		p    *profile.Profile
		want time.Duration
	}{
		{"nil profile defaults to 5s", nil, 5 * time.Second},
		{"unset defaults to 5s", &profile.Profile{}, 5 * time.Second},
		{"zero disables the timer", &profile.Profile{StatusAutoReturnSeconds: ptr(0)}, 0},
		{"negative disables the timer", &profile.Profile{StatusAutoReturnSeconds: ptr(-3)}, 0},
		{"explicit positive value", &profile.Profile{StatusAutoReturnSeconds: ptr(10)}, 10 * time.Second},
	}
	for _, c := range cases {
		if got := c.p.StatusAutoReturnDuration(); got != c.want {
			t.Errorf("%s: StatusAutoReturnDuration() = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestResolvedScope(t *testing.T) {
	cases := []struct {
		name string
		p    *profile.Profile
		want string
	}{
		{"nil profile defaults to global", nil, profile.ScopeGlobal},
		{"unset defaults to global", &profile.Profile{}, profile.ScopeGlobal},
		{"explicit global", &profile.Profile{DefaultScope: profile.ScopeGlobal}, profile.ScopeGlobal},
		{"explicit user", &profile.Profile{DefaultScope: profile.ScopeUser}, profile.ScopeUser},
		{"explicit project", &profile.Profile{DefaultScope: profile.ScopeProject}, profile.ScopeProject},
		{"explicit adapter-default", &profile.Profile{DefaultScope: profile.ScopeAdapterDefault}, profile.ScopeAdapterDefault},
	}
	for _, c := range cases {
		if got := c.p.ResolvedScope(); got != c.want {
			t.Errorf("%s: ResolvedScope() = %q, want %q", c.name, got, c.want)
		}
	}
}

func boolPtr(b bool) *bool { return &b }
