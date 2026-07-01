package selfupdate

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
)

// ErrBrewNotFound is returned by Upgrade when a Homebrew-managed install was
// detected but the brew binary itself isn't on PATH.
var ErrBrewNotFound = errors.New("brew not found on PATH")

// IsHomebrewManaged reports whether exePath resolves (after following
// symlinks, the way Homebrew's opt/Cellar layout works) into a Homebrew
// Cellar or Caskroom — the canonical signal that this install is managed by
// brew and shouldn't be overwritten by a self-download/swap, which would
// leave brew's own bookkeeping (and future `brew upgrade`/`brew uninstall`)
// broken.
func IsHomebrewManaged(exePath string) bool {
	resolved, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		resolved = exePath
	}
	return strings.Contains(resolved, "/Cellar/") || strings.Contains(resolved, "/Caskroom/")
}

// Runner abstracts process construction so tests can stub out the real
// `brew` binary. Defaults to exec.CommandContext.
type Runner func(ctx context.Context, name string, args ...string) *exec.Cmd

// Upgrade refreshes Homebrew's local tap metadata (`brew update`) and then
// runs `brew upgrade humblskills`, streaming both commands' own output to
// stdout/stderr live so the user sees Homebrew's real progress instead of a
// reimplementation of it. run defaults to exec.CommandContext when nil.
//
// The `brew update` step exists because Homebrew throttles its own
// opportunistic tap refresh (HOMEBREW_AUTO_UPDATE_SECS, 24h by default) —
// without it, `brew upgrade` can silently no-op against a stale tap and
// still exit 0, leaving the caller believing an upgrade happened when
// nothing changed. A `brew update` failure is logged to stderr but not
// treated as fatal on its own: `brew upgrade` still runs, and the caller's
// own post-upgrade version check (VerifyInstalledVersion) is what actually
// decides whether the upgrade succeeded.
func Upgrade(ctx context.Context, run Runner, stdout, stderr io.Writer, sink EventSink) error {
	if run == nil {
		run = exec.CommandContext
	}

	sink.emit(Event{Phase: PhaseBrewUpdating})
	if err := runBrew(ctx, run, stdout, stderr, "update"); err != nil {
		if errors.Is(err, ErrBrewNotFound) {
			sink.emit(Event{Phase: PhaseError, Err: err})
			return err
		}
		fmt.Fprintf(stderr, "warning: brew update failed, continuing with brew upgrade anyway: %v\n", err)
	}

	sink.emit(Event{Phase: PhaseBrewUpgrading})
	if err := runBrew(ctx, run, stdout, stderr, "upgrade", "humblskills"); err != nil {
		sink.emit(Event{Phase: PhaseError, Err: err})
		return err
	}
	return nil
}

// runBrew runs one `brew <args...>` invocation via run, streaming its
// output to stdout/stderr and normalizing a missing `brew` binary to
// ErrBrewNotFound.
func runBrew(ctx context.Context, run Runner, stdout, stderr io.Writer, args ...string) error {
	cmd := run(ctx, "brew", args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return ErrBrewNotFound
		}
		return fmt.Errorf("brew %s: %w", strings.Join(args, " "), err)
	}
	return nil
}
