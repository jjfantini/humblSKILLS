package selfupdate

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"runtime"
	"testing"
)

func TestIsHomebrewManaged(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"/opt/homebrew/Cellar/humblskills/2.17.0/bin/humblskills", true},
		{"/usr/local/Cellar/humblskills/2.17.0/bin/humblskills", true},
		{"/opt/homebrew/Caskroom/humblskills/2.17.0/humblskills", true},
		{"/usr/local/bin/humblskills", false},
		{"/home/user/go/bin/humblskills", false},
	}
	for _, c := range cases {
		// These paths don't exist on disk, so EvalSymlinks will fail and
		// IsHomebrewManaged falls back to checking the raw path — which is
		// exactly what we want to assert here (the substring check itself).
		if got := IsHomebrewManaged(c.path); got != c.want {
			t.Errorf("IsHomebrewManaged(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestUpgrade_InvokesBrewWithExpectedArgs(t *testing.T) {
	var capturedName string
	var capturedArgs []string
	runner := func(ctx context.Context, name string, args ...string) *exec.Cmd {
		capturedName = name
		capturedArgs = append([]string{}, args...)
		// Substitute a real, always-succeeding command so .Run() doesn't
		// actually need a `brew` binary in the test environment.
		if runtime.GOOS == "windows" {
			return exec.CommandContext(ctx, "cmd", "/c", "exit 0")
		}
		return exec.CommandContext(ctx, "true")
	}

	var stdout, stderr bytes.Buffer
	if err := Upgrade(context.Background(), runner, &stdout, &stderr); err != nil {
		t.Fatalf("Upgrade: %v", err)
	}
	if capturedName != "brew" {
		t.Errorf("runner called with name = %q, want brew", capturedName)
	}
	want := []string{"upgrade", "humblskills"}
	if len(capturedArgs) != len(want) || capturedArgs[0] != want[0] || capturedArgs[1] != want[1] {
		t.Errorf("runner called with args = %v, want %v", capturedArgs, want)
	}
}

func TestUpgrade_BrewNotFound(t *testing.T) {
	runner := func(ctx context.Context, name string, args ...string) *exec.Cmd {
		// A bare (no path separator) binary name guaranteed not to exist on
		// PATH, so exec.Command goes through the same LookPath resolution
		// real production code hits when it runs "brew" (also a bare
		// name) and gets exec.ErrNotFound.
		return exec.CommandContext(ctx, "definitely-not-a-real-binary-xyz")
	}

	err := Upgrade(context.Background(), runner, &bytes.Buffer{}, &bytes.Buffer{})
	if !errors.Is(err, ErrBrewNotFound) {
		t.Errorf("expected ErrBrewNotFound, got %v", err)
	}
}

func TestUpgrade_NonZeroExit(t *testing.T) {
	runner := func(ctx context.Context, name string, args ...string) *exec.Cmd {
		if runtime.GOOS == "windows" {
			return exec.CommandContext(ctx, "cmd", "/c", "exit 1")
		}
		return exec.CommandContext(ctx, "false")
	}

	err := Upgrade(context.Background(), runner, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for non-zero brew exit code")
	}
	if errors.Is(err, ErrBrewNotFound) {
		t.Errorf("non-zero exit shouldn't be classified as ErrBrewNotFound: %v", err)
	}
}
