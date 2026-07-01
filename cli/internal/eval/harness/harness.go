// Package harness orchestrates an eval run: for each scenario, for each
// configuration (arm), run every session in order, restoring the brain
// snapshot between sessions so longitudinal state compounds. Grading
// and aggregation happen after each session so the TUI progress view
// can render frame-accurate trajectories.
//
// The harness exposes an event stream (Events()) consumed by the live
// progress TUI, and an aggregated result via Run().
package harness

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/brain"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/grader"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/metrics"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/report"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/runner"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/scenarios"
	"github.com/jjfantini/humblSKILLS/cli/internal/eval/workspace"
	"github.com/jjfantini/humblSKILLS/cli/internal/fsutil"
	"github.com/jjfantini/humblSKILLS/cli/internal/jsonutil"
)

// Options controls one full eval invocation.
type Options struct {
	// SkillDir is the absolute path to the installed skill directory. The
	// harness derives the flat-skill variant from this and leaves the
	// source dir untouched.
	SkillDir string

	// Scenarios is the loaded scenarios file.
	Scenarios *scenarios.File

	// Arms is the list of configurations to run. Subset of
	// Scenarios.Configurations.
	Arms []string

	// Runner is the engine that executes sessions.
	Runner runner.Runner

	// Grader is the LLM-judge used for "llm" assertions. May be nil - those
	// assertions then auto-fail with a clear evidence note.
	Grader grader.LLMJudge

	// Workspace root (<root>/<skill>/iteration-N/...).
	WorkspaceRoot string

	// RunsPerConfiguration: repeat each scenario N times for variance /
	// pass^k computation. Defaults to Scenarios.RunsPerConfiguration.
	RunsPerConfiguration int

	// Parallel caps the concurrent runner.Execute calls. 0 -> 1
	// (sequential) since mixing arm ordering within a session would break
	// longitudinal guarantees. Concurrency still runs *across* scenarios.
	Parallel int

	// ScenarioFilter restricts which scenarios run (nil = all).
	ScenarioFilter []string
}

// Event is emitted on the Events channel so the TUI can render progress.
type Event struct {
	Kind       EventKind
	When       time.Time
	Arm        string
	Scenario   string
	Session    int
	Run        int
	Message    string
	Totals     EventTotals
	AssertText string
	AssertPass bool
}

// EventKind enumerates harness event types.
type EventKind string

const (
	EvtStart       EventKind = "start"
	EvtSession     EventKind = "session"
	EvtSessionDone EventKind = "session_done"
	EvtArmDone     EventKind = "arm_done"
	EvtAssertion   EventKind = "assertion"
	EvtProgress    EventKind = "progress"
	EvtDone        EventKind = "done"
	EvtError       EventKind = "error"
)

// EventTotals carries the running totals attached to progress events.
type EventTotals struct {
	CompletedSessions int
	TotalSessions     int
	Tokens            int
	CostUSD           float64
	ElapsedSec        float64
}

// Result is what Run returns when everything is done.
type Result struct {
	Iteration   int
	IterDir     string
	Trajectory  *metrics.Trajectory
	Benchmark   *metrics.Benchmark
	ReportHTML  string
	ReportMD    string
	ReportJSON  string
	Started     time.Time
	Completed   time.Time
}

// Harness is one eval invocation. Construct via New and drive with Run.
type Harness struct {
	opts    Options
	events  chan Event
	rows    []metrics.TrajectoryRow
	rowsMu  sync.Mutex
	iter    int
	iterDir string
	started time.Time
}

// New constructs a Harness. Validates options; returns an error if the
// config is incoherent (no runner, bad scenarios, unknown arms).
func New(opts Options) (*Harness, error) {
	if opts.SkillDir == "" {
		return nil, errors.New("SkillDir is required")
	}
	if opts.Scenarios == nil || len(opts.Scenarios.Scenarios) == 0 {
		return nil, errors.New("no scenarios loaded")
	}
	if opts.Runner == nil {
		return nil, errors.New("runner is required")
	}
	if opts.WorkspaceRoot == "" {
		return nil, errors.New("workspace root is required")
	}
	if len(opts.Arms) == 0 {
		opts.Arms = opts.Scenarios.Configurations
	}
	if opts.RunsPerConfiguration <= 0 {
		opts.RunsPerConfiguration = opts.Scenarios.RunsPerConfiguration
		if opts.RunsPerConfiguration <= 0 {
			opts.RunsPerConfiguration = 1
		}
	}
	if opts.Parallel <= 0 {
		opts.Parallel = 1
	}
	// Validate arms.
	known := map[string]bool{
		scenarios.ArmSmartSkill:    true,
		scenarios.ArmFlatSkillWiki: true,
		scenarios.ArmFlatSkill:     true,
		scenarios.ArmNoSkill:       true,
	}
	for _, a := range opts.Arms {
		if !known[a] {
			return nil, fmt.Errorf("unknown arm %q", a)
		}
	}
	return &Harness{
		opts:   opts,
		events: make(chan Event, 256),
	}, nil
}

// Events returns the channel for live progress. Closes when Run returns.
func (h *Harness) Events() <-chan Event { return h.events }

// Run executes the full eval pipeline and returns the aggregated Result.
// Closes the Events channel before returning.
func (h *Harness) Run(ctx context.Context) (*Result, error) {
	defer close(h.events)
	h.started = time.Now()
	skillName := h.opts.Scenarios.SkillName
	n, iterDir, err := workspace.BeginIteration(
		h.opts.WorkspaceRoot, skillName, h.opts.Runner.Name(), h.opts.Arms, scenarioIDs(h.opts.Scenarios, h.opts.ScenarioFilter))
	if err != nil {
		return nil, fmt.Errorf("begin iteration: %w", err)
	}
	h.iter = n
	h.iterDir = iterDir

	h.emit(Event{Kind: EvtStart, Message: fmt.Sprintf("iteration-%d", n)})

	totalSessions := h.totalSessionsToRun()
	completed := 0

	for _, arm := range h.opts.Arms {
		skillForArm, cleanup, err := h.prepareSkillForArm(arm)
		if err != nil {
			_ = workspace.MarkIteration(h.opts.WorkspaceRoot, skillName, n, workspace.StatusFailed)
			return nil, fmt.Errorf("prepare skill for arm %s: %w", arm, err)
		}
		defer cleanup()

		for _, sc := range h.selectedScenarios() {
			for run := 1; run <= h.opts.RunsPerConfiguration; run++ {
				// New brain snapshot restore chain per (arm, scenario, run).
				brainState := ""
				if arm == scenarios.ArmSmartSkill && skillForArm != "" {
					// Reset brain to the pristine source before each run.
					_ = brain.Restore(
						filepath.Join(h.opts.SkillDir, "references"),
						skillForArm,
					)
				}
				for _, sess := range sc.Sessions {
					if err := ctx.Err(); err != nil {
						_ = workspace.MarkIteration(h.opts.WorkspaceRoot, skillName, n, workspace.StatusAborted)
						return nil, err
					}
					// Flat skill must be stateless per session: re-derive the
					// flat variant before each session so nothing the agent
					// appended to patterns.md / decisions.md / log.md in the
					// previous session leaks forward. Smart skill uses its
					// own snapshot chain (prevSnapAfter) and is unaffected.
					if arm == scenarios.ArmFlatSkill && skillForArm != "" {
						if _, derr := brain.DeriveFlat(h.opts.SkillDir, skillForArm); derr != nil {
							h.emit(Event{Kind: EvtError, Arm: arm, Scenario: sc.ID, Session: sess.N, Message: "re-derive flat: " + derr.Error()})
						}
					}
					if arm == scenarios.ArmFlatSkillWiki && skillForArm != "" {
						if _, derr := brain.DeriveFlatWithWiki(h.opts.SkillDir, skillForArm); derr != nil {
							h.emit(Event{Kind: EvtError, Arm: arm, Scenario: sc.ID, Session: sess.N, Message: "re-derive flat+wiki: " + derr.Error()})
						}
					}
					row, snapDir, err := h.runSession(ctx, arm, skillForArm, sc, sess, run, brainState)
					if err != nil {
						h.emit(Event{Kind: EvtError, Arm: arm, Scenario: sc.ID, Session: sess.N, Message: err.Error()})
					}
					if row != nil {
						h.rowsMu.Lock()
						h.rows = append(h.rows, *row)
						h.rowsMu.Unlock()
					}
					if arm == scenarios.ArmSmartSkill {
						brainState = snapDir
					}
					completed++
					h.emit(Event{
						Kind: EvtSessionDone, Arm: arm, Scenario: sc.ID, Session: sess.N, Run: run,
						Totals: EventTotals{
							CompletedSessions: completed, TotalSessions: totalSessions,
							Tokens: totalTokens(h.rowsLocked()), CostUSD: totalCost(h.rowsLocked()),
							ElapsedSec: time.Since(h.started).Seconds(),
						},
					})
				}
			}
		}
		h.emit(Event{Kind: EvtArmDone, Arm: arm})
	}

	traj := metrics.AggregateTrajectory(skillName, h.opts.Runner.Name(), h.rowsLocked())
	bench := metrics.AggregateBenchmark(skillName, n, h.rowsLocked())
	if err := metrics.Write(filepath.Join(iterDir, "trajectory.json"), traj); err != nil {
		return nil, err
	}
	if err := metrics.Write(filepath.Join(iterDir, "benchmark.json"), bench); err != nil {
		return nil, err
	}
	scenarioIDs := make([]string, 0, len(h.selectedScenarios()))
	for _, sc := range h.selectedScenarios() {
		scenarioIDs = append(scenarioIDs, sc.ID)
	}
	sort.Strings(scenarioIDs)
	htmlPath, mdPath, jsonPath, err := report.RenderAll(iterDir, &report.Bundle{
		SkillName:   skillName,
		Iteration:   n,
		Runner:      h.opts.Runner.Name(),
		ScenarioIDs: scenarioIDs,
		Trajectory:  traj,
		Benchmark:   bench,
	})
	if err != nil {
		return nil, err
	}

	passRates := map[string]float64{}
	tokens := map[string]int{}
	for arm, rs := range bench.RunSummary {
		passRates[arm] = rs.PassRate.Mean
		tokens[arm] = int(rs.Tokens.Mean)
	}
	_ = workspace.CompleteIteration(h.opts.WorkspaceRoot, skillName, n, passRates, tokens)

	res := &Result{
		Iteration:  n,
		IterDir:    iterDir,
		Trajectory: traj,
		Benchmark:  bench,
		ReportHTML: htmlPath,
		ReportMD:   mdPath,
		ReportJSON: jsonPath,
		Started:    h.started,
		Completed:  time.Now(),
	}
	h.emit(Event{Kind: EvtDone, Totals: EventTotals{
		CompletedSessions: completed, TotalSessions: totalSessions,
		ElapsedSec: time.Since(h.started).Seconds(),
	}})
	return res, nil
}

// prepareSkillForArm returns a path suitable for passing to the runner:
//   - smart_skill:      a fresh copy of the source skill that the harness can
//                       mutate between sessions
//   - flat_skill_wiki:  the derive-flat+wiki variant (SKILL.md + wiki, brain reset)
//   - flat_skill:       the derive-flat variant (SKILL.md only, brain reset)
//   - no_skill:         empty string (runner skips the system prompt for skills)
func (h *Harness) prepareSkillForArm(arm string) (skillPath string, cleanup func(), err error) {
	switch arm {
	case scenarios.ArmNoSkill:
		return "", func() {}, nil
	case scenarios.ArmFlatSkill:
		dst := filepath.Join(h.iterDir, "derived-flat-skill")
		if _, err := brain.DeriveFlat(h.opts.SkillDir, dst); err != nil {
			return "", nil, err
		}
		return dst, func() {}, nil
	case scenarios.ArmFlatSkillWiki:
		dst := filepath.Join(h.iterDir, "derived-flat-skill-wiki")
		if _, err := brain.DeriveFlatWithWiki(h.opts.SkillDir, dst); err != nil {
			return "", nil, err
		}
		return dst, func() {}, nil
	case scenarios.ArmSmartSkill:
		// Copy the skill into a working dir so the harness can mutate
		// references/ between sessions without touching the source.
		dst := filepath.Join(h.iterDir, "smart-skill-working")
		if err := fsutil.CopyTree(h.opts.SkillDir, dst, fsutil.Options{}); err != nil {
			return "", nil, err
		}
		return dst, func() {}, nil
	}
	return "", nil, fmt.Errorf("unknown arm %s", arm)
}

// runSession executes one session. Returns the collected trajectory row
// and the path of the brain-after snapshot (empty for non-smart arms).
func (h *Harness) runSession(
	ctx context.Context,
	arm, skillPath string,
	sc scenarios.Scenario,
	sess scenarios.Session,
	run int,
	prevSnapAfter string,
) (*metrics.TrajectoryRow, string, error) {
	sessDir := filepath.Join(h.iterDir, arm, fmt.Sprintf("session-%02d-%s", sess.N, sc.ID), fmt.Sprintf("run-%d", run))
	outputDir := filepath.Join(sessDir, "outputs")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, "", err
	}

	// Snapshot brain BEFORE the session (smart arm only).
	var snapBefore string
	if arm == scenarios.ArmSmartSkill {
		snapBefore = filepath.Join(sessDir, "brain-snapshot-before")
		if prevSnapAfter != "" {
			if err := brain.Restore(prevSnapAfter, skillPath); err != nil {
				return nil, "", fmt.Errorf("restore snapshot: %w", err)
			}
		}
		_ = brain.Snapshot(skillPath, snapBefore)
	}

	h.emit(Event{Kind: EvtSession, Arm: arm, Scenario: sc.ID, Session: sess.N, Run: run, Message: sess.Prompt})

	req := runner.Request{
		SkillDir:   skillPath,
		Prompt:     sess.Prompt,
		InputFiles: sess.Files,
		OutputDir:  outputDir,
		Timeout:    sess.Timeout,
	}
	start := time.Now()
	res, err := h.opts.Runner.Execute(ctx, req)
	if err != nil {
		return nil, "", err
	}
	duration := time.Since(start)
	if res.DurationMs == 0 {
		res.DurationMs = duration.Milliseconds()
	}
	// Runners report transport/API errors on Result.Err rather than the
	// returned error (so partial transcripts still survive). Surface
	// those to the event stream - otherwise a bad API key looks like
	// a successful zero-token session and the grader happily passes
	// every negation-style assertion.
	if res.Err != nil {
		h.emit(Event{
			Kind:    EvtError,
			Arm:     arm,
			Scenario: sc.ID,
			Session: sess.N,
			Run:     run,
			Message: "runner: " + res.Err.Error(),
		})
	}

	// Persist timing / transcript / metrics sidecars.
	_ = jsonutil.WriteFile(filepath.Join(sessDir, "timing.json"), map[string]any{
		"total_tokens":      res.TotalTokens,
		"prompt_tokens":     res.PromptTokens,
		"completion_tokens": res.CompletionTokens,
		"duration_ms":       res.DurationMs,
		"cost_usd":          res.CostUSD,
	})
	if len(res.Transcript) > 0 {
		_ = os.WriteFile(filepath.Join(sessDir, "transcript.txt"), res.Transcript, 0o644)
	}
	_ = jsonutil.WriteFile(filepath.Join(sessDir, "metrics.json"), map[string]any{
		"tool_calls":      res.ToolCalls,
		"brain_reads":     brain.ReadsFromBrain(res.Transcript),
		"output_files":    res.OutputFiles,
	})

	// Snapshot brain AFTER the session (smart arm only).
	var snapAfter string
	if arm == scenarios.ArmSmartSkill {
		snapAfter = filepath.Join(sessDir, "brain-snapshot-after")
		_ = brain.Snapshot(skillPath, snapAfter)
		if g, err := brain.ComputeGrowth(snapBefore, snapAfter); err == nil {
			_ = jsonutil.WriteFile(filepath.Join(sessDir, "growth.json"), g)
		}
	}

	// Grade.
	gReq := grader.Request{
		EvalPrompt: sess.Prompt,
		Assertions: sess.Assertions,
		OutputDir:  outputDir,
		WorkDir:    sessDir,
		Transcript: res.Transcript,
		LLMJudge:   h.opts.Grader,
		ToolCalls:  res.ToolCalls,
	}
	g, gerr := grader.Grade(ctx, gReq)
	var passRate float64
	if g != nil {
		passRate = g.Summary.PassRate
		if err := grader.Write(filepath.Join(sessDir, "grading.json"), g); err != nil {
			return nil, "", err
		}
		for _, e := range g.Expectations {
			h.emit(Event{Kind: EvtAssertion, Arm: arm, Scenario: sc.ID, Session: sess.N, Run: run,
				AssertText: e.Text, AssertPass: e.Passed})
		}
	} else if gerr != nil {
		h.emit(Event{Kind: EvtError, Arm: arm, Scenario: sc.ID, Session: sess.N, Run: run, Message: "grade: " + gerr.Error()})
	}

	row := &metrics.TrajectoryRow{
		Arm:          arm,
		Scenario:     sc.ID,
		Session:      sess.N,
		RunIdx:       run,
		PassRate:     passRate,
		PromptTokens: res.PromptTokens,
		Tokens:       res.TotalTokens,
		DurationMs:   res.DurationMs,
		CostUSD:      res.CostUSD,
		ToolCalls:    totalToolCalls(res.ToolCalls),
		Violations:   sumCheckerViolations(outputDir),
	}
	// Brain stats (smart only).
	if snapAfter != "" {
		stats, _ := brain.ComputeGrowth("", snapAfter)
		if stats != nil {
			row.WikiConcepts = stats.WikiConcepts.Total
			row.PatternsCount = stats.PatternsEntries.Total
			row.BrainBytes = stats.BrainBytes.Total
		}
		row.ReadsFromBrain = brain.ReadsFromBrain(res.Transcript)
	}
	return row, snapAfter, nil
}

// sumCheckerViolations walks outputDir for files matching `*-check.json`,
// parses each as {"count": N}, and returns the sum. Scenarios that drop a
// deterministic rule-checker sidecar get per-session violation counts
// automatically surfaced in the trajectory + report.
func sumCheckerViolations(outputDir string) int {
	if outputDir == "" {
		return 0
	}
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return 0
	}
	total := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, "-check.json") {
			continue
		}
		body, err := os.ReadFile(filepath.Join(outputDir, name))
		if err != nil {
			continue
		}
		var v struct {
			Count int `json:"count"`
		}
		if err := json.Unmarshal(body, &v); err != nil {
			continue
		}
		total += v.Count
	}
	return total
}

// --- helpers ----------------------------------------------------------------

func (h *Harness) emit(e Event) {
	e.When = time.Now()
	select {
	case h.events <- e:
	default:
		// Consumer fell behind; drop the event rather than stall. Progress
		// consumers see this as a dropped frame; correctness is unaffected
		// because event state is derivable from on-disk artifacts.
	}
}

func (h *Harness) selectedScenarios() []scenarios.Scenario {
	if len(h.opts.ScenarioFilter) == 0 {
		return h.opts.Scenarios.Scenarios
	}
	want := map[string]bool{}
	for _, id := range h.opts.ScenarioFilter {
		want[id] = true
	}
	var out []scenarios.Scenario
	for _, s := range h.opts.Scenarios.Scenarios {
		if want[s.ID] {
			out = append(out, s)
		}
	}
	return out
}

func (h *Harness) totalSessionsToRun() int {
	n := 0
	for _, sc := range h.selectedScenarios() {
		n += len(sc.Sessions) * h.opts.RunsPerConfiguration * len(h.opts.Arms)
	}
	return n
}

func (h *Harness) rowsLocked() []metrics.TrajectoryRow {
	h.rowsMu.Lock()
	defer h.rowsMu.Unlock()
	out := make([]metrics.TrajectoryRow, len(h.rows))
	copy(out, h.rows)
	return out
}

func scenarioIDs(f *scenarios.File, filter []string) []string {
	want := map[string]bool{}
	for _, id := range filter {
		want[id] = true
	}
	var ids []string
	for _, s := range f.Scenarios {
		if len(filter) == 0 || want[s.ID] {
			ids = append(ids, s.ID)
		}
	}
	sort.Strings(ids)
	return ids
}

func totalTokens(rows []metrics.TrajectoryRow) int {
	n := 0
	for _, r := range rows {
		n += r.Tokens
	}
	return n
}

func totalCost(rows []metrics.TrajectoryRow) float64 {
	var n float64
	for _, r := range rows {
		n += r.CostUSD
	}
	return n
}

func totalToolCalls(m map[string]int) int {
	n := 0
	for _, v := range m {
		n += v
	}
	return n
}

