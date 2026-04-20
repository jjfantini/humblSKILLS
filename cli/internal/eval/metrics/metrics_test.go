package metrics

import (
	"math"
	"testing"
)

func TestSlope(t *testing.T) {
	xs := []float64{1, 2, 3, 4, 5}
	ys := []float64{2, 4, 6, 8, 10}
	if got := slope(xs, ys); math.Abs(got-2.0) > 1e-9 {
		t.Fatalf("slope = %v, want 2", got)
	}
}

func TestStatPair(t *testing.T) {
	got := stat([]float64{1, 2, 3, 4, 5})
	if math.Abs(got.Mean-3.0) > 1e-9 {
		t.Fatalf("mean = %v", got.Mean)
	}
	if math.Abs(got.StdDev-1.5811388300841898) > 1e-9 {
		t.Fatalf("stddev = %v", got.StdDev)
	}
	if stat([]float64{7}).Mean != 7 {
		t.Fatalf("single-value mean")
	}
}

func TestAggregateTrajectoryLearningVelocity(t *testing.T) {
	rows := []TrajectoryRow{
		{Arm: "smart_skill", Session: 1, PassRate: 0.5, Tokens: 1000},
		{Arm: "smart_skill", Session: 2, PassRate: 0.7, Tokens: 800},
		{Arm: "smart_skill", Session: 3, PassRate: 0.9, Tokens: 600},
	}
	tr := AggregateTrajectory("demo", "", rows)
	d := tr.Derived["smart_skill"]
	if math.Abs(d.LearningVelocity-0.2) > 1e-9 {
		t.Fatalf("learning_velocity = %v, want 0.2", d.LearningVelocity)
	}
	if math.Abs(d.TokenDecay-0.6) > 1e-9 {
		t.Fatalf("token_decay = %v, want 0.6", d.TokenDecay)
	}
}

func TestAggregateBenchmarkDeltas(t *testing.T) {
	rows := []TrajectoryRow{
		{Arm: "smart_skill", Session: 1, PassRate: 0.9, Tokens: 1000},
		{Arm: "flat_skill", Session: 1, PassRate: 0.5, Tokens: 1500},
		{Arm: "no_skill", Session: 1, PassRate: 0.2, Tokens: 400},
	}
	b := AggregateBenchmark("demo", 1, rows)
	if d, ok := b.Delta["smart_vs_flat"]; !ok || math.Abs(d.PassRate-0.4) > 1e-9 {
		t.Fatalf("smart_vs_flat: %+v", d)
	}
	if d, ok := b.Delta["smart_vs_none"]; !ok || math.Abs(d.PassRate-0.7) > 1e-9 {
		t.Fatalf("smart_vs_none: %+v", d)
	}
}
