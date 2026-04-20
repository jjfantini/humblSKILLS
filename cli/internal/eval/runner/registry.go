package runner

import (
	"context"
	"fmt"
	"sort"
)

// Registry is the ordered list of runners the CLI knows about. The order is
// meaningful: auto-detect walks the list and returns the first runner whose
// DoctorCheck reports Available.
type Registry struct {
	runners []Runner
}

// NewRegistry builds a registry from the given runners in the order given.
// Use eval.DefaultRegistry() for the canonical six-runner lineup.
func NewRegistry(rs ...Runner) *Registry { return &Registry{runners: rs} }

// All returns every registered runner in priority order.
func (r *Registry) All() []Runner { return r.runners }

// ByName returns the runner with the given Name, or an error if none matches.
func (r *Registry) ByName(name string) (Runner, error) {
	for _, rn := range r.runners {
		if rn.Name() == name {
			return rn, nil
		}
	}
	return nil, fmt.Errorf("unknown runner %q (known: %v)", name, r.Names())
}

// Names returns the sorted list of registered runner names (helpful for
// flag validation messages).
func (r *Registry) Names() []string {
	out := make([]string, 0, len(r.runners))
	for _, rn := range r.runners {
		out = append(out, rn.Name())
	}
	sort.Strings(out)
	return out
}

// Detect runs DoctorCheck on every registered runner and returns a per-runner
// availability report.
type DetectResult struct {
	Name    string
	Version string
	Check   DoctorCheck
}

func (r *Registry) Detect(ctx context.Context) []DetectResult {
	out := make([]DetectResult, 0, len(r.runners))
	for _, rn := range r.runners {
		c := rn.DoctorCheck(ctx)
		out = append(out, DetectResult{Name: rn.Name(), Version: c.Version, Check: c})
	}
	return out
}

// AutoPick returns the first Available runner in priority order. Returns
// an error if none is available (so the caller can surface a helpful setup
// hint, e.g. "run humblskills eval set-key anthropic").
func (r *Registry) AutoPick(ctx context.Context) (Runner, error) {
	for _, rn := range r.runners {
		if rn.DoctorCheck(ctx).Available {
			return rn, nil
		}
	}
	return nil, fmt.Errorf("no eval runner is available - run `humblskills doctor` for setup hints")
}
