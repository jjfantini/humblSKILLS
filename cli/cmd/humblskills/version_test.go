package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

func TestVersion_JSONOutputsSchema(t *testing.T) {
	testutil.NewSandbox(t)

	res := runCLIWithStdoutCapture(t, "version", "--json")
	if res.RunErr != nil {
		t.Fatalf("run: %v\nerr: %s", res.RunErr, res.Err)
	}
	// Extract just the JSON object (the UI may prefix with other output).
	out := strings.TrimSpace(res.Out)
	if !strings.HasPrefix(out, "{") {
		t.Fatalf("expected JSON object, got:\n%s", out)
	}

	var v struct {
		Version string `json:"version"`
		Commit  string `json:"commit"`
	}
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("unmarshal: %v\nbody: %s", err, out)
	}
	if v.Version == "" {
		t.Error("Version field empty")
	}
	if v.Commit == "" {
		t.Error("Commit field empty")
	}
}

func TestVersion_TextOutputIncludesBrand(t *testing.T) {
	testutil.NewSandbox(t)

	// --yes disables TUI mode (see tui.ShouldUseTUI) so we get the plain
	// pipe-friendly output which is the part tests should lock down.
	res := runCLIWithStdoutCapture(t, "version", "--yes")
	if res.RunErr != nil {
		t.Fatalf("run: %v\nerr: %s", res.RunErr, res.Err)
	}
	assertContains(t, res.Out, "humblskills")
}

func TestResolveVersion_ReturnsDevDefaults(t *testing.T) {
	info := resolveVersion()
	// Even in a test binary, Version and Commit are non-empty.
	if info.Version == "" {
		t.Error("Version empty")
	}
	if info.Commit == "" {
		t.Error("Commit empty")
	}
}

func TestShortCommit(t *testing.T) {
	cases := map[string]string{
		"abc":                    "abc",
		"abcdef1234567":          "abcdef123456",
		"abcdef1234567890abcdef": "abcdef123456",
	}
	for in, want := range cases {
		if got := shortCommit(in); got != want {
			t.Errorf("shortCommit(%q) = %q, want %q", in, got, want)
		}
	}
}
