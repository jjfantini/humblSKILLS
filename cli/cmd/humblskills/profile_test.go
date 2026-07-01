package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

func TestProfileShow_EmptyJSON(t *testing.T) {
	s := testutil.NewSandbox(t)

	res := runCLIWithStdoutCapture(t,
		"profile", "show",
		"--profile", s.ProfilePath,
		"--json",
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v\n%s", res.RunErr, res.Err)
	}
	idx := strings.Index(res.Out, "{")
	var p struct {
		SchemaVersion int `json:"schema_version"`
	}
	if err := json.Unmarshal([]byte(res.Out[idx:]), &p); err != nil {
		t.Fatalf("parse: %v\n%s", err, res.Out)
	}
	if p.SchemaVersion != profile.SchemaVersion {
		t.Errorf("schema = %d", p.SchemaVersion)
	}
}

func TestProfileSet_ValidScope(t *testing.T) {
	s := testutil.NewSandbox(t)

	res := runCLIWithStdoutCapture(t,
		"profile", "set", "scope", "user",
		"--profile", s.ProfilePath,
		"--json",
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v", res.RunErr)
	}
	p, err := profile.Load(s.ProfilePath)
	if err != nil {
		t.Fatal(err)
	}
	if p.DefaultScope != "user" {
		t.Errorf("scope = %q", p.DefaultScope)
	}
}

func TestProfileSet_ValidScope_GlobalAndAdapterDefault(t *testing.T) {
	s := testutil.NewSandbox(t)

	for _, scope := range []string{"global", "adapter-default", "project"} {
		res := runCLIWithStdoutCapture(t,
			"profile", "set", "scope", scope,
			"--profile", s.ProfilePath,
			"--json",
		)
		if res.RunErr != nil {
			t.Fatalf("scope=%q: run: %v", scope, res.RunErr)
		}
		p, err := profile.Load(s.ProfilePath)
		if err != nil {
			t.Fatal(err)
		}
		if p.DefaultScope != scope {
			t.Errorf("scope=%q: DefaultScope = %q", scope, p.DefaultScope)
		}
	}
}

func TestProfileSet_InvalidScope(t *testing.T) {
	s := testutil.NewSandbox(t)

	res := runCLIWithStdoutCapture(t,
		"profile", "set", "scope", "bogus-value",
		"--profile", s.ProfilePath,
	)
	if res.RunErr == nil {
		t.Fatal("expected error for invalid scope")
	}
}

func TestProfileSet_UnknownKey(t *testing.T) {
	s := testutil.NewSandbox(t)

	res := runCLIWithStdoutCapture(t,
		"profile", "set", "bogus-key", "x",
		"--profile", s.ProfilePath,
	)
	if res.RunErr == nil {
		t.Fatal("expected error for unknown key")
	}
	if !strings.Contains(res.RunErr.Error(), "unknown key") {
		t.Errorf("err = %v", res.RunErr)
	}
}

func TestProfileSet_ValidPlatforms(t *testing.T) {
	s := testutil.NewSandbox(t)
	res := runCLIWithStdoutCapture(t,
		"profile", "set", "platforms", "claude-code,cursor",
		"--profile", s.ProfilePath,
		"--json",
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v", res.RunErr)
	}
	p, _ := profile.Load(s.ProfilePath)
	if len(p.DefaultPlatforms) != 2 {
		t.Errorf("platforms = %v", p.DefaultPlatforms)
	}
}

func TestProfileSet_UnknownPlatform(t *testing.T) {
	s := testutil.NewSandbox(t)
	res := runCLIWithStdoutCapture(t,
		"profile", "set", "platforms", "claude-code,ghost-platform",
		"--profile", s.ProfilePath,
	)
	if res.RunErr == nil {
		t.Fatal("expected unknown platform error")
	}
}

func TestProfileSet_EmptyPlatformsClears(t *testing.T) {
	s := testutil.NewSandbox(t)
	// First set to something.
	_ = runCLIWithStdoutCapture(t,
		"profile", "set", "platforms", "claude-code",
		"--profile", s.ProfilePath, "--json",
	)
	// Now clear it.
	r := runCLIWithStdoutCapture(t,
		"profile", "set", "platforms", "",
		"--profile", s.ProfilePath, "--json",
	)
	if r.RunErr != nil {
		t.Fatalf("run: %v", r.RunErr)
	}
	p, _ := profile.Load(s.ProfilePath)
	if len(p.DefaultPlatforms) != 0 {
		t.Errorf("not cleared: %v", p.DefaultPlatforms)
	}
}

func TestProfileSet_StatusAutoReturnSeconds(t *testing.T) {
	s := testutil.NewSandbox(t)

	res := runCLIWithStdoutCapture(t,
		"profile", "set", "status_auto_return_seconds", "10",
		"--profile", s.ProfilePath,
		"--json",
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v", res.RunErr)
	}
	p, err := profile.Load(s.ProfilePath)
	if err != nil {
		t.Fatal(err)
	}
	if p.StatusAutoReturnSeconds == nil || *p.StatusAutoReturnSeconds != 10 {
		t.Errorf("StatusAutoReturnSeconds = %v", p.StatusAutoReturnSeconds)
	}
}

func TestProfileSet_StatusAutoReturnSeconds_Off(t *testing.T) {
	s := testutil.NewSandbox(t)

	res := runCLIWithStdoutCapture(t,
		"profile", "set", "status_auto_return_seconds", "off",
		"--profile", s.ProfilePath,
		"--json",
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v", res.RunErr)
	}
	p, err := profile.Load(s.ProfilePath)
	if err != nil {
		t.Fatal(err)
	}
	if p.StatusAutoReturnSeconds == nil || *p.StatusAutoReturnSeconds != 0 {
		t.Errorf("StatusAutoReturnSeconds = %v, want 0", p.StatusAutoReturnSeconds)
	}
	if p.StatusAutoReturnDuration() != 0 {
		t.Errorf("StatusAutoReturnDuration() = %v, want 0 (disabled)", p.StatusAutoReturnDuration())
	}
}

func TestProfileSet_StatusAutoReturnSeconds_Default(t *testing.T) {
	s := testutil.NewSandbox(t)

	_ = runCLIWithStdoutCapture(t,
		"profile", "set", "status_auto_return_seconds", "10",
		"--profile", s.ProfilePath, "--json",
	)
	res := runCLIWithStdoutCapture(t,
		"profile", "set", "status_auto_return_seconds", "default",
		"--profile", s.ProfilePath, "--json",
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v", res.RunErr)
	}
	p, err := profile.Load(s.ProfilePath)
	if err != nil {
		t.Fatal(err)
	}
	if p.StatusAutoReturnSeconds != nil {
		t.Errorf("StatusAutoReturnSeconds = %v, want nil (unset/default)", *p.StatusAutoReturnSeconds)
	}
}

func TestProfileSet_StatusAutoReturnSeconds_Invalid(t *testing.T) {
	s := testutil.NewSandbox(t)

	res := runCLIWithStdoutCapture(t,
		"profile", "set", "status_auto_return_seconds", "not-a-number",
		"--profile", s.ProfilePath,
	)
	if res.RunErr == nil {
		t.Fatal("expected error for invalid status_auto_return_seconds value")
	}
}

func TestProfileReset_RemovesFile(t *testing.T) {
	s := testutil.NewSandbox(t)

	// Seed a profile.
	_ = profile.Save(s.ProfilePath, &profile.Profile{DefaultScope: "user"})
	if _, err := os.Stat(s.ProfilePath); err != nil {
		t.Fatalf("precondition: profile missing: %v", err)
	}

	res := runCLIWithStdoutCapture(t,
		"profile", "reset",
		"--profile", s.ProfilePath,
		"--json",
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v", res.RunErr)
	}
	if _, err := os.Stat(s.ProfilePath); err == nil {
		t.Error("profile file should be removed after reset")
	}
}

func TestProfilePath_PrintsResolvedPath(t *testing.T) {
	s := testutil.NewSandbox(t)
	res := runCLIWithStdoutCapture(t,
		"profile", "path",
		"--profile", s.ProfilePath,
	)
	if res.RunErr != nil {
		t.Fatalf("run: %v", res.RunErr)
	}
	if !strings.Contains(res.Out, s.ProfilePath) {
		t.Errorf("expected path in output, got:\n%s", res.Out)
	}
}

func TestParseCSV(t *testing.T) {
	cases := map[string][]string{
		"":          nil,
		"   ":       nil,
		"a":         {"a"},
		"a,b":       {"a", "b"},
		"a,  b ,c":  {"a", "b", "c"},
		", a ,,b ,": {"a", "b"},
	}
	for in, want := range cases {
		got := parseCSV(in)
		if len(got) != len(want) {
			t.Errorf("parseCSV(%q) = %v, want %v", in, got, want)
			continue
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("parseCSV(%q)[%d] = %q, want %q", in, i, got[i], want[i])
			}
		}
	}
}

func TestFormatPlatforms(t *testing.T) {
	if got := formatPlatforms(nil); got != "(all detected)" {
		t.Errorf("got %q", got)
	}
	if got := formatPlatforms([]string{"a", "b"}); got != "a, b" {
		t.Errorf("got %q", got)
	}
}

func TestFormatScope(t *testing.T) {
	if got := formatScope(""); got != "global humblskills (default)" {
		t.Errorf("got %q", got)
	}
	if got := formatScope(profile.ScopeGlobal); got != "global humblskills" {
		t.Errorf("got %q", got)
	}
	if got := formatScope("user"); got != "user" {
		t.Errorf("got %q", got)
	}
	if got := formatScope("project"); got != "project" {
		t.Errorf("got %q", got)
	}
	if got := formatScope(profile.ScopeAdapterDefault); got != "adapter default" {
		t.Errorf("got %q", got)
	}
}
