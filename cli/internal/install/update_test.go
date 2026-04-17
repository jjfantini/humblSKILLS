package install

import (
	"testing"

	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
)

func TestPlanUpdates(t *testing.T) {
	reg := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/example/repo", SHA: "newSourceSHA"},
		Skills: []registry.Skill{
			{Name: "foo", Version: "0.2.0", DirSHA: "dirSHA-foo-new"},
			{Name: "bar", Version: "1.0.0", DirSHA: "dirSHA-bar-v1"},
			{Name: "baz", Version: "0.1.0", DirSHA: "dirSHA-baz"},
		},
	}

	m := &manifest.Manifest{
		SchemaVersion: manifest.SchemaVersion,
		Installations: []manifest.Installation{
			// foo: version drift + dir_sha drift across two targets.
			{
				Skill: "foo", Version: "0.1.0", Platform: "claude", Scope: "user",
				Path: "/u/foo", SourceSHA: "oldSourceSHA", RegistryRef: "dirSHA-foo-old",
			},
			{
				Skill: "foo", Version: "0.1.0", Platform: "cursor", Scope: "project",
				Path: "/p/foo", SourceSHA: "oldSourceSHA", RegistryRef: "dirSHA-foo-old",
			},
			// bar: already up-to-date — source_sha matches current registry.
			{
				Skill: "bar", Version: "1.0.0", Platform: "claude", Scope: "user",
				Path: "/u/bar", SourceSHA: "newSourceSHA", RegistryRef: "dirSHA-bar-v1",
			},
			// baz: same version but source_sha drift should still trigger an update.
			{
				Skill: "baz", Version: "0.1.0", Platform: "claude", Scope: "user",
				Path: "/u/baz", SourceSHA: "oldSourceSHA", RegistryRef: "dirSHA-baz",
			},
			// orphan: installed skill removed from registry — must be skipped.
			{
				Skill: "orphan", Version: "0.1.0", Platform: "claude", Scope: "user",
				Path: "/u/orphan", SourceSHA: "oldSourceSHA", RegistryRef: "dirSHA-orphan",
			},
		},
	}

	plans := PlanUpdates(reg, m, nil)
	byName := map[string]UpdatePlan{}
	for _, p := range plans {
		byName[p.Skill] = p
	}

	if _, ok := byName["foo"]; !ok {
		t.Error("foo should be in plans")
	}
	if _, ok := byName["baz"]; !ok {
		t.Error("baz should be in plans (source_sha drift)")
	}
	if _, ok := byName["bar"]; ok {
		t.Error("bar should NOT be in plans (up-to-date)")
	}
	if _, ok := byName["orphan"]; ok {
		t.Error("orphan should NOT be in plans (no registry entry)")
	}

	foo := byName["foo"]
	if foo.FromVersion != "0.1.0" || foo.ToVersion != "0.2.0" {
		t.Errorf("foo version range wrong: %+v", foo)
	}
	if foo.FromSHA != "oldSourceSHA" || foo.ToSHA != "newSourceSHA" {
		t.Errorf("foo source sha range wrong: %+v", foo)
	}
	if len(foo.Targets) != 2 {
		t.Errorf("foo should have 2 targets, got %d", len(foo.Targets))
	}

	// Filter to "baz": should exclude foo.
	only := PlanUpdates(reg, m, []string{"baz"})
	if len(only) != 1 || only[0].Skill != "baz" {
		t.Errorf("filter failed: %+v", only)
	}
}

func TestPlanUpdates_NilInputs(t *testing.T) {
	if got := PlanUpdates(nil, nil, nil); got != nil {
		t.Errorf("nil inputs: expected nil, got %+v", got)
	}
}
