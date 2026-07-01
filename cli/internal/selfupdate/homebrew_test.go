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

// stubBrewRunner returns a Runner that always succeeds (or always fails,
// per succeed) regardless of args, recording every invocation's args in
// order.
func stubBrewRunner(t *testing.T, invocations *[][]string, succeed bool) Runner {
	t.Helper()
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		*invocations = append(*invocations, append([]string{}, args...))
		if succeed {
			if runtime.GOOS == "windows" {
				return exec.CommandContext(ctx, "cmd", "/c", "exit 0")
			}
			return exec.CommandContext(ctx, "true")
		}
		if runtime.GOOS == "windows" {
			return exec.CommandContext(ctx, "cmd", "/c", "exit 1")
		}
		return exec.CommandContext(ctx, "false")
	}
}

func TestUpgrade_RunsBrewUpdateBeforeUpgrade(t *testing.T) {
	var invocations [][]string
	runner := stubBrewRunner(t, &invocations, true)

	var stdout, stderr bytes.Buffer
	if err := Upgrade(context.Background(), runner, &stdout, &stderr, nil); err != nil {
		t.Fatalf("Upgrade: %v", err)
	}

	if len(invocations) != 2 {
		t.Fatalf("invocations = %v, want 2 calls", invocations)
	}
	if len(invocations[0]) != 1 || invocations[0][0] != "update" {
		t.Errorf("first call = %v, want [update]", invocations[0])
	}
	if len(invocations[1]) != 2 || invocations[1][0] != "upgrade" || invocations[1][1] != "humblskills" {
		t.Errorf("second call = %v, want [upgrade humblskills]", invocations[1])
	}
}

func TestUpgrade_BrewUpdateFailureDoesNotBlockUpgrade(t *testing.T) {
	var invocations [][]string
	callN := 0
	runner := func(ctx context.Context, name string, args ...string) *exec.Cmd {
		invocations = append(invocations, append([]string{}, args...))
		callN++
		// Only the first call (brew update) fails; brew upgrade succeeds.
		if callN == 1 {
			if runtime.GOOS == "windows" {
				return exec.CommandContext(ctx, "cmd", "/c", "exit 1")
			}
			return exec.CommandContext(ctx, "false")
		}
		if runtime.GOOS == "windows" {
			return exec.CommandContext(ctx, "cmd", "/c", "exit 0")
		}
		return exec.CommandContext(ctx, "true")
	}

	var stdout, stderr bytes.Buffer
	if err := Upgrade(context.Background(), runner, &stdout, &stderr, nil); err != nil {
		t.Fatalf("Upgrade should not fail when only brew update fails: %v", err)
	}
	if len(invocations) != 2 {
		t.Fatalf("invocations = %v, want 2 calls (upgrade must still run)", invocations)
	}
	if stderr.Len() == 0 {
		t.Error("expected a warning written to stderr about the failed brew update")
	}
}

func TestUpgrade_EmitsBrewUpdatingThenBrewUpgradingPhases(t *testing.T) {
	var invocations [][]string
	runner := stubBrewRunner(t, &invocations, true)

	var phases []Phase
	sink := EventSink(func(ev Event) { phases = append(phases, ev.Phase) })

	if err := Upgrade(context.Background(), runner, &bytes.Buffer{}, &bytes.Buffer{}, sink); err != nil {
		t.Fatalf("Upgrade: %v", err)
	}
	if len(phases) != 2 || phases[0] != PhaseBrewUpdating || phases[1] != PhaseBrewUpgrading {
		t.Errorf("phases = %v, want [%s %s]", phases, PhaseBrewUpdating, PhaseBrewUpgrading)
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

	err := Upgrade(context.Background(), runner, &bytes.Buffer{}, &bytes.Buffer{}, nil)
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

	// Both brew update and brew upgrade fail here, so the overall call must
	// still surface the (second, fatal) failure.
	err := Upgrade(context.Background(), runner, &bytes.Buffer{}, &bytes.Buffer{}, nil)
	if err == nil {
		t.Fatal("expected error for non-zero brew exit code")
	}
	if errors.Is(err, ErrBrewNotFound) {
		t.Errorf("non-zero exit shouldn't be classified as ErrBrewNotFound: %v", err)
	}
}
