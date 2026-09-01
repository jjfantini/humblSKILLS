package main

import (
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/selfupdate"
	"github.com/jjfantini/humblSKILLS/cli/v2/internal/tui"
)

// checkUpgradeNotice resolves the profile/--channel stream the same way
// upgrade does and asks selfupdate.Check whether a newer GitHub release
// exists. JSON and quiet stay silent; failures stay silent.
func (a *App) checkUpgradeNotice() selfupdate.Notice {
	if a == nil || a.Config.JSON || a.Config.Quiet {
		return selfupdate.Notice{}
	}
	exePath, err := osExecutable()
	if err != nil {
		return selfupdate.Notice{}
	}
	return selfupdate.Check(selfupdate.CheckOptions{
		Client:         selfupdate.NewNoticeHTTPClient(),
		CurrentVersion: resolveVersion().Version,
		ExePath:        exePath,
		Channel:        a.resolvedNoticeChannel(),
		CacheDir:       a.Config.CacheDir,
	})
}

// resolvedNoticeChannel is upgrade --channel → profile → stable. Profile
// load errors fall back to stable so a notice check never fails a command.
func (a *App) resolvedNoticeChannel() string {
	if flag := a.upgradeChannelFlag(); flag != "" {
		return flag
	}
	p, err := profile.Load(a.Config.ProfilePath)
	if err != nil || p == nil {
		return profile.ChannelStable
	}
	return p.ResolvedChannel()
}

// upgradeChannelFlag is the persistent --channel override when present.
// Empty means "use the profile".
func (a *App) upgradeChannelFlag() string {
	return a.Config.Channel
}

// printUpgradeNotice writes the CLI banner to stderr (via Warn) when a
// newer channel release exists. --json and --quiet suppress it.
func (a *App) printUpgradeNotice() {
	n := a.checkUpgradeNotice()
	if !n.Available {
		return
	}
	a.UI.Warn("%s", n.CLILine())
}

// tuiVersionNotice is the dashboard banner payload, or nil when quiet.
func (a *App) tuiVersionNotice() *tui.VersionNotice {
	n := a.checkUpgradeNotice()
	if !n.Available {
		return nil
	}
	return &tui.VersionNotice{
		Current: selfupdate.DisplayVersion(n.CurrentVersion),
		Latest:  selfupdate.DisplayVersion(n.LatestVersion),
		Channel: n.Channel,
		Command: n.UpdateCommand(),
	}
}
