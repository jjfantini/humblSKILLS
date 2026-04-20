package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/evalruntime"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/grader"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/grader/anthropicjudge"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/harness"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/metrics"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/report"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/scenarios"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/workspace"
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/profile"
	"github.com/jjfantini/humblSKILLS/cli/internal/secrets"
	"github.com/jjfantini/humblSKILLS/cli/internal/tui"
)

func newEvalCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "eval",
		Short: "Benchmark skills across no_skill / flat_skill / smart_skill arms",
		Long: "eval runs a full cross-sectional + longitudinal benchmark of a skill " +
			"against three arms (no / flat / smart), grades the outputs, and emits " +
			"a single-file HTML dashboard. Six runners available (claudecode, " +
			"cursor-agent, codex, anthropic-api, openai-api, mock).",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runEvalTUI(app)
		},
	}
	cmd.AddCommand(
		newEvalRunCmd(app),
		newEvalInitCmd(app),
		newEvalGradeCmd(app),
		newEvalAggregateCmd(app),
		newEvalReportCmd(app),
		newEvalCompareCmd(app),
		newEvalShowcaseCmd(app),
		newEvalRunnersCmd(app),
		newEvalSetKeyCmd(app),
		newEvalLsCmd(app),
		newEvalPruneCmd(app),
		newEvalWhereCmd(app),
	)
	return cmd
}

// --- flag bag ---------------------------------------------------------------

type evalRunFlags struct {
	scenarioIDs  []string
	configs      []string
	sessionsCap  int
	runs         int
	parallel     int
	workspace    string
	runnerName   string
	executor     string
	grader       string
	noGrade      bool
	noReport     bool
	resume       bool
	minPassRate  float64
	open         bool
}

// --- eval run ---------------------------------------------------------------

func newEvalRunCmd(app *App) *cobra.Command {
	var f evalRunFlags
	cmd := &cobra.Command{
		Use:   "run <skill>",
		Short: "Run the full eval pipeline for a skill",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEvalRun(app, args[0], f)
		},
	}
	cmd.Flags().StringSliceVar(&f.scenarioIDs, "scenario", nil, "subset of scenario IDs to run (default: all)")
	cmd.Flags().StringSliceVar(&f.configs, "config", nil, "subset of arms: smart_skill, flat_skill, no_skill (default: all)")
	cmd.Flags().IntVar(&f.sessionsCap, "sessions", 0, "cap session count per scenario (0 = all)")
	cmd.Flags().IntVar(&f.runs, "runs", 0, "runs per configuration (0 = scenarios.json default)")
	cmd.Flags().IntVar(&f.parallel, "parallel", 0, "parallel workers (0 = 1 — sequential preserves longitudinal ordering)")
	cmd.Flags().StringVar(&f.workspace, "workspace", "", "workspace root (default: XDG_STATE_HOME/humblskills/evals)")
	cmd.Flags().StringVar(&f.runnerName, "runner", "", "runner id (claudecode|cursor-agent|codex|anthropic-api|openai-api|mock; auto-detects when empty)")
	cmd.Flags().StringVar(&f.executor, "executor-model", "", "model for the executor runner (default: runner's DefaultModel)")
	cmd.Flags().StringVar(&f.grader, "grader-model", "", "model for LLM-judge grading (default: anthropic opus-4-5)")
	cmd.Flags().BoolVar(&f.noGrade, "no-grade", false, "skip grading (useful to batch grade later)")
	cmd.Flags().BoolVar(&f.noReport, "no-report", false, "skip report rendering")
	cmd.Flags().BoolVar(&f.resume, "resume", false, "(reserved) resume an interrupted iteration")
	cmd.Flags().Float64Var(&f.minPassRate, "min-pass-rate", 0, "exit non-zero if mean pass_rate on the primary arm is below this")
	cmd.Flags().BoolVar(&f.open, "open", false, "open report.html in the OS default browser after completion")
	return cmd
}

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

// --- eval init --------------------------------------------------------------

func newEvalInitCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "init [skill]",
		Short: "Scaffold an evals/ directory inside a skill",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var skill string
			if len(args) == 1 {
				skill = args[0]
			}
			return runEvalInit(app, skill)
		},
	}
}

func runEvalInit(app *App, skill string) error {
	skillDir, _, err := resolveSkill(app, skill)
	if err != nil {
		return err
	}
	return scaffoldEvalsDir(filepath.Join(skillDir, "evals"), skillBasename(skillDir))
}

// --- eval grade / aggregate / report ---------------------------------------

func newEvalGradeCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "grade <iteration-dir>",
		Short: "Re-grade every session in an existing iteration directory",
		Args:  cobra.ExactArgs(1),
		RunE:  func(cmd *cobra.Command, args []string) error { return fmt.Errorf("grade: not implemented in v1 — re-run `eval run` to regrade") },
	}
}

func newEvalAggregateCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "aggregate <iteration-dir>",
		Short: "Rebuild trajectory.json / benchmark.json from grading files",
		Args:  cobra.ExactArgs(1),
		RunE:  func(cmd *cobra.Command, args []string) error { return runEvalAggregate(app, args[0]) },
	}
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
	traj := metrics.AggregateTrajectory("", rows)
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

func newEvalReportCmd(app *App) *cobra.Command {
	var open bool
	cmd := &cobra.Command{
		Use:   "report <iteration-dir>",
		Short: "Rebuild report.{html,md,json} for an existing iteration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEvalReport(app, args[0], open)
		},
	}
	cmd.Flags().BoolVar(&open, "open", false, "open report.html in the default browser")
	return cmd
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
		SkillName:  traj.SkillName,
		Iteration:  bench.Iteration,
		Trajectory: traj,
		Benchmark:  bench,
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

func newEvalCompareCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "compare <iter-a> <iter-b>",
		Short: "Diff two iterations on pass_rate + tokens + time",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEvalCompare(app, args[0], args[1])
		},
	}
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

func newEvalShowcaseCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "showcase",
		Short: "Run the canonical demo on use-smart-skill",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runEvalRun(app, "use-smart-skill", evalRunFlags{open: true})
		},
	}
}

// --- eval runners -----------------------------------------------------------

func newEvalRunnersCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "runners",
		Short: "List runner availability (same data as `doctor` Eval section)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runEvalRunners(app)
		},
	}
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

// --- eval set-key -----------------------------------------------------------

func newEvalSetKeyCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "set-key <provider>",
		Short: "Store an API key via the OS keyring (anthropic, openai, ...)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEvalSetKey(app, args[0])
		},
	}
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

// --- eval ls + prune + where ------------------------------------------------

func newEvalLsCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "ls [skill]",
		Short: "List eval iterations per skill",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skill := ""
			if len(args) == 1 {
				skill = args[0]
			}
			return runEvalLs(app, skill)
		},
	}
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

func newEvalPruneCmd(app *App) *cobra.Command {
	var (
		keepLast  int
		olderThan time.Duration
		dryRun    bool
		all       bool
	)
	cmd := &cobra.Command{
		Use:   "prune <skill>",
		Short: "Retention policy for eval iterations",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws := resolveWorkspace(app, "")
			res, err := workspace.Prune(ws, args[0], workspace.PruneOpts{
				KeepLast: keepLast, OlderThan: olderThan, DryRun: dryRun, All: all,
			})
			if err != nil {
				return err
			}
			if len(res.Removed) == 0 {
				app.UI.Info("nothing to prune")
				return nil
			}
			label := "pruned"
			if dryRun {
				label = "would prune"
			}
			app.UI.Success("%s %d iteration(s), freeing %s",
				label, len(res.Removed), workspace.HumanSize(res.BytesFreed))
			return nil
		},
	}
	cmd.Flags().IntVar(&keepLast, "keep-last", 0, "retain the N most recent iterations")
	cmd.Flags().DurationVar(&olderThan, "older-than", 0, "drop iterations older than this (e.g. 720h)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be removed")
	cmd.Flags().BoolVar(&all, "all", false, "drop every iteration (use with --yes)")
	return cmd
}

func newEvalWhereCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "where",
		Short: "Print the resolved eval workspace path",
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintln(app.UI.Out(), resolveWorkspace(app, ""))
			return nil
		},
	}
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
				configs:      choices.Arms,
				scenarioIDs:  choices.Scenarios,
				runnerName:   choices.Runner,
				runs:         choices.Runs,
				parallel:     choices.Parallel,
				open:         true,
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

// --- helpers ----------------------------------------------------------------

func resolveSkill(app *App, skill string) (skillDir string, f *scenarios.File, err error) {
	if skill == "" {
		return "", nil, errors.New("skill name required")
	}
	// Collect every candidate location. Order is a preference: a local
	// dev copy with evals/ wins over an installed copy without one, so
	// authoring scenarios against the repo checkout "just works".
	var candidates []string
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(cwd, "skills", skill),
			filepath.Join(cwd, skill),
		)
	}
	if m, err := manifest.Load(app.Config.ManifestPath); err == nil && m != nil {
		for _, inst := range m.FindAll(skill) {
			candidates = append(candidates, inst.Path)
		}
	}
	// Prefer candidates whose scenarios.json parses cleanly.
	var firstExisting string
	for _, c := range candidates {
		if _, err := os.Stat(c); err != nil {
			continue
		}
		if firstExisting == "" {
			firstExisting = c
		}
		sf, serr := scenarios.LoadFromSkill(c)
		if serr == nil {
			return c, sf, nil
		}
		if app.Config.Verbose {
			app.UI.Warn("skip %s: %v", c, serr)
		}
	}
	if firstExisting == "" {
		return "", nil, fmt.Errorf("skill %q not found in manifest and no local copy at ./skills/%s", skill, skill)
	}
	// Found the skill but no valid evals — return dir + nil file so the
	// caller can `eval init` into it.
	return firstExisting, nil, nil
}

func resolveWorkspace(app *App, override string) string {
	r := workspace.Resolver{
		FlagOverride: override,
		EnvOverride:  os.Getenv("HUMBLSKILLS_EVAL_WORKSPACE"),
	}
	if p, err := profile.Load(app.Config.ProfilePath); err == nil && p != nil && p.Eval != nil {
		r.ProfileDefault = p.Eval.DefaultWorkspace
	}
	root, err := r.Root()
	if err != nil {
		root, _ = workspace.DefaultRoot()
	}
	return root
}

func pickRunner(reg *runner.Registry, name string) (runner.Runner, error) {
	// Prefer flag, then env, then profile, then auto-detect.
	if name == "" {
		name = os.Getenv("HUMBLSKILLS_EVAL_RUNNER")
	}
	if name == "" {
		return reg.AutoPick(context.Background())
	}
	return reg.ByName(name)
}

func openInBrowser(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}

func loadJSON[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func skillBasename(dir string) string { return filepath.Base(dir) }

func scaffoldEvalsDir(dir, skillName string) error {
	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("%s already exists — delete it first to re-scaffold", dir)
	}
	if err := os.MkdirAll(filepath.Join(dir, "files"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "assertions"), 0o755); err != nil {
		return err
	}
	scenariosBody := fmt.Sprintf(`{
  "skill_name": %q,
  "schema_version": 1,
  "configurations": ["smart_skill", "flat_skill", "no_skill"],
  "runs_per_configuration": 1,
  "scenarios": [
    {
      "id": "starter",
      "family": "generic",
      "sessions": [
        {
          "n": 1,
          "prompt": "Describe the first task you want %s to handle.",
          "assertions": [
            {"text": "agent produced at least one output file", "check": "path_exists:mock-output.txt"}
          ]
        }
      ]
    }
  ]
}
`, skillName, skillName)
	evalsBody := fmt.Sprintf(`{
  "skill_name": %q,
  "evals": [
    {"id": 1, "prompt": "single-session eval compatible with agentskills.io",
     "expected_output": "describe the expected output",
     "assertions": [{"text": "agent produced output"}]}
  ]
}
`, skillName)
	readme := fmt.Sprintf(`# evals/ for %s

scenarios.json    — humblSKILLS longitudinal + multi-arm scenarios
evals.json        — agentskills.io-compatible single-session evals (optional)
files/            — input fixtures referenced by prompts
assertions/       — optional shell/python scripts for deterministic checks

Run:  humblskills eval run %s
`, skillName, skillName)
	files := map[string]string{
		"scenarios.json": scenariosBody,
		"evals.json":     evalsBody,
		"README.md":      readme,
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func sortedStrKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func unionKeys[V any](a, b map[string]V) []string {
	set := map[string]struct{}{}
	for k := range a {
		set[k] = struct{}{}
	}
	for k := range b {
		set[k] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func knownProviders() []string {
	ps := secrets.Providers()
	out := make([]string, 0, len(ps))
	for _, p := range ps {
		out = append(out, p.Name)
	}
	return out
}

func firstNonEmptyStr(vs ...string) string {
	for _, v := range vs {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func abbreviate(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
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
