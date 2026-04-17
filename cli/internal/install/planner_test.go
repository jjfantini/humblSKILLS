package install

import (
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
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
