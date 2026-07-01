package profile_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

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

func TestDefaultPath_ResolvesUnderSandboxedXDG(t *testing.T) {
	s := testutil.NewSandbox(t)
	got, err := profile.DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}
	if !strings.HasPrefix(got, s.XDGConfigHome) {
		t.Errorf("DefaultPath = %q, want under %q", got, s.XDGConfigHome)
	}
	if filepath.Base(got) != "config.json" {
		t.Errorf("DefaultPath basename = %q, want config.json", filepath.Base(got))
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
