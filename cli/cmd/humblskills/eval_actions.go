package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/evalruntime"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/grader"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/grader/anthropicjudge"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/harness"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/metrics"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/report"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/workspace"
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/secrets"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
)

// This file holds the run/behavior logic for the eval command family. Cobra
// wiring lives in eval.go; shared helpers in eval_helpers.go.

func runEvalRun(app *App, skill string, f evalRunFlags) error {
	skillDir, evalsFile, err := resolveSkill(app, skill)
	if err != nil {
		return err
	}
	if evalsFile == nil {
		return fmt.Errorf("skill %q has no evals/ directory — run `humblskills eval init %s` to scaffold one", skill, skill)
	}
	store, err := secrets.NewStore("")
	if err != nil {
		return err
	}
	reg := evalruntime.DefaultRegistry(store)
	rn, err := pickRunner(reg, f.runnerName)
	if err != nil {
		return err
	}
	ws := resolveWorkspace(app, f.workspace)
	arms := f.configs
	if len(arms) == 0 {
		arms = evalsFile.Configurations
	}
	runs := f.runs
	if runs == 0 {
		runs = evalsFile.RunsPerConfiguration
	}
	// LLM judge: construct one when an Anthropic key is available
	// (env / keyring / file). "llm" assertions otherwise auto-fail with a
	// clear evidence note, which is the documented no-judge behaviour.
	var judge grader.LLMJudge
	if key, _, err := store.Get("anthropic"); err == nil && key != "" {
		judge = anthropicjudge.New(key, f.grader)
		if !app.Config.JSON {
			model := f.grader
			if model == "" {
				model = anthropicjudge.DefaultModel
			}
			app.UI.Detail("LLM judge: %s", model)
		}
	} else if !app.Config.JSON {
		app.UI.Detail("LLM judge: disabled (no anthropic key) — 'llm' assertions will auto-fail")
	}

	opts := harness.Options{
		SkillDir:             skillDir,
		Scenarios:            evalsFile,
		Arms:                 arms,
		Runner:               rn,
		Grader:               judge,
		WorkspaceRoot:        ws,
		RunsPerConfiguration: runs,
		Parallel:             f.parallel,
		ScenarioFilter:       f.scenarioIDs,
	}
	h, err := harness.New(opts)
	if err != nil {
		return err
	}
	// Drain events into the Printer when not in JSON mode so users see progress.
	go func() {
		for ev := range h.Events() {
			if app.Config.JSON {
				continue
			}
			switch ev.Kind {
			case harness.EvtSession:
				app.UI.Detail("▸ %s · %s · session %d (run %d): %s",
					ev.Arm, ev.Scenario, ev.Session, ev.Run, abbreviate(ev.Message, 80))
			case harness.EvtSessionDone:
				app.UI.Detail("  done (%d/%d · %d tokens · %.0fs)",
					ev.Totals.CompletedSessions, ev.Totals.TotalSessions,
					ev.Totals.Tokens, ev.Totals.ElapsedSec)
			case harness.EvtError:
				app.UI.Warn("%s · %s session %d: %s", ev.Arm, ev.Scenario, ev.Session, ev.Message)
			}
		}
	}()
	res, err := h.Run(context.Background())
	if err != nil {
		return err
	}
	if app.Config.JSON {
		return app.UI.JSON(map[string]any{
			"iteration":   res.Iteration,
			"iter_dir":    res.IterDir,
			"report_html": res.ReportHTML,
			"report_md":   res.ReportMD,
			"report_json": res.ReportJSON,
			"benchmark":   res.Benchmark,
		})
	}
	app.UI.Success("iteration-%d complete → %s", res.Iteration, res.IterDir)
	app.UI.Info("report: %s", res.ReportHTML)
	printBenchmarkSummary(app, res.Benchmark)
	if f.open {
		_ = openInBrowser(res.ReportHTML)
	}
	if f.minPassRate > 0 {
		if b, ok := res.Benchmark.RunSummary["smart_skill"]; ok && b.PassRate.Mean < f.minPassRate {
			return fmt.Errorf("smart_skill pass_rate %.3f < min %.3f", b.PassRate.Mean, f.minPassRate)
		}
	}
	return nil
}

func printBenchmarkSummary(app *App, b *metrics.Benchmark) {
	if b == nil {
		return
	}
	arms := make([]string, 0, len(b.RunSummary))
	for a := range b.RunSummary {
		arms = append(arms, a)
	}
	sort.Strings(arms)
	app.UI.Section("cross-section")
	for _, a := range arms {
		s := b.RunSummary[a]
		app.UI.Info("  %-14s pass %.3f · tokens %d · time %.1fs",
			a, s.PassRate.Mean, int(s.Tokens.Mean), s.TimeSeconds.Mean)
	}
	if len(b.Delta) > 0 {
		app.UI.Section("deltas")
		for _, k := range sortedStrKeys(b.Delta) {
			d := b.Delta[k]
			app.UI.Info("  %-14s Δ pass %+.3f · Δ tokens %+d · Δ time %+.1fs",
				k, d.PassRate, int(d.Tokens), d.TimeSeconds)
		}
	}
}

func runEvalInit(app *App, skill string) error {
	skillDir, _, err := resolveSkill(app, skill)
	if err != nil {
		return err
	}
	return scaffoldEvalsDir(filepath.Join(skillDir, "evals"), skillBasename(skillDir))
}

func runEvalAggregate(app *App, iterDir string) error {
	// Walk grading.json files and synthesize trajectory rows.
	var rows []metrics.TrajectoryRow
	_ = filepath.Walk(iterDir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if filepath.Base(p) != "grading.json" {
			return nil
		}
		var g struct {
			Summary struct {
				PassRate float64 `json:"pass_rate"`
			}
		}
		if data, e := os.ReadFile(p); e == nil {
			_ = json.Unmarshal(data, &g)
		}
		// Parse path to recover arm/scenario/session.
		rel, _ := filepath.Rel(iterDir, p)
		parts := strings.Split(rel, string(os.PathSeparator))
		if len(parts) < 3 {
			return nil
		}
		row := metrics.TrajectoryRow{
			Arm:      parts[0],
			Scenario: parts[1],
			PassRate: g.Summary.PassRate,
		}
		// Pull timing.json from the session dir.
		if data, err := os.ReadFile(filepath.Join(filepath.Dir(p), "timing.json")); err == nil {
			var t struct {
				TotalTokens int     `json:"total_tokens"`
				DurationMs  int64   `json:"duration_ms"`
				CostUSD     float64 `json:"cost_usd"`
			}
			_ = json.Unmarshal(data, &t)
			row.Tokens = t.TotalTokens
			row.DurationMs = t.DurationMs
			row.CostUSD = t.CostUSD
		}
		rows = append(rows, row)
		return nil
	})
	traj := metrics.AggregateTrajectory("", "", rows)
	bench := metrics.AggregateBenchmark("", 0, rows)
	if err := metrics.Write(filepath.Join(iterDir, "trajectory.json"), traj); err != nil {
		return err
	}
	if err := metrics.Write(filepath.Join(iterDir, "benchmark.json"), bench); err != nil {
		return err
	}
	app.UI.Success("rebuilt trajectory.json / benchmark.json from %d grading files", len(rows))
	return nil
}

func runEvalReport(app *App, iterDir string, open bool) error {
	traj, err := loadJSON[metrics.Trajectory](filepath.Join(iterDir, "trajectory.json"))
	if err != nil {
		return err
	}
	bench, err := loadJSON[metrics.Benchmark](filepath.Join(iterDir, "benchmark.json"))
	if err != nil {
		return err
	}
	html, _, _, err := report.RenderAll(iterDir, &report.Bundle{
		SkillName:   traj.SkillName,
		Iteration:   bench.Iteration,
		Runner:      traj.Runner,
		ScenarioIDs: report.InferScenarioIDs(traj.Rows),
		Trajectory:  traj,
		Benchmark:   bench,
	})
	if err != nil {
		return err
	}
	app.UI.Success("rebuilt report at %s", html)
	if open {
		_ = openInBrowser(html)
	}
	return nil
}

func runEvalCompare(app *App, a, b string) error {
	ba, err := loadJSON[metrics.Benchmark](filepath.Join(a, "benchmark.json"))
	if err != nil {
		return err
	}
	bb, err := loadJSON[metrics.Benchmark](filepath.Join(b, "benchmark.json"))
	if err != nil {
		return err
	}
	if app.Config.JSON {
		return app.UI.JSON(map[string]any{"a": ba, "b": bb})
	}
	app.UI.Section(fmt.Sprintf("compare %s vs %s", a, b))
	arms := unionKeys(ba.RunSummary, bb.RunSummary)
	for _, arm := range arms {
		aS := ba.RunSummary[arm]
		bS := bb.RunSummary[arm]
		app.UI.Info("  %-14s pass %.3f → %.3f (Δ %+.3f) · tokens %d → %d (Δ %+d)",
			arm, aS.PassRate.Mean, bS.PassRate.Mean, bS.PassRate.Mean-aS.PassRate.Mean,
			int(aS.Tokens.Mean), int(bS.Tokens.Mean), int(bS.Tokens.Mean-aS.Tokens.Mean))
	}
	return nil
}

func runEvalRunners(app *App) error {
	store, err := secrets.NewStore("")
	if err != nil {
		return err
	}
	reg := evalruntime.DefaultRegistry(store)
	det := reg.Detect(context.Background())
	if app.Config.JSON {
		return app.UI.JSON(det)
	}
	app.UI.Section("runners")
	for _, r := range det {
		status := "missing"
		if r.Check.Available {
			status = "ready"
		}
		app.UI.Info("  %-16s %s  %s", r.Name, status, firstNonEmptyStr(r.Check.Version, r.Check.Reason))
		if !r.Check.Available && r.Check.Fix != "" {
			app.UI.Detail("    fix: %s", r.Check.Fix)
		}
	}
	return nil
}

func runEvalSetKey(app *App, provider string) error {
	if _, ok := secrets.ProviderByName(provider); !ok {
		return fmt.Errorf("unknown provider %q — known: %v", provider, knownProviders())
	}
	val, err := app.Prompt.Secret(fmt.Sprintf("API key for %s", provider))
	if err != nil {
		return err
	}
	if strings.TrimSpace(val) == "" {
		return errors.New("refusing to store empty secret")
	}
	store, err := secrets.NewStore("")
	if err != nil {
		return err
	}
	src, err := store.Set(provider, val)
	if err != nil {
		return err
	}
	app.UI.Success("stored %s key in %s", provider, src)
	return nil
}

func runEvalLs(app *App, skill string) error {
	ws := resolveWorkspace(app, "")
	skills := []string{skill}
	if skill == "" {
		all, err := workspace.ListSkills(ws)
		if err != nil {
			return err
		}
		skills = all
	}
	for _, s := range skills {
		reg, err := workspace.LoadRegistry(ws, s)
		if err != nil {
			return err
		}
		app.UI.Section(s)
		for _, it := range reg.Iterations {
			size, _ := workspace.SizeBytes(workspace.IterationDir(ws, s, it.N))
			passStr := ""
			if smart, ok := it.HeadlinePassRt["smart_skill"]; ok {
				passStr = fmt.Sprintf("smart %.2f ", smart)
			}
			app.UI.Info("  iteration-%-3d %s %s %s%s",
				it.N, it.Status, it.StartedAt.Format("2006-01-02 15:04"),
				passStr, workspace.HumanSize(size))
		}
	}
	return nil
}

// --- TUI default entry (no-arg `humblskills eval`) -------------------------

func runEvalTUI(app *App) error {
	if !tui.ShouldUseTUI(app.Config.JSON, app.Config.Quiet, app.Config.Yes) {
		return runEvalRunners(app)
	}
	// TUI defers to tui.RunEvalHome when present; fall back to runners
	// listing until the TUI screen lands (same package).
	return tui.RunEvalHomeOr(app.UI.Theme(), app.headerSection("Eval"),
		func() (tui.EvalHomeData, error) {
			return buildEvalHomeData(app)
		},
		func(skill string) error {
			// Between Home and Run: open the Config Modal so the user
			// picks arms / scenarios / runner / runs / parallel.
			choices, err := runEvalConfigModal(app, skill)
			if err != nil {
				return err
			}
			if !choices.Confirmed {
				return nil
			}
			flags := evalRunFlags{
				configs:     choices.Arms,
				scenarioIDs: choices.Scenarios,
				runnerName:  choices.Runner,
				runs:        choices.Runs,
				parallel:    choices.Parallel,
				open:        true,
			}
			return runEvalRun(app, skill, flags)
		})
}

func runEvalConfigModal(app *App, skill string) (tui.EvalConfigChoices, error) {
	skillDir, file, err := resolveSkill(app, skill)
	if err != nil {
		return tui.EvalConfigChoices{}, err
	}
	if file == nil {
		return tui.EvalConfigChoices{}, fmt.Errorf("skill %q has no evals/ — run `humblskills eval init %s`", skill, skill)
	}
	_ = skillDir
	// Detect runners.
	store, err := secrets.NewStore("")
	if err != nil {
		return tui.EvalConfigChoices{}, err
	}
	reg := evalruntime.DefaultRegistry(store)
	det := reg.Detect(context.Background())
	runners := make([]tui.EvalRunnerEntry, 0, len(det))
	for _, d := range det {
		runners = append(runners, tui.EvalRunnerEntry{
			Name: d.Name, Available: d.Check.Available,
			Version: d.Check.Version, Reason: d.Check.Reason,
		})
	}
	scens := make([]tui.EvalScenarioEntry, 0, len(file.Scenarios))
	for _, s := range file.Scenarios {
		scens = append(scens, tui.EvalScenarioEntry{
			ID: s.ID, Family: s.Family, Sessions: len(s.Sessions),
		})
	}
	inputs := tui.EvalConfigInputs{
		Skill:      skill,
		Arms:       file.Configurations,
		Scenarios:  scens,
		Runners:    runners,
		Runs:       []int{1, 3, 5, 10},
		Parallel:   []int{1, 2, 4, 8},
		DefaultRun: file.RunsPerConfiguration,
		DefaultPar: 1,
	}
	return tui.RunEvalConfigModal(app.UI.Theme(), app.headerSection("Eval > Configure"), inputs)
}

// buildEvalHomeData scans the workspace to populate the TUI Home screen.
func buildEvalHomeData(app *App) (tui.EvalHomeData, error) {
	ws := resolveWorkspace(app, "")
	var items []tui.EvalHomeItem
	// List skills with evals/ on disk (registry-manifest + ./skills/).
	candidates := []string{}
	m, err := manifest.Load(app.Config.ManifestPath)
	if err == nil && m != nil {
		for _, inst := range m.Installations {
			candidates = append(candidates, inst.Skill)
		}
	}
	// Also include ./skills/<name>/ for dev checkouts.
	if entries, err := os.ReadDir("skills"); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				candidates = append(candidates, e.Name())
			}
		}
	}
	seen := map[string]bool{}
	for _, name := range candidates {
		if seen[name] {
			continue
		}
		seen[name] = true
		skillDir, f, err := resolveSkill(app, name)
		if err != nil {
			continue
		}
		it := tui.EvalHomeItem{Skill: name, SkillDir: skillDir}
		if f != nil {
			it.HasEvals = true
			it.ScenarioCount = len(f.Scenarios)
			it.Configurations = f.Configurations
		}
		reg, err := workspace.LoadRegistry(ws, name)
		if err == nil {
			it.IterationCount = len(reg.Iterations)
			if len(reg.Iterations) > 0 {
				last := reg.Iterations[len(reg.Iterations)-1]
				it.LastRun = last.StartedAt
				it.LastPassRates = last.HeadlinePassRt
				it.Runner = last.Runner
			}
		}
		items = append(items, it)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Skill < items[j].Skill })
	return tui.EvalHomeData{
		Workspace: ws,
		Items:     items,
	}, nil
}
