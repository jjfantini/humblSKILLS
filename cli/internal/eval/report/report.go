// Package report renders eval trajectories into shareable artifacts.
//
// Three output shapes per iteration:
//
//	report.html  - single-file Plotly dashboard (embedded via go:embed)
//	report.md    - plaintext mirror for CI / PR comments
//	report.json  - machine-readable, consumed by `eval compare` / CI
//
// The TUI results screen reuses the Sparkline helpers so the in-terminal
// view and the HTML dashboard share the same data contract.
package report

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/metrics"
)

//go:embed assets/report.html.tmpl
var reportHTMLTemplate string

// Bundle is the full set of inputs a Render call needs.
type Bundle struct {
	SkillName   string
	Iteration   int
	Runner      string
	ScenarioIDs []string
	Trajectory  *metrics.Trajectory
	Benchmark   *metrics.Benchmark
}

// RenderAll writes report.{html,md,json} into iterDir and returns paths.
// Callers pass iterDir = <workspace>/<skill>/iteration-<N>/.
func RenderAll(iterDir string, b *Bundle) (html, md, jsonPath string, err error) {
	if err = os.MkdirAll(iterDir, 0o755); err != nil {
		return
	}
	html = filepath.Join(iterDir, "report.html")
	md = filepath.Join(iterDir, "report.md")
	jsonPath = filepath.Join(iterDir, "report.json")
	if err = writeHTML(html, b); err != nil {
		return
	}
	if err = writeMarkdown(md, b); err != nil {
		return
	}
	if err = writeJSON(jsonPath, b); err != nil {
		return
	}
	return
}

// --- HTML -------------------------------------------------------------------

type templatePayload struct {
	SkillName   string
	Iteration   int
	Runner      string
	ScenarioIDs []string
	DataJSON    template.JS
	Arms        []string
}

func writeHTML(path string, b *Bundle) error {
	tmpl, err := template.New("report").Parse(reportHTMLTemplate)
	if err != nil {
		return err
	}
	payload := buildPayload(b)
	data, err := json.Marshal(payload.rawData())
	if err != nil {
		return err
	}
	runner := b.Runner
	if runner == "" && b.Trajectory != nil {
		runner = b.Trajectory.Runner
	}
	scIDs := b.ScenarioIDs
	if len(scIDs) == 0 && b.Trajectory != nil {
		scIDs = InferScenarioIDs(b.Trajectory.Rows)
	}
	p := templatePayload{
		SkillName:   b.SkillName,
		Iteration:   b.Iteration,
		Runner:      runner,
		ScenarioIDs: scIDs,
		DataJSON:    template.JS(data), //nolint:gosec // trusted JSON
		Arms:        payload.arms(),
	}
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(f, p); err != nil {
		f.Close()
		return err
	}
	f.Close()
	return os.Rename(tmp, path)
}

type payload struct {
	Skill         string
	Iteration     int
	Runner        string
	ScenarioIDs   []string
	Series        map[string]seriesPayload
	Summary       map[string]metrics.RunSummary
	Delta         map[string]metrics.DeltaSummary
	Derived       map[string]metrics.Derived
	LearningSpans []learningSpan
	Narrative     []string
}

type learningSpan struct {
	Scenario   string  `json:"scenario"`
	Arm        string  `json:"arm"`
	SessionMin int     `json:"session_min"`
	SessionMax int     `json:"session_max"`
	FirstPass  float64 `json:"first_pass_mean"`
	LastPass   float64 `json:"last_pass_mean"`
	Delta      float64 `json:"delta_pass"`
}

type seriesPayload struct {
	Sessions []int     `json:"sessions"`
	PassRate []float64 `json:"pass_rate"`
	PassAtK  []float64 `json:"pass_at_k"`
	Tokens   []int     `json:"tokens"`
	Wiki     []int64   `json:"wiki_concepts"`
	Patterns []int64   `json:"patterns_entries"`
}

// rawData is the subset serialized into the HTML for Plotly. Keeps the
// payload tight so the single-file report stays lean.
func (p payload) rawData() map[string]any {
	return map[string]any{
		"skill":          p.Skill,
		"iteration":      p.Iteration,
		"runner":         p.Runner,
		"scenarios":      p.ScenarioIDs,
		"series":         p.Series,
		"summary":        p.Summary,
		"delta":          p.Delta,
		"derived":        p.Derived,
		"learning_spans": p.LearningSpans,
		"narrative":      p.Narrative,
	}
}

func (p payload) arms() []string {
	out := make([]string, 0, len(p.Series))
	for a := range p.Series {
		out = append(out, a)
	}
	sort.Strings(out)
	return out
}

func buildPayload(b *Bundle) payload {
	runner := b.Runner
	if runner == "" && b.Trajectory != nil {
		runner = b.Trajectory.Runner
	}
	scenarios := b.ScenarioIDs
	if len(scenarios) == 0 && b.Trajectory != nil {
		scenarios = InferScenarioIDs(b.Trajectory.Rows)
	}
	p := payload{
		Skill:       b.SkillName,
		Iteration:   b.Iteration,
		Runner:      runner,
		ScenarioIDs: scenarios,
		Series:      map[string]seriesPayload{},
	}
	if b.Benchmark != nil {
		p.Summary = b.Benchmark.RunSummary
		p.Delta = b.Benchmark.Delta
	}
	if b.Trajectory == nil {
		return p
	}
	p.Derived = b.Trajectory.Derived
	p.LearningSpans = learningSpans(b.Trajectory.Rows)
	p.Narrative = narrativeBullets(b, p.LearningSpans)
	byArm := map[string][]metrics.TrajectoryRow{}
	for _, r := range b.Trajectory.Rows {
		byArm[r.Arm] = append(byArm[r.Arm], r)
	}
	for arm, rows := range byArm {
		// Group by session, average.
		bySession := map[int][]metrics.TrajectoryRow{}
		for _, r := range rows {
			bySession[r.Session] = append(bySession[r.Session], r)
		}
		sessions := make([]int, 0, len(bySession))
		for s := range bySession {
			sessions = append(sessions, s)
		}
		sort.Ints(sessions)
		sp := seriesPayload{}
		for _, s := range sessions {
			var pr float64
			var tok int64
			var wiki int64
			var patt int64
			var allPassed int
			nRuns := len(bySession[s])
			for _, r := range bySession[s] {
				pr += r.PassRate
				tok += int64(r.Tokens)
				if r.WikiConcepts > wiki {
					wiki = r.WikiConcepts
				}
				if r.PatternsCount > patt {
					patt = r.PatternsCount
				}
				if r.PassRate >= 0.999 {
					allPassed++
				}
			}
			n := float64(nRuns)
			sp.Sessions = append(sp.Sessions, s)
			sp.PassRate = append(sp.PassRate, pr/n)
			sp.PassAtK = append(sp.PassAtK, float64(allPassed)/n)
			sp.Tokens = append(sp.Tokens, int(tok/int64(nRuns)))
			sp.Wiki = append(sp.Wiki, wiki)
			sp.Patterns = append(sp.Patterns, patt)
		}
		p.Series[arm] = sp
	}
	return p
}

// InferScenarioIDs returns sorted unique scenario ids from trajectory rows.
func InferScenarioIDs(rows []metrics.TrajectoryRow) []string {
	seen := map[string]bool{}
	for _, r := range rows {
		if r.Scenario != "" {
			seen[r.Scenario] = true
		}
	}
	out := make([]string, 0, len(seen))
	for s := range seen {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func learningSpans(rows []metrics.TrajectoryRow) []learningSpan {
	type key struct {
		arm, scenario string
	}
	groups := map[key]map[int][]float64{}
	for _, r := range rows {
		k := key{r.Arm, r.Scenario}
		if groups[k] == nil {
			groups[k] = map[int][]float64{}
		}
		groups[k][r.Session] = append(groups[k][r.Session], r.PassRate)
	}
	out := make([]learningSpan, 0, len(groups))
	for k, bySess := range groups {
		if len(bySess) < 2 {
			continue
		}
		sessions := make([]int, 0, len(bySess))
		for s := range bySess {
			sessions = append(sessions, s)
		}
		sort.Ints(sessions)
		minS := sessions[0]
		maxS := sessions[len(sessions)-1]
		mean := func(xs []float64) float64 {
			var s float64
			for _, x := range xs {
				s += x
			}
			return s / float64(len(xs))
		}
		out = append(out, learningSpan{
			Scenario:   k.scenario,
			Arm:        k.arm,
			SessionMin: minS,
			SessionMax: maxS,
			FirstPass:  mean(bySess[minS]),
			LastPass:   mean(bySess[maxS]),
			Delta:      mean(bySess[maxS]) - mean(bySess[minS]),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Scenario != out[j].Scenario {
			return out[i].Scenario < out[j].Scenario
		}
		return out[i].Arm < out[j].Arm
	})
	return out
}

func narrativeBullets(b *Bundle, spans []learningSpan) []string {
	var lines []string
	if b.Benchmark != nil {
		if d, ok := b.Benchmark.Delta["smart_vs_none"]; ok {
			lines = append(lines, fmt.Sprintf(
				"Cross-section: smart_skill vs no_skill mean pass rate delta %+0.3f (all sessions and scenarios in this run).",
				d.PassRate))
		}
		if d, ok := b.Benchmark.Delta["smart_vs_flat"]; ok {
			lines = append(lines, fmt.Sprintf(
				"Cross-section: smart_skill vs flat_skill mean pass rate delta %+0.3f.",
				d.PassRate))
		}
	}
	if b.Trajectory != nil && b.Trajectory.Derived != nil {
		if d, ok := b.Trajectory.Derived["smart_skill"]; ok {
			if d.LearningVelocity != 0 {
				lines = append(lines, fmt.Sprintf(
					"smart_skill learning velocity (OLS slope of mean pass rate vs session index): %+0.4f per session.",
					d.LearningVelocity))
			}
			if d.TokenDecay > 0 && d.TokenDecay != 1 {
				lines = append(lines, fmt.Sprintf(
					"smart_skill token ratio (last session / first session mean): %0.3f (below 1.0 means less token use late in the run).",
					d.TokenDecay))
			}
		}
	}
	byScenario := map[string][]learningSpan{}
	for _, sp := range spans {
		byScenario[sp.Scenario] = append(byScenario[sp.Scenario], sp)
	}
	for sc, arr := range byScenario {
		var sd, fd, nd float64
		var hasS, hasF, hasN bool
		var smin, smax int
		for _, sp := range arr {
			switch sp.Arm {
			case "smart_skill":
				sd, hasS = sp.Delta, true
				smin, smax = sp.SessionMin, sp.SessionMax
			case "flat_skill":
				fd, hasF = sp.Delta, true
			case "no_skill":
				nd, hasN = sp.Delta, true
			}
		}
		if hasS && hasF {
			lines = append(lines, fmt.Sprintf(
				"Longitudinal %q (sessions %d-%d): pass rate change from first to last session: smart_skill %+0.3f, flat_skill %+0.3f.",
				sc, smin, smax, sd, fd))
			if sd > fd+1e-6 {
				lines = append(lines, fmt.Sprintf(
					"Longitudinal %q: smart_skill gained %+0.3f more pass rate over the run than flat_skill (compounding / brain advantage signal).",
					sc, sd-fd))
			}
		}
		if hasS && hasN {
			lines = append(lines, fmt.Sprintf(
				"Longitudinal %q: first-to-last pass rate change smart_skill %+0.3f vs no_skill %+0.3f.",
				sc, sd, nd))
		}
	}
	if b.Trajectory != nil {
		sessPat := map[int]int64{}
		for _, r := range b.Trajectory.Rows {
			if r.Arm != "smart_skill" {
				continue
			}
			if r.PatternsCount > sessPat[r.Session] {
				sessPat[r.Session] = r.PatternsCount
			}
		}
		sessions := make([]int, 0, len(sessPat))
		for s := range sessPat {
			sessions = append(sessions, s)
		}
		sort.Ints(sessions)
		if len(sessions) >= 2 {
			minS, maxS := sessions[0], sessions[len(sessions)-1]
			firstPat := sessPat[minS]
			lastPat := sessPat[maxS]
			if lastPat > firstPat {
				lines = append(lines, fmt.Sprintf(
					"Brain compounding: patterns.md entry count (from brain snapshot) from session %d to %d: %d -> %d (smart_skill only).",
					minS, maxS, firstPat, lastPat))
			}
		}
	}
	return lines
}

// --- Markdown + JSON --------------------------------------------------------

func writeMarkdown(path string, b *Bundle) error {
	runner := b.Runner
	if runner == "" && b.Trajectory != nil {
		runner = b.Trajectory.Runner
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "# Eval report: %s\n\n", b.SkillName)
	fmt.Fprintf(&sb, "**iteration** %d · **runner** %s\n\n", b.Iteration, runner)
	scen := b.ScenarioIDs
	if len(scen) == 0 && b.Trajectory != nil {
		scen = InferScenarioIDs(b.Trajectory.Rows)
	}
	if len(scen) > 0 {
		fmt.Fprintf(&sb, "**scenarios** %s\n\n", strings.Join(scen, ", "))
	}
	if b.Trajectory != nil {
		spans := learningSpans(b.Trajectory.Rows)
		if len(spans) > 0 {
			fmt.Fprintln(&sb, "## Longitudinal (first vs last session)")
			fmt.Fprintln(&sb, "")
			fmt.Fprintln(&sb, "| scenario | arm | first | last | delta | sessions |")
			fmt.Fprintln(&sb, "|----------|-----|-------|------|-------|----------|")
			for _, s := range spans {
				fmt.Fprintf(&sb, "| %s | %s | %.3f | %.3f | %+.3f | %d-%d |\n",
					s.Scenario, s.Arm, s.FirstPass, s.LastPass, s.Delta, s.SessionMin, s.SessionMax)
			}
			fmt.Fprintln(&sb, "")
		}
		bullets := narrativeBullets(b, spans)
		if len(bullets) > 0 {
			fmt.Fprintln(&sb, "## Summary bullets")
			fmt.Fprintln(&sb, "")
			for _, line := range bullets {
				fmt.Fprintf(&sb, "- %s\n", line)
			}
			fmt.Fprintln(&sb, "")
		}
	}
	if b.Benchmark != nil {
		fmt.Fprintln(&sb, "## Cross-section")
		fmt.Fprintln(&sb, "")
		fmt.Fprintln(&sb, "| arm | pass_rate | tokens | time (s) |")
		fmt.Fprintln(&sb, "|-----|-----------|--------|----------|")
		arms := sortedKeys(b.Benchmark.RunSummary)
		for _, a := range arms {
			s := b.Benchmark.RunSummary[a]
			fmt.Fprintf(&sb, "| %s | %.3f | %d | %.1f |\n",
				a, s.PassRate.Mean, int(s.Tokens.Mean), s.TimeSeconds.Mean)
		}
		fmt.Fprintln(&sb, "")
		if len(b.Benchmark.Delta) > 0 {
			fmt.Fprintln(&sb, "## Deltas")
			fmt.Fprintln(&sb, "")
			fmt.Fprintln(&sb, "| pair | Δ pass_rate | Δ tokens | Δ time (s) |")
			fmt.Fprintln(&sb, "|------|-------------|----------|------------|")
			deltas := sortedKeys(b.Benchmark.Delta)
			for _, k := range deltas {
				d := b.Benchmark.Delta[k]
				fmt.Fprintf(&sb, "| %s | %+.3f | %+d | %+.1f |\n",
					k, d.PassRate, int(d.Tokens), d.TimeSeconds)
			}
			fmt.Fprintln(&sb, "")
		}
	}
	if b.Trajectory != nil {
		fmt.Fprintln(&sb, "## Trajectory")
		fmt.Fprintln(&sb, "")
		byArm := map[string][]metrics.TrajectoryRow{}
		for _, r := range b.Trajectory.Rows {
			byArm[r.Arm] = append(byArm[r.Arm], r)
		}
		for _, a := range sortedKeys(byArm) {
			fmt.Fprintf(&sb, "### %s\n\n", a)
			fmt.Fprintln(&sb, "```")
			vals := make([]float64, 0, len(byArm[a]))
			for _, r := range byArm[a] {
				vals = append(vals, r.PassRate)
			}
			fmt.Fprintln(&sb, "pass_rate  "+Sparkline(vals))
			toks := make([]float64, 0, len(byArm[a]))
			for _, r := range byArm[a] {
				toks = append(toks, float64(r.Tokens))
			}
			fmt.Fprintln(&sb, "tokens     "+Sparkline(toks))
			fmt.Fprintln(&sb, "```")
			fmt.Fprintln(&sb, "")
		}
		if len(b.Trajectory.Derived) > 0 {
			fmt.Fprintln(&sb, "## Derived metrics")
			fmt.Fprintln(&sb, "")
			fmt.Fprintln(&sb, "| arm | learning_velocity | token_decay |")
			fmt.Fprintln(&sb, "|-----|-------------------|-------------|")
			for _, a := range sortedKeys(b.Trajectory.Derived) {
				d := b.Trajectory.Derived[a]
				fmt.Fprintf(&sb, "| %s | %.3f | %.3f |\n", a, d.LearningVelocity, d.TokenDecay)
			}
		}
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(sb.String()), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func writeJSON(path string, b *Bundle) error {
	out := map[string]any{
		"skill":      b.SkillName,
		"iteration":  b.Iteration,
		"runner":     b.Runner,
		"trajectory": b.Trajectory,
		"benchmark":  b.Benchmark,
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// --- ASCII sparkline --------------------------------------------------------

var sparkChars = []rune("▁▂▃▄▅▆▇█")

// Sparkline returns a unicode bar chart for a series. Empty series returns
// an empty string. Used by report.md, the TUI results screen, and the Eval
// Home per-skill detail pane.
func Sparkline(vals []float64) string {
	if len(vals) == 0 {
		return ""
	}
	min, max := vals[0], vals[0]
	for _, v := range vals {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	rng := max - min
	var b strings.Builder
	for _, v := range vals {
		i := 0
		if rng > 0 {
			i = int(math.Round(((v - min) / rng) * float64(len(sparkChars)-1)))
			if i < 0 {
				i = 0
			}
			if i > len(sparkChars)-1 {
				i = len(sparkChars) - 1
			}
		}
		b.WriteRune(sparkChars[i])
	}
	return b.String()
}

// --- helpers ----------------------------------------------------------------

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
