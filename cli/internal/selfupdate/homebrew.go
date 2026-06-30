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

// Upgrade runs `brew upgrade humblskills`, streaming its own output to
// stdout/stderr live so the user sees Homebrew's real progress instead of a
// reimplementation of it. run defaults to exec.CommandContext when nil.
func Upgrade(ctx context.Context, run Runner, stdout, stderr io.Writer) error {
	if run == nil {
		run = exec.CommandContext
	}
	cmd := run(ctx, "brew", "upgrade", "humblskills")
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return ErrBrewNotFound
		}
		return fmt.Errorf("brew upgrade humblskills: %w", err)
	}
	return nil
}
