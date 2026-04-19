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
			// baz: version and dir_sha match — a stale source_sha alone
			// must NOT flag the skill as drifted. Source.SHA advances on
			// every humblSKILLS repo commit whether or not this skill
			// changed, so consulting it here produces false positives
			// after every CLI release.
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
	if _, ok := byName["baz"]; ok {
		t.Error("baz should NOT be in plans (source_sha differs but version + dir_sha match)")
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

	// Filter to "foo": should exclude every other skill.
	only := PlanUpdates(reg, m, []string{"foo"})
	if len(only) != 1 || only[0].Skill != "foo" {
		t.Errorf("filter failed: %+v", only)
	}
}

// TestPlanUpdates_StaleSourceSHAIsNotDrift reproduces the dashboard bug where
// every installation was flagged as drifted after a CLI release even though
// no skill content had changed. Source.SHA is the humblSKILLS repo commit
// SHA; it advances on every commit, so consulting it as a drift signal
// produces false positives on every new release. Drift must key only on
// per-skill signals (version + DirSHA).
func TestPlanUpdates_StaleSourceSHAIsNotDrift(t *testing.T) {
	reg := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/example/repo", SHA: "sha-after-cli-release"},
		Skills: []registry.Skill{
			{Name: "use-smart-humanize-text", Version: "2.0.0", DirSHA: "dirSHA-humanize-v2"},
			{Name: "use-smart-skill", Version: "1.1.0", DirSHA: "dirSHA-smart-skill-v1-1"},
		},
	}
	m := &manifest.Manifest{
		SchemaVersion: manifest.SchemaVersion,
		Installations: []manifest.Installation{
			{
				Skill: "use-smart-humanize-text", Version: "2.0.0",
				Platform: "claude-code", Scope: "user", Path: "/u/humanize",
				SourceSHA: "sha-before-cli-release", RegistryRef: "dirSHA-humanize-v2",
			},
			{
				Skill: "use-smart-humanize-text", Version: "2.0.0",
				Platform: "cursor", Scope: "user", Path: "/u/humanize-cursor",
				SourceSHA: "sha-before-cli-release", RegistryRef: "dirSHA-humanize-v2",
			},
			{
				Skill: "use-smart-skill", Version: "1.1.0",
				Platform: "claude-code", Scope: "user", Path: "/u/smart-skill",
				SourceSHA: "sha-before-cli-release", RegistryRef: "dirSHA-smart-skill-v1-1",
			},
		},
	}

	plans := PlanUpdates(reg, m, nil)
	if len(plans) != 0 {
		t.Errorf("no skill should drift purely from a stale repo SourceSHA; got %+v", plans)
	}
}

func TestPlanUpdates_NilInputs(t *testing.T) {
	if got := PlanUpdates(nil, nil, nil); got != nil {
		t.Errorf("nil inputs: expected nil, got %+v", got)
	}
}
