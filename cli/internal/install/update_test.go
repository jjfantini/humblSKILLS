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
			{Name: "smart-humanize-text", Version: "2.0.0", DirSHA: "dirSHA-humanize-v2"},
			{Name: "smart-skill", Version: "1.1.0", DirSHA: "dirSHA-smart-skill-v1-1"},
		},
	}
	m := &manifest.Manifest{
		SchemaVersion: manifest.SchemaVersion,
		Installations: []manifest.Installation{
			{
				Skill: "smart-humanize-text", Version: "2.0.0",
				Platform: "claude-code", Scope: "user", Path: "/u/humanize",
				SourceSHA: "sha-before-cli-release", RegistryRef: "dirSHA-humanize-v2",
			},
			{
				Skill: "smart-humanize-text", Version: "2.0.0",
				Platform: "cursor", Scope: "user", Path: "/u/humanize-cursor",
				SourceSHA: "sha-before-cli-release", RegistryRef: "dirSHA-humanize-v2",
			},
			{
				Skill: "smart-skill", Version: "1.1.0",
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

// regSkill is a tiny helper for the rename-planning tests below.
func regSkillNamed(name, version, dirSHA string, previous ...string) registry.Skill {
	return registry.Skill{
		Name: name, Version: version, Path: "skills/" + name,
		DirSHA: dirSHA, PreviousNames: previous,
	}
}

func instNamed(skill, version, dirSHA string) manifest.Installation {
	return manifest.Installation{
		Skill: skill, Version: version, RegistryRef: dirSHA,
		Platform: "test", Scope: "user", Path: "/tmp/" + skill,
	}
}

// TestPlanUpdates_FollowsRename: an installation whose name the registry no
// longer publishes must upgrade to the skill that claims it, rather than being
// skipped as "withdrawn" and left behind forever.
func TestPlanUpdates_FollowsRename(t *testing.T) {
	reg := &registry.Registry{Skills: []registry.Skill{
		regSkillNamed("foo", "2.0.0", "sha-new", "use-foo"),
	}}
	m := &manifest.Manifest{Installations: []manifest.Installation{
		instNamed("use-foo", "1.0.0", "sha-old"),
	}}

	plans := PlanUpdates(reg, m, nil)
	if len(plans) != 1 {
		t.Fatalf("want 1 plan, got %d (%+v)", len(plans), plans)
	}
	if plans[0].Skill != "foo" {
		t.Errorf("plan must target the NEW name, got %q", plans[0].Skill)
	}
	if plans[0].RenamedFrom != "use-foo" {
		t.Errorf("plan must record the old name, got %q", plans[0].RenamedFrom)
	}
	if len(plans[0].Targets) != 1 {
		t.Errorf("targets should carry over from the old installation, got %d", len(plans[0].Targets))
	}
}

// TestPlanUpdates_RenameGuards covers every way this could install the wrong
// skill or retire a real one. Each case must produce no rename plan.
func TestPlanUpdates_RenameGuards(t *testing.T) {
	tests := []struct {
		name      string
		reg       []registry.Skill
		installed []manifest.Installation
		wantPlans int
		wantSkill string
	}{
		{
			// The claimed name is still published. A live skill must always win
			// over another skill's claim on its name.
			name: "live skill is never hijacked by a claim",
			reg: []registry.Skill{
				regSkillNamed("foo", "2.0.0", "sha-new", "use-foo"),
				regSkillNamed("use-foo", "1.0.0", "sha-old"),
			},
			installed: []manifest.Installation{instNamed("use-foo", "1.0.0", "sha-old")},
			wantPlans: 0,
		},
		{
			// Two skills claim the same previous name: ambiguous, so refuse
			// rather than resolve by map/slice ordering.
			name: "ambiguous claim is dropped",
			reg: []registry.Skill{
				regSkillNamed("foo", "2.0.0", "sha-a", "use-foo"),
				regSkillNamed("bar", "2.0.0", "sha-b", "use-foo"),
			},
			installed: []manifest.Installation{instNamed("use-foo", "1.0.0", "sha-old")},
			wantPlans: 0,
		},
		{
			// Withdrawn, not renamed: nobody claims it, so there is nothing to
			// upgrade to and the old behaviour (skip) must hold.
			name: "withdrawn skill is still skipped",
			reg: []registry.Skill{
				regSkillNamed("other", "1.0.0", "sha-x"),
			},
			installed: []manifest.Installation{instNamed("use-foo", "1.0.0", "sha-old")},
			wantPlans: 0,
		},
		{
			// Half-finished migration: both names installed. The direct match
			// already plans the new name; planning it twice would run the same
			// install twice.
			name: "both names installed plans the new name once",
			reg: []registry.Skill{
				regSkillNamed("foo", "2.0.0", "sha-new", "use-foo"),
			},
			installed: []manifest.Installation{
				instNamed("use-foo", "1.0.0", "sha-old"),
				instNamed("foo", "1.0.0", "sha-old"),
			},
			wantPlans: 1,
			wantSkill: "foo",
		},
		{
			// A skill listing itself must not become its own rename target.
			name: "self-claim is ignored",
			reg: []registry.Skill{
				regSkillNamed("foo", "2.0.0", "sha-new", "foo"),
			},
			installed: []manifest.Installation{instNamed("use-foo", "1.0.0", "sha-old")},
			wantPlans: 0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plans := PlanUpdates(&registry.Registry{Skills: tc.reg},
				&manifest.Manifest{Installations: tc.installed}, nil)
			if len(plans) != tc.wantPlans {
				t.Fatalf("want %d plans, got %d (%+v)", tc.wantPlans, len(plans), plans)
			}
			if tc.wantSkill != "" && plans[0].Skill != tc.wantSkill {
				t.Errorf("want plan for %q, got %q", tc.wantSkill, plans[0].Skill)
			}
			for _, p := range plans {
				if p.RenamedFrom != "" && tc.wantPlans == 0 {
					t.Errorf("unexpected rename plan: %+v", p)
				}
			}
		})
	}
}

// TestPlanUpdates_RenameReachableByEitherName: the user may type the name they
// now know the skill by, even though the manifest still records the old one.
func TestPlanUpdates_RenameReachableByEitherName(t *testing.T) {
	reg := &registry.Registry{Skills: []registry.Skill{
		regSkillNamed("foo", "2.0.0", "sha-new", "use-foo"),
	}}
	mk := func() *manifest.Manifest {
		return &manifest.Manifest{Installations: []manifest.Installation{
			instNamed("use-foo", "1.0.0", "sha-old"),
		}}
	}
	for _, only := range []string{"foo", "use-foo"} {
		plans := PlanUpdates(reg, mk(), []string{only})
		if len(plans) != 1 || plans[0].Skill != "foo" {
			t.Errorf("update %q should reach the rename, got %+v", only, plans)
		}
	}
	if plans := PlanUpdates(reg, mk(), []string{"unrelated"}); len(plans) != 0 {
		t.Errorf("an unrelated filter must match nothing, got %+v", plans)
	}
}

// TestPlanUpdatesFor_PlatformBackfill: a skill whose content is current but
// which is missing a wanted platform still yields a plan, flagged LinkOnly so
// no caller describes it as an upgrade.
func TestPlanUpdatesFor_PlatformBackfill(t *testing.T) {
	reg := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/example/repo", SHA: "src"},
		Skills: []registry.Skill{
			{Name: "current", Version: "1.0.0", DirSHA: "sha-current"},
			{Name: "drifted", Version: "2.0.0", DirSHA: "sha-drifted-new"},
			{Name: "picky", Version: "1.0.0", DirSHA: "sha-picky", Platforms: []string{"claude-code"}},
		},
	}
	m := &manifest.Manifest{
		SchemaVersion: manifest.SchemaVersion,
		Installations: []manifest.Installation{
			{Skill: "current", Version: "1.0.0", Platform: "claude-code", Scope: "user",
				RegistryRef: "sha-current", InstallMode: InstallModeLinked},
			{Skill: "drifted", Version: "1.0.0", Platform: "claude-code", Scope: "user",
				RegistryRef: "sha-drifted-old", InstallMode: InstallModeLinked},
			{Skill: "picky", Version: "1.0.0", Platform: "claude-code", Scope: "user",
				RegistryRef: "sha-picky", InstallMode: InstallModeLinked},
		},
	}

	byName := map[string]UpdatePlan{}
	for _, p := range PlanUpdatesFor(reg, m, nil, []string{"claude-code", "codex"}) {
		byName[p.Skill] = p
	}

	cur, ok := byName["current"]
	if !ok {
		t.Fatal("a current skill missing a wanted platform must still be planned")
	}
	if !cur.LinkOnly {
		t.Error("no content drift, so the plan must be LinkOnly")
	}
	if len(cur.AddPlatforms) != 1 || cur.AddPlatforms[0] != "codex" {
		t.Errorf("AddPlatforms = %v, want [codex]", cur.AddPlatforms)
	}

	dr, ok := byName["drifted"]
	if !ok {
		t.Fatal("drifted skill missing from plans")
	}
	if dr.LinkOnly {
		t.Error("drifted skill must not be LinkOnly")
	}
	if len(dr.AddPlatforms) != 1 || dr.AddPlatforms[0] != "codex" {
		t.Errorf("drift + backfill should fold into one plan, got %v", dr.AddPlatforms)
	}

	// `picky` declares platforms: [claude-code], so codex is not a legal
	// target and must not be promised.
	if p, planned := byName["picky"]; planned {
		t.Errorf("allow-list violated: %+v", p)
	}

	// Without wantPlatforms the behaviour is unchanged: drift only.
	plans := PlanUpdatesFor(reg, m, nil, nil)
	if len(plans) != 1 || plans[0].Skill != "drifted" {
		t.Errorf("no wantPlatforms should mean drift-only, got %+v", plans)
	}
}

// A --global install bypasses the platforms[] allow-list in the engine, so the
// plan must bypass it too or it would hide a target Execute accepts.
func TestPlanUpdatesFor_GlobalIgnoresAllowList(t *testing.T) {
	reg := &registry.Registry{
		SchemaVersion: registry.SchemaVersion,
		Source:        registry.Source{Repo: "github.com/example/repo", SHA: "src"},
		Skills: []registry.Skill{
			{Name: "picky", Version: "1.0.0", DirSHA: "sha", Platforms: []string{"claude-code"}},
		},
	}
	m := &manifest.Manifest{
		SchemaVersion: manifest.SchemaVersion,
		Installations: []manifest.Installation{
			{Skill: "picky", Version: "1.0.0", Platform: "claude-code", Scope: "user",
				RegistryRef: "sha", InstallMode: InstallModeGlobal},
		},
	}
	plans := PlanUpdatesFor(reg, m, nil, []string{"codex"})
	if len(plans) != 1 || len(plans[0].AddPlatforms) != 1 || plans[0].AddPlatforms[0] != "codex" {
		t.Fatalf("global install should allow any platform, got %+v", plans)
	}
}
