package install

import (
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/v2/internal/registry"
)

func reg(skills ...registry.Skill) *registry.Registry {
	return &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/example/repo", SHA: "deadbeef"},
		Skills:        skills,
	}
}

func TestPlan_Simple(t *testing.T) {
	r := reg(
		registry.Skill{Name: "a", Version: "0.1.0"},
	)
	steps, err := Plan(r, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 1 || steps[0].Skill.Name != "a" || steps[0].IsDep {
		t.Errorf("unexpected plan: %+v", steps)
	}
}

func TestPlan_TransitiveDeps(t *testing.T) {
	r := reg(
		registry.Skill{Name: "a", Version: "0.1.0", Requires: []string{"b"}},
		registry.Skill{Name: "b", Version: "0.1.0", Requires: []string{"c@>=0.1.0"}},
		registry.Skill{Name: "c", Version: "0.2.0"},
		registry.Skill{Name: "unrelated", Version: "1.0.0"},
	)
	steps, err := Plan(r, "a")
	if err != nil {
		t.Fatal(err)
	}
	got := []string{}
	for _, s := range steps {
		got = append(got, s.Skill.Name)
	}
	want := []string{"c", "b", "a"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("order[%d]=%s want %s (full=%v)", i, got[i], want[i], got)
		}
	}
	for _, s := range steps {
		if (s.Skill.Name == "a") == s.IsDep {
			t.Errorf("IsDep wrong for %s", s.Skill.Name)
		}
	}
}

func TestPlan_MissingDep(t *testing.T) {
	r := reg(registry.Skill{Name: "a", Requires: []string{"ghost"}})
	if _, err := Plan(r, "a"); err == nil {
		t.Fatal("expected error")
	}
}

func TestPlan_UnsatisfiedPin(t *testing.T) {
	r := reg(
		registry.Skill{Name: "a", Requires: []string{"b@>=1.0.0"}},
		registry.Skill{Name: "b", Version: "0.1.0"},
	)
	if _, err := Plan(r, "a"); err == nil {
		t.Fatal("expected error")
	}
}

func TestPlan_UnknownRoot(t *testing.T) {
	r := reg(registry.Skill{Name: "a"})
	if _, err := Plan(r, "ghost"); err == nil {
		t.Fatal("expected error")
	}
}

// --- PlanAll (batch install) --------------------------------------------------

func planNames(steps []Step) []string {
	out := make([]string, 0, len(steps))
	for _, s := range steps {
		out = append(out, s.Skill.Name)
	}
	return out
}

// A dep shared by two roots must be planned once, or it gets fetched and placed
// once per root that wants it.
func TestPlanAll_SharedDepAppearsOnce(t *testing.T) {
	r := reg(
		registry.Skill{Name: "common", Version: "1.0.0"},
		registry.Skill{Name: "a", Version: "1.0.0", Requires: []string{"common"}},
		registry.Skill{Name: "b", Version: "1.0.0", Requires: []string{"common"}},
	)
	steps, err := PlanAll(r, []string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	got := planNames(steps)
	if len(got) != 3 {
		t.Fatalf("expected 3 steps (a, b, and common once), got %v", got)
	}
	n := 0
	for _, name := range got {
		if name == "common" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("shared dep planned %d times, want 1: %v", n, got)
	}
}

// Deps must precede the skills that need them across the whole batch, not just
// within one root's own subtree.
func TestPlanAll_DepsOrderedBeforeDependents(t *testing.T) {
	r := reg(
		registry.Skill{Name: "base", Version: "1.0.0"},
		registry.Skill{Name: "mid", Version: "1.0.0", Requires: []string{"base"}},
		registry.Skill{Name: "leaf", Version: "1.0.0", Requires: []string{"mid"}},
		registry.Skill{Name: "solo", Version: "1.0.0"},
	)
	steps, err := PlanAll(r, []string{"leaf", "solo"})
	if err != nil {
		t.Fatal(err)
	}
	names := planNames(steps)
	pos := map[string]int{}
	for i, name := range names {
		pos[name] = i
	}
	if pos["base"] > pos["mid"] || pos["mid"] > pos["leaf"] {
		t.Errorf("dep order violated: %v", names)
	}
	if _, ok := pos["solo"]; !ok {
		t.Errorf("second root missing from plan: %v", names)
	}
}

// IsDep drives the "root"/"dep" labelling, and must follow what the caller asked
// for by name: a root stays a root even when another root also depends on it.
func TestPlanAll_RootStaysRootWhenAlsoADep(t *testing.T) {
	r := reg(
		registry.Skill{Name: "base", Version: "1.0.0"},
		registry.Skill{Name: "top", Version: "1.0.0", Requires: []string{"base"}},
	)
	steps, err := PlanAll(r, []string{"top", "base"})
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range steps {
		if s.IsDep {
			t.Errorf("%q marked as a dep, but both names were requested", s.Skill.Name)
		}
	}
}

func TestPlanAll_RepeatedRootIsNotAnError(t *testing.T) {
	r := reg(registry.Skill{Name: "a", Version: "1.0.0"})
	steps, err := PlanAll(r, []string{"a", "a"})
	if err != nil {
		t.Fatal(err)
	}
	if got := planNames(steps); len(got) != 1 {
		t.Errorf("expected 1 step, got %v", got)
	}
}

// An unknown name fails the whole batch: a partially installed batch is worse
// than an unstarted one, so resolution happens before any work.
func TestPlanAll_UnknownRootFailsWholeBatch(t *testing.T) {
	r := reg(registry.Skill{Name: "a", Version: "1.0.0"})
	if _, err := PlanAll(r, []string{"a", "nope"}); err == nil {
		t.Fatal("expected an error for an unknown root")
	}
}

func TestPlanAll_NoRootsIsEmpty(t *testing.T) {
	r := reg(registry.Skill{Name: "a", Version: "1.0.0"})
	steps, err := PlanAll(r, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 0 {
		t.Errorf("expected an empty plan, got %v", planNames(steps))
	}
}
