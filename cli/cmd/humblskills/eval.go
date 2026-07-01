package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/workspace"
)

// The eval command family is split across three files:
//   - eval.go         (this file): the cobra command tree + flag bag
//   - eval_actions.go: the run/behavior logic (runEval*)
//   - eval_helpers.go: shared helpers (skill/workspace resolution, scaffolding)

func newEvalCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "eval",
		Short: "Benchmark skills across no_skill / flat_skill / smart_skill arms",
		Long: "eval runs a full cross-sectional + longitudinal benchmark of a skill " +
			"against three arms (no / flat / smart), grades the outputs, and emits " +
			"a single-file HTML dashboard. Six runners available (claudecode, " +
			"cursor-agent, codex, anthropic-api, openai-api, mock). " +
			"Skill resolution checks $HUMBLSKILLS_ROOT/skills/<name> first, then cwd/skills/<name>, " +
			"and if the current directory is the cli module, ../skills/<name> (so `go run` from cli/ still finds repo skills).",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runEvalTUI(app)
		},
	}
	cmd.AddCommand(
		newEvalRunCmd(app),
		newEvalInitCmd(app),
		newEvalAggregateCmd(app),
		newEvalReportCmd(app),
		newEvalCompareCmd(app),
		newEvalShowcaseCmd(app),
		newEvalBrandVoiceCmd(app),
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
	scenarioIDs []string
	configs     []string
	sessionsCap int
	runs        int
	parallel    int
	workspace   string
	runnerName  string
	executor    string
	grader      string
	noGrade     bool
	noReport    bool
	resume      bool
	minPassRate float64
	open        bool
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
	cmd.Flags().StringVar(&f.grader, "grader-model", "", "model for LLM-judge grading (default: anthropic sonnet-4-6)")
	cmd.Flags().BoolVar(&f.noGrade, "no-grade", false, "skip grading (useful to batch grade later)")
	cmd.Flags().BoolVar(&f.noReport, "no-report", false, "skip report rendering")
	cmd.Flags().BoolVar(&f.resume, "resume", false, "(reserved) resume an interrupted iteration")
	cmd.Flags().Float64Var(&f.minPassRate, "min-pass-rate", 0, "exit non-zero if mean pass_rate on the primary arm is below this")
	cmd.Flags().BoolVar(&f.open, "open", false, "open report.html in the OS default browser after completion")
	return cmd
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

// --- eval aggregate / report ------------------------------------------------

func newEvalAggregateCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "aggregate <iteration-dir>",
		Short: "Rebuild trajectory.json / benchmark.json from grading files",
		Args:  cobra.ExactArgs(1),
		RunE:  func(cmd *cobra.Command, args []string) error { return runEvalAggregate(app, args[0]) },
	}
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

func newEvalShowcaseCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "showcase",
		Short: "Run the canonical demo on use-smart-skill",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runEvalRun(app, "use-smart-skill", evalRunFlags{open: true})
		},
	}
}

// newEvalBrandVoiceCmd runs the adaptive-brand-voice-discovery scenario on
// use-smart-humanize-text. This is the canonical 3-arm (smart / flat / no)
// compounding-learning showcase: 6 sessions over a fictional company's 10
// idiosyncratic style rules, with per-session violation charts and deltas.
// Opens the single-file HTML report in the browser when complete.
func newEvalBrandVoiceCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "brand-voice",
		Short: "Run the adaptive-brand-voice-discovery showcase (3-arm compounding demo) and open the report",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runEvalRun(app, "use-smart-humanize-text", evalRunFlags{
				scenarioIDs: []string{"adaptive-brand-voice-discovery"},
				open:        true,
			})
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
