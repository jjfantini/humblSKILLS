package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/selfupdate"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
)

// errUpgradeSkipped signals the user declined the upgrade confirmation
// prompt. It's not a failure — runUpgrade reports it as "skipped" rather
// than surfacing a Cobra error.
var errUpgradeSkipped = errors.New("upgrade skipped")

// osExecutable and brewRunner are test seams: production code always uses
// the zero-value defaults (os.Executable, and selfupdate.Upgrade's own
// exec.CommandContext), but command-level tests override them so they can
// exercise the Homebrew-detection and brew-invocation branches without
// touching the real running test binary or requiring a real `brew` on PATH.
var (
	osExecutable = os.Executable
	brewRunner   selfupdate.Runner
)

// applyPhaseSteps maps internal/selfupdate's download/verify/install/brew
// phases onto the upgrade command's own themed step list.
var applyPhaseSteps = map[selfupdate.Phase]tui.UpgradeStep{
	selfupdate.PhaseDownloading:   tui.UpgradeStepDownloading,
	selfupdate.PhaseVerifyingSum:  tui.UpgradeStepVerifyingChecksum,
	selfupdate.PhaseInstalling:    tui.UpgradeStepInstalling,
	selfupdate.PhaseBrewUpdating:  tui.UpgradeStepBrewUpdating,
	selfupdate.PhaseBrewUpgrading: tui.UpgradeStepBrewUpgrading,
}

type upgradeFlags struct {
	dryRun bool
}

// upgradeResult is both the --json payload and the source of truth the
// human-readable summary renders from, so the two presentations can never
// disagree about what actually happened.
type upgradeResult struct {
	CurrentVersion   string `json:"currentVersion"`
	LatestVersion    string `json:"latestVersion"`
	UpgradeAvailable bool   `json:"upgradeAvailable"`
	Homebrew         bool   `json:"homebrew,omitempty"`
	DryRun           bool   `json:"dryRun,omitempty"`
	Applied          bool   `json:"applied"`
	Skipped          bool   `json:"skipped,omitempty"`
	InstalledVersion string `json:"installedVersion,omitempty"`
}

func newUpgradeCmd(app *App) *cobra.Command {
	var f upgradeFlags
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade the humblskills CLI itself to the latest release",
		Long: "upgrade checks GitHub releases for a newer humblskills CLI build, " +
			"downloads it, verifies its checksum, and swaps it onto the running " +
			"binary's path. Installs managed by Homebrew are upgraded via " +
			"`brew upgrade humblskills` instead, so Homebrew's own bookkeeping " +
			"stays correct. --dry-run reports the version you'd upgrade to " +
			"without changing anything.\n\n" +
			"This upgrades the CLI binary itself. To upgrade installed skills, " +
			"use `humblskills update`.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runUpgrade(app, f)
		},
	}
	cmd.Flags().BoolVar(&f.dryRun, "dry-run", false, "show the version you'd upgrade to without changing anything")
	return cmd
}

func runUpgrade(app *App, f upgradeFlags) error {
	current := resolveVersion().Version
	exePath, err := osExecutable()
	if err != nil {
		return fmt.Errorf("resolve own executable path: %w", err)
	}

	client := selfupdate.NewHTTPClient()
	th := app.UI.Theme()

	var plan *selfupdate.Plan
	resolveErr := tui.RunWithSpinner(th, "checking latest release…", func() error {
		p, err := selfupdate.ResolvePlan(client, selfupdate.DefaultRepo, current, exePath, nil)
		plan = p
		return err
	})
	if resolveErr != nil {
		return fmt.Errorf("check latest release: %w", resolveErr)
	}

	res := upgradeResult{
		CurrentVersion:   current,
		LatestVersion:    plan.LatestVersion,
		UpgradeAvailable: plan.UpgradeAvailable,
		Homebrew:         plan.Homebrew,
		DryRun:           f.dryRun,
	}

	if !plan.UpgradeAvailable || f.dryRun {
		return finishUpgrade(app, res)
	}

	var installed string
	if plan.Homebrew {
		installed, err = applyHomebrewUpgrade(app, plan, exePath)
	} else {
		installed, err = applySelfUpgrade(app, client, plan, exePath)
	}
	if err != nil {
		if errors.Is(err, errUpgradeSkipped) {
			res.Skipped = true
			return finishUpgrade(app, res)
		}
		return err
	}

	res.Applied = true
	res.InstalledVersion = installed
	return finishUpgrade(app, res)
}

// finishUpgrade is the single place runUpgrade's outcome turns into output —
// JSON for --json, otherwise a themed one-liner. Called even after a
// TUI/confetti run closes, since alt-screen content doesn't persist in
// scrollback once the program exits.
func finishUpgrade(app *App, res upgradeResult) error {
	if app.Config.JSON {
		return app.UI.JSON(res)
	}
	switch {
	case res.Skipped:
		app.UI.Info("skipped — humblskills is still %s", versionTag(res.CurrentVersion))
	case res.Applied:
		app.UI.Success("humblskills is now %s", versionTag(res.InstalledVersion))
	case !res.UpgradeAvailable:
		app.UI.Success("humblskills is already up to date (%s)", versionTag(res.CurrentVersion))
	case res.DryRun && res.Homebrew:
		app.UI.Info("%s → %s available — Homebrew-managed install detected; would run `brew upgrade humblskills`", versionTag(res.CurrentVersion), versionTag(res.LatestVersion))
	case res.DryRun:
		app.UI.Info("%s → %s available — run without --dry-run to upgrade", versionTag(res.CurrentVersion), versionTag(res.LatestVersion))
	}
	return nil
}

// versionTag renders a version for display: "v2.17.0" for ordinary semver
// builds, but the bare string itself for non-numeric local builds ("dev")
// so messages don't read as the slightly silly "vdev".
func versionTag(v string) string {
	if v != "" && v[0] >= '0' && v[0] <= '9' {
		return "v" + v
	}
	return v
}

// applyHomebrewUpgrade defers to `brew upgrade humblskills` instead of
// self-downloading/swapping, so Homebrew's own Cellar bookkeeping stays
// correct. Returns the verified installed version on success.
func applyHomebrewUpgrade(app *App, plan *selfupdate.Plan, exePath string) (string, error) {
	app.UI.Info("Homebrew-managed install detected — upgrading via `brew upgrade humblskills`")

	ok, err := app.Prompt.Confirm("Run brew upgrade humblskills now?", true)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errUpgradeSkipped
	}

	steps := []tui.UpgradeStep{tui.UpgradeStepBrewUpdating, tui.UpgradeStepBrewUpgrading, tui.UpgradeStepVerifyingInstall}
	useTUI := tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes)

	var installed string
	run := func(sink func(tui.UpgradeEvent)) error {
		var stdout, stderr io.Writer = app.UI.Out(), app.UI.Err()
		if useTUI {
			// brew's own progress output would corrupt our alt-screen.
			stdout, stderr = io.Discard, io.Discard
		}
		brewSink := func(ev selfupdate.Event) {
			if step, ok := applyPhaseSteps[ev.Phase]; ok {
				sink(tui.UpgradeEvent{Step: step})
			}
		}
		if err := selfupdate.Upgrade(context.Background(), brewRunner, stdout, stderr, brewSink); err != nil {
			if errors.Is(err, selfupdate.ErrBrewNotFound) {
				return fmt.Errorf("brew not found on PATH — run `brew update && brew upgrade humblskills` yourself: %w", err)
			}
			return fmt.Errorf("brew upgrade humblskills: %w", err)
		}

		sink(tui.UpgradeEvent{Step: tui.UpgradeStepVerifyingInstall})
		v, err := selfupdate.VerifyInstalledVersion(exePath)
		if err != nil {
			return fmt.Errorf("verify installed version: %w", err)
		}
		if v != plan.LatestVersion {
			return fmt.Errorf("brew reported success but installed version is v%s, expected v%s", v, plan.LatestVersion)
		}
		installed = v
		return nil
	}

	if useTUI {
		if err := tui.ExecuteUpgrade(app.UI.Theme(), plan.CurrentVersion, plan.LatestVersion, steps, run); err != nil {
			return "", err
		}
		return installed, nil
	}
	if err := run(func(tui.UpgradeEvent) {}); err != nil {
		return "", err
	}
	return installed, nil
}

// applySelfUpgrade downloads, checksum-verifies, and swaps the CLI binary
// onto exePath. Returns the verified installed version on success.
func applySelfUpgrade(app *App, client *http.Client, plan *selfupdate.Plan, exePath string) (string, error) {
	ok, err := app.Prompt.Confirm(fmt.Sprintf("Upgrade humblskills %s → %s now?", versionTag(plan.CurrentVersion), versionTag(plan.LatestVersion)), true)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errUpgradeSkipped
	}

	steps := []tui.UpgradeStep{
		tui.UpgradeStepDownloading,
		tui.UpgradeStepVerifyingChecksum,
		tui.UpgradeStepInstalling,
		tui.UpgradeStepVerifyingInstall,
	}
	useTUI := tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes)
	cacheDir := filepath.Join(app.Config.CacheDir, "selfupdate")

	var installed string
	run := func(sink func(tui.UpgradeEvent)) error {
		applySink := func(ev selfupdate.Event) {
			if step, ok := applyPhaseSteps[ev.Phase]; ok {
				sink(tui.UpgradeEvent{Step: step})
			}
		}
		if err := selfupdate.Apply(client, plan, cacheDir, exePath, applySink); err != nil {
			if selfupdate.IsPermissionError(err) {
				return fmt.Errorf("%w — try `sudo humblskills upgrade`, or reinstall via Homebrew", err)
			}
			return err
		}

		sink(tui.UpgradeEvent{Step: tui.UpgradeStepVerifyingInstall})
		v, err := selfupdate.VerifyInstalledVersion(exePath)
		if err != nil {
			return fmt.Errorf("verify installed version: %w", err)
		}
		if v != plan.LatestVersion {
			return fmt.Errorf("swap succeeded but installed version is v%s, expected v%s", v, plan.LatestVersion)
		}
		installed = v
		return nil
	}

	if useTUI {
		if err := tui.ExecuteUpgrade(app.UI.Theme(), plan.CurrentVersion, plan.LatestVersion, steps, run); err != nil {
			return "", err
		}
		return installed, nil
	}
	if err := run(func(tui.UpgradeEvent) {}); err != nil {
		return "", err
	}
	return installed, nil
}
