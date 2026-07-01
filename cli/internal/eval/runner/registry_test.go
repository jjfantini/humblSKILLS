package runner

import (
	"context"
	"math"
	"reflect"
	"testing"
)

// fakeRunner is a minimal Runner used to exercise the registry without any
// external processes or API calls.
type fakeRunner struct {
	name      string
	available bool
	version   string
}

func (f fakeRunner) Name() string               { return f.name }
func (f fakeRunner) Capabilities() Capabilities { return Capabilities{} }
func (f fakeRunner) DoctorCheck(context.Context) DoctorCheck {
	return DoctorCheck{Available: f.available, Version: f.version}
}
func (f fakeRunner) Execute(context.Context, Request) (*Result, error) { return &Result{}, nil }

func TestRegistry_AllPreservesOrder(t *testing.T) {
	a := fakeRunner{name: "a"}
	b := fakeRunner{name: "b"}
	reg := NewRegistry(b, a) // intentionally not sorted
	got := reg.All()
	if len(got) != 2 || got[0].Name() != "b" || got[1].Name() != "a" {
		t.Fatalf("All did not preserve insertion order: %v", names(got))
	}
}

func TestRegistry_ByName(t *testing.T) {
	reg := NewRegistry(fakeRunner{name: "mock"}, fakeRunner{name: "codex"})

	r, err := reg.ByName("codex")
	if err != nil {
		t.Fatalf("ByName(codex): %v", err)
	}
	if r.Name() != "codex" {
		t.Errorf("ByName returned %q", r.Name())
	}

	if _, err := reg.ByName("nope"); err == nil {
		t.Error("expected error for unknown runner")
	}
}

func TestRegistry_NamesSorted(t *testing.T) {
	reg := NewRegistry(fakeRunner{name: "zeta"}, fakeRunner{name: "alpha"}, fakeRunner{name: "mid"})
	got := reg.Names()
	want := []string{"alpha", "mid", "zeta"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Names() = %v, want %v", got, want)
	}
}

func TestRegistry_Detect(t *testing.T) {
	reg := NewRegistry(
		fakeRunner{name: "up", available: true, version: "1.2.3"},
		fakeRunner{name: "down", available: false},
	)
	det := reg.Detect(context.Background())
	if len(det) != 2 {
		t.Fatalf("Detect returned %d results", len(det))
	}
	if det[0].Name != "up" || !det[0].Check.Available || det[0].Version != "1.2.3" {
		t.Errorf("unexpected first detect result: %+v", det[0])
	}
	if det[1].Name != "down" || det[1].Check.Available {
		t.Errorf("unexpected second detect result: %+v", det[1])
	}
}

func TestRegistry_AutoPickReturnsFirstAvailable(t *testing.T) {
	reg := NewRegistry(
		fakeRunner{name: "down", available: false},
		fakeRunner{name: "first-up", available: true},
		fakeRunner{name: "second-up", available: true},
	)
	r, err := reg.AutoPick(context.Background())
	if err != nil {
		t.Fatalf("AutoPick: %v", err)
	}
	if r.Name() != "first-up" {
		t.Errorf("AutoPick chose %q, want first available in order", r.Name())
	}
}

func TestRegistry_AutoPickNoneAvailable(t *testing.T) {
	reg := NewRegistry(fakeRunner{name: "down", available: false})
	if _, err := reg.AutoPick(context.Background()); err == nil {
		t.Error("expected error when no runner is available")
	}
}

func TestEstimateCost(t *testing.T) {
	if got := EstimateCost(nil, 1000, 1000); got != 0 {
		t.Errorf("nil pricing should cost 0, got %v", got)
	}
	p := &Pricing{PromptUSDPerMtok: 3, CompletionUSDPerMtok: 15}
	// 1M prompt tokens @ $3 + 1M completion @ $15 = $18.
	if got := EstimateCost(p, 1_000_000, 1_000_000); math.Abs(got-18) > 1e-9 {
		t.Errorf("EstimateCost = %v, want 18", got)
	}
	// Fractional: 500k prompt @ $3 = $1.50.
	if got := EstimateCost(p, 500_000, 0); math.Abs(got-1.5) > 1e-9 {
		t.Errorf("EstimateCost = %v, want 1.5", got)
	}
}

func names(rs []Runner) []string {
	out := make([]string, 0, len(rs))
	for _, r := range rs {
		out = append(out, r.Name())
	}
	return out
}
