package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/eval/metrics"
)

func TestSparkline(t *testing.T) {
	s := Sparkline([]float64{1, 2, 3, 4, 5})
	if n := len([]rune(s)); n != 5 {
		t.Fatalf("expected 5 glyphs, got %d in %q", n, s)
	}
	if got := Sparkline(nil); got != "" {
		t.Fatalf("empty input should give empty string: %q", got)
	}
}

func TestRenderAll(t *testing.T) {
	rows := []metrics.TrajectoryRow{
		{Arm: "smart_skill", Session: 1, PassRate: 0.5, Tokens: 1000},
		{Arm: "smart_skill", Session: 2, PassRate: 0.8, Tokens: 800},
		{Arm: "flat_skill", Session: 1, PassRate: 0.5, Tokens: 1100},
		{Arm: "flat_skill", Session: 2, PassRate: 0.55, Tokens: 1100},
	}
	traj := metrics.AggregateTrajectory("demo", rows)
	bench := metrics.AggregateBenchmark("demo", 1, rows)
	dir := t.TempDir()
	html, md, js, err := RenderAll(dir, &Bundle{
		SkillName:  "demo",
		Iteration:  1,
		Runner:     "mock",
		Trajectory: traj,
		Benchmark:  bench,
	})
	if err != nil {
		t.Fatalf("RenderAll: %v", err)
	}
	htmlBody, _ := os.ReadFile(html)
	if !strings.Contains(string(htmlBody), "demo") {
		t.Fatalf("html missing skill name")
	}
	if !strings.Contains(string(htmlBody), "Plotly.newPlot") {
		t.Fatalf("html missing Plotly init")
	}
	mdBody, _ := os.ReadFile(md)
	if !strings.Contains(string(mdBody), "# Eval report") {
		t.Fatalf("md missing header")
	}
	jsBody, _ := os.ReadFile(js)
	if !strings.Contains(string(jsBody), `"skill": "demo"`) {
		t.Fatalf("json missing skill key")
	}
	// Each artifact lives at the expected path.
	for _, p := range []string{html, md, js} {
		if filepath.Dir(p) != dir {
			t.Fatalf("artifact outside dir: %s", p)
		}
	}
}
