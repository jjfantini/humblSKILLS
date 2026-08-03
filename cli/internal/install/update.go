package install

import (
	"github.com/jjfantini/humblSKILLS/cli/internal/manifest"
	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
)

// UpdatePlan describes one skill whose manifest entry has drifted from the
// registry. Targets lists every (platform, scope) the skill is currently
// installed onto.
type UpdatePlan struct {
	Skill string `json:"skill"`
	// RenamedFrom is set when the installed skill has been renamed upstream.
	// Skill then holds the NEW name, which is what gets installed; the engine
	// carries the old installation's preserved data forward and retires it.
	RenamedFrom string          `json:"renamed_from,omitempty"`
	FromVersion string          `json:"from_version"`
	ToVersion   string          `json:"to_version"`
	FromSHA     string          `json:"from_source_sha,omitempty"`
	ToSHA       string          `json:"to_source_sha,omitempty"`
	FromDirSHA  string          `json:"from_dir_sha,omitempty"`
	ToDirSHA    string          `json:"to_dir_sha,omitempty"`
	Targets     []ManifestEntry `json:"targets"`
	// AddPlatforms are platforms this skill should also be installed onto but
	// isn't yet (see PlanUpdatesFor). Adding one is a symlink plus a manifest
	// row; it never changes skill content.
	AddPlatforms []string `json:"add_platforms,omitempty"`
	// LinkOnly is true when the skill's content is already current and the only
	// work is AddPlatforms. Callers must not describe such a plan as an
	// upgrade: nothing is fetched and no file changes.
	LinkOnly bool `json:"link_only,omitempty"`
}

// ManifestEntry mirrors the subset of manifest.Installation a caller needs to
// decide what to update.
type ManifestEntry struct {
	Platform    string `json:"platform"`
	Scope       string `json:"scope"`
	Path        string `json:"path"`
	InstallMode string `json:"install_mode,omitempty"`
}

// PlanUpdates compares every manifest entry against the registry and returns
// one UpdatePlan per skill with drift. A skill whose registry entry is
// missing is skipped (there's nothing to upgrade to). When `only` is
// non-empty, only those skill names are considered.
func PlanUpdates(reg *registry.Registry, m *manifest.Manifest, only []string) []UpdatePlan {
	return PlanUpdatesFor(reg, m, only, nil)
}

// PlanUpdatesFor is PlanUpdates plus platform reconciliation: any platform in
// wantPlatforms that an installed skill doesn't target yet is reported in the
// plan's AddPlatforms.
//
// A skill with no content drift but a missing platform still yields a plan
// (LinkOnly). That's deliberate — it's what makes a backfill visible in
// --check, --json and the TUI picker instead of happening invisibly inside
// whichever command the user happened to run.
func PlanUpdatesFor(reg *registry.Registry, m *manifest.Manifest, only, wantPlatforms []string) []UpdatePlan {
	if reg == nil || m == nil {
		return nil
	}
	regIndex := make(map[string]registry.Skill, len(reg.Skills))
	for _, s := range reg.Skills {
		regIndex[s.Name] = s
	}

	renameIndex := buildRenameIndex(reg, regIndex)

	filter := map[string]struct{}{}
	for _, n := range only {
		filter[n] = struct{}{}
	}

	// Group manifest entries by skill.
	bySkill := map[string][]manifest.Installation{}
	for _, inst := range m.Installations {
		if len(filter) > 0 {
			_, wanted := filter[inst.Skill]
			// `update <new-name>` should also reach an installation still
			// recorded under the old name, or the rename would be unreachable
			// by the name the user now knows the skill by.
			if !wanted {
				if renamed, ok := renameIndex[inst.Skill]; ok {
					_, wanted = filter[renamed.Name]
				}
			}
			if !wanted {
				continue
			}
		}
		bySkill[inst.Skill] = append(bySkill[inst.Skill], inst)
	}

	var out []UpdatePlan
	for name, insts := range bySkill {
		renamedFrom := ""
		regSkill, ok := regIndex[name]
		if !ok {
			// The installed name is gone from the registry. It is a rename only
			// if some registry skill explicitly claims it; otherwise the skill
			// was withdrawn and there is nothing to upgrade to.
			regSkill, ok = renameIndex[name]
			if !ok {
				continue
			}
			// Both names installed at once (a half-finished migration): the
			// direct match already plans the new name, so planning it again
			// from the old entry would execute the same install twice.
			if _, alsoInstalled := bySkill[regSkill.Name]; alsoInstalled {
				continue
			}
			renamedFrom = name
		}
		// Any target that's drifted triggers an UpdatePlan for the skill.
		//
		// Drift is keyed on the per-skill signals: version and DirSHA
		// (RegistryRef). The repo-wide Source.SHA is NOT consulted — it
		// advances on every commit to the humblSKILLS repo, including
		// commits that don't touch this skill, which would flag every
		// installation as drifted after each CLI release. Source.SHA is
		// kept in the manifest purely as install-time metadata.
		//
		// A rename is drift by definition — the installed directory carries a
		// name the registry no longer publishes — so it short-circuits the
		// version/DirSHA comparison, which cannot decide it.
		drifted := renamedFrom != ""
		for _, i := range insts {
			if drifted {
				break
			}
			if i.Version != regSkill.Version ||
				i.RegistryRef != regSkill.DirSHA {
				drifted = true
				break
			}
		}
		add := missingPlatforms(insts, regSkill, wantPlatforms)
		if !drifted && len(add) == 0 {
			continue
		}

		first := insts[0]
		plan := UpdatePlan{
			Skill:        regSkill.Name,
			RenamedFrom:  renamedFrom,
			FromVersion:  first.Version,
			ToVersion:    regSkill.Version,
			FromSHA:      first.SourceSHA,
			ToSHA:        reg.Source.SHA,
			FromDirSHA:   first.RegistryRef,
			ToDirSHA:     regSkill.DirSHA,
			AddPlatforms: add,
			LinkOnly:     !drifted,
		}
		for _, i := range insts {
			plan.Targets = append(plan.Targets, ManifestEntry{
				Platform: i.Platform, Scope: i.Scope, Path: i.Path, InstallMode: i.InstallMode,
			})
		}
		out = append(out, plan)
	}
	return out
}

// missingPlatforms returns the wanted platforms this skill isn't installed on
// yet, in the caller's order.
//
// The skill's own platforms[] allow-list is honoured so a plan never promises
// a target the engine will drop on the floor — except for global installs,
// which bypass the allow-list in installOne, and must bypass it here too or
// the plan would hide a target the engine does accept.
func missingPlatforms(insts []manifest.Installation, s registry.Skill, want []string) []string {
	if len(want) == 0 {
		return nil
	}
	have := make(map[string]struct{}, len(insts))
	global := false
	for _, i := range insts {
		have[i.Platform] = struct{}{}
		if i.InstallMode == InstallModeGlobal {
			global = true
		}
	}
	allow := make(map[string]struct{}, len(s.Platforms))
	for _, p := range s.Platforms {
		allow[p] = struct{}{}
	}
	var out []string
	for _, p := range want {
		if _, installed := have[p]; installed {
			continue
		}
		if !global && len(allow) > 0 {
			if _, ok := allow[p]; !ok {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

// buildRenameIndex maps a superseded skill name to the registry skill that
// claims it via previous_names.
//
// Two guards keep this from ever guessing, because a wrong answer here installs
// the wrong skill and retires a real one:
//
//   - A name that is still published in its own right is never treated as a
//     rename target. A live skill always wins over someone else's claim on its
//     name, so a stale or mistaken previous_names entry cannot hijack it.
//   - A name claimed by two or more skills is ambiguous and is dropped
//     entirely rather than resolved by ordering.
func buildRenameIndex(reg *registry.Registry, regIndex map[string]registry.Skill) map[string]registry.Skill {
	claims := map[string][]registry.Skill{}
	for _, s := range reg.Skills {
		for _, prev := range s.PreviousNames {
			if prev == "" || prev == s.Name {
				continue
			}
			if _, live := regIndex[prev]; live {
				continue
			}
			claims[prev] = append(claims[prev], s)
		}
	}
	out := make(map[string]registry.Skill, len(claims))
	for prev, cs := range claims {
		if len(cs) == 1 {
			out[prev] = cs[0]
		}
	}
	return out
}
