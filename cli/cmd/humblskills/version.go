package main

import (
	"runtime/debug"

	"github.com/spf13/cobra"
)

// These are optionally overridden via -ldflags at release time.
var (
	version = ""
	commit  = ""
)

type versionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Dirty   bool   `json:"dirty,omitempty"`
}

func newVersionCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print humblskills version and commit",
		RunE: func(cmd *cobra.Command, _ []string) error {
			info := resolveVersion()
			if app.Config.JSON {
				return app.UI.JSON(info)
			}
			suffix := ""
			if info.Dirty {
				suffix = " (dirty)"
			}
			app.UI.Info("humblskills %s  commit %s%s", info.Version, info.Commit, suffix)
			return nil
		},
	}
}

func resolveVersion() versionInfo {
	info := versionInfo{Version: "dev", Commit: "unknown"}
	if version != "" {
		info.Version = version
	}
	if commit != "" {
		info.Commit = commit
	}
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return info
	}
	if version == "" && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		info.Version = bi.Main.Version
	}
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			if commit == "" && s.Value != "" {
				info.Commit = shortCommit(s.Value)
			}
		case "vcs.modified":
			if s.Value == "true" {
				info.Dirty = true
			}
		}
	}
	return info
}

func shortCommit(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}
