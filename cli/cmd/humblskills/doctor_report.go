package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jjfantini/humblSKILLS/cli/internal/adapters"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/evalruntime"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/workspace"
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
	"github.com/jjfantini/humblSKILLS/cli/internal/secrets"
)

// This file is the doctor "service" layer: the report data model plus the
// logic that gathers it. Rendering (TUI list items + static output) lives in
// doctor_view.go; command wiring lives in doctor.go.

type doctorReport struct {
	Adapters   []adapterReport  `json:"adapters"`
	Manifest   manifestReport   `json:"manifest"`
	Registries []registryReport `json:"registries"`
	Updates    updatesReport    `json:"updates"`
	Eval       evalReport       `json:"eval"`
	Issues     []string         `json:"issues,omitempty"`
}

// evalReport is the eval-prerequisite block: per-runner availability,
// workspace writability, per-provider API key source (never the value),
// and the count of installed skills with evals/ on disk.
type evalReport struct {
	Runners       []runnerCheck  `json:"runners"`
	Workspace     workspaceCheck `json:"workspace"`
	APIKeys       []apiKeyCheck  `json:"api_keys"`
	DefaultRunner string         `json:"default_runner"`
	EvalSkills    int            `json:"eval_skills"`
}

type runnerCheck struct {
	Name        string `json:"name"`
	Available   bool   `json:"available"`
	Version     string `json:"version,omitempty"`
	Reason      string `json:"reason,omitempty"`
	Fix         string `json:"fix,omitempty"`
	RequiresKey string `json:"requires_key,omitempty"`
}

type workspaceCheck struct {
	Path           string `json:"path"`
	Exists         bool   `json:"exists"`
	Writable       bool   `json:"writable"`
	SizeBytes      int64  `json:"size_bytes"`
	IterationCount int    `json:"iteration_count"`
	SkillsWithRuns int    `json:"skills_with_runs"`
}

type apiKeyCheck struct {
	Provider string `json:"provider"`
	Present  bool   `json:"present"`
	Source   string `json:"source"` // "env" | "keyring" | "file" | "absent"
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
	Name      string        `json:"name"`
	URL       string        `json:"url"`
	Source    string        `json:"source"`
	Cached    bool          `json:"cached"`
	FetchedAt time.Time     `json:"fetched_at,omitempty"`
	Age       time.Duration `json:"age_seconds,omitempty"`
	Skills    int           `json:"skills"`
	DepIssues []string      `json:"dep_issues,omitempty"`
	Error     string        `json:"error,omitempty"`
}

func buildDoctorReport(app *App) (doctorReport, error) {
	report := doctorReport{}
	adapterList, err := app.Adapters()
	if err != nil {
		return report, fmt.Errorf("load adapters: %w", err)
	}

	results := adapters.Detect(adapterList)
	byName := make(map[string]adapters.Adapter, len(adapterList))
	for _, a := range adapterList {
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

	// One report per configured registry. No spinner here - the caller
	// (runDoctor) already wraps this whole function in one loading screen.
	ix := newSkillIndex()
	anyRegOK := false
	for _, r := range app.resolvedRegistries() {
		f := app.fetcherForRegistry(r)
		reg, origin, rErr := f.Load()
		rr := registryReport{Name: r.Name, URL: r.URL, Source: string(origin)}
		if rErr != nil {
			rr.Error = rErr.Error()
			report.Issues = append(report.Issues, fmt.Sprintf("registry %q: %s", r.Name, rErr))
		} else {
			anyRegOK = true
			rr.Skills = len(reg.Skills)
			ix.add(r.Name, reg.Skills)
			for _, issue := range registry.ValidateDeps(reg) {
				rr.DepIssues = append(rr.DepIssues, issue.Error())
			}
		}
		info := f.Inspect()
		rr.Cached = info.Exists
		rr.FetchedAt = info.FetchedAt
		rr.Age = info.Age
		report.Registries = append(report.Registries, rr)
	}

	// Update count: each installed skill checked against its ORIGIN registry.
	if mErr == nil && anyRegOK {
		seen := map[string]bool{}
		for _, inst := range m.Installations {
			if seen[inst.Skill] {
				continue
			}
			origin := inst.RegistryName
			if origin == "" {
				origin = ix.registryOf(inst.Skill)
			}
			rs, ok := ix.find(origin, inst.Skill)
			if ok && (inst.Version != rs.Version || inst.RegistryRef != rs.DirSHA) {
				seen[inst.Skill] = true
				report.Updates.Count++
				report.Updates.Skills = append(report.Updates.Skills, inst.Skill)
			}
		}
	}

	report.Eval = buildEvalReport(app)
	return report, nil
}

// buildEvalReport collects per-runner availability + API-key presence +
// workspace writability for the doctor report. Never touches the wire -
// runner DoctorCheck is designed to be a fast local probe.
func buildEvalReport(app *App) evalReport {
	store, _ := secrets.NewStore("")
	reg := evalruntime.DefaultRegistry(store)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	det := reg.Detect(ctx)

	er := evalReport{}
	for _, d := range det {
		er.Runners = append(er.Runners, runnerCheck{
			Name:        d.Name,
			Available:   d.Check.Available,
			Version:     d.Check.Version,
			Reason:      d.Check.Reason,
			Fix:         d.Check.Fix,
			RequiresKey: d.Check.RequiresKey,
		})
		if er.DefaultRunner == "" && d.Check.Available {
			er.DefaultRunner = d.Name
		}
	}
	// API keys.
	if store != nil {
		for _, p := range secrets.Providers() {
			_, src, err := store.Get(p.Name)
			present := err == nil && src != secrets.SourceAbsent
			er.APIKeys = append(er.APIKeys, apiKeyCheck{
				Provider: p.Name,
				Present:  present,
				Source:   string(src),
			})
		}
	}
	// Workspace.
	wsRoot := resolveWorkspace(app, "")
	ws := workspaceCheck{Path: wsRoot}
	if fi, err := os.Stat(wsRoot); err == nil && fi.IsDir() {
		ws.Exists = true
		probe := filepath.Join(wsRoot, ".writable-probe")
		if err := os.WriteFile(probe, []byte("ok"), 0o644); err == nil {
			_ = os.Remove(probe)
			ws.Writable = true
		}
		ws.SizeBytes, _ = workspace.SizeBytes(wsRoot)
		skills, _ := workspace.ListSkills(wsRoot)
		ws.SkillsWithRuns = len(skills)
		for _, s := range skills {
			if n, _ := workspace.MaxIterationN(wsRoot, s); n > 0 {
				ws.IterationCount += n
			}
		}
	} else {
		// Non-existent is expected on first run; writability reflects the
		// parent dir's state so we can surface real permission issues.
		if parent := filepath.Dir(wsRoot); parent != "" {
			if _, err := os.Stat(parent); err == nil {
				ws.Writable = true
			}
		}
	}
	er.Workspace = ws
	// Count installed skills with evals/ directories.
	m, err := manifest.Load(app.Config.ManifestPath)
	if err == nil && m != nil {
		for _, inst := range m.Installations {
			if _, err := os.Stat(filepath.Join(inst.Path, "evals")); err == nil {
				er.EvalSkills++
			}
		}
	}
	return er
}

var errDoctorFailed = errors.New("doctor found issues")

func hasFailures(r doctorReport) bool {
	if len(r.Issues) > 0 {
		return true
	}
	for _, rr := range r.Registries {
		if rr.Error != "" || len(rr.DepIssues) > 0 {
			return true
		}
	}
	// Eval: default runner must be available for CI to claim "healthy".
	if r.Eval.DefaultRunner == "" && len(r.Eval.Runners) > 0 {
		return true
	}
	return false
}
