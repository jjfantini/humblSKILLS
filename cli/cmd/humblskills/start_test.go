package main

import (
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/testutil"
)

func TestCrumbLabel(t *testing.T) {
	cases := map[string]string{
		"install": "Install",
		"eval":    "Eval",
		"doctor":  "Doctor",
		// Unknown commands fall through unchanged.
		"registry refresh": "registry refresh",
	}
	for in, want := range cases {
		if got := crumbLabel(in); got != want {
			t.Errorf("crumbLabel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestStart_NonTTYFallbackListsCommands(t *testing.T) {
	testutil.NewSandbox(t)
	// Tests don't run on a TTY, so `start` prints the command table
	// fallback instead of opening the dashboard.
	res := runCLIWithStdoutCapture(t, "start")
	if res.RunErr != nil {
		t.Fatalf("start: %v", res.RunErr)
	}
	assertContains(t, res.Out, "COMMANDS")
	assertContains(t, res.Out, "install")
	assertContains(t, res.Out, "doctor")
}

func TestStart_FullscreenNonTTYErrors(t *testing.T) {
	testutil.NewSandbox(t)
	res := runCLI(t, "start", "--fullscreen")
	if res.RunErr == nil {
		t.Fatal("expected --fullscreen to error without a TTY")
	}
	assertContains(t, res.RunErr.Error(), "interactive terminal")
}
