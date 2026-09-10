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

// Homebrew formula names in jjfantini/homebrew-humbl. `humblskills@beta` is
// illegal (Homebrew only maps `@` + digits to a Ruby class). The pre
// formula is therefore `humblskills-pre`.
const (
	FormulaStable = "humblskills"
	FormulaPre    = "humblskills-pre"
)

// FormulaForChannel is the *default* tap formula for a channel when no
// release has been resolved yet. Beta's default is the pre formula, but
// the winning beta release may be a stable — callers that already have a
// version must use FormulaForVersion so Homebrew users are not told to
// `brew upgrade humblskills-pre` when that formula cannot reach it.
func FormulaForChannel(channel string) string {
	if NormalizeChannel(channel) == ChannelBeta {
		return FormulaPre
	}
	return FormulaStable
}

// FormulaForVersion returns the tap formula that publishes version.
// Prerelease tags (`vX.Y.Z-pre.N`) map to humblskills-pre; everything
// else maps to humblskills. Channel is not consulted — the winning
// release decides, so beta can land on the stable formula.
func FormulaForVersion(version string) string {
	if isPreTag(version) {
		return FormulaPre
	}
	return FormulaStable
}

// FormulaForRelease is FormulaForVersion using the GitHub prerelease
// flag or a `-pre` tag.
func FormulaForRelease(rel *Release) string {
	if rel == nil {
		return FormulaStable
	}
	if rel.Prerelease || isPreTag(rel.TagName) {
		return FormulaPre
	}
	return FormulaStable
}

// InstalledFormula returns the Homebrew formula name encoded in a
// Cellar/Caskroom path, or "" when exePath is not brew-managed.
// `/opt/homebrew/Cellar/humblskills-pre/2.52.0-pre/bin/humblskills` →
// `humblskills-pre`.
func InstalledFormula(exePath string) string {
	resolved, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		resolved = exePath
	}
	if f := formulaInPath(resolved, "/Cellar/"); f != "" {
		return f
	}
	return formulaInPath(resolved, "/Caskroom/")
}

func formulaInPath(path, marker string) string {
	i := strings.Index(path, marker)
	if i < 0 {
		return ""
	}
	rest := path[i+len(marker):]
	formula, _, ok := strings.Cut(rest, "/")
	if !ok || formula == "" {
		return ""
	}
	return formula
}

// HomebrewLinkedBinary is the brew prefix `bin/` symlink for a Cellar
// install (`…/Cellar/humblskills-pre/2.52.0-pre/bin/humblskills` →
// `…/bin/humblskills`). After a formula switch the old Cellar path is
// gone; this is the path that still exists.
func HomebrewLinkedBinary(exePath string) string {
	resolved, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		resolved = exePath
	}
	for _, marker := range []string{"/Cellar/", "/Caskroom/"} {
		i := strings.Index(resolved, marker)
		if i < 0 {
			continue
		}
		return filepath.Join(resolved[:i], "bin", filepath.Base(resolved))
	}
	return ""
}

// BrewAction is the Homebrew work an upgrade (or a notice) should run.
// Same-formula: `brew upgrade <target>`. Cross-formula (beta picked a
// stable while the user is on humblskills-pre, or the reverse):
// `brew uninstall <replace> && brew install <target>`.
type BrewAction struct {
	Target  string
	Replace string
}

// NeedsSwitch reports whether the user must change formulas to reach Target.
func (a BrewAction) NeedsSwitch() bool {
	return a.Replace != "" && a.Target != "" && a.Replace != a.Target
}

// Hint is the exact shell the notice / dry-run / confirm prompt prints.
func (a BrewAction) Hint() string {
	target := a.Target
	if target == "" {
		target = FormulaStable
	}
	if a.NeedsSwitch() {
		return fmt.Sprintf("brew uninstall %s && brew install %s", a.Replace, target)
	}
	return "brew upgrade " + target
}

// PlanBrewAction builds the Homebrew action for a resolved target formula
// and the formula currently installed (if any).
func PlanBrewAction(currentFormula, targetFormula string) BrewAction {
	if targetFormula == "" {
		targetFormula = FormulaStable
	}
	a := BrewAction{Target: targetFormula}
	if currentFormula != "" && currentFormula != targetFormula {
		a.Replace = currentFormula
	}
	return a
}

// RecommendedUpgradeCommand is the command notices, the dashboard banner,
// and upgrade --dry-run all print. One source of truth: GitHub installs
// get `humblskills upgrade`; brew stays on `brew upgrade <formula>` unless
// the winning version lives on the other formula, in which case it tells
// the user to uninstall/install.
func RecommendedUpgradeCommand(homebrew bool, currentFormula, targetFormula string) string {
	if !homebrew {
		return "humblskills upgrade"
	}
	return PlanBrewAction(currentFormula, targetFormula).Hint()
}

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
// runs `brew upgrade <formula>`, streaming both commands' own output to
// stdout/stderr live so the user sees Homebrew's real progress instead of a
// reimplementation of it. formula defaults to FormulaStable when empty.
// run defaults to exec.CommandContext when nil.
//
// The `brew update` step exists because Homebrew throttles its own
// opportunistic tap refresh (HOMEBREW_AUTO_UPDATE_SECS, 24h by default) —
// without it, `brew upgrade` can silently no-op against a stale tap and
// still exit 0, leaving the caller believing an upgrade happened when
// nothing changed. A `brew update` failure is logged to stderr but not
// treated as fatal on its own: `brew upgrade` still runs, and the caller's
// own post-upgrade version check (VerifyInstalledVersion) is what actually
// decides whether the upgrade succeeded.
func Upgrade(ctx context.Context, run Runner, stdout, stderr io.Writer, sink EventSink, formula string) error {
	return ApplyBrew(ctx, run, stdout, stderr, sink, BrewAction{Target: formula})
}

// ApplyBrew runs `brew update` then either `brew upgrade <target>` or
// `brew uninstall <replace> && brew install <target>` when the winning
// version lives on the other formula.
func ApplyBrew(ctx context.Context, run Runner, stdout, stderr io.Writer, sink EventSink, action BrewAction) error {
	if run == nil {
		run = exec.CommandContext
	}
	if action.Target == "" {
		action.Target = FormulaStable
	}

	sink.emit(Event{Phase: PhaseBrewUpdating})
	if err := runBrew(ctx, run, stdout, stderr, "update"); err != nil {
		if errors.Is(err, ErrBrewNotFound) {
			sink.emit(Event{Phase: PhaseError, Err: err})
			return err
		}
		next := "brew upgrade"
		if action.NeedsSwitch() {
			next = "the formula switch"
		}
		fmt.Fprintf(stderr, "warning: brew update failed, continuing with %s anyway: %v\n", next, err)
	}

	if action.NeedsSwitch() {
		sink.emit(Event{Phase: PhaseBrewUninstalling})
		if err := runBrew(ctx, run, stdout, stderr, "uninstall", action.Replace); err != nil {
			sink.emit(Event{Phase: PhaseError, Err: err})
			return err
		}
		sink.emit(Event{Phase: PhaseBrewInstalling})
		if err := runBrew(ctx, run, stdout, stderr, "install", action.Target); err != nil {
			sink.emit(Event{Phase: PhaseError, Err: err})
			return err
		}
		return nil
	}

	sink.emit(Event{Phase: PhaseBrewUpgrading})
	if err := runBrew(ctx, run, stdout, stderr, "upgrade", action.Target); err != nil {
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
