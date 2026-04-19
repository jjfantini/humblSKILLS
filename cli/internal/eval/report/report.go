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

//go:embed assets/plotly.min.js
var plotlyJS string

// Bundle is the full set of inputs a Render call needs.
type Bundle struct {
	SkillName  string
	Iteration  int
	Runner     string
	Trajectory *metrics.Trajectory
	Benchmark  *metrics.Benchmark
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
	SkillName string
	Iteration int
	Runner    string
	PlotlyJS  template.JS
	DataJSON  template.JS
	Arms      []string
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
	p := templatePayload{
		SkillName: b.SkillName,
		Iteration: b.Iteration,
		Runner:    b.Runner,
		PlotlyJS:  template.JS(plotlyJS), //nolint:gosec // first-party embed
		DataJSON:  template.JS(data),     //nolint:gosec // trusted JSON
		Arms:      payload.arms(),
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
	Skill     string                   `json:"skill"`
	Iteration int                      `json:"iteration"`
	Runner    string                   `json:"runner"`
	Series    map[string]seriesPayload `json:"series"` // keyed by arm
	Summary   map[string]metrics.RunSummary
	Delta     map[string]metrics.DeltaSummary
}

type seriesPayload struct {
	Sessions []int     `json:"sessions"`
	PassRate []float64 `json:"pass_rate"`
	PassAtK  []float64 `json:"pass_at_k"`
	Tokens   []int     `json:"tokens"`
	Wiki     []int64   `json:"wiki_concepts"`
}

// rawData is the subset serialized into the HTML for Plotly. Keeps the
// payload tight so the single-file report stays lean.
func (p payload) rawData() map[string]any {
	return map[string]any{
		"skill":     p.Skill,
		"iteration": p.Iteration,
		"runner":    p.Runner,
		"series":    p.Series,
		"summary":   p.Summary,
		"delta":     p.Delta,
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
	p := payload{
		Skill:     b.SkillName,
		Iteration: b.Iteration,
		Runner:    b.Runner,
		Series:    map[string]seriesPayload{},
	}
	if b.Benchmark != nil {
		p.Summary = b.Benchmark.RunSummary
		p.Delta = b.Benchmark.Delta
	}
	if b.Trajectory == nil {
		return p
	}
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
			var allPassed int
			for _, r := range bySession[s] {
				pr += r.PassRate
				tok += int64(r.Tokens)
				if r.WikiConcepts > wiki {
					wiki = r.WikiConcepts
				}
				if r.PassRate >= 0.999 {
					allPassed++
				}
			}
			n := float64(len(bySession[s]))
			sp.Sessions = append(sp.Sessions, s)
			sp.PassRate = append(sp.PassRate, pr/n)
			sp.PassAtK = append(sp.PassAtK, float64(allPassed)/n)
			sp.Tokens = append(sp.Tokens, int(tok/int64(len(bySession[s]))))
			sp.Wiki = append(sp.Wiki, wiki)
		}
		p.Series[arm] = sp
	}
	return p
}

// --- Markdown + JSON --------------------------------------------------------

func writeMarkdown(path string, b *Bundle) error {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# Eval report: %s\n\n", b.SkillName)
	fmt.Fprintf(&sb, "**iteration** %d · **runner** %s\n\n", b.Iteration, b.Runner)
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
