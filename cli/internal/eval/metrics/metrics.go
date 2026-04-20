// Package metrics aggregates per-run grading.json / timing.json / growth.json
// into the trajectory, benchmark, and report sidecar files.
//
// Three shapes land on disk per iteration:
//
//	trajectory.json  per-session * per-configuration time series
//	benchmark.json   cross-section summary with deltas
//	growth.json      per-session brain-growth deltas (smart_skill only)
//
// All computation is pure + synchronous - the harness can call Aggregate
// between sessions to refresh live trajectory data for the TUI.
package metrics

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sort"
)

// Trajectory is the time-series artifact. One Row per (arm, session).
type Trajectory struct {
	SkillName string               `json:"skill_name"`
	Runner    string               `json:"runner,omitempty"`
	Rows      []TrajectoryRow      `json:"rows"`
	Derived   map[string]Derived   `json:"derived"` // keyed by arm
}

// TrajectoryRow is one sampling point.
type TrajectoryRow struct {
	Arm            string  `json:"arm"`
	Scenario       string  `json:"scenario"`
	Session        int     `json:"session"`
	RunIdx         int     `json:"run"`
	PassRate       float64 `json:"pass_rate"`
	PromptTokens   int     `json:"prompt_tokens"`
	Tokens         int     `json:"tokens"`
	DurationMs     int64   `json:"duration_ms"`
	CostUSD        float64 `json:"cost_usd"`
	WikiConcepts   int64   `json:"wiki_concepts,omitempty"`
	PatternsCount  int64   `json:"patterns_entries,omitempty"`
	BrainBytes     int64   `json:"brain_bytes,omitempty"`
	ReadsFromBrain int     `json:"reads_from_brain,omitempty"`
	ToolCalls      int     `json:"tool_calls,omitempty"`
}

// Derived surfaces aggregate insights - learning velocity, token decay,
// retention, transfer - per arm.
type Derived struct {
	LearningVelocity float64            `json:"learning_velocity"`
	TokenDecay       float64            `json:"token_decay"`
	PassAtK          map[int]float64    `json:"pass_at_k,omitempty"`
	RetentionAtK     map[int]float64    `json:"retention_at_k,omitempty"`
	TransferDelta    map[string]float64 `json:"transfer_delta,omitempty"` // keyed by family
}

// Benchmark is the cross-section summary file.
type Benchmark struct {
	SkillName  string                  `json:"skill_name"`
	Iteration  int                     `json:"iteration"`
	RunSummary map[string]RunSummary   `json:"run_summary"` // keyed by arm
	Delta      map[string]DeltaSummary `json:"delta"`       // "smart_vs_flat" etc
}

// RunSummary is per-arm aggregate stats.
type RunSummary struct {
	PassRate    StatPair `json:"pass_rate"`
	TimeSeconds StatPair `json:"time_seconds"`
	Tokens      StatPair `json:"tokens"`
	CostUSD     StatPair `json:"cost_usd"`
}

// StatPair is a mean + stddev bundle.
type StatPair struct {
	Mean   float64 `json:"mean"`
	StdDev float64 `json:"stddev"`
}

// DeltaSummary reports deltas between two arms.
type DeltaSummary struct {
	PassRate    float64 `json:"pass_rate"`
	TimeSeconds float64 `json:"time_seconds"`
	Tokens      float64 `json:"tokens"`
	CostUSD     float64 `json:"cost_usd"`
}

// AggregateTrajectory builds the trajectory from rows. Rows should already
// be collected by the harness as sessions finish. Runner is stored for report
// rebuilds (eval report) and may be empty for legacy artifacts.
func AggregateTrajectory(skill, runner string, rows []TrajectoryRow) *Trajectory {
	t := &Trajectory{SkillName: skill, Runner: runner, Rows: append([]TrajectoryRow(nil), rows...)}
	sort.Slice(t.Rows, func(i, j int) bool {
		if t.Rows[i].Arm != t.Rows[j].Arm {
			return t.Rows[i].Arm < t.Rows[j].Arm
		}
		if t.Rows[i].Scenario != t.Rows[j].Scenario {
			return t.Rows[i].Scenario < t.Rows[j].Scenario
		}
		if t.Rows[i].Session != t.Rows[j].Session {
			return t.Rows[i].Session < t.Rows[j].Session
		}
		return t.Rows[i].RunIdx < t.Rows[j].RunIdx
	})
	t.Derived = computeDerived(t.Rows)
	return t
}

// AggregateBenchmark builds the cross-section summary.
func AggregateBenchmark(skill string, iteration int, rows []TrajectoryRow) *Benchmark {
	b := &Benchmark{
		SkillName:  skill,
		Iteration:  iteration,
		RunSummary: map[string]RunSummary{},
		Delta:      map[string]DeltaSummary{},
	}
	byArm := groupByArm(rows)
	for arm, arows := range byArm {
		b.RunSummary[arm] = summarize(arows)
	}
	// Compute deltas vs the two baseline arms when present.
	smart, hasSmart := b.RunSummary["smart_skill"]
	flat, hasFlat := b.RunSummary["flat_skill"]
	none, hasNone := b.RunSummary["no_skill"]
	if hasSmart && hasFlat {
		b.Delta["smart_vs_flat"] = DeltaSummary{
			PassRate:    smart.PassRate.Mean - flat.PassRate.Mean,
			TimeSeconds: smart.TimeSeconds.Mean - flat.TimeSeconds.Mean,
			Tokens:      smart.Tokens.Mean - flat.Tokens.Mean,
			CostUSD:     smart.CostUSD.Mean - flat.CostUSD.Mean,
		}
	}
	if hasSmart && hasNone {
		b.Delta["smart_vs_none"] = DeltaSummary{
			PassRate:    smart.PassRate.Mean - none.PassRate.Mean,
			TimeSeconds: smart.TimeSeconds.Mean - none.TimeSeconds.Mean,
			Tokens:      smart.Tokens.Mean - none.Tokens.Mean,
			CostUSD:     smart.CostUSD.Mean - none.CostUSD.Mean,
		}
	}
	return b
}

// Write serializes v to path atomically.
func Write(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// --- per-arm summary --------------------------------------------------------

func summarize(rows []TrajectoryRow) RunSummary {
	pass := make([]float64, 0, len(rows))
	time := make([]float64, 0, len(rows))
	tok := make([]float64, 0, len(rows))
	cost := make([]float64, 0, len(rows))
	for _, r := range rows {
		pass = append(pass, r.PassRate)
		time = append(time, float64(r.DurationMs)/1000.0)
		tok = append(tok, float64(r.Tokens))
		cost = append(cost, r.CostUSD)
	}
	return RunSummary{
		PassRate:    stat(pass),
		TimeSeconds: stat(time),
		Tokens:      stat(tok),
		CostUSD:     stat(cost),
	}
}

func stat(xs []float64) StatPair {
	if len(xs) == 0 {
		return StatPair{}
	}
	sum := 0.0
	for _, x := range xs {
		sum += x
	}
	mean := sum / float64(len(xs))
	if len(xs) == 1 {
		return StatPair{Mean: mean}
	}
	var sq float64
	for _, x := range xs {
		sq += (x - mean) * (x - mean)
	}
	return StatPair{Mean: mean, StdDev: math.Sqrt(sq / float64(len(xs)-1))}
}

func groupByArm(rows []TrajectoryRow) map[string][]TrajectoryRow {
	out := map[string][]TrajectoryRow{}
	for _, r := range rows {
		out[r.Arm] = append(out[r.Arm], r)
	}
	return out
}

// --- derived metrics --------------------------------------------------------

func computeDerived(rows []TrajectoryRow) map[string]Derived {
	out := map[string]Derived{}
	for arm, arows := range groupByArm(rows) {
		d := Derived{}
		// Aggregate by session: mean pass_rate / tokens per session.
		bySession := map[int][]TrajectoryRow{}
		for _, r := range arows {
			bySession[r.Session] = append(bySession[r.Session], r)
		}
		sessions := sortedKeys(bySession)
		if len(sessions) >= 2 {
			xs := make([]float64, 0, len(sessions))
			pys := make([]float64, 0, len(sessions))
			tys := make([]float64, 0, len(sessions))
			for _, s := range sessions {
				rs := bySession[s]
				meanPass := 0.0
				meanTok := 0.0
				for _, r := range rs {
					meanPass += r.PassRate
					meanTok += float64(r.Tokens)
				}
				meanPass /= float64(len(rs))
				meanTok /= float64(len(rs))
				xs = append(xs, float64(s))
				pys = append(pys, meanPass)
				tys = append(tys, meanTok)
			}
			d.LearningVelocity = slope(xs, pys)
			if tys[0] > 0 {
				d.TokenDecay = tys[len(tys)-1] / tys[0]
			}
		}
		// pass^k by session: for each session, compute product of runs
		// that passed, if all runs passed (pass rate is 1.0 for that run);
		// approximated as pass_rate ^ k where k = runs at that session.
		d.PassAtK = map[int]float64{}
		for s, rs := range bySession {
			passed := 0
			for _, r := range rs {
				if r.PassRate >= 0.999 {
					passed++
				}
			}
			d.PassAtK[s] = float64(passed) / float64(len(rs))
		}
		// Retention: sessions that declared a retention_check are marked
		// in the harness via RunIdx semantics; left to future iteration.
		// Transfer_delta: requires cross-scenario pairing; computed at
		// harness level when scenarios declare transfer_from. Stubbed
		// here for consumers.
		out[arm] = d
	}
	return out
}

func sortedKeys(m map[int][]TrajectoryRow) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}

// slope returns the least-squares slope of y on x. Zero for degenerate
// inputs.
func slope(x, y []float64) float64 {
	if len(x) < 2 || len(x) != len(y) {
		return 0
	}
	var sumX, sumY, sumXY, sumXX float64
	n := float64(len(x))
	for i := range x {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumXX += x[i] * x[i]
	}
	denom := n*sumXX - sumX*sumX
	if denom == 0 {
		return 0
	}
	return (n*sumXY - sumX*sumY) / denom
}
