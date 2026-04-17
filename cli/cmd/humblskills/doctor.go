package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/install"
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/platform"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
)

type doctorReport struct {
	Adapters []adapterReport `json:"adapters"`
	Manifest manifestReport  `json:"manifest"`
	Registry registryReport  `json:"registry"`
	Updates  updatesReport   `json:"updates"`
	Issues   []string        `json:"issues,omitempty"`
}

type updatesReport struct {
	Count  int      `json:"count"`
	Skills []string `json:"skills,omitempty"`
}

type adapterReport struct {
	Name     string         `json:"name"`
	Detected bool           `json:"detected"`
	Reason   string         `json:"reason"`
	Targets  []targetReport `json:"targets"`
}

type targetReport struct {
	Scope    string `json:"scope"`
	Path     string `json:"path"`
	Writable bool   `json:"writable"`
}

type manifestReport struct {
	Path          string `json:"path"`
	SchemaVersion int    `json:"schema_version"`
	Installs      int    `json:"installs"`
}

type registryReport struct {
	URL       string        `json:"url"`
	Source    string        `json:"source"`
	Cached    bool          `json:"cached"`
	FetchedAt time.Time     `json:"fetched_at,omitempty"`
	Age       time.Duration `json:"age_seconds,omitempty"`
	Skills    int           `json:"skills"`
	DepIssues []string      `json:"dep_issues,omitempty"`
	Error     string        `json:"error,omitempty"`
}

func newDoctorCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check environment, targets, manifest, and registry health",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDoctor(app)
		},
	}
}

func runDoctor(app *App) error {
	report := doctorReport{}

	adapters, err := app.Adapters()
	if err != nil {
		return fmt.Errorf("load adapters: %w", err)
	}

	results := platform.Detect(adapters)
	byName := make(map[string]platform.Adapter, len(adapters))
	for _, a := range adapters {
		byName[a.Name] = a
	}
	for _, r := range results {
		ar := adapterReport{Name: r.Name, Detected: r.Detected, Reason: r.Reason}
		for _, t := range byName[r.Name].Targets() {
			ar.Targets = append(ar.Targets, targetReport{Scope: t.Scope, Path: t.Path, Writable: t.Writable})
		}
		report.Adapters = append(report.Adapters, ar)
	}

	mpath := app.Config.ManifestPath
	m, mErr := manifest.Load(mpath)
	if mErr != nil {
		report.Issues = append(report.Issues, fmt.Sprintf("manifest: %s", mErr))
		report.Manifest = manifestReport{Path: mpath}
	} else {
		report.Manifest = manifestReport{
			Path:          mpath,
			SchemaVersion: m.SchemaVersion,
			Installs:      len(m.Installations),
		}
	}

	f := registry.NewFetcher(app.Config.RegistryURL, app.Config.CacheDir)
	reg, origin, rErr := f.Load()
	rr := registryReport{URL: app.Config.RegistryURL, Source: string(origin)}
	if rErr != nil {
		rr.Error = rErr.Error()
		report.Issues = append(report.Issues, fmt.Sprintf("registry: %s", rErr))
	} else {
		rr.Skills = len(reg.Skills)
		for _, issue := range registry.ValidateDeps(reg) {
			rr.DepIssues = append(rr.DepIssues, issue.Error())
		}
	}
	info := f.Inspect()
	rr.Cached = info.Exists
	rr.FetchedAt = info.FetchedAt
	rr.Age = info.Age
	report.Registry = rr

	if mErr == nil && rErr == nil {
		plans := install.PlanUpdates(reg, m, nil)
		report.Updates.Count = len(plans)
		for _, p := range plans {
			report.Updates.Skills = append(report.Updates.Skills, p.Skill)
		}
	}

	if app.Config.JSON {
		if err := app.UI.JSON(report); err != nil {
			return err
		}
		if hasFailures(report) {
			return errDoctorFailed
		}
		return nil
	}

	printDoctor(app, report)
	if hasFailures(report) {
		return errDoctorFailed
	}
	return nil
}

var errDoctorFailed = errors.New("doctor found issues")

func hasFailures(r doctorReport) bool {
	if len(r.Issues) > 0 {
		return true
	}
	if r.Registry.Error != "" || len(r.Registry.DepIssues) > 0 {
		return true
	}
	return false
}

func printDoctor(app *App, r doctorReport) {
	app.UI.Info("Adapters:")
	anyDetected := false
	for _, a := range r.Adapters {
		if a.Detected {
			anyDetected = true
			app.UI.Success("%s — %s", a.Name, a.Reason)
		} else {
			app.UI.Detail("  %s (not detected): %s", a.Name, a.Reason)
		}
		for _, t := range a.Targets {
			mark := "rw"
			if !t.Writable {
				mark = "ro"
			}
			app.UI.Detail("    %s [%s] %s", t.Scope, mark, t.Path)
		}
	}
	if !anyDetected {
		app.UI.Warn("no agent platform detected — run inside a project that uses Claude Code, Cursor, etc.")
	}

	app.UI.Info("")
	app.UI.Info("Manifest:")
	if _, err := os.Stat(r.Manifest.Path); err == nil {
		app.UI.Info("  %s (%d install%s)", r.Manifest.Path, r.Manifest.Installs, plural(r.Manifest.Installs))
	} else {
		app.UI.Detail("  %s (not yet created)", r.Manifest.Path)
	}

	app.UI.Info("")
	app.UI.Info("Registry:")
	app.UI.Info("  URL: %s", r.Registry.URL)
	if r.Registry.Error != "" {
		app.UI.Error("registry unreachable: %s", r.Registry.Error)
	} else {
		app.UI.Success("%d skill%s available (via %s)", r.Registry.Skills, plural(r.Registry.Skills), r.Registry.Source)
		if r.Registry.Cached {
			app.UI.Detail("  cache fetched %s ago", r.Registry.Age.Round(time.Second))
		}
		for _, issue := range r.Registry.DepIssues {
			app.UI.Warn("dep issue: %s", issue)
		}
	}

	if r.Updates.Count > 0 {
		app.UI.Info("")
		app.UI.Warn("%d skill%s can be updated — run 'humblskills update'", r.Updates.Count, plural(r.Updates.Count))
		for _, name := range r.Updates.Skills {
			app.UI.Detail("  %s", name)
		}
	}

	for _, i := range r.Issues {
		app.UI.Warn(i)
	}
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
